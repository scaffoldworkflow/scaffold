package runbook

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/mongodb"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Runbook struct {
	ID           string   `json:"id" bson:"id" yaml:"id"`
	Name         string   `json:"name" bson:"name" yaml:"name"`
	Category     string   `json:"category" bson:"category" yaml:"category"`
	Groups       []string `json:"groups" bson:"groups" yaml:"groups"`
	Created      string   `json:"created" bson:"created" yaml:"created"`
	Updated      string   `json:"updated" bson:"updated" yaml:"updated"`
	Requirements string   `json:"requirements" bson:"requirements" yaml:"requirements"`
	Blocks       []Block  `json:"blocks" bson:"blocks" yaml:"blocks"`
}

type Block struct {
	Contents string `json:"contents" bson:"contents" yaml:"contents"`
	Output   string `json:"output" bson:"output" yaml:"output"`
	Type     string `json:"type" bson:"type" yaml:"type"`
}

// Create runbook object
func CreateRunbook(r *Runbook) error {
	currentTime := time.Now().UTC()
	r.Created = currentTime.Format("2006-01-02T15:04:05Z")
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	rr, err := GetRunbookByID(r.ID)
	if err != nil {
		return fmt.Errorf("error getting runbooks: %s", err.Error())
	}
	if rr != nil {
		return fmt.Errorf("runbook already exists with ID %s", r.ID)
	}

	_, err = mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME].InsertOne(mongodb.Ctx, r)
	return err
}

// Delete a runbook by its ID
func DeleteRunbookByID(id string) error {
	filter := bson.M{"id": id}

	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no runbook found with id %s", id)
	}

	return nil

}

// Get all the runbooks
func GetAllRunbooks() ([]*Runbook, error) {
	filter := bson.D{{}}

	runbooks, err := FilterRunbooks(filter)

	return runbooks, err
}

// Get a runbook by its ID
func GetRunbookByID(id string) (*Runbook, error) {
	filter := bson.M{"id": id}

	runbooks, err := FilterRunbooks(filter)

	if err != nil {
		return nil, err
	}

	if len(runbooks) == 0 {
		return nil, nil
	}

	if len(runbooks) > 1 {
		return nil, fmt.Errorf("multiple runbooks found with ID %s", id)
	}

	return runbooks[0], nil
}

// Update a datastore for a particular workflow
func UpdateRunbookByID(r *Runbook) error {
	filter := bson.M{"id": r.ID}

	currentTime := time.Now().UTC()
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, r, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		logger.Debugf("", "Could not update runbook %s", r.ID)
		return CreateRunbook(r)
	}

	return nil
}

// Filter the MongoDB response from a query
func FilterRunbooks(filter interface{}) ([]*Runbook, error) {
	// A slice of tasks for storing the decoded documents
	var runbooks []*Runbook

	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return runbooks, err
	}

	for cur.Next(ctx) {
		var d Runbook
		err := cur.Decode(&d)
		if err != nil {
			return runbooks, err
		}

		runbooks = append(runbooks, &d)
	}

	if err := cur.Err(); err != nil {
		return runbooks, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return runbooks, nil
}
