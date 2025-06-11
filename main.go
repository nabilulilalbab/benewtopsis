package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/nabilulilalbab/TopsisByme/controllers/topsiscontroller"
	usercontroller "github.com/nabilulilalbab/TopsisByme/controllers/userController"
	"github.com/nabilulilalbab/TopsisByme/database"
	_ "github.com/nabilulilalbab/TopsisByme/docs" // Import docs
	"github.com/nabilulilalbab/TopsisByme/middleware"
)

// @title TOPSIS API
// @version 1.0
// @description This is a TOPSIS (Technique for Order Preference by Similarity to Ideal Solution) API server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func init() {
	log.Println("Hit init")

	// Load environment variables
	// initializers.LoadVariables()

	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize database
	database.InitDB()
	if database.DB == nil {
		panic("Database not initialized")
	}

	log.Println("Done Exec init")
}

func main() {
	// Initialize Gin router
	router := gin.Default()

	// Trust all proxies (allow any proxy)
	router.SetTrustedProxies(nil)

	// Configure CORS
	setupCORS(router)

	// Add database middleware
	router.Use(func(c *gin.Context) {
		c.Set("db", database.DB)
		c.Next()
	})

	// Global OPTIONS handler
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup routes
	setupRoutes(router)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Gunakan 8080 sebagai default jika tidak ada (untuk lokal)
	}

	// Start server
	log.Println("Server starting on 0.0.0.0:8080")
	log.Println("Swagger UI available at: http://localhost:8080/swagger/index.html")
	router.Run(":" + port)
}

// setupCORS configures CORS settings
func setupCORS(router *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Content-Length",
		"Accept",
		"Accept-Encoding",
		"Authorization",
		"X-CSRF-Token",
		"X-Requested-With",
	}
	config.ExposeHeaders = []string{"Content-Length", "Content-Type"}
	config.AllowCredentials = false // Must be false when AllowAllOrigins is true
	config.MaxAge = 86400           // 24 hours

	router.Use(cors.New(config))
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine) {
	// User authentication routes
	authRoutes := router.Group("/api")
	{
		authRoutes.POST("/signup", usercontroller.Signup)
		authRoutes.POST("/login", usercontroller.Login)
		authRoutes.POST("/logout", middleware.RequireAuth, usercontroller.Logout)
		authRoutes.GET("/validate", middleware.RequireAuth, usercontroller.Validate)
	}

	// TOPSIS routes (all require authentication)
	topsisRoutes := router.Group("/api/topsis")
	topsisRoutes.Use(middleware.RequireAuth)
	{
		topsisRoutes.POST("/", topsiscontroller.HandleTopsis)
		topsisRoutes.POST("/save", topsiscontroller.SaveTopsisResult)
		topsisRoutes.GET("/history", topsiscontroller.GetAllTopsisHistory)
		topsisRoutes.GET("/:id", topsiscontroller.TopsisGetById)
		topsisRoutes.PUT("/:id", topsiscontroller.UpdateTopsisResult)
	}
}
