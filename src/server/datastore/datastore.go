package datastore

import (
	"fmt"
	"scaffold/server/constants"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type DataStore struct {
	Name    string            `json:"name" bson:"name"`
	Env     map[string]string `json:"env" bson:"env"`
	Files   []string          `json:"files" bson:"files"`
	Created string            `json:"created" bson:"created"`
	Updated string            `json:"updated" bson:"updated"`
}

func CreateDataStore(d *DataStore) error {
	currentTime := time.Now().UTC()
	d.Created = currentTime.Format("2006-01-02T15:04:05Z")
	d.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	if _, err := GetDataStoreByName(d.Name); err == nil {
		return fmt.Errorf("datastore already exists with name %s", d.Name)
	}

	_, err := mongodb.Collections[constants.MONGODB_DATASTORE_COLLECTION_NAME].InsertOne(mongodb.Ctx, d)
	return err
}

func DeleteDataStoreByName(name string) error {
	filter := bson.M{"name": name}

	collection := mongodb.Collections[constants.MONGODB_DATASTORE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no datastore found with name %s", name)
	}

	return nil

}

func GetAllDataStores() ([]*DataStore, error) {
	filter := bson.D{{}}

	datastores, err := FilterDataStores(filter)

	return datastores, err
}

func GetDataStoreByName(name string) (*DataStore, error) {
	filter := bson.M{"name": name}

	datastores, err := FilterDataStores(filter)

	if err != nil {
		return nil, err
	}

	if len(datastores) == 0 {
		return nil, fmt.Errorf("no datastore found with name %s", name)
	}

	if len(datastores) > 1 {
		return nil, fmt.Errorf("multiple datastores found with name %s", name)
	}

	return datastores[0], nil
}

func UpdateDataStoreByName(name string, d *DataStore) error {
	filter := bson.M{"name": name}

	currentTime := time.Now().UTC()
	d.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_DATASTORE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, d, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no datastore found with name %s", name)
	}

	return nil
}

func FilterDataStores(filter interface{}) ([]*DataStore, error) {
	// A slice of tasks for storing the decoded documents
	var datastores []*DataStore

	collection := mongodb.Collections[constants.MONGODB_DATASTORE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return datastores, err
	}

	for cur.Next(ctx) {
		var d DataStore
		err := cur.Decode(&d)
		if err != nil {
			return datastores, err
		}

		datastores = append(datastores, &d)
	}

	if err := cur.Err(); err != nil {
		return datastores, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(datastores) == 0 {
		return datastores, mongo.ErrNoDocuments
	}

	return datastores, nil
}
