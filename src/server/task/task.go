package task

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/state"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type TaskDependsOn struct {
	Success []string `json:"success" bson:"success"`
	Error   []string `json:"error" bson:"error"`
	Always  []string `json:"always" bson:"always"`
}

type TaskLoadStore struct {
	Env            []string `json:"env" bson:"env"`
	File           []string `json:"file" bson:"file"`
	EnvPassthrough []string `json:"env_passthrough" bson:"env_passthrough"`
	Mounts         []string `json:"mounts" bson:"mounts"`
}

type TaskCheck struct {
	Cron      string            `json:"cron" bson:"cron"`
	Image     string            `json:"image" bson:"image"`
	Run       string            `json:"run" bson:"run"`
	Store     TaskLoadStore     `json:"store" bson:"store"`
	Load      TaskLoadStore     `json:"load" bson:"load"`
	Env       map[string]string `json:"env" bson:"env"`
	Inputs    map[string]string `json:"inputs" bson:"inputs"`
	Updated   string            `json:"updated" bson:"updated"`
	RunNumber int               `json:"run_number" bson:"run_number"`
}

type Task struct {
	Name                  string            `json:"name" bson:"name"`
	Cron                  string            `json:"cron" bson:"cron"`
	Cascade               string            `json:"cascade" bson:"cascade"`
	Verb                  string            `json:"verb" bson:"verb"`
	DependsOn             TaskDependsOn     `json:"depends_on" bson:"depends_on"`
	Image                 string            `json:"image" bson:"image"`
	Run                   string            `json:"run" bson:"run"`
	Store                 TaskLoadStore     `json:"store" bson:"store"`
	Load                  TaskLoadStore     `json:"load" bson:"load"`
	Env                   map[string]string `json:"env" bson:"env"`
	Inputs                map[string]string `json:"inputs" bson:"inputs"`
	Updated               string            `json:"updated" bson:"updated"`
	RunNumber             int               `json:"run_number" bson:"run_number"`
	ShouldRM              bool              `json:"should_rm" bson:"should_rm"`
	AutoExecute           bool              `json:"auto_execute" bson:"auto_execute"`
	Disabled              bool              `json:"disabled" bson:"disabled"`
	Check                 TaskCheck         `json:"check" bson:"check"`
	ContainerLoginCommand string            `json:"container_login_command" bson:"container_login_command"`
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
		Number:   t.RunNumber,
		Display:  make([]map[string]interface{}, 0),
		Killed:   false,
	}
	if err := state.CreateState(&s); err != nil {
		return err
	}

	// sc := state.State{
	// 	Task:     fmt.Sprintf("SCAFFOLD_CHECK-%s", t.Name),
	// 	Cascade:  t.Cascade,
	// 	Status:   constants.STATE_STATUS_NOT_STARTED,
	// 	Started:  "",
	// 	Finished: "",
	// 	Output:   "",
	// 	Number:   t.RunNumber,
	// 	Display:  make([]map[string]interface{}, 0),
	// }
	// if err := state.CreateState(&sc); err != nil {
	// 	return err
	// }

	// sp := state.State{
	// 	Task:     fmt.Sprintf("SCAFFOLD_PREVIOUS-%s", t.Name),
	// 	Cascade:  t.Cascade,
	// 	Status:   constants.STATE_STATUS_NOT_STARTED,
	// 	Started:  "",
	// 	Finished: "",
	// 	Output:   "",
	// 	Number:   0,
	// 	Display:  make([]map[string]interface{}, 0),
	// }
	// if err := state.CreateState(&sp); err != nil {
	// 	return err
	// }

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
	currentTime := time.Now().UTC()
	t.Updated = currentTime.Format("2006-01-02T15:04:05Z")
	t.Cascade = cascade

	logger.Debugf("", "Updating task %v", *t)

	collection := mongodb.Collections[constants.MONGODB_TASK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, t, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no task found with names %s, %s", cascade, task)
	}

	logger.Debugf("", "Update result: %v", result)

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

func VerifyDepends(cn, tn string) (bool, error) {
	t, err := GetTaskByNames(cn, tn)
	if err != nil {
		return false, err
	}
	for _, n := range t.DependsOn.Success {
		s, err := state.GetStateByNames(cn, n)
		if err != nil {
			return false, err
		}
		if s.Status != constants.STATE_STATUS_SUCCESS {
			return false, nil
		}
	}
	for _, n := range t.DependsOn.Error {
		s, err := state.GetStateByNames(cn, n)
		if err != nil {
			return false, err
		}
		if s.Status != constants.STATE_STATUS_ERROR {
			return false, nil
		}
	}
	return true, nil
}
