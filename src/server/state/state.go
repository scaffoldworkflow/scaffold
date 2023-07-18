package state

import (
	"fmt"
	"scaffold/server/constants"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"scaffold/server/mongodb"
)

type StateTask struct {
	Name     string `json:"name" bson:"name"`
	Status   string `json:"status" bson:"status"`
	Started  string `json:"started" bson:"started"`
	Finished string `json:"finished" bson:"finished"`
	Error    string `json:"error" bson:"error"`
	Output   string `json:"output" bson:"output"`
}

type State struct {
	Name    string      `json:"name" bson:"name"`
	Tasks   []StateTask `json:"tasks" bson:"tasks"`
	Created string      `json:"created" bson:"created"`
	Updated string      `json:"updated" bson:"updated"`
}

func CreateState(s *State) error {
	currentTime := time.Now().UTC()
	s.Created = currentTime.Format("2006-01-02T15:04:05Z")
	s.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	if _, err := GetStateByName(s.Name); err == nil {
		return fmt.Errorf("state already exists with name %s", s.Name)
	}

	_, err := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME].InsertOne(mongodb.Ctx, s)
	return err
}

func DeleteStateByName(name string) error {
	filter := bson.M{"name": name}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no state found with name %s", name)
	}

	return nil

}

func GetAllStates() ([]*State, error) {
	filter := bson.D{{}}

	states, err := FilterStates(filter)

	return states, err
}

func GetStateByName(name string) (*State, error) {
	filter := bson.M{"name": name}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no state found with name %s", name)
	}

	if len(states) > 1 {
		return nil, fmt.Errorf("multiple states found with name %s", name)
	}

	return states[0], nil
}

func UpdateStateByName(name string, s *State) error {
	filter := bson.M{"name": name}

	currentTime := time.Now().UTC()
	s.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.ReplaceOne(ctx, filter, s)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no state found with name %s", name)
	}

	return nil
}

func FilterStates(filter interface{}) ([]*State, error) {
	// A slice of tasks for storing the decoded documents
	var states []*State

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return states, err
	}

	for cur.Next(ctx) {
		var s State
		err := cur.Decode(&s)
		if err != nil {
			return states, err
		}

		states = append(states, &s)
	}

	if err := cur.Err(); err != nil {
		return states, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(states) == 0 {
		return states, mongo.ErrNoDocuments
	}

	return states, nil
}
