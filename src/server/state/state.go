package state

import (
	"fmt"
	"scaffold/server/constants"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type State struct {
	Task     string                   `json:"task" bson:"task"`
	Cascade  string                   `json:"cascade" bson:"cascade"`
	Status   string                   `json:"status" bson:"status"`
	Started  string                   `json:"started" bson:"started"`
	Finished string                   `json:"finished" bson:"finished"`
	Output   string                   `json:"output" bson:"output"`
	Display  []map[string]interface{} `json:"display" bson:"display"`
	Number   int                      `json:"number" bson:"number"`
}

func CreateState(s *State) error {
	if _, err := GetStateByNames(s.Cascade, s.Task); err == nil {
		return fmt.Errorf("state already exists with names %s, %s", s.Cascade, s.Task)
	}

	_, err := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME].InsertOne(mongodb.Ctx, s)
	return err
}

func DeleteStateByNames(cascade, task string) error {
	filter := bson.M{"cascade": cascade, "task": task}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no state found with names %s, %s", cascade, task)
	}

	return nil

}

func DeleteStatesByCascade(cascade string) error {
	filter := bson.M{"cascade": cascade}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no states found with cascade %s", cascade)
	}

	return nil

}

func GetAllStates() ([]*State, error) {
	filter := bson.D{{}}

	states, err := FilterStates(filter)

	return states, err
}

func GetStateByNames(cascade, task string) (*State, error) {
	filter := bson.M{"cascade": cascade, "task": task}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no state found with names %s, %s", cascade, task)
	}

	if len(states) > 1 {
		return nil, fmt.Errorf("multiple states found with names %s, %s", cascade, task)
	}

	return states[0], nil
}

func GetStatesByCascade(cascade string) ([]*State, error) {
	filter := bson.M{"cascade": cascade}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	return states, nil
}

func UpdateStateByNames(cascade, task string, s *State) error {
	filter := bson.M{"cascade": cascade, "task": task}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, s, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no state found with names %s, %s", cascade, task)
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
