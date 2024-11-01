package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// HealthCheck hanterar health check-förfrågningar
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
} 