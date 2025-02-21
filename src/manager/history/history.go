package history

import (
	"fmt"
	"scaffold/manager/constants"
	"scaffold/manager/project"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"scaffold/manager/mongodb"

	logger "github.com/jfcarter2358/go-logger"
)

type History struct {
	RunID       string                 `json:"run_id" bson:"run_id" yaml:"run_id"`
	States      []State                `json:"states" bson:"states" yaml:"states"`
	Project     string                 `json:"project" bson:"project" yaml:"project"`
	Environment string                 `json:"environment" bson:"environment" yaml:"environment"`
	Service     string                 `json:"service" bson:"service" yaml:"service"`
	Created     string                 `json:"created" bson:"created" yaml:"created"`
	Updated     string                 `json:"updated" bson:"updated" yaml:"updated"`
	Context     map[string]interface{} `json:"context" bson:"context" yaml:"context"`
	Team        string                 `json:"team" bson:"team" yaml:"team"`
	Workflow    project.Workflow       `json:"workflow" bson:"workflow" yaml:"workflow"`
	ReleaseID   string                 `json:"release_id" bson:"release_id" yaml:"release_id"`
}

var locks = map[string]*sync.RWMutex{}

// func PruneHistories() {
// 	now := time.Now()
// 	histories, err := GetAllHistories()
// 	if err != nil {
// 		logger.Errorf("", "Cannot get histories: %s", err.Error())
// 		return
// 	}
// 	for _, h := range histories {
// 		t, err := time.Parse(h.Updated, "2006-01-02 15:04:05")
// 		if err != nil {
// 			logger.Errorf("", "Cannot parse history updated timestamp of %s: %s", h.Updated, err.Error())
// 		}
// 		diff := now.Sub(t)
// 		if diff.Hours() > float64(config.Config.RunPruneDuration) {
// 			if err := DeleteHistoryByRunID(h.RunID); err != nil {
// 				logger.Errorf("", "Cannot delete history with run ID %s: %s", h.RunID, err.Error())
// 			}
// 		}
// 	}
// }

func AddStateToHistory(runID string, s State) error {
	if _, ok := locks[runID]; !ok {
		locks[runID] = &sync.RWMutex{}
	}
	locks[runID].Lock()
	defer locks[runID].Unlock()
	h, err := GetHistoryByRunID(runID)
	if err != nil {
		return err
	}
	h.States = append(h.States, s)
	filter := bson.M{"run_id": runID}
	update := bson.M{"$set": bson.M{"states": h.States}}

	collection := mongodb.Collections[constants.MONGODB_HISTORY_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err = collection.UpdateOne(ctx, filter, update)

	if err != nil {
		logger.Errorf("", "Could not update history %s: %s", runID, err.Error())
		return err
	}

	// if result.UpsertedCount != 1 {
	// 	logger.Errorf("", "No history updated with run ID %s: got upserted count %d", runID, result.UpsertedCount)
	// 	return fmt.Errorf("no history updated with run ID %s", runID)
	// }

	return nil
}

func UpdateContext(runID string, newContext map[string]interface{}) error {
	if _, ok := locks[runID]; !ok {
		locks[runID] = &sync.RWMutex{}
	}
	locks[runID].Lock()
	defer locks[runID].Unlock()
	h, err := GetHistoryByRunID(runID)
	if err != nil {
		logger.Errorf("", "Cannot get history %s to update last state: %s", runID, err.Error())
		return err
	}

	for key, val := range newContext {
		h.Context[key] = val
	}

	filter := bson.M{"run_id": runID}
	update := bson.M{"$set": bson.M{"context": h.Context}}

	collection := mongodb.Collections[constants.MONGODB_HISTORY_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err = collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	// if result.UpsertedCount != 1 {
	// 	logger.Errorf("", "No history updated with run ID %s: got upserted count %d", runID, result.UpsertedCount)
	// 	return fmt.Errorf("no history updated with run ID %s", runID)
	// }

	return nil
}

func UpdateState(runID string, idx int, s State) error {
	if _, ok := locks[runID]; !ok {
		locks[runID] = &sync.RWMutex{}
	}
	locks[runID].Lock()
	defer locks[runID].Unlock()
	h, err := GetHistoryByRunID(runID)
	if err != nil {
		logger.Errorf("", "Cannot get history %s to update last state: %s", runID, err.Error())
		return err
	}
	if h.States == nil {
		h.States = []State{
			{
				Step:     "",
				Status:   "",
				Started:  "",
				Finished: "",
				Output:   "",
				Idx:      idx,
			},
		}
	}
	for idx >= len(h.States) {
		h.States = append(h.States, State{
			Step:     "",
			Status:   "",
			Started:  "",
			Finished: "",
			Output:   "",
			Idx:      idx,
		})
	}
	h.States[idx] = s

	filter := bson.M{"run_id": runID}
	update := bson.M{"$set": bson.M{"states": h.States}}

	collection := mongodb.Collections[constants.MONGODB_HISTORY_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err = collection.UpdateOne(ctx, filter, update)

	if err != nil {
		logger.Errorf("", "Could not update history %s", runID)
		return err
	}

	// if result.UpsertedCount != 1 {
	// 	logger.Debugf("", "%d", result.UpsertedCount)
	// 	logger.Errorf("", "No history updated with run ID %s", runID)
	// 	return fmt.Errorf("no history updated with run ID %s", runID)
	// }

	return nil
}

func GetHistories(filter bson.M) ([]*History, error) {

	histories, err := FilterHistories(filter)

	if err != nil {
		logger.Errorf("", "Could not get histories with filter %v: %s", filter, err.Error())
		return nil, err
	}

	return histories, nil
}

func CreateHistory(h *History) error {
	currentTime := time.Now().UTC()
	h.Created = currentTime.Format("2006-01-02T15:04:05Z")
	h.Updated = currentTime.Format("2006-01-02T15:04:05Z")
	h.Context = make(map[string]interface{})

	logger.Debugf("", "Creating history for %s", h.RunID)

	hh, err := GetHistoryByRunID(h.RunID)
	if err != nil {
		return fmt.Errorf("error getting histories: %s", err.Error())
	}
	if hh != nil {
		logger.Errorf("history already exists with run ID %s", hh.RunID)
	}

	_, err = mongodb.Collections[constants.MONGODB_HISTORY_COLLECTION_NAME].InsertOne(mongodb.Ctx, h)
	return err
}

func DeleteHistoryByRunID(runID string) error {
	filter := bson.M{"run_id": runID}

	collection := mongodb.Collections[constants.MONGODB_HISTORY_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no history found with run ID %s", runID)
	}

	return nil

}

func GetAllHistories() ([]*History, error) {
	filter := bson.D{{}}

	datastores, err := FilterHistories(filter)

	return datastores, err
}

func GetHistoryByRunID(runID string) (*History, error) {
	filter := bson.M{"run_id": runID}

	histories, err := FilterHistories(filter)

	if err != nil {
		return nil, err
	}

	if len(histories) == 0 {
		return nil, nil
	}

	if len(histories) > 1 {
		return nil, fmt.Errorf("multiple history found with run ID %s", runID)
	}

	return histories[0], nil
}

func (h *History) GetHistoryStateByStepName(stepName string) *State {
	for i := range h.States {
		s := h.States[len(h.States)-1-i]
		if s.Step == stepName {
			return &s
		}
	}

	return nil
}

func UpdateHistoryByRunID(runID string, h *History) error {
	if _, ok := locks[runID]; !ok {
		locks[runID] = &sync.RWMutex{}
	}
	locks[runID].Lock()
	defer locks[runID].Unlock()
	filter := bson.M{"run_id": h.RunID}

	currentTime := time.Now().UTC()
	h.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.ReplaceOne(ctx, filter, h)

	if err != nil {
		logger.Errorf("", "Encountered error in update: %s", err.Error())
		return err
	}

	if result.ModifiedCount == 0 {
		logger.Debugf("", "History %s does not exist, creating...", h.RunID)
		return CreateHistory(h)
	}

	return nil
}

func FilterHistories(filter interface{}) ([]*History, error) {
	// A slice of tasks for storing the decoded documents
	var histories []*History

	collection := mongodb.Collections[constants.MONGODB_HISTORY_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return histories, err
	}

	for cur.Next(ctx) {
		var h History
		err := cur.Decode(&h)
		if err != nil {
			return histories, err
		}

		histories = append(histories, &h)
	}

	if err := cur.Err(); err != nil {
		return histories, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return histories, nil
}
