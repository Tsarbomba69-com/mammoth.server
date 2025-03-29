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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	repositories.ConnectDatabase()
	repositories.Context.AutoMigrate(&models.DBConnection{}, &models.Project{})
	r := routes.SetupRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))
	r.Run(":" + port)
}
