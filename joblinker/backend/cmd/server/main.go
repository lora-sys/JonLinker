package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"joblinker/internal/handler"
	"joblinker/internal/middleware"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/internal/service"
	"joblinker/pkg/rabbitmq"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	// Load .env file if present (from same directory as binary)
	godotenv.Load()

	// Log environment for debugging
	log.Printf("AI_API_KEY length: %d", len(os.Getenv("AI_API_KEY")))
	log.Printf("AI_BASE_URL: %s", os.Getenv("AI_BASE_URL"))

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

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

	// Initialize RabbitMQ
	var rmq *rabbitmq.RabbitMQ
	var mqSvc *service.MessageQueueService
	rmq, err = rabbitmq.New(nil)
	if err != nil {
		log.Printf("RabbitMQ not available: %v (continuing without queue)", err)
	} else {
		log.Printf("Connected to RabbitMQ")
		mqSvc = service.NewMessageQueueService(rmq, messageRepo, matchRepo, agentRepo, jobRepo, offerRepo, interviewRepo, metricsRepo, auditRepo, observabilityErrorRepo)
	}

	agentSvc := service.NewAgentService(agentRepo, userRepo, securityRepo)
	matchSvc := service.NewMatchService(matchRepo, agentRepo, jobRepo)
	messageSvc := service.NewMessageService(messageRepo, matchRepo, agentRepo)
	securitySvc := service.NewSecurityService(securityRepo)
	interviewSvc := service.NewInterviewService(interviewRepo, matchRepo, messageSvc, securitySvc)
	offerSvc := service.NewOfferService(offerRepo, matchRepo, jobRepo, securitySvc)
	privacySvc := service.NewPrivacyService(userRepo, agentRepo, matchRepo)
	adminHandler := handler.NewAdminHandler(metricsRepo, auditRepo, observabilityErrorRepo)

	authHandler := handler.NewAuthHandler(userRepo, securitySvc)
	agentHandler := handler.NewAgentHandler(agentSvc)
	jobHandler := handler.NewJobHandler(jobRepo, agentRepo)
	matchHandler := handler.NewMatchHandler(matchSvc)
	interviewHandler := handler.NewInterviewHandler(interviewSvc)
	offerHandler := handler.NewOfferHandler(offerSvc)
	privacyHandler := handler.NewPrivacyHandler(privacySvc)
	messageHandler := handler.NewMessageHandler(messageRepo, matchRepo, agentRepo, rmq)
	healthHandler := handler.NewHealthHandler()

	r.GET("/health", healthHandler.Health)

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	api := r.Group("/api")
	api.Use(middleware.Auth())
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

		api.GET("/interviews", interviewHandler.List)
		api.POST("/interviews", interviewHandler.Create)
		api.PATCH("/interviews/:id", interviewHandler.Update)
		api.GET("/interviews/:matchId", interviewHandler.GetByMatchID)
		api.POST("/interviews/:matchId/confirm", interviewHandler.Confirm)
		api.POST("/interviews/:matchId/cancel", interviewHandler.Cancel)

		api.GET("/offers/:matchId", offerHandler.GetByMatchID)
		api.POST("/offers", offerHandler.Create)
		api.POST("/offers/:matchId/accept", offerHandler.Accept)
		api.POST("/offers/:matchId/decline", offerHandler.Decline)

		api.POST("/privacy/export", privacyHandler.Export)
		api.DELETE("/privacy/account", privacyHandler.DeleteAccount)

		// Admin endpoints
		api.GET("/admin/metrics", adminHandler.GetMetrics)
		api.GET("/admin/agent-metrics", adminHandler.GetAllAgentMetrics)
		api.GET("/admin/audit", adminHandler.GetAuditLogs)
		api.GET("/admin/errors", adminHandler.GetErrors)
		api.POST("/admin/errors/:id/resolve", adminHandler.ResolveError)

		// Messages REST (auth required)
		api.GET("/messages/:matchId", messageHandler.GetConversation)
		api.POST("/messages/:matchId", messageHandler.SendMessage)
	}

	// Messages WebSocket (token in query param, no auth middleware)
	r.GET("/api/messages/:matchId/ws", messageHandler.HandleWebSocket)

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
