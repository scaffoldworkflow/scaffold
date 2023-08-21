package api

import (
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/health"

	"github.com/gin-gonic/gin"
)

//	@summary		Check if a node is healthy
//	@description	Get node health
//	@tags			manager
//	@tags			worker
//	@tags			health
//	@success		200
//	@failure		503
//	@router			/health/healthy [get]
func Healthy(c *gin.Context) {
	if health.IsHealthy {
		c.JSON(http.StatusOK, gin.H{"version": constants.VERSION})
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

//	@summary		Check if a node is ready
//	@description	Get node readiness
//	@tags			manager
//	@tags			worker
//	@tags			health
//	@success		200
//	@failure		503
//	@router			/health/ready [get]
func Ready(c *gin.Context) {
	if health.IsReady {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

//	@summary		Check if a worker node is available
//	@description	Get status from node succeeding if not containers are running
//	@tags			worker
//	@tags			health
//	@success		200
//	@failure		503
//	@router			/health/available [get]
func Available(c *gin.Context) {
	if health.IsAvailable {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}
