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
	Task           string                   `json:"task" bson:"task" yaml:"task"`
	Workflow       string                   `json:"workflow" bson:"workflow" yaml:"workflow"`
	Status         string                   `json:"status" bson:"status" yaml:"status"`
	Started        string                   `json:"started" bson:"started" yaml:"started"`
	Finished       string                   `json:"finished" bson:"finished" yaml:"finished"`
	Output         string                   `json:"output" bson:"output" yaml:"output"`
	OutputChecksum string                   `json:"output_checksum" bson:"output_checksum" yaml:"output_checksum"`
	Display        []map[string]interface{} `json:"display" bson:"display" yaml:"display"`
	Worker         string                   `json:"worker" bson:"worker" yaml:"worker"`
	Number         int                      `json:"number" bson:"number" yaml:"number"`
	Disabled       bool                     `json:"disabled" bson:"disabled" yaml:"disabled"`
	Killed         bool                     `json:"killed" bson:"killed" yaml:"killed"`
	PID            int                      `json:"pid" bson:"pid" yaml:"pid"`
	History        []string                 `json:"history" bson:"history" yaml:"history"`
	Context        map[string]string        `json:"context" bson:"context" yaml:"context"`
}

func CreateState(s *State) error {
	ss, err := GetStateByNames(s.Workflow, s.Task)
	if err != nil {
		return fmt.Errorf("error getting states: %s", err.Error())
	}
	if ss != nil {
		return nil
	}

	_, err = mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME].InsertOne(mongodb.Ctx, s)
	return err
}

func DeleteStateByNames(workflow, task string) error {
	filter := bson.M{"workflow": workflow, "task": task}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no state found with names %s, %s", workflow, task)
	}

	return nil

}

func DeleteStatesByWorkflow(workflow string) error {
	filter := bson.M{"workflow": workflow}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no states found with workflow %s", workflow)
	}

	return nil

}

func GetAllStates() ([]*State, error) {
	filter := bson.D{{}}

	states, err := FilterStates(filter)

	return states, err
}

func CopyStatesByNames(workflow, task1, task2 string) error {
	filter1 := bson.M{"workflow": workflow, "task": task1}
	states, err := FilterStates(filter1)
	if err != nil {
		return err
	}
	if len(states) == 0 {
		return fmt.Errorf("no state found with names %s, %s", workflow, task1)
	}
	if len(states) > 1 {
		return fmt.Errorf("multiple states found with names %s, %s", workflow, task1)
	}

	filter2 := bson.M{"workflow": workflow, "task": task2}
	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx
	opts := options.Replace().SetUpsert(true)

	_, err = collection.ReplaceOne(ctx, filter2, states[0], opts)

	return err
}

func GetStateByNames(workflow, task string) (*State, error) {
	filter := bson.M{"workflow": workflow, "task": task}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, nil
	}

	if len(states) > 1 {
		return nil, fmt.Errorf("multiple states found with names %s, %s", workflow, task)
	}

	return states[0], nil
}

func GetStateByNamesAndRunID(workflow, task, runID string) (*State, error) {
	filter := bson.M{"workflow": workflow, "task": task}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, nil
	}

	if len(states) > 1 {
		return nil, fmt.Errorf("multiple states found with names %s, %s", workflow, task)
	}

	return states[0], nil
}

func GetStateByNamesNumber(workflow, task string, number int) (*State, error) {
	filter := bson.M{"workflow": workflow, "task": task, "number": number}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no state found with names %s, %s", workflow, task)
	}

	if len(states) > 1 {
		return nil, fmt.Errorf("multiple states found with names %s, %s", workflow, task)
	}

	return states[0], nil
}

func GetStatesByWorkflow(workflow string) ([]*State, error) {
	filter := bson.M{"workflow": workflow}

	states, err := FilterStates(filter)

	if err != nil {
		return nil, err
	}

	return states, nil
}

func UpdateStateByNames(workflow, task string, s *State) error {
	filter := bson.M{"workflow": workflow, "task": task}

	collection := mongodb.Collections[constants.MONGODB_STATE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, s, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return CreateState(s)
	}

	return nil
}

func UpdateStateKilledByNames(workflow, task string, killed bool) error {
	filter := bson.M{"workflow": workflow, "task": task}

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
		logger.Tracef("", "no state found with names %s, %s", workflow, task)
	}

	return nil
}

func UpdateStateRunByNames(workflow, task string, s State) error {

	ss, err := GetStateByNames(workflow, task)
	if err != nil {
		return err
	}

	// checksum := md5.Sum([]byte(s.Output))
	// s.OutputChecksum = string(checksum[:])

	ss.Status = s.Status
	ss.Started = s.Started
	ss.Finished = s.Finished
	ss.Output = s.Output
	// ss.OutputChecksum = s.OutputChecksum
	ss.Display = s.Display
	ss.PID = s.PID

	logger.Tracef("", "Updating state by names")

	return UpdateStateByNames(workflow, task, ss)
}

func ClearStateByNames(workflow, task string, runNumber int) error {
	s := &State{
		Task:           task,
		Workflow:       workflow,
		Status:         constants.STATE_STATUS_NOT_STARTED,
		Started:        "",
		Finished:       "",
		Output:         "",
		OutputChecksum: "",
		Number:         runNumber,
		Worker:         "",
		Display:        make([]map[string]interface{}, 0),
	}

	if err := UpdateStateByNames(workflow, task, s); err != nil {
		logger.Errorf("", "Cannot update state %s.%s: %s", workflow, task, err.Error())
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
