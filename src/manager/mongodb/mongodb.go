package mongodb

import (
	"context"
	"scaffold/manager/config"
	"scaffold/manager/constants"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collectionNames = []string{
	constants.MONGODB_USER_COLLECTION_NAME,
	constants.MONGODB_HISTORY_COLLECTION_NAME,
	constants.MONGODB_RUNBOOK_COLLECTION_NAME,
	constants.MONGODB_ALERT_COLLECTION_NAME,
	constants.MONGODB_PROJECT_COLLECTION_NAME,
	constants.MONGODB_ENVIRONMENT_COLLECTION_NAME,
	constants.MONGODB_SERVICE_COLLECTION_NAME,
	constants.MONGODB_RELEASE_COLLECTION_NAME,
}
var Collections map[string]*mongo.Collection
var Ctx = context.TODO()

func InitCollections() {
	clientOptions := options.Client().ApplyURI(config.Config.DBConnectionString)
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		logger.Fatalf("", "Unable to connect to MongoDB: %s", err)
	}

	err = client.Ping(Ctx, nil)
	if err != nil {
		logger.Fatalf("", "Unable to ping to MongoDB: %s", err)
	}

	Collections = make(map[string]*mongo.Collection)

	for _, collection := range collectionNames {
		logger.Debugf("", "Connecting to collection %s", collection)
		Collections[collection] = client.Database(config.Config.DB.Name).Collection(collection)
	}
}
