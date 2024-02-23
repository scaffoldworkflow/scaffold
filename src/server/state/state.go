package state

import (
	"fmt"
	"scaffold/server/constants"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type State struct {
	Task     string                   `json:"task" bson:"task" yaml:"task"`
	Cascade  string                   `json:"cascade" bson:"cascade" yaml:"cascade"`
	Status   string                   `json:"status" bson:"status" yaml:"status"`
	Started  string                   `json:"started" bson:"started" yaml:"started"`
	Finished string                   `json:"finished" bson:"finished" yaml:"finished"`
	Output   string                   `json:"output" bson:"output" yaml:"output"`
	Display  []map[string]interface{} `json:"display" bson:"display" yaml:"display"`
	Worker   string                   `json:"worker" bson:"worker" yaml:"worker"`
	Number   int                      `json:"number" bson:"number" yaml:"number"`
	Disabled bool                     `json:"disabled" bson:"disabled" yaml:"disabled"`
	Killed   bool                     `json:"killed" bson:"killed" yaml:"killed"`
}

func CreateState(s *State) error {
	ss, err := GetStateByNames(s.Cascade, s.Task)
	if err != nil {
		return fmt.Errorf("error getting states: %s", err.Error())
	}
	if ss != nil {
		return fmt.Errorf("state already exists with names %s, %s", s.Cascade, s.Task)
	}

	_, err = mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME].InsertOne(mongodb.Ctx, s)
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

func CopyStatesByNames(cascade, task1, task2 string) error {
	filter1 := bson.M{"cascade": cascade, "task": task1}
	states, err := FilterStates(filter1)
	if err != nil {
		return err
	}
	if len(states) == 0 {
		return fmt.Errorf("no state found with names %s, %s", cascade, task1)
	}
	if len(states) > 1 {
		return fmt.Errorf("multiple states found with names %s, %s", cascade, task1)
	}

	filter2 := bson.M{"cascade": cascade, "task": task2}
	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx
	opts := options.Replace().SetUpsert(true)

	_, err = collection.ReplaceOne(ctx, filter2, states[0], opts)

	return err
}

func GetStateByNames(cascade, task string) (*State, error) {
	filter := bson.M{"cascade": cascade, "task": task}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, nil
	}

	if len(states) > 1 {
		return nil, fmt.Errorf("multiple states found with names %s, %s", cascade, task)
	}

	return states[0], nil
}

func GetStateByNamesNumber(cascade, task string, number int) (*State, error) {
	filter := bson.M{"cascade": cascade, "task": task, "number": number}

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

func UpdateStateKilledByNames(cascade, task string, killed bool) error {
	filter := bson.M{"cascade": cascade, "task": task}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	update := bson.D{
		{"$set", bson.D{{"killed", killed}}},
	}

	result, err := collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		// return fmt.Errorf("no state found with names %s, %s", cascade, task)
		logger.Tracef("no state found with names %s, %s", cascade, task)
	}

	return nil
}

func UpdateStateRunByNames(cascade, task string, s State) error {
	filter := bson.M{"cascade": cascade, "task": task}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	update := bson.D{
		{"$set", bson.D{{"status", s.Status}}},
		{"$set", bson.D{{"started", s.Started}}},
		{"$set", bson.D{{"finished", s.Finished}}},
		{"$set", bson.D{{"output", s.Output}}},
		{"$set", bson.D{{"display", s.Display}}},
	}

	result, err := collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		// return fmt.Errorf("no state found with names %s, %s", cascade, task)
		logger.Tracef("no state found with names %s, %s", cascade, task)
	}

	return nil
}

func ClearStateByNames(cascade, task string, runNumber int) error {
	s := &State{
		Task:     task,
		Cascade:  cascade,
		Status:   constants.STATE_STATUS_NOT_STARTED,
		Started:  "",
		Finished: "",
		Output:   "",
		Number:   runNumber,
		Worker:   "",
		Display:  make([]map[string]interface{}, 0),
	}

	if err := UpdateStateByNames(cascade, task, s); err != nil {
		logger.Errorf("", "Cannot update state %s.%s: %s", cascade, task, err.Error())
		return err
	}

	return nil
}

func GetStatesByWorker(worker string) ([]*State, error) {
	filter := bson.M{"worker": worker}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	return states, nil
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

	return states, nil
}
