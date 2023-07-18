// main.go

package main

import (
	"log"
	"scaffold/server/config"
	"scaffold/server/manager"
	"strconv"

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

	go manager.Run()

	// Start serving the application
	router.Run(routerPort)
}
