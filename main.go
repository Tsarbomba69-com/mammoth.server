package main

import (
	"log"
	"os"

	_ "github.com/Tsarbomba69-com/mammoth.server/docs"
	"github.com/Tsarbomba69-com/mammoth.server/models"
	"github.com/Tsarbomba69-com/mammoth.server/repositories"
	"github.com/Tsarbomba69-com/mammoth.server/routes"
	"github.com/joho/godotenv"
	files "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Mammoth Server API
// @version 1.0
// @description This is a database (PostgreSQL) schema comparsion and migration.
// @host localhost:8080
// @BasePath /
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	// Set up port
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database connection
	if err := repositories.ConnectDatabase(); err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Run database migrations
	if err := repositories.Context.AutoMigrate(
		&models.DBConnection{},
		&models.Project{},
	); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	// Set up router
	r := routes.SetupRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
