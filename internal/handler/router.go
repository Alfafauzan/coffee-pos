package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter initializes and configures the Gin engine router.
func NewRouter(appEnv string) *gin.Engine {
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Route group /api/v1
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "server is running",
			})
		})
	}

	return router
}
