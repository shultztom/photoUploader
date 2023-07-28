package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Health is a controller that returns a 200 OK
// response to indicate that the service is up and running.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}
