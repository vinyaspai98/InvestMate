package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"investmate-backend/internal/handlers"
	"investmate-backend/internal/middleware"
	"investmate-backend/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize Firebase
	ctx := context.Background()
	firebaseApp, err := config.InitializeFirebase(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Initialize Firestore client
	firestoreClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer firestoreClient.Close()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(firebaseApp, firestoreClient)
	dashboardHandler := handlers.NewDashboardHandler(firestoreClient)
	investmentHandler := handlers.NewInvestmentHandler(firestoreClient, cfg)
	loanHandler := handlers.NewLoanHandler(firestoreClient)
	transactionHandler := handlers.NewTransactionHandler(firestoreClient)
	chartHandler := handlers.NewChartHandler(firestoreClient)
	gmailHandler := handlers.NewGmailHandler(firestoreClient)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Disable automatic trailing slash redirect
	router.RedirectTrailingSlash = false

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORSAllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour

	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "timestamp": time.Now().Unix()})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Authentication routes (no middleware needed)
		auth := api.Group("/auth")
		{
			auth.POST("/verify-token", authHandler.VerifyToken)
			auth.GET("/user", middleware.AuthMiddleware(firebaseApp), authHandler.GetProfile)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(firebaseApp))
		{
			// Dashboard routes
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/summary", dashboardHandler.GetDashboardSummary)
			}

			// User profile routes
			users := protected.Group("/users")
			{
				users.GET("/profile", authHandler.GetProfile)
				users.PUT("/profile", authHandler.UpdateProfile)
				users.PUT("/preferences", authHandler.UpdatePreferences)
			}

			// Investment routes
			investments := protected.Group("/investments")
			{
				// Category-based routes (should come first)
				investments.GET("/category/:category", investmentHandler.GetInvestmentsByCategory)
				investments.GET("/category/:category/summary", investmentHandler.GetCategorySummary)

				// CRUD routes for specific investments
				investments.GET("/detail/:id", investmentHandler.GetInvestment)
				investments.PUT("/:id", investmentHandler.UpdateInvestment)
				investments.DELETE("/:id", investmentHandler.DeleteInvestment)

				// Alpha Vantage proxy routes
				investments.GET("/search", investmentHandler.SearchSymbols)
				investments.GET("/price/:symbol", investmentHandler.GetPriceData)
			}

			// Loan routes
			loans := protected.Group("/loans")
			{
				loans.GET("/", loanHandler.GetAllLoans)
				loans.GET("/summary", loanHandler.GetLoanSummary)
				loans.GET("/:id", loanHandler.GetLoan)
				loans.POST("", loanHandler.CreateLoan)
				loans.PUT("/:id", loanHandler.UpdateLoan)
				loans.DELETE("/:id", loanHandler.DeleteLoan)
			}

			// Transaction routes
			transactions := protected.Group("/transactions")
			{
				transactions.GET("/", transactionHandler.GetAllTransactions)
				transactions.GET("/category/:category", transactionHandler.GetTransactionsByCategory)
				transactions.GET("/related/:relatedId", transactionHandler.GetTransactionsByRelatedID)
				transactions.POST("/", transactionHandler.CreateTransaction)
			}

			// Chart routes
			charts := protected.Group("/charts")
			{
				charts.GET("/:category/:period", chartHandler.GetCategoryChart)
				charts.GET("/portfolio/:period", chartHandler.GetPortfolioChart)
			}

			// Gmail sync routes
			gmail := protected.Group("/gmail")
			{
				gmail.POST("/sync", gmailHandler.SyncGmail)
			}
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful server shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 InvestMate API server started on port %s", cfg.Port)
	log.Printf("📊 Dashboard: http://localhost:%s/api/v1/dashboard/summary", cfg.Port)
	log.Printf("🔍 Health check: http://localhost:%s/health", cfg.Port)
	log.Printf("🌍 Environment: %s", cfg.Environment)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited gracefully")
}
