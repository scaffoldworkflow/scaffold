package mongodb

import (
	"context"
	"log"
	"scaffold/server/config"
	"scaffold/server/constants"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collectionNames = []string{
	constants.MONGODB_CASCADE_COLLECTION_NAME,
	constants.MONGODB_DATASTORE_COLLECTION_NAME,
	constants.MONGODB_STATE_COLLECTION_NAME,
	constants.MONGODB_USER_COLLECTION_NAME,
	constants.MONGODB_TASK_COLLECTION_NAME,
	constants.MONGODB_INPUT_COLLECTION_NAME,
	constants.MONGODB_WEBHOOK_COLLECTION_NAME,
	constants.MONGODB_HISTORY_COLLECTION_NAME,
}
var Collections map[string]*mongo.Collection
var Ctx = context.TODO()

func InitCollections() {
	clientOptions := options.Client().ApplyURI(config.Config.DBConnectionString)
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(Ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	Collections = make(map[string]*mongo.Collection)

	for _, collection := range collectionNames {
		Collections[collection] = client.Database(config.Config.DB.Name).Collection(collection)
	}
}
