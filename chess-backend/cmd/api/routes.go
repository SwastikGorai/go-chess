package main

import (
	"net/http"

	"chess-backend/internal/api"

	"github.com/gin-gonic/gin"
)

func (app *app) routes() http.Handler {
	g := gin.Default()

	health := g.Group("/health")
	{
		health.GET("", healthHandler)
	}

	v1 := g.Group("/api/v1")
	{
		v1.Use(api.NoopAuthMiddleware())
		handlers := api.NewHandlers(app.store)
		v1.POST("/games", handlers.CreateGame)
		v1.GET("/games/:id", handlers.GetGame)
		v1.GET("/games/:id/legal-moves", handlers.LegalMoves)
		v1.POST("/games/:id/moves", handlers.MakeMove)
		v1.GET("/games/:id/status", handlers.Status)
		v1.GET("/games/:id/history", handlers.History)
		v1.POST("/games/:id/resign", handlers.Resign)
		v1.POST("/games/:id/offer-draw", handlers.OfferDraw)
		v1.POST("/games/:id/accept-draw", handlers.AcceptDraw)
	}

	return g
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
