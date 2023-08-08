// main.go

package main

import (
	"fmt"
	"math/rand"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/logger"
	"scaffold/server/manager"
	"scaffold/server/worker"
	"time"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	config.LoadConfig()
	logger.SetLevel(config.Config.LogLevel)

	router = gin.New()
	router.Use(gin.LoggerWithFormatter(logger.ConsoleLogFormatter))
	router.Use(gin.Recovery())

	logger.Infof("", "Running with port: %d", config.Config.HTTPPort)

	router.LoadHTMLGlob("templates/*")
	initializeRoutes()

	rand.Seed(time.Now().UnixNano())

	if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
		go manager.Run()
	} else {
		go worker.Run()
	}

	routerPort := fmt.Sprintf(":%d", config.Config.HTTPPort)
	router.Run(routerPort)
}
