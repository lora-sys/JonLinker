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
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"joblinker/internal/cache"
	"joblinker/internal/handler"
	"joblinker/internal/middleware"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/internal/service"
	"joblinker/pkg/ai"
	"joblinker/pkg/rabbitmq"

	"joblinker/internal/eino/agent"
	"joblinker/internal/eino/chatmodel"
	"joblinker/internal/eino/memory"
	eino_runner "joblinker/internal/eino/runner"
	"joblinker/internal/eino/sessionstore"
	"joblinker/internal/eino/tools"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/components/tool"
	localbk "github.com/cloudwego/eino-ext/adk/backend/local"
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
		&model.AgentToolCall{},
		&model.ConfirmationRequest{},
		&model.SessionMeta{},
		&model.SessionMessage{},
		&model.SessionSummary{},
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
	rmq, err = rabbitmq.New(nil)
	if err != nil {
		log.Printf("RabbitMQ not available: %v (continuing without queue)", err)
		rmq = nil
	} else {
		log.Printf("Connected to RabbitMQ")
	}

	var mqSvc *service.MessageQueueService
	if rmq == nil {
		mqSvc = service.NewMessageQueueService(nil, messageRepo, matchRepo, agentRepo, jobRepo, offerRepo, interviewRepo, metricsRepo, auditRepo, observabilityErrorRepo, toolCache)
	} else {
		mqSvc = service.NewMessageQueueService(rmq, messageRepo, matchRepo, agentRepo, jobRepo, offerRepo, interviewRepo, metricsRepo, auditRepo, observabilityErrorRepo, toolCache)
	}

	agentSvc := service.NewAgentService(agentRepo, userRepo, securityRepo)
	matchSvc := service.NewMatchService(matchRepo, agentRepo, jobRepo, rmq, mqSvc)

	// ── Background auto-matcher ──
	{
		autoMatchInterval := getEnvInt("AUTO_MATCH_INTERVAL", 60) // seconds
		matchSvc.StartAutoMatcher(context.Background(), time.Duration(autoMatchInterval)*time.Second)
		log.Printf("Auto-matcher started (interval: %ds)", autoMatchInterval)
	}
	messageSvc := service.NewMessageService(messageRepo, matchRepo, agentRepo)
	securitySvc := service.NewSecurityService(securityRepo)
	interviewSvc := service.NewInterviewService(interviewRepo, matchRepo, messageSvc, securitySvc)
	offerSvc := service.NewOfferService(offerRepo, matchRepo, jobRepo, securitySvc)

	// Initialize SessionStore (JSONL for now, PG when PostgresStore is production-ready)
	var sessionStore sessionstore.Store
	sessionDataDir := getEnv("SESSION_DATA_DIR", "./data/sessions")
	sessionStore, sessionStoreErr := sessionstore.NewJSONLSessionStore(sessionDataDir)
	if sessionStoreErr != nil {
		log.Printf("WARNING: failed to create JSONL session store: %v (continuing without)", sessionStoreErr)
		sessionStore = nil
	} else {
		log.Printf("JSONL SessionStore initialized at %s", sessionDataDir)
	}

	privacySvc := service.NewPrivacyService(userRepo, agentRepo, matchRepo, messageRepo, interviewRepo, offerRepo)
	adminHandler := handler.NewAdminHandler(metricsRepo, auditRepo, observabilityErrorRepo)

	// Initialize Eino Agent Runner (for T029 integration)
	aiClient := ai.NewClient()
	einoRunner := eino_runner.NewAgentRunner(aiClient, nil)
	log.Printf("Eino AgentRunner initialized with pool config: MaxAgents=%d, MinAgents=%d",
		100, 5)

	// Initialize SessionService for session-based conversation management
	var sessionSvc *sessionstore.SessionService
	if sessionStore != nil {
		legacySeeker := agent.NewSeekerAgent(aiClient)
		legacyRecruiter := agent.NewRecruiterAgent(aiClient)
		sessionSvc = sessionstore.NewSessionService(sessionStore, legacySeeker, legacyRecruiter)
		log.Printf("SessionService initialized with session store")
	}

	// Initialize ADK components
	var adkRunner *eino_runner.ADKRunner
	{
		chatModel := chatmodel.NewEinoChatModel(aiClient)
		einoTools := tools.NewRealTools(jobRepo, agentRepo, matchRepo, offerRepo, interviewRepo)
		baseTools := make([]tool.BaseTool, len(einoTools))
		for i, t := range einoTools {
			baseTools[i] = t
		}

		seekerAgent, err := agent.NewSeekerChatModelAgent(context.Background(), chatModel, baseTools)
		if err != nil {
			log.Printf("WARNING: failed to create seeker ADK agent: %v", err)
		} else {
			recruiterAgent, err := agent.NewRecruiterChatModelAgent(context.Background(), chatModel, baseTools)
			if err != nil {
				log.Printf("WARNING: failed to create recruiter ADK agent: %v", err)
			} else {
				supervisor, err := agent.NewA2ASupervisor(context.Background(), seekerAgent, recruiterAgent)
				if err != nil {
					log.Printf("WARNING: failed to create A2A supervisor: %v", err)
				} else {
					redisClient := redis.NewClient(&redis.Options{
						Addr: getEnv("REDIS_ADDR", "localhost:6379"),
					})
					cpStore := memory.NewRedisCheckPointStore(redisClient, "adk:cp:")
					adkRunner = eino_runner.NewADKRunner(context.Background(), supervisor, cpStore)
					log.Printf("ADK Runner initialized with Supervisor + CheckPointStore")

					// ── Phase 2: DeepAgent + Routing Supervisor ──
					deepRunner := initDeepAgentAndRouting(context.Background(), chatModel, baseTools, adkRunner)

					// A2A Deep SSE endpoint
					if deepRunner != nil {
						a2aDeepHandler := handler.NewA2ASSEHandler(deepRunner)
						r.POST("/api/a2a/deep/chat", a2aDeepHandler.Chat)
						log.Printf("A2A Deep SSE endpoint registered at POST /api/a2a/deep/chat")
					}
				}
			}
		}
	}

	authHandler := handler.NewAuthHandler(userRepo, securitySvc)
	agentHandler := handler.NewAgentHandler(agentSvc)
	jobHandler := handler.NewJobHandler(jobRepo, agentRepo)
	matchHandler := handler.NewMatchHandler(matchSvc)
	interviewHandler := handler.NewInterviewHandler(interviewSvc)
	offerHandler := handler.NewOfferHandler(offerSvc)
	privacyHandler := handler.NewPrivacyHandler(privacySvc)
	messageHandler := handler.NewMessageHandler(messageRepo, matchRepo, agentRepo, rmq, mqSvc, sessionStore)
	a2aHandler := handler.NewA2AHandler(matchRepo, agentRepo, messageRepo, rmq)
	resumeHandler := handler.NewResumeHandler()
	sqlDB, _ := db.DB()
	healthHandler := handler.NewHealthHandler(sqlDB)

	// Wire Eino Runner + ADK Runner + Broadcast to MessageQueueService
	if mqSvc != nil {
		mqSvc.SetEinoRunner(einoRunner)
		log.Printf("Eino Runner wired to MessageQueueService")

		if adkRunner != nil {
			mqSvc.SetADKRunner(adkRunner)
			log.Printf("ADK Runner wired to MessageQueueService")
		}

		// Initialize StateGraph Runner (autonomous negotiation pipeline)
		stateGraphRunner := eino_runner.NewStateGraphRunner(einoRunner)
		mqSvc.SetStateGraphRunner(stateGraphRunner)
		log.Printf("StateGraph Runner wired to MessageQueueService")

		ctxOptimizer := service.NewContextOptimizerService(toolCache, messageRepo, matchRepo, agentRepo)
		mqSvc.SetContextOptimizer(ctxOptimizer)
		log.Printf("Context Optimizer wired to MessageQueueService")

		mqSvc.SetBroadcastCallback(messageHandler.BroadcastAgentResponse)
		log.Printf("Broadcast callback wired to MessageQueueService")

		if sessionStore != nil {
			mqSvc.SetSessionStore(sessionStore)
			log.Printf("SessionStore wired to MessageQueueService")

			if sessionSvc != nil {
				mqSvc.SetSessionService(sessionSvc)
				log.Printf("SessionService wired to MessageQueueService")
			}
		}
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
		api.DELETE("/jobs/:id", jobHandler.Delete)

		api.GET("/matches", matchHandler.List)
		api.GET("/matches/:id", matchHandler.Get)
		api.POST("/matches/auto", matchHandler.AutoCreate)
		api.POST("/matches/:id/confirm", matchHandler.Confirm)
		api.POST("/matches/:id/decline", matchHandler.Decline)
		api.POST("/matches/:id/human-confirm", messageHandler.HandleHumanConfirm)

		api.GET("/messages", messageHandler.GetMessages)
		api.GET("/messages/:matchId", messageHandler.GetMessages)
		api.POST("/messages/:matchId", messageHandler.SendMessage)
		api.GET("/conversation/:matchId", messageHandler.GetConversation)

		api.GET("/sessions/:matchId", messageHandler.GetSessionInfo)
		api.GET("/sessions/:matchId/messages", messageHandler.GetSessionMessages)
		api.GET("/sessions/:matchId/summary", messageHandler.GetSessionSummary)
		api.POST("/sessions/:matchId/reopen", messageHandler.ReopenSession)

		api.GET("/interviews", interviewHandler.List)
		api.POST("/interviews", interviewHandler.Create)
		api.PATCH("/interviews/:id", interviewHandler.Update)
		api.GET("/interviews/match/:matchId", interviewHandler.GetByMatchID)
		api.GET("/interviews/:matchId", interviewHandler.GetByMatchID) // also support short form
		api.POST("/interviews/match/:matchId/confirm", interviewHandler.Confirm)
		api.POST("/interviews/:matchId/confirm", interviewHandler.Confirm) // short form
		api.POST("/interviews/match/:matchId/cancel", interviewHandler.Cancel)
		api.POST("/interviews/:matchId/cancel", interviewHandler.Cancel) // short form

		api.GET("/offers", offerHandler.List)
		api.GET("/offers/:matchId", offerHandler.GetByMatchID)
		api.GET("/offers/item/:id", offerHandler.Get)
		api.POST("/offers", offerHandler.Create)
		api.POST("/offers/:matchId/accept", offerHandler.Accept)
		api.POST("/offers/:matchId/decline", offerHandler.Decline)
		api.POST("/offers/item/:id/respond", offerHandler.Respond)

		api.POST("/privacy/export", privacyHandler.Export)
		api.DELETE("/privacy/account", privacyHandler.DeleteAccount)

		// Resume AI generation endpoint
		api.POST("/resume/generate", resumeHandler.Generate)
		api.POST("/resume/parse", agentHandler.ParseResume)

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

	// A2A SSE endpoint (ADK Runner streaming)
	if adkRunner != nil {
		a2aSSEHandler := handler.NewA2ASSEHandler(adkRunner)
		r.POST("/api/a2a/chat", a2aSSEHandler.Chat)
		log.Printf("A2A SSE endpoint registered at POST /api/a2a/chat")
	}

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

// initDeepAgentAndRouting initializes the Phase 2 components:
//   - RoutingSupervisor with screening/interview/offer/general agents
//   - DeepAgent for task decomposition
//   - HiringGraph (compose.NewGraph) wired into the ADK runner
//
// This augments the basic A2A supervisor with intent-based routing and
// deep task decomposition capabilities.
func initDeepAgentAndRouting(ctx context.Context, chatModel *chatmodel.EinoChatModel, baseTools []tool.BaseTool, adkRunner *eino_runner.ADKRunner) *eino_runner.ADKRunner {
	// Create workflow agents for routing supervisor
	screeningAgent, err := agent.NewScreeningAgent(ctx, chatModel, baseTools)
	if err != nil {
		log.Printf("WARNING: failed to create screening agent: %v", err)
		return nil
	}

	interviewAgent, err := agent.NewInterviewAgent(ctx, chatModel, baseTools)
	if err != nil {
		log.Printf("WARNING: failed to create interview agent: %v", err)
		return nil
	}

	offerAgent, err := agent.NewOfferAgent(ctx, chatModel, baseTools)
	if err != nil {
		log.Printf("WARNING: failed to create offer agent: %v", err)
		return nil
	}

	generalAgent, err := agent.NewGeneralRecruiterAgent(ctx, chatModel, baseTools)
	if err != nil {
		log.Printf("WARNING: failed to create general recruiter agent: %v", err)
		return nil
	}

	// Create routing supervisor (intent-based delegation)
	routingSupervisor, err := agent.NewRoutingSupervisor(ctx, screeningAgent, interviewAgent, offerAgent, generalAgent)
	if err != nil {
		log.Printf("WARNING: failed to create routing supervisor: %v", err)
		return nil
	}
	log.Printf("Routing Supervisor initialized with screening/interview/offer/general agents")

	// Initialize LocalBackend for filesystem access (read_file, write_file, etc.)
	var backend filesystem.Backend
	localBackend, err := localbk.NewBackend(ctx, &localbk.Config{})
	if err != nil {
		log.Printf("WARNING: failed to create LocalBackend: %v", err)
	} else {
		backend = localBackend
		log.Printf("LocalBackend created for DeepAgent filesystem tools")
	}

	// Create DeepAgent for task decomposition (wraps the routing supervisor as a sub-agent)
	deepInstruction := `You are a deep recruiting agent that decomposes complex hiring tasks into steps.
For each user request, break it down into sub-tasks and delegate to the appropriate sub-agent.
Use the routing supervisor for standard recruitment tasks.`
	deepRecruiter, err := agent.NewDeepRecruiterAgent(ctx, chatModel, deepInstruction, []adk.Agent{routingSupervisor}, baseTools, backend)
	if err != nil {
		log.Printf("WARNING: failed to create deep recruiter: %v", err)
		return nil
	}
	log.Printf("DeepRecruiterAgent initialized with routing supervisor as sub-agent")

	// Wrap DeepAgent into ADKRunner for streaming
	deepCpStore := memory.NewRedisCheckPointStore(redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
	}), "adk:deep:cp:")
	deepRunner := eino_runner.NewADKRunner(ctx, deepRecruiter, deepCpStore)
	log.Printf("DeepAgent Runner initialized")

	log.Printf("Phase 2 components: RoutingSupervisor + DeepAgent + HiringGraph initialized")
	return deepRunner
}
