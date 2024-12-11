package workflow

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/input"
	"scaffold/server/task"
	"sync"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"

	"scaffold/server/mongodb"
)

type Workflow struct {
	Version string            `json:"version" bson:"version" yaml:"version"`
	Name    string            `json:"name" bson:"name" yaml:"name"`
	Inputs  []input.Input     `json:"inputs" bson:"inputs" yaml:"inputs"`
	Tasks   []task.Task       `json:"tasks" bson:"tasks" yaml:"tasks"`
	Created string            `json:"created" bson:"created" yaml:"created"`
	Updated string            `json:"updated" bson:"updated" yaml:"updated"`
	Groups  []string          `json:"groups" bson:"groups" yaml:"groups"`
	Links   map[string]string `json:"links" bson:"links" yaml:"links"`
}

type cacheObj struct {
	Workflows map[string]Workflow
	Lock      *sync.RWMutex
}

var cache = cacheObj{
	Workflows: make(map[string]Workflow),
	Lock:      &sync.RWMutex{},
}

func SetCache(ws []*Workflow) {
	cache.Lock.Lock()
	defer cache.Lock.Unlock()

	for _, w := range ws {
		cache.Workflows[w.Name] = *w
	}
}

func AddCache(w Workflow) {
	cache.Lock.Lock()
	defer cache.Lock.Unlock()

	cache.Workflows[w.Name] = w
}

func DeleteCache(name string) {
	cache.Lock.Lock()
	defer cache.Lock.Unlock()

	delete(cache.Workflows, name)
}

func GetCacheAll() map[string]Workflow {
	cache.Lock.Lock()
	defer cache.Lock.Unlock()

	return cache.Workflows
}

func GetCacheSingle(name string) Workflow {
	cache.Lock.Lock()
	defer cache.Lock.Unlock()

	if w, ok := cache.Workflows[name]; ok {
		return w
	}
	return Workflow{}
}

func CreateWorkflow(w *Workflow) error {
	currentTime := time.Now().UTC()
	w.Created = currentTime.Format("2006-01-02T15:04:05Z")
	w.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	ww, err := GetWorkflowByName(w.Name)
	if err != nil {
		return fmt.Errorf("error getting workflows: %s", err.Error())
	}
	if ww != nil {
		return fmt.Errorf("workflow already exists with name %s", w.Name)
	}

	_, err = mongodb.Collections[constants.MONGODB_WORKFLOW_COLLECTION_NAME].InsertOne(mongodb.Ctx, w)

	if err != nil {
		return err
	}

	for _, t := range w.Tasks {
		t.Workflow = w.Name
		t.RunNumber = 0
		if err := task.CreateTask(&t); err != nil {
			return err
		}
	}

	for _, i := range w.Inputs {
		i.Workflow = w.Name
		if err := input.CreateInput(&i); err != nil {
			return err
		}
	}

	d := &datastore.DataStore{
		Name:    w.Name,
		Env:     make(map[string]string),
		Files:   make([]string, 0),
		Created: w.Created,
		Updated: w.Updated,
	}

	for _, val := range w.Inputs {
		d.Env[val.Name] = val.Default
	}

	err = datastore.CreateDataStore(d)

	if err == nil {
		AddCache(*w)
	}
	return err
}

func DeleteWorkflowByName(name string) error {
	filter := bson.M{"name": name}

	collection := mongodb.Collections[constants.MONGODB_WORKFLOW_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	DeleteCache(name)

	if result.DeletedCount != 1 {
		return fmt.Errorf("unable to delete workflow %s, doesn't exist", name)
	}

	if err := task.DeleteTasksByWorkflow(name); err != nil {
		return err
	}

	if err := input.DeleteInputsByWorkflow(name); err != nil {
		return err
	}

	err = datastore.DeleteDataStoreByWorkflow(name)

	return err

}

func GetAllWorkflows() ([]*Workflow, error) {
	filter := bson.M{}

	workflows, err := FilterWorkflows(filter)

	return workflows, err
}

func GetWorkflowByName(name string) (*Workflow, error) {
	filter := bson.M{"name": name}

	workflows, err := FilterWorkflows(filter)

	if err != nil {
		logger.Errorf("", "filter workflows returned error %s", err.Error())
		return nil, err
	}

	if len(workflows) == 0 {
		return nil, nil
	}

	if len(workflows) > 1 {
		return nil, fmt.Errorf("multiple workflows found with name %s", name)
	}

	return workflows[0], nil
}

func UpdateWorkflowByName(name string, w *Workflow) error {
	// filter := bson.M{"name": name}

	// currentTime := time.Now().UTC()
	// w.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	// collection := mongodb.Collections[constants.MONGODB_WORKFLOW_COLLECTION_NAME]
	// ctx := mongodb.Ctx

	// opts := options.Replace().SetUpsert(true)

	// states, err := state.GetStatesByWorkflow(name)
	// if err != nil {
	// 	return err
	// }
	// tasks, err := task.GetTasksByWorkflow(name)
	// if err != nil {
	// 	return nil
	// }

	// result, err := collection.ReplaceOne(ctx, filter, w, opts)

	// if err != nil {
	// 	return err
	// }

	// if result.ModifiedCount != 1 {
	// 	AddCache(*w)
	// 	return CreateWorkflow(w)
	// 	// return fmt.Errorf("no workflow found with name %s", name)
	// }

	// AddCache(*w)

	// taskNames := make([]string, len(states))

	// for idx, t := range tasks {
	// 	taskNames[idx] = t.Name
	// }

	// newNames := make([]string, len(w.Tasks))

	// for idx, t := range w.Tasks {
	// 	if !utils.Contains(taskNames, t.Name) {
	// 		t.Workflow = name
	// 		logger.Debugf("", "Creating task %s with workflow %s", t.Name, t.Workflow)
	// 		if err := task.CreateTask(&t); err != nil {
	// 			return err
	// 		}
	// 		continue
	// 	}
	// 	logger.Debugf("", "Updating task %s with workflow %s", t.Name, name)
	// 	if err := task.UpdateTaskByNames(name, t.Name, &t); err != nil {
	// 		return err
	// 	}
	// 	newNames[idx] = t.Name
	// }

	// logger.Debugf("", "Old tasks: %v", taskNames)
	// logger.Debugf("", "New tasks: %v", newNames)

	// for _, t := range tasks {
	// 	if !utils.Contains(newNames, t.Name) {
	// 		logger.Debugf("", "Removing task %s", t.Name)
	// 		if err := task.DeleteTaskByNames(name, t.Name); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	// return err

	// filter := bson.M{"name": name}
	// currentTime := time.Now().UTC()
	// w.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	// logger.Debugf("", "Updating workflow %v", *w)

	// collection := mongodb.Collections[constants.MONGODB_WORKFLOW_COLLECTION_NAME]
	// ctx := mongodb.Ctx

	// opts := options.Replace().SetUpsert(true)

	// result, err := collection.ReplaceOne(ctx, filter, w, opts)

	// if err != nil {
	// 	return err
	// }

	// if result.ModifiedCount != 1 {
	// 	return CreateWorkflow(w)
	// }

	// logger.Debugf("", "Update result: %v", result)

	// return nil

	if err := DeleteWorkflowByName(name); err != nil {
		logger.Warnf("", "Got error doing workflow update delete: %s", err.Error())
	}
	return CreateWorkflow(w)
}

func FilterWorkflows(filter interface{}) ([]*Workflow, error) {
	// A slice of tasks for storing the decoded documents
	var workflows []*Workflow

	collection := mongodb.Collections[constants.MONGODB_WORKFLOW_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return workflows, err
	}

	for cur.Next(ctx) {
		var c Workflow
		err := cur.Decode(&c)
		if err != nil {
			return workflows, err
		}

		workflows = append(workflows, &c)
	}

	if err := cur.Err(); err != nil {
		return workflows, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return workflows, nil
}
