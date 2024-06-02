package task

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/state"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type TaskDependsOn struct {
	Success []string `json:"success" bson:"success" yaml:"success"`
	Error   []string `json:"error" bson:"error" yaml:"error"`
	Always  []string `json:"always" bson:"always" yaml:"always"`
}

type TaskLoadStore struct {
	Env            []string `json:"env" bson:"env" yaml:"env"`
	File           []string `json:"file" bson:"file" yaml:"file"`
	EnvPassthrough []string `json:"env_passthrough" bson:"env_passthrough" yaml:"env_passthrough"`
	Mounts         []string `json:"mounts" bson:"mounts" yaml:"mounts"`
}

type TaskCheck struct {
	Cron      string            `json:"cron" bson:"cron" yaml:"cron"`
	Image     string            `json:"image" bson:"image" yaml:"image"`
	Run       string            `json:"run" bson:"run" yaml:"run"`
	Store     TaskLoadStore     `json:"store" bson:"store" yaml:"store"`
	Load      TaskLoadStore     `json:"load" bson:"load" yaml:"load"`
	Env       map[string]string `json:"env" bson:"env" yaml:"env"`
	Inputs    map[string]string `json:"inputs" bson:"inputs" yaml:"inputs"`
	Updated   string            `json:"updated" bson:"updated" yaml:"updated"`
	RunNumber int               `json:"run_number" bson:"run_number" yaml:"run_number"`
}

type Task struct {
	Name        string            `json:"name" bson:"name" yaml:"name"`
	Kind        string            `json:"kind" bson:"kind" yaml:"kind"`
	Cron        string            `json:"cron" bson:"cron" yaml:"cron"`
	Cascade     string            `json:"cascade" bson:"cascade" yaml:"cascade"`
	Verb        string            `json:"verb" bson:"verb" yaml:"verb"`
	DependsOn   TaskDependsOn     `json:"depends_on" bson:"depends_on" yaml:"depends_on"`
	Image       string            `json:"image" bson:"image" yaml:"image"`
	Run         string            `json:"run" bson:"run" yaml:"run"`
	Store       TaskLoadStore     `json:"store" bson:"store" yaml:"store"`
	Load        TaskLoadStore     `json:"load" bson:"load" yaml:"load"`
	Env         map[string]string `json:"env" bson:"env" yaml:"env"`
	Inputs      map[string]string `json:"inputs" bson:"inputs" yaml:"inputs"`
	Updated     string            `json:"updated" bson:"updated" yaml:"updated"`
	RunNumber   int               `json:"run_number" bson:"run_number" yaml:"run_number"`
	ShouldRM    bool              `json:"should_rm" bson:"should_rm" yaml:"should_rm"`
	AutoExecute bool              `json:"auto_execute" bson:"auto_execute" yaml:"auto_execute"`
	Disabled    bool              `json:"disabled" bson:"disabled" yaml:"disabled"`
	// Check                 TaskCheck         `json:"check" bson:"check" yaml:"check"`
	ContainerLoginCommand string `json:"container_login_command" bson:"container_login_command" yaml:"container_login_command"`
}

func CreateTask(t *Task) error {
	tt, err := GetTaskByNames(t.Cascade, t.Name)
	if err != nil {
		return fmt.Errorf("error getting tasks: %s", err.Error())
	}
	if tt != nil {
		return fmt.Errorf("task already exists with names %s, %s", t.Cascade, t.Name)
	}

	if t.Kind == "" {
		t.Kind = constants.TASK_KIND_LOCAL
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
		History:  make([]string, 0),
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

	_, err = mongodb.Collections[constants.MONGODB_TASK_COLLECTION_NAME].InsertOne(mongodb.Ctx, t)
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
		return nil, nil
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

	// tt, err := GetTaskByNames(cascade, task)
	// if err != nil {
	// 	return fmt.Errorf("task %s does not exist in cascade %s", task, cascade)
	// }
	// if t.Cron == "" {
	// 	t.Cron = tt.Cron
	// }
	// if t.Verb == "" {
	// 	t.Verb = tt.Verb
	// }
	// if t.Image == "" {
	// 	t.Image = tt.Image
	// }
	// if t.Run == "" {
	// 	t.Run = tt.Run
	// }
	// if t.Env == nil {
	// 	t.Env = tt.Env
	// }
	// if t.Inputs == nil {
	// 	t.Inputs = tt.Inputs
	// }
	// if t.RunNumber == 0 {
	// 	t.RunNumber = tt.RunNumber
	// }
	// if t.ContainerLoginCommand == "" {
	// 	t.ContainerLoginCommand = tt.ContainerLoginCommand
	// }

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
