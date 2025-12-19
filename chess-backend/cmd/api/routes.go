package main

import (
	"net/http"

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
		// #TODO: apis and all
	}

	auth_group := v1.Group("/")
	auth_group.Use()

	return g
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
