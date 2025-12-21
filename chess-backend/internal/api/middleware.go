package api

import "github.com/gin-gonic/gin"

func NoopAuthMiddleware() gin.HandlerFunc { //TODO: some auth middleware
	return func(c *gin.Context) {
		c.Next()
	}
}
