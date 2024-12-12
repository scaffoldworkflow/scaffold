// main.go

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/filestore"
	"scaffold/server/manager"
	"scaffold/server/mongodb"
	"scaffold/server/rabbitmq"
	"scaffold/server/worker"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func run(ctx context.Context, channel chan struct{}) {
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

	initializeRoutes()

	rand.Seed(time.Now().UnixNano())

	mongodb.InitCollections()
	filestore.InitBucket()

	if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
		rabbitmq.RunManagerProducer()
		go manager.Run()
		go rabbitmq.RunConsumer(manager.QueueDataReceive, config.Config.ManagerQueueName)
	} else {
		rabbitmq.RunWorkerProducer()
		go worker.Run()
		go rabbitmq.RunConsumer(worker.QueueDataReceive, config.Config.WorkerQueueName)
		// go rabbitmq.RunConsumer(worker.)
	}

	routerPort := fmt.Sprintf(":%d", config.Config.Port)
	if config.Config.TLSEnabled {
		logger.Infof("", "Running with TLS loaded from %s and %s", config.Config.TLSCrtPath, config.Config.TLSKeyPath)
		router.RunTLS(routerPort, config.Config.TLSCrtPath, config.Config.TLSKeyPath)
	} else {
		router.Run(routerPort)
	}
	// for {
	// 	select {
	// 	case <-ctx.Done(): // if cancel() execute
	// 		channel <- struct{}{}
	// 		return
	// 	default:
	// 		// foobar
	// 	}

	// 	time.Sleep(1 * time.Second)
	// }
}

//	@title			Scaffold Swagger API
//	@version		2.0
//	@description	Scaffold workflow tool
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Scaffold
//	@contact.url	https://github.com/scaffoldworkflow/scaffold/issues
//	@contact.email	scaffoldworkflow@gmail.com

//	@license.name	MIT
//	@license.url	https://opensource.org/license/mit/
func main() {
	channel := make(chan struct{})
	ctx, _ := context.WithCancel(context.Background())
	run(ctx, channel)
}
