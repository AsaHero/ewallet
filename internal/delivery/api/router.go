package api

import (
	"net/http"
	"strings"

	"github.com/AsaHero/e-wallet/internal/delivery"
	"github.com/AsaHero/e-wallet/internal/delivery/api/handlers"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/validation"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/AsaHero/e-wallet/internal/delivery/api/docs"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// @title Autocollection Backend Docs
// @version 0.0.1
// @description Autocollection Backend API Documentation

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @basePath /api
// @securityDefinitions.basic 	BasicAuth
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description					API Токен используется для авторизации
func NewRouter(opts *delivery.Options) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
	router.Use(otelgin.Middleware(
		"ewallet",
		otelgin.WithFilter(func(r *http.Request) bool {
			if r.URL.Path == "/ping" || strings.HasSuffix(r.URL.Path, "/api/swagger") {
				return false
			}

			return true
		}),
	))

	h := &handlers.Handlers{
		Config:              opts.Config,
		Validator:           validation.NewValidator(),
		Logger:              opts.Logger,
		UsersUsecase:        opts.UsersUsecase,
		AccountsUsecase:     opts.AccountsUsecase,
		TransactionsUsecase: opts.TransactionsUsecase,
		CategoriesUsecase:   opts.CategoriesUsecase,
		ParserUsecase:       opts.ParserUsecase,
	}

	// API routes
	api := router.Group("/api")
	{
		// Authentication (no auth required)
		api.POST("/auth/telegram", h.AuthTelegram)
		api.POST("/parse/image", h.ParseImage)
		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			protected.GET("/users/me", h.GetMe)
			protected.PATCH("/users/me", h.UpdateMe)

			// Account routes
			protected.GET("/accounts", h.GetAccounts)
			protected.POST("/accounts", h.CreateAccount)
			protected.PATCH("/accounts/:id", h.UpdateAccount)
			protected.DELETE("/accounts/:id", h.DeleteAccount)

			// Parsers routes
			protected.POST("/parse/text", h.ParseText)
			protected.POST("/parse/voice", h.ParseVoice)

			// Transaction routes
			protected.POST("/transactions", h.CreateTransaction)
			protected.GET("/transactions", h.GetTransactions)
			protected.GET("/transactions/:id", h.GetTransaction)
			protected.PUT("/transactions/:id", h.UpdateTransaction)
			protected.DELETE("/transactions/:id", h.DeleteTransaction)

			// Category routes
			protected.GET("/categories", h.GetCategories)
			protected.GET("/subcategories", h.GetSubcategories)

			// Stats routes
			protected.GET("/stats/summary", h.GetStats)
		}
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// swagger handler
	router.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
