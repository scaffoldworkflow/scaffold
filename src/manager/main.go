// main.go

package main

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"scaffold/manager/config"
	"scaffold/manager/manager"
	"scaffold/manager/mongodb"
	"scaffold/manager/monitor"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func run() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	config.LoadConfig()
	logger.SetLevel(config.Config.LogLevel)
	logger.SetFormat(config.Config.LogFormat)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: config.Config.TLSSkipVerify}

	router = gin.New()
	router.Use(gin.LoggerWithFormatter(logger.ConsoleLogFormatter))
	router.Use(gin.Recovery())

	logger.Infof("", "Running with port: %d", config.Config.Port)

	logger.Debugf("", "Initializing routes")
	initializeRoutes()

	rand.Seed(time.Now().UnixNano())

	logger.Debugf("", "Initializing mongo collections")
	mongodb.InitCollections()

	logger.Debugf("", "Starting manager")
	manager.Run()

	if err := monitor.MonitorDrift(); err != nil {
		logger.Warnf("", "Unable to monitor run drift: %s", err)
	}

	logger.Infof("", "Starting router at %d", config.Config.Port)
	routerPort := fmt.Sprintf(":%d", config.Config.Port)
	if config.Config.TLSEnabled {
		logger.Infof("", "Running with TLS loaded from %s and %s", config.Config.TLSCrtPath, config.Config.TLSKeyPath)
		router.RunTLS(routerPort, config.Config.TLSCrtPath, config.Config.TLSKeyPath)
	} else {
		logger.Tracef("", "Running without TLS")
		if err := router.Run(routerPort); err != nil {
			logger.Fatalf("", "Router failed to run: %s", err)
		}
	}
	logger.Warnf("", "Router has shut down")
}

//	@title			Scaffold Swagger API
//	@version		2.0
//	@description	Scaffold workflow tool
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Scaffold
//	@contact.url	https://github.com/scaffoldworkflow/scaffold/issues
//	@contact.email	scaffoldworkflow@gmail.com

// @license.name	MIT
// @license.url	https://opensource.org/license/mit/
func main() {
	run()
}
