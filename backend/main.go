package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"live-polling-backend/config"
	"live-polling-backend/db"
	"live-polling-backend/handlers"
	"live-polling-backend/middleware"
)

func main() {
	cfg := config.Load()

	db.ConnectMongo(cfg)
	db.ConnectRedis(cfg)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		polls := api.Group("/polls")
		{
			// Public: fetch poll details + live counts, and cast a vote.
			polls.GET("/:id", handlers.GetPoll)
			polls.POST("/:id/vote", handlers.CastVote)

			// Protected: creating and listing your own polls.
			polls.POST("", middleware.AuthRequired(), handlers.CreatePoll)
			polls.GET("", middleware.AuthRequired(), handlers.ListMyPolls)
		}
	}

	// Public real-time voting feed for a poll.
	router.GET("/ws/polls/:id", handlers.PollWebSocket)

	log.Printf("server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
