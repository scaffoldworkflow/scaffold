package cascade

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/input"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/utils"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type Cascade struct {
	Version string            `json:"version" bson:"version"`
	Name    string            `json:"name" bson:"name"`
	Inputs  []input.Input     `json:"inputs" bson:"inputs"`
	Tasks   []task.Task       `json:"tasks" bson:"tasks"`
	Created string            `json:"created" bson:"created"`
	Updated string            `json:"updated" bson:"updated"`
	Groups  []string          `json:"groups" bson:"groups"`
	Links   map[string]string `json:"links" bson:"links"`
}

func CreateCascade(c *Cascade) error {
	currentTime := time.Now().UTC()
	c.Created = currentTime.Format("2006-01-02T15:04:05Z")
	c.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	if _, err := GetCascadeByName(c.Name); err == nil {
		return fmt.Errorf("cascade already exists with name %s", c.Name)
	}

	_, err := mongodb.Collections[constants.MONGODB_CASCADE_COLLECTION_NAME].InsertOne(mongodb.Ctx, c)

	if err != nil {
		return err
	}

	for _, t := range c.Tasks {
		t.Cascade = c.Name
		t.RunNumber = 0
		if err := task.CreateTask(&t); err != nil {
			return err
		}
	}

	for _, i := range c.Inputs {
		i.Cascade = c.Name
		if err := input.CreateInput(&i); err != nil {
			return err
		}
	}

	d := &datastore.DataStore{
		Name:    c.Name,
		Env:     make(map[string]string),
		Files:   make([]string, 0),
		Created: c.Created,
		Updated: c.Updated,
	}

	for _, val := range c.Inputs {
		d.Env[val.Name] = val.Default
	}

	err = datastore.CreateDataStore(d)
	return err
}

func DeleteCascadeByName(name string) error {
	filter := bson.M{"name": name}

	collection := mongodb.Collections[constants.MONGODB_CASCADE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no cascade found with name %s", name)
	}

	if err := task.DeleteTasksByCascade(name); err != nil {
		return err
	}

	if err := input.DeleteInputsByCascade(name); err != nil {
		return err
	}

	err = datastore.DeleteDataStoreByName(name)
	return err

}

func GetAllCascades() ([]*Cascade, error) {
	filter := bson.M{}

	cascades, err := FilterCascades(filter)

	return cascades, err
}

func GetCascadeByName(name string) (*Cascade, error) {
	filter := bson.M{"name": name}

	cascades, err := FilterCascades(filter)

	if err != nil {
		return nil, err
	}

	if len(cascades) == 0 {
		return nil, fmt.Errorf("no cascade found with name %s", name)
	}

	if len(cascades) > 1 {
		return nil, fmt.Errorf("multiple cascades found with name %s", name)
	}

	return cascades[0], nil
}

func UpdateCascadeByName(name string, c *Cascade) error {
	filter := bson.M{"name": name}

	currentTime := time.Now().UTC()
	c.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_CASCADE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	states, err := state.GetStatesByCascade(name)
	if err != nil {
		return err
	}
	tasks, err := task.GetTasksByCascade(name)
	if err != nil {
		return nil
	}

	result, err := collection.ReplaceOne(ctx, filter, c, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no cascade found with name %s", name)
	}

	taskNames := make([]string, len(states))

	for idx, t := range tasks {
		taskNames[idx] = t.Name
	}

	newNames := make([]string, len(c.Tasks))

	for idx, t := range c.Tasks {
		if !utils.Contains(taskNames, t.Name) {
			t.Cascade = name
			logger.Debugf("", "Creating task %s with cascade %s", t.Name, t.Cascade)
			if err := task.CreateTask(&t); err != nil {
				return err
			}
			continue
		}
		logger.Debugf("", "Updating task %s with cascade %s", t.Name, name)
		if err := task.UpdateTaskByNames(name, t.Name, &t); err != nil {
			return err
		}
		newNames[idx] = t.Name
	}

	logger.Debugf("", "Old tasks: %v", taskNames)
	logger.Debugf("", "New tasks: %v", newNames)

	for _, t := range tasks {
		if !utils.Contains(newNames, t.Name) {
			logger.Debugf("", "Removing task %s", t.Name)
			if err := task.DeleteTaskByNames(name, t.Name); err != nil {
				return err
			}
		}
	}

	return err
}

func FilterCascades(filter interface{}) ([]*Cascade, error) {
	// A slice of tasks for storing the decoded documents
	var cascades []*Cascade

	collection := mongodb.Collections[constants.MONGODB_CASCADE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return cascades, err
	}

	for cur.Next(ctx) {
		var c Cascade
		err := cur.Decode(&c)
		if err != nil {
			return cascades, err
		}

		cascades = append(cascades, &c)
	}

	if err := cur.Err(); err != nil {
		return cascades, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(cascades) == 0 {
		return cascades, mongo.ErrNoDocuments
	}

	return cascades, nil
}
