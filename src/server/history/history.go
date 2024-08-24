package history

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/state"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"

	logger "github.com/jfcarter2358/go-logger"
)

type History struct {
	RunID    string        `json:"run_id" bson:"run_id" yaml:"run_id"`
	States   []state.State `json:"states" bson:"states" yaml:"states"`
	Workflow string        `json:"workflow" bson:"workflow" yaml:"workflow"`
	Created  string        `json:"created" bson:"created" yaml:"created"`
	Updated  string        `json:"updated" bson:"updated" yaml:"updated"`
}

func AddStateToHistory(runID string, s state.State) error {
	h, err := GetHistoryByRunID(runID)
	if err != nil {
		return err
	}
	if h == nil {
		h = &History{
			Workflow: s.Workflow,
			States:   make([]state.State, 0),
			RunID:    runID,
		}
	}
	for idx, ss := range h.States {
		if ss.Workflow == s.Workflow && ss.Task == s.Task {
			h.States[idx] = s
			if err := UpdateHistoryByRunID(runID, h); err != nil {
				return err
			}
			return nil
		}
	}
	h.States = append(h.States, s)
	err = UpdateHistoryByRunID(runID, h)
	return err
}

func CreateHistory(h *History) error {
	currentTime := time.Now().UTC()
	h.Created = currentTime.Format("2006-01-02T15:04:05Z")
	h.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	logger.Errorf("", "Creating history for %s", h.RunID)

	hh, err := GetHistoryByRunID(h.RunID)
	if err != nil {
		return fmt.Errorf("error getting histories: %s", err.Error())
	}
	if hh != nil {
		logger.Errorf("history already exists with run ID %s", hh.RunID)
		return UpdateHistoryByRunID(h.RunID, h)
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

func UpdateHistoryByRunID(runID string, h *History) error {
	filter := bson.M{"run_id": runID}

	currentTime := time.Now().UTC()
	h.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_HISTORY_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	_, err := collection.ReplaceOne(ctx, filter, h, opts)

	if err != nil {
		return err
	}

	logger.Errorf("", "Inserted history with runID %s", runID)

	hh, err := GetAllHistories()
	logger.Errorf("", "Current histories: %v", hh)

	// if result.ModifiedCount == 0 {
	// 	return CreateHistory(h)
	// 	// return fmt.Errorf("no datastore found with name %s", name)
	// }

	return err
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
