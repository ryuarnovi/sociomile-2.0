package main

import (
	"log"
	"os"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Determine DSN
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// build DSN from individual env vars if DATABASE_URL not set
		user := os.Getenv("DB_USER")
		if user == "" {
			user = cfg.DBUser
		}
		pass := os.Getenv("DB_PASSWORD")
		if pass == "" {
			pass = cfg.DBPassword
		}
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = cfg.DBHost
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = cfg.DBPort
		}
		name := os.Getenv("DB_NAME")
		if name == "" {
			name = cfg.DBName
		}
		if user != "" && host != "" && name != "" {
			dsn = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=disable"
		}
	}
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set and could not be constructed from env")
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database (%s): %v", dsn, err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient := config.NewRedisClient(cfg)
	defer redisClient.Close()

	// Initialize RabbitMQ
	rabbitConn, rabbitCh := config.NewRabbitMQ(cfg)
	if rabbitConn != nil {
		defer rabbitConn.Close()
		defer rabbitCh.Close()
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	eventRepo := repository.NewEventRepository(db)
	channelRepo := repository.NewChannelRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	conversationService := service.NewConversationService(conversationRepo, messageRepo, customerRepo, eventRepo, ticketRepo, redisClient, rabbitCh)
	ticketService := service.NewTicketService(ticketRepo, conversationRepo, eventRepo, rabbitCh)
	userService := service.NewUserService(userRepo)
	channelService := service.NewChannelService(channelRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	conversationHandler := handler.NewConversationHandler(conversationService)
	ticketHandler := handler.NewTicketHandler(ticketService)
	webhookHandler := handler.NewWebhookHandler(conversationService)
	userHandler := handler.NewUserHandler(userService)
	channelHandler := handler.NewChannelHandler(channelService)

	messageHandler := handler.NewMessageHandler(conversationService)

	// WebSocket handler (for realtime)
	websocketHandler := handler.NewWebsocketHandler(rabbitConn, cfg.JWTSecret)

	// Initialize Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/auth/register", authHandler.Register)

		// Webhook (simulated channel)
		v1.POST("/channel/webhook", webhookHandler.HandleWebhook)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Channels
			protected.GET("/channels", channelHandler.List)
			protected.GET("/channels/:id", channelHandler.GetByID)
			protected.POST("/channels", channelHandler.Create)
			protected.PUT("/channels/:id", channelHandler.Update)
			protected.DELETE("/channels/:id", channelHandler.Delete)

			// Additional conversation CRUD for frontend
			protected.POST("/conversations", conversationHandler.Create)
			protected.PUT("/conversations/:id", conversationHandler.Update)
			protected.DELETE("/conversations/:id", conversationHandler.Delete)
			// Conversations
			protected.GET("/conversations", conversationHandler.List)
			protected.GET("/conversations/:id", conversationHandler.GetByID)
			protected.POST("/conversations/:id/messages", conversationHandler.SendMessage)
			protected.POST("/conversations/:id/assign", conversationHandler.Assign)
			protected.POST("/conversations/:id/close", conversationHandler.Close)
			// Tickets per conversation and selection
			protected.GET("/conversations/:id/tickets", conversationHandler.ListTickets)
			protected.PUT("/conversations/:id/selected-ticket", conversationHandler.SetSelectedTicket)

			// Messages
			protected.DELETE("/messages/:id", messageHandler.Delete)

			// Tickets
			protected.GET("/tickets", ticketHandler.List)
			protected.GET("/tickets/:id", ticketHandler.GetByID)
			protected.POST("/conversations/:id/escalate", ticketHandler.Escalate)
			// Extra ticket CRUD for frontend
			protected.POST("/tickets", ticketHandler.Create)
			protected.PUT("/tickets/:id", ticketHandler.Update)
			protected.DELETE("/tickets/:id", ticketHandler.Delete)

			// Admin only routes
			admin := protected.Group("")
			admin.Use(middleware.AdminOnly())
			{
				admin.PUT("/tickets/:id/status", ticketHandler.UpdateStatus)
				admin.GET("/users", userHandler.List)
				admin.POST("/users", userHandler.Create)
				admin.PUT("/users/:id", userHandler.Update)
				admin.DELETE("/users/:id", userHandler.Delete)
			}
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// WebSocket endpoint (upgrades outside /api path)
	router.GET("/ws", websocketHandler.Handle)

	// Start server
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
