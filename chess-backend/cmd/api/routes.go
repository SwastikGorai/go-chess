package main

import (
	"context"
	"net/http"
	"time"

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
	v1.Use(api.NoopAuthMiddleware())
	handlers := api.NewHandlers(app.store)
	const generalTimeout = 7*time.Second
	{
		v1.POST("/games", withTimeout(generalTimeout, handlers.CreateGame))
		v1.GET("/games/:id", withTimeout(generalTimeout, handlers.GetGame))
		v1.GET("/games/:id/legal-moves", withTimeout(generalTimeout, handlers.LegalMoves))
		v1.POST("/games/:id/moves", withTimeout(generalTimeout, handlers.MakeMove))
		v1.GET("/games/:id/status", withTimeout(generalTimeout, handlers.Status))
		v1.GET("/games/:id/history", withTimeout(generalTimeout, handlers.History))
		v1.POST("/games/:id/resign", withTimeout(generalTimeout, handlers.Resign))
		v1.POST("/games/:id/offer-draw", withTimeout(generalTimeout, handlers.OfferDraw))
		v1.POST("/games/:id/accept-draw", withTimeout(generalTimeout, handlers.AcceptDraw))
	}

	return g
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func withTimeout(d time.Duration, fn gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		fn(c)
	}
}
