package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/abhishek-sengar/ytmanager/internal/api"
	"github.com/abhishek-sengar/ytmanager/internal/db"
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

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	router.POST("/signup", api.Signup)
	router.GET("/login", api.Login)

	protected := router.Group("/")
	protected.Use(api.AuthMiddleware())

	protected.POST("/projects", api.CreateProject)
	protected.GET("/projects", api.GetProjects)
	protected.POST("/projects/:id/notes", api.AddNote)
	protected.POST("/projects/:id/approve", api.ApproveProject)
	protected.POST("/projects/:id/reject", api.RejectProject)

	router.GET("/youtube/auth", api.YoutubeAuth)
	router.GET("/youtube/callback", api.YoutubeCallback)

	// etc.

	port := os.Getenv("PORT")
	router.Run(":" + port)
}
