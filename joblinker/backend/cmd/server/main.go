package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"joblinker/internal/cache"
	"joblinker/internal/eino/runner"
	"joblinker/internal/handler"
	"joblinker/internal/middleware"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/internal/service"
	"joblinker/pkg/ai"
	"joblinker/pkg/chroma"
	"joblinker/pkg/rabbitmq"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvOrFail(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func main() {
	// Load .env file if present (from same directory as binary)
	godotenv.Load()

	// Validate required environment variables at startup
	getEnvOrFail("JWT_SECRET")

	// Initialize Chroma client
	chromaHost := getEnv("CHROMA_HOST", "localhost")
	chromaPort := getEnvInt("CHROMA_PORT", 8000)
	chromaClient := chroma.NewClient(chromaHost, chromaPort)
	if err := chromaClient.Heartbeat(); err != nil {
		log.Printf("WARNING: Chroma not reachable at %s:%d: %v", chromaHost, chromaPort, err)
	} else {
		log.Printf("Connected to Chroma at %s:%d", chromaHost, chromaPort)
		if _, err := chromaClient.EnsureCollection("agent_memories"); err != nil {
			log.Printf("WARNING: failed to ensure agent_memories collection: %v", err)
		}
		if _, err := chromaClient.EnsureCollection("user_preferences"); err != nil {
			log.Printf("WARNING: failed to ensure user_preferences collection: %v", err)
		}
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())
	r.Use(middleware.GatewayMiddleware())
	r.Use(middleware.ErrorHandler())

	dsn := getEnv("DATABASE_URL", "host=localhost user=joblinker password=joblinker_dev dbname=joblinker port=5432 sslmode=disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to database")

	// Auto migrate models
	if err := db.AutoMigrate(
		&model.User{},
		&model.Organization{},
		&model.Agent{},
		&model.Resume{},
		&model.Job{},
		&model.Match{},
		&model.Message{},
		&model.Interview{},
		&model.Offer{},
		&model.SecurityEvent{},
		&model.ErrorLog{},
		&model.RateLimitCounter{},
		&model.AgentMetrics{},
		&model.AuditLog{},
		&model.ErrorEvent{},
		&model.AgentMemory{},
		&model.ConversationSummary{},
		&model.AgentToolCall{},
		&model.ConfirmationRequest{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Printf("Database migrated")

	userRepo := repository.NewUserRepository().WithDB(db)
	agentRepo := repository.NewAgentRepository().WithDB(db)
	jobRepo := repository.NewJobRepository().WithDB(db)
	matchRepo := repository.NewMatchRepository().WithDB(db)
	messageRepo := repository.NewMessageRepository().WithDB(db)
	interviewRepo := repository.NewInterviewRepository().WithDB(db)
	offerRepo := repository.NewOfferRepository().WithDB(db)
	securityRepo := repository.NewSecurityEventRepository().WithDB(db)
	errorLogRepo := repository.NewErrorLogRepository().WithDB(db)
	_ = errorLogRepo // used by middleware via LogError
	// Observability repositories
	metricsRepo := repository.NewAgentMetricsRepository().WithDB(db)
	auditRepo := repository.NewAuditLogRepository().WithDB(db)
	observabilityErrorRepo := repository.NewErrorEventRepository().WithDB(db)

	// Initialize tool cache (100 entries max, 5 min TTL)
	toolCache := cache.NewToolCache(100, 5*time.Minute)

	// Initialize RabbitMQ
	var rmq *rabbitmq.RabbitMQ
	var mqSvc *service.MessageQueueService
	rmq, err = rabbitmq.New(nil)
	if err != nil {
		log.Printf("RabbitMQ not available: %v (continuing without queue)", err)
	} else {
		log.Printf("Connected to RabbitMQ")
		mqSvc = service.NewMessageQueueService(rmq, messageRepo, matchRepo, agentRepo, jobRepo, offerRepo, interviewRepo, metricsRepo, auditRepo, observabilityErrorRepo, toolCache)
	}

	agentSvc := service.NewAgentService(agentRepo, userRepo, securityRepo)
	matchSvc := service.NewMatchService(matchRepo, agentRepo, jobRepo)
	messageSvc := service.NewMessageService(messageRepo, matchRepo, agentRepo)
	securitySvc := service.NewSecurityService(securityRepo)
	interviewSvc := service.NewInterviewService(interviewRepo, matchRepo, messageSvc, securitySvc)
	offerSvc := service.NewOfferService(offerRepo, matchRepo, jobRepo, securitySvc)
	privacySvc := service.NewPrivacyService(userRepo, agentRepo, matchRepo)
	adminHandler := handler.NewAdminHandler(metricsRepo, auditRepo, observabilityErrorRepo)

	// Initialize Eino Agent Runner (for T029 integration)
	aiClient := ai.NewClient()
	einoRunner := runner.NewAgentRunner(aiClient, nil)
	log.Printf("Eino AgentRunner initialized with pool config: MaxAgents=%d, MinAgents=%d",
		100, 5)

	authHandler := handler.NewAuthHandler(userRepo, securitySvc)
	agentHandler := handler.NewAgentHandler(agentSvc)
	jobHandler := handler.NewJobHandler(jobRepo, agentRepo)
	matchHandler := handler.NewMatchHandler(matchSvc)
	interviewHandler := handler.NewInterviewHandler(interviewSvc)
	offerHandler := handler.NewOfferHandler(offerSvc)
	privacyHandler := handler.NewPrivacyHandler(privacySvc)
	messageHandler := handler.NewMessageHandler(messageRepo, matchRepo, agentRepo, rmq)
	a2aHandler := handler.NewA2AHandler(matchRepo, agentRepo, messageRepo, rmq)
	_ = handler.NewResumeHandler()
	sqlDB, _ := db.DB()
	healthHandler := handler.NewHealthHandler(sqlDB)

	// Wire Eino Runner to MessageQueueService (T029)
	if mqSvc != nil {
		mqSvc.SetEinoRunner(einoRunner)
		log.Printf("Eino Runner wired to MessageQueueService")

		ctxOptimizer := service.NewContextOptimizerService(toolCache, messageRepo, matchRepo, agentRepo)
		mqSvc.SetContextOptimizer(ctxOptimizer)
		log.Printf("Context Optimizer wired to MessageQueueService")
	}

	r.GET("/health", healthHandler.Health)

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Refresh requires authentication
	authProtected := r.Group("/api/auth")
	authProtected.Use(middleware.Auth())
	{
		authProtected.POST("/refresh", authHandler.Refresh)
	}

	api := r.Group("/api")
	api.Use(middleware.Auth())
	api.Use(middleware.RateLimit())
	api.Use(middleware.ProtoResponseMiddleware())
	api.Use(middleware.ProtoMiddleware())
	{
		api.GET("/agents", agentHandler.List)
		api.POST("/agents", agentHandler.Create)
		api.GET("/agents/:id", agentHandler.Get)
		api.PATCH("/agents/:id", agentHandler.Update)
		api.DELETE("/agents/:id", agentHandler.Delete)

		api.GET("/jobs", jobHandler.List)
		api.POST("/jobs", jobHandler.Create)
		api.GET("/jobs/:id", jobHandler.Get)
		api.PATCH("/jobs/:id", jobHandler.Update)

		api.GET("/matches", matchHandler.List)
		api.GET("/matches/:id", matchHandler.Get)
		api.POST("/matches/auto", matchHandler.AutoCreate)
		api.POST("/matches/:id/confirm", matchHandler.Confirm)

		api.GET("/messages", messageHandler.GetMessages)
		api.GET("/messages/:matchId", messageHandler.GetMessages)
		api.POST("/messages/:matchId", messageHandler.SendMessage)
		api.GET("/conversation/:matchId", messageHandler.GetConversation)

		api.GET("/interviews", interviewHandler.List)
		api.POST("/interviews", interviewHandler.Create)
		api.PATCH("/interviews/:id", interviewHandler.Update)
		api.GET("/interviews/match/:matchId", interviewHandler.GetByMatchID)
		api.POST("/interviews/match/:matchId/confirm", interviewHandler.Confirm)
		api.POST("/interviews/match/:matchId/cancel", interviewHandler.Cancel)

		api.GET("/offers/:matchId", offerHandler.GetByMatchID)
		api.POST("/offers", offerHandler.Create)
		api.POST("/offers/:matchId/accept", offerHandler.Accept)
		api.POST("/offers/:matchId/decline", offerHandler.Decline)

		api.POST("/privacy/export", privacyHandler.Export)
		api.DELETE("/privacy/account", privacyHandler.DeleteAccount)

		// Admin endpoints (require admin role)
		admin := api.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/metrics", adminHandler.GetMetrics)
			admin.GET("/agent-metrics", adminHandler.GetAllAgentMetrics)
			admin.GET("/audit", adminHandler.GetAuditLogs)
			admin.GET("/errors", adminHandler.GetErrors)
			admin.POST("/errors/:id/resolve", adminHandler.ResolveError)
		}
	}

	// Human WebSocket (JWT auth via query param fallback)
	r.GET("/api/messages/ws", messageHandler.HandleWebSocket)
	r.GET("/api/messages/:matchId/ws", messageHandler.HandleWebSocket)

	// A2A Agent WebSocket (HMAC internal auth, no rate limit, XML protocol)
	r.GET("/api/a2a/:matchId/ws", a2aHandler.HandleA2AWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start message queue consumer
	if mqSvc != nil {
		ctx := context.Background()
		if err := mqSvc.StartConsuming(ctx); err != nil {
			log.Printf("Failed to start message queue consumer: %v", err)
		}
	}

log.Printf("Server starting on :%s", port)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
