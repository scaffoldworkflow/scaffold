package mongodb

import (
	"context"
	"fmt"
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
}
var Collections map[string]*mongo.Collection
var Ctx = context.TODO()

func InitCollections() {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", config.Config.DB.Username, config.Config.DB.Password, config.Config.DB.Host, config.Config.DB.Port, config.Config.DB.Name))
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
