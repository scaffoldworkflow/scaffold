package task

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/state"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"scaffold/server/mongodb"
)

type TaskLoadStore struct {
	Env  []string `json:"env" bson:"env"`
	File []string `json:"file" bson:"file"`
}

type Task struct {
	Name      string            `json:"name" bson:"name"`
	Cascade   string            `json:"cascade" bson:"cascade"`
	Verb      string            `json:"verb" bson:"verb"`
	DependsOn []string          `json:"depends_on" bson:"depends_on"`
	Image     string            `json:"image" bson:"image"`
	Run       string            `json:"run" bson:"run"`
	Store     TaskLoadStore     `json:"store" bson:"store"`
	Load      TaskLoadStore     `json:"load" bson:"load"`
	Outputs   map[string]string `json:"outputs" bson:"outputs"`
	Inputs    map[string]string `json:"inputs" bson:"inputs"`
	Group     string            `json:"group" bson:"group"`
}

func CreateTask(t *Task) error {
	if _, err := GetTaskByNames(t.Cascade, t.Name); err == nil {
		return fmt.Errorf("task already exists with names %s, %s", t.Cascade, t.Name)
	}

	s := state.State{
		Task:     t.Name,
		Cascade:  t.Cascade,
		Status:   constants.STATE_STATUS_NOT_STARTED,
		Started:  "",
		Finished: "",
		Output:   "",
	}
	if err := state.CreateState(&s); err != nil {
		return err
	}

	_, err := mongodb.Collections[constants.MONGODB_TASK_COLLECTION_NAME].InsertOne(mongodb.Ctx, t)
	return err
}

func DeleteTaskByNames(cascade, task string) error {
	filter := bson.M{"cascade": cascade, "name": task}

	collection := mongodb.Collections[constants.MONGODB_TASK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no task found with names %s, %s", cascade, task)
	}

	if err := state.DeleteStateByNames(cascade, task); err != nil {
		return err
	}

	return nil

}

func DeleteTasksByCascade(cascade string) error {
	filter := bson.M{"cascade": cascade}

	collection := mongodb.Collections[constants.MONGODB_TASK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no tasks found with cascade %s", cascade)
	}

	if err := state.DeleteStatesByCascade(cascade); err != nil {
		return err
	}

	return nil

}

func GetAllTasks() ([]*Task, error) {
	filter := bson.D{{}}

	tasks, err := FilterTasks(filter)

	return tasks, err
}

func GetTaskByNames(cascade, task string) (*Task, error) {
	filter := bson.M{"cascade": cascade, "name": task}

	tasks, err := FilterTasks(filter)

	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no task found with names %s, %s", cascade, task)
	}

	if len(tasks) > 1 {
		return nil, fmt.Errorf("multiple tasks found with names %s, %s", cascade, task)
	}

	return tasks[0], nil
}

func GetTasksByCascade(cascade string) ([]*Task, error) {
	filter := bson.M{"cascade": cascade}

	tasks, err := FilterTasks(filter)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func UpdateTaskByNames(cascade, task string, t *Task) error {
	filter := bson.M{"cascade": cascade, "name": task}

	collection := mongodb.Collections[constants.MONGODB_TASK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.ReplaceOne(ctx, filter, t)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no task found with names %s, %s", cascade, task)
	}

	return nil
}

func FilterTasks(filter interface{}) ([]*Task, error) {
	// A slice of tasks for storing the decoded documents
	var tasks []*Task

	collection := mongodb.Collections[constants.MONGODB_TASK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return tasks, err
	}

	for cur.Next(ctx) {
		var s Task
		err := cur.Decode(&s)
		if err != nil {
			return tasks, err
		}

		tasks = append(tasks, &s)
	}

	if err := cur.Err(); err != nil {
		return tasks, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(tasks) == 0 {
		return tasks, mongo.ErrNoDocuments
	}

	return tasks, nil
}
