package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	"github.com/dastanchynybek/rag-chat/backend/internal/config"
	"github.com/dastanchynybek/rag-chat/backend/internal/handler"
	"github.com/dastanchynybek/rag-chat/backend/internal/middleware"
	"github.com/dastanchynybek/rag-chat/backend/internal/pkg/llm"
	"github.com/dastanchynybek/rag-chat/backend/internal/repository"
	"github.com/dastanchynybek/rag-chat/backend/internal/service"

	_ "github.com/dastanchynybek/rag-chat/backend/docs"
)

// @title Mini RAG Chat API
// @version 1.0
// @description RAG-powered chat API with document upload, embedding, and LLM-based Q&A
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}" to authenticate
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Database
	pool, err := repository.NewPostgresPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database")

	// Migrations
	if err := repository.RunMigrations(cfg.Database.URL, migrationsPath()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations applied")

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	docRepo := repository.NewDocumentRepository(pool)
	chunkRepo := repository.NewChunkRepository(pool)
	chatRepo := repository.NewChatRepository(pool)

	// LLM Client
	var llmClient llm.LLMClient
	switch cfg.LLMProvider {
	case "gemini":
		geminiClient, err := llm.NewGeminiClient(
			ctx,
			cfg.Gemini.APIKey,
			cfg.Gemini.ChatModel,
			cfg.Gemini.EmbeddingModel,
		)
		if err != nil {
			log.Fatalf("Failed to create Gemini client: %v", err)
		}
		llmClient = geminiClient
	case "openai":
		llmClient = llm.NewOpenAIClient(
			cfg.OpenAI.APIKey,
			cfg.OpenAI.ChatModel,
			cfg.OpenAI.EmbeddingModel,
		)
	}

	// Services
	authSvc := service.NewAuthService(
		userRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	docSvc := service.NewDocumentService(docRepo, chunkRepo, llmClient)
	chatSvc := service.NewChatService(chatRepo, chunkRepo, llmClient)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, cfg.JWT.RefreshTokenTTL)
	docHandler := handler.NewDocumentHandler(docSvc)
	chatHandler := handler.NewChatHandler(chatSvc)

	// Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
		BodyLimit:    11 * 1024 * 1024, // 11MB (slightly above 10MB file limit)
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.CORS())

	// Routes
	api := app.Group("/api")

	// Auth routes (no auth required)
	auth := api.Group("/auth")
	auth.Use(middleware.AuthRateLimiter())
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// Protected routes
	protected := api.Group("", middleware.Auth(cfg.JWT.Secret))

	// Document routes
	docs := protected.Group("/documents")
	docs.Post("", middleware.UploadRateLimiter(), docHandler.Upload)
	docs.Get("", middleware.DefaultRateLimiter(), docHandler.List)
	docs.Delete("/:id", middleware.DefaultRateLimiter(), docHandler.Delete)

	// Chat routes
	chat := protected.Group("/chat")
	chat.Post("", middleware.ChatRateLimiter(), chatHandler.Ask)
	chat.Get("/history", middleware.DefaultRateLimiter(), chatHandler.GetHistory)
	chat.Get("/history/:sessionId", middleware.DefaultRateLimiter(), chatHandler.GetMessages)

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + cfg.Server.Port
		log.Printf("Server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func migrationsPath() string {
	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		return path
	}
	if _, err := os.Stat("/migrations"); err == nil {
		return "/migrations"
	}
	return "migrations"
}
