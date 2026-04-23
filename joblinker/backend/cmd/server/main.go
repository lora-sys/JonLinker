package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"joblinker/internal/handler"
	"joblinker/internal/middleware"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/internal/service"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
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

	agentSvc := service.NewAgentService(agentRepo, userRepo, securityRepo)
	matchSvc := service.NewMatchService(matchRepo, agentRepo, jobRepo)
	messageSvc := service.NewMessageService(messageRepo, matchRepo, agentRepo)
	securitySvc := service.NewSecurityService(securityRepo)
	interviewSvc := service.NewInterviewService(interviewRepo, matchRepo, messageSvc, securitySvc)
	offerSvc := service.NewOfferService(offerRepo, matchRepo, jobRepo, securitySvc)
	privacySvc := service.NewPrivacyService(userRepo, agentRepo, matchRepo)
	adminHandler := handler.NewAdminHandler()

	authHandler := handler.NewAuthHandler(userRepo, securitySvc)
	agentHandler := handler.NewAgentHandler(agentSvc)
	jobHandler := handler.NewJobHandler(jobRepo, agentRepo)
	matchHandler := handler.NewMatchHandler(matchSvc)
	interviewHandler := handler.NewInterviewHandler(interviewSvc)
	offerHandler := handler.NewOfferHandler(offerSvc)
	privacyHandler := handler.NewPrivacyHandler(privacySvc)
	messageHandler := handler.NewMessageHandler(messageRepo, matchRepo, agentRepo)
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
		api.POST("/matches/:id/confirm", matchHandler.Confirm)

		api.GET("/interviews", interviewHandler.List)
		api.POST("/interviews", interviewHandler.Create)
		api.PATCH("/interviews/:id", interviewHandler.Update)

		api.GET("/offers/:id", offerHandler.Get)
		api.POST("/offers/:id/respond", offerHandler.Respond)

		api.POST("/privacy/export", privacyHandler.Export)
		api.DELETE("/privacy/account", privacyHandler.DeleteAccount)

		// Admin endpoints
		api.GET("/admin/dashboard", adminHandler.GetDashboard)

		// Messages WebSocket and REST
		api.GET("/messages/:matchId/ws", messageHandler.HandleWebSocket)
		api.GET("/messages/:matchId", messageHandler.GetConversation)
		api.POST("/messages/:matchId", messageHandler.SendMessage)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
