package cascade

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/state"
	"scaffold/server/utils"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"scaffold/server/mongodb"
)

type CascadeLoadStore struct {
	Env  []string `json:"env" bson:"env"`
	File []string `json:"file" bson:"file"`
}

type CascadeInput struct {
	Description string `json:"description" bson:"description"`
	Default     string `json:"default" bson:"default"`
	Type        string `json:"type" bson:"type"`
}

type CascadeTask struct {
	Name      string            `json:"name" bson:"name"`
	Verb      string            `json:"verb" bson:"verb"`
	DependsOn []string          `json:"depends_on" bson:"depends_on"`
	Image     string            `json:"image" bson:"image"`
	Run       string            `json:"run" bson:"run"`
	Store     CascadeLoadStore  `json:"store" bson:"store"`
	Load      CascadeLoadStore  `json:"load" bson:"load"`
	Outputs   map[string]string `json:"outputs" bson:"outputs"`
	Inputs    map[string]string `json:"inputs" bson:"inputs"`
	Group     string            `json:"group" bson:"group"`
}

type Cascade struct {
	Version string         `json:"version" bson:"version"`
	Name    string         `json:"name" bson:"name"`
	Inputs  []CascadeInput `json:"inputs" bson:"inputs"`
	Tasks   []CascadeTask  `json:"tasks" bson:"tasks"`
	Created string         `json:"created" bson:"created"`
	Updated string         `json:"updated" bson:"updated"`
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

	ts := make([]state.StateTask, len(c.Tasks))
	for idx, task := range c.Tasks {
		t := state.StateTask{
			Name:     task.Name,
			Status:   constants.STATE_STATUS_NOT_STARTED,
			Started:  "",
			Finished: "",
			Error:    "",
			Output:   "",
		}
		ts[idx] = t
	}

	s := &state.State{
		Name:    c.Name,
		Created: c.Created,
		Updated: c.Updated,
		Tasks:   ts,
	}

	err = state.CreateState(s)
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

	err = state.DeleteStateByName(name)

	return err

}

func GetAllCascades() ([]*Cascade, error) {
	filter := bson.D{{}}

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

	result, err := collection.ReplaceOne(ctx, filter, c)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no cascade found with name %s", name)
	}

	s, err := state.GetStateByName(name)

	if err != nil {
		return err
	}

	taskNames := make([]string, len(c.Tasks))
	for idx, task := range c.Tasks {
		taskNames[idx] = task.Name
	}

	toRemove := make([]int, 0)
	for idx, stateTask := range s.Tasks {
		if !utils.Contains(taskNames, stateTask.Name) {
			toRemove = append(toRemove, idx)
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(toRemove)))

	for _, idx := range toRemove {
		s.Tasks = append(s.Tasks[:idx], s.Tasks[idx+1:]...)
	}

	stateTaskNames := make([]string, len(s.Tasks))
	for idx, stateTask := range s.Tasks {
		stateTaskNames[idx] = stateTask.Name
	}

	for idx, task := range c.Tasks {
		if !utils.Contains(stateTaskNames, task.Name) {
			temp := s.Tasks[idx:]
			s.Tasks = append(s.Tasks[:idx], state.StateTask{
				Name:     task.Name,
				Status:   constants.STATE_STATUS_NOT_STARTED,
				Started:  "",
				Finished: "",
				Error:    "",
				Output:   "",
			})
			s.Tasks = append(s.Tasks, temp...)
		}
	}

	err = state.UpdateStateByName(name, s)

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
