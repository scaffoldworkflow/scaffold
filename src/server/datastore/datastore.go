package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/input"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"

	logger "github.com/jfcarter2358/go-logger"
)

type DataStore struct {
	Name    string            `json:"name" bson:"name" yaml:"name"`
	Env     map[string]string `json:"env" bson:"env" yaml:"env"`
	Files   []string          `json:"files" bson:"files" yaml:"files"`
	Created string            `json:"created" bson:"created" yaml:"created"`
	Updated string            `json:"updated" bson:"updated" yaml:"updated"`
}

func CreateDataStore(d *DataStore) error {
	currentTime := time.Now().UTC()
	d.Created = currentTime.Format("2006-01-02T15:04:05Z")
	d.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	dd, err := GetDataStoreByWorkflow(d.Name)
	if err != nil {
		return fmt.Errorf("error getting datastores: %s", err.Error())
	}
	if dd != nil {
		return fmt.Errorf("datastore already exists with name %s", d.Name)
	}

	_, err = mongodb.Collections[constants.MONGODB_DATASTORE_COLLECTION_NAME].InsertOne(mongodb.Ctx, d)
	return err
}

func DeleteDataStoreByWorkflow(name string) error {
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

func GetDataStoreByWorkflow(name string) (*DataStore, error) {
	filter := bson.M{"name": name}

	datastores, err := FilterDataStores(filter)

	if err != nil {
		return nil, err
	}

	if len(datastores) == 0 {
		return nil, nil
	}

	if len(datastores) > 1 {
		return nil, fmt.Errorf("multiple datastores found with name %s", name)
	}

	return datastores[0], nil
}

func UpdateDataStoreByWorkflow(name string, d *DataStore, is []input.Input) error {
	filter := bson.M{"name": name}

	currentTime := time.Now().UTC()
	d.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	// if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
	// 	logger.Infof("", "Node is of type %s", constants.NODE_TYPE_MANAGER)
	toChange := []string{}
	old, err := GetDataStoreByWorkflow(name)
	if err != nil {
		logger.Errorf("", "Error getting datastore %s: %s\n", name, err.Error())
		return err
	}

	for key, val := range d.Env {
		if old.Env[key] != val {
			toChange = append(toChange, key)
		}
	}
	postBody, _ := json.Marshal(toChange)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s://%s:%d/api/v1/input/%s/update", config.Config.Node.ManagerProtocol, config.Config.Node.ManagerHost, config.Config.Node.ManagerPort, d.Name)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("received input update status code %d", resp.StatusCode)
	}
	// }

	collection := mongodb.Collections[constants.MONGODB_DATASTORE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, d, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return CreateDataStore(d)
		// return fmt.Errorf("no datastore found with name %s", name)
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

	return datastores, nil
}
