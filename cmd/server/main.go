package main

import (
	"log"
	"os"

	"github.com/abhishek-sengar/ytmanager/internal/api"
	"github.com/abhishek-sengar/ytmanager/internal/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	err = db.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Setup Gin router
	router := gin.Default()

	// Allow requests from your React frontend
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Root endpoint for health check
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "YouTube Manager API is running",
		})
	})

	// Ping endpoint for testing
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Auth routes (login/signup)
	router.POST("/signup", api.Signup)
	router.POST("/login", api.Login)

	// Public YouTube auth route (needed for OAuth flow)
	router.GET("/api/youtube/auth", api.YoutubeAuth)
	router.GET("/api/youtube/callback", api.YoutubeCallback)

	// Protected routes
	protected := router.Group("/")
	protected.Use(api.AuthMiddleware())

	// Project related routes
	protected.POST("/projects", api.CreateProject)
	protected.GET("/projects", api.GetUserProjects) // Generalized endpoint for owner/editor
	protected.GET("/projects/:id", api.GetProjectDetailsByID)
	protected.POST("/projects/:id/notes", api.AddNote)
	protected.POST("/projects/:id/approve", api.ApproveProject)
	protected.POST("/projects/:id/reject", api.RejectProject)
	protected.GET("/projects/recent", api.GetRecentProjects)

	// Sidebar data for both owners and editors
	protected.GET("/sidebar-data", api.GetSidebarData)

	// Protected YouTube integration routes
	protected.GET("/api/youtube/unattached-channels", api.GetUnattachedChannels)
	protected.POST("/api/youtube/add-channels", api.AddChannelsToDashboard)

	// Start the server
	port := os.Getenv("PORT")
	router.Run(":" + port)
}
