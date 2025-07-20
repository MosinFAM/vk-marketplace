package main

import (
	"log"
	"os"

	"github.com/MosinFAM/vk-marketplace/internal/db"
	"github.com/MosinFAM/vk-marketplace/internal/handlers"
	"github.com/MosinFAM/vk-marketplace/internal/middleware"
	"github.com/MosinFAM/vk-marketplace/internal/repository"
	"github.com/gin-gonic/gin"

	_ "github.com/MosinFAM/vk-marketplace/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Marketplace API
// @version 1.0
// @description REST API for a marketplace with user auth and ads

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	conn, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewPostgresRepo(conn)
	h := &handlers.Handler{Repo: repo}

	r := gin.Default()
	r.Use(middleware.AuthMiddleware())

	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	r.POST("/ads", h.CreateAd)
	r.GET("/ads", h.ListAds)
	r.GET("/ads/:id", h.GetAdByID)

	// Swagger docs only in non-prod
	if os.Getenv("ENV") != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	log.Println("Listening on :8080")
	if err := r.Run(":8080"); err != nil {
		os.Exit(1)
	}
}
