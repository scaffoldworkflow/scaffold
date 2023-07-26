// main.go

package main

import (
	"log"
	"math/rand"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/manager"
	"scaffold/server/worker"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	config.LoadConfig()

	routerPort := ":" + strconv.Itoa(config.Config.HTTPPort)

	log.Print("Running with port: " + strconv.Itoa(config.Config.HTTPPort))

	router = gin.Default()
	router.LoadHTMLGlob("templates/*")

	// Initialize the routes
	initializeRoutes()

	rand.Seed(time.Now().UnixNano())

	if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
		go manager.Run()
	} else {
		go worker.Run()
	}

	// Start serving the application
	router.Run(routerPort)
}
