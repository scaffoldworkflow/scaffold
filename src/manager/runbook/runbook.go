package runbook

import (
	"fmt"
	"scaffold/manager/constants"
	"scaffold/manager/mongodb"
	"time"

	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"
)

type Runbook struct {
	ID           string  `json:"id" bson:"id" yaml:"id"`
	Name         string  `json:"name" bson:"name" yaml:"name"`
	Created      string  `json:"created" bson:"created" yaml:"created"`
	Updated      string  `json:"updated" bson:"updated" yaml:"updated"`
	Requirements string  `json:"requirements" bson:"requirements" yaml:"requirements"`
	Blocks       []Block `json:"blocks" bson:"blocks" yaml:"blocks"`
	Team         string  `json:"team" bson:"team" yaml:"team"`
	Project      string  `json:"project" bson:"project" yaml:"project"`
}

type Block struct {
	Contents string `json:"contents" bson:"contents" yaml:"contents"`
	Output   string `json:"output" bson:"output" yaml:"output"`
	Type     string `json:"type" bson:"type" yaml:"type"`
	Language string `json:"language" bson:"language" yaml:"language"`
}

func (r *Runbook) Load() error {
	currentTime := time.Now().UTC()

	// Create empty structures if not provided to avoid nils
	if r.Blocks == nil {
		r.Blocks = make([]Block, 0)
	}

	r.Created = currentTime.Format("2006-01-02T15:04:05Z")
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")
	r.ID = uuid.NewString()

	return nil
}

// // Create runbook object
// func CreateRunbook(r *Runbook) error {
// 	currentTime := time.Now().UTC()
// 	r.Created = currentTime.Format("2006-01-02T15:04:05Z")
// 	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

// 	rr, err := GetRunbookByID(r.ID)
// 	if err != nil {
// 		return fmt.Errorf("error getting runbooks: %s", err.Error())
// 	}
// 	if rr != nil {
// 		return fmt.Errorf("runbook already exists with ID %s", r.ID)
// 	}

// 	_, err = mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME].InsertOne(mongodb.Ctx, r)
// 	return err
// }

// // Delete a runbook by its ID
// func DeleteRunbookByID(id string) error {
// 	filter := bson.M{"id": id}

// 	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
// 	ctx := mongodb.Ctx

// 	result, err := collection.DeleteOne(ctx, filter)

// 	if err != nil {
// 		return err
// 	}

// 	if result.DeletedCount != 1 {
// 		return fmt.Errorf("no runbook found with id %s", id)
// 	}

// 	return nil

// }

// // Get all the runbooks
// func GetAllRunbooks() ([]*Runbook, error) {
// 	filter := bson.D{{}}

// 	runbooks, err := FilterRunbooks(filter)

// 	return runbooks, err
// }

// // Get a runbook by its ID
// func GetRunbookByID(id string) (*Runbook, error) {
// 	filter := bson.M{"id": id}

// 	runbooks, err := FilterRunbooks(filter)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(runbooks) == 0 {
// 		return nil, nil
// 	}

// 	if len(runbooks) > 1 {
// 		return nil, fmt.Errorf("multiple runbooks found with ID %s", id)
// 	}

// 	return runbooks[0], nil
// }

// // Update a datastore for a particular workflow
// func UpdateRunbookByID(r *Runbook) error {
// 	filter := bson.M{"id": r.ID}

// 	currentTime := time.Now().UTC()
// 	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

// 	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
// 	ctx := mongodb.Ctx

// 	// opts := options.Replace().SetUpsert(true)

// 	result, err := collection.ReplaceOne(ctx, filter, r)

// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// if result.ModifiedCount != 1 {
// 	// 	logger.Debugf("", "Could not update runbook %s", r.ID)
// 	// 	return CreateRunbook(r)
// 	// }

// 	if err != nil {
// 		logger.Errorf("", "Encountered error in update: %s", err.Error())
// 		// if result.ModifiedCount == 0 {
// 		// 	logger.Debugf("", "Monitor %s does not exist, creating...", m.ID)
// 		// 	return CreateMonitor(m)
// 		// }
// 		return err
// 	}

// 	if result.ModifiedCount == 0 {
// 		logger.Debugf("", "Runbook %s does not exist, creating...", r.ID)
// 		return CreateRunbook(r)
// 	}

// 	return nil
// }

func (r *Runbook) Create() error {
	currentTime := time.Now().UTC()
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	logger.Debugf("", "Creating runbook at %s/%s/%s", r.Team, r.Project, r.Name)

	rs, err := GetRunbooks(bson.M{"name": r.Name, "project": r.Project, "team": r.Team})
	if err != nil {
		logger.Errorf("", "Error getting runbooks: %s", err.Error())
		return err
	}
	if rs != nil {
		logger.Errorf("Runbook already exists at %s/%s/%s", r.Team, r.Project, r.Name)
		return fmt.Errorf("runbook already exists at %s/%s/%s", r.Team, r.Project, r.Name)
	}

	if err := r.Load(); err != nil {
		logger.Errorf("", "Unable to load runbook %s/%s/%s: %s", r.Team, r.Project, r.Name, err)
		return err
	}

	_, err = mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME].InsertOne(mongodb.Ctx, r)
	return err
}

func (r *Runbook) Update() error {
	filter := bson.M{"name": r.Name, "project": r.Project, "team": r.Team}

	rbs, err := GetRunbooks(filter)
	if err != nil {
		logger.Errorf("", "Could not get runbooks with filter %v: %s", filter, err)
		return err
	}

	if len(rbs) == 0 {
		logger.Debug("", "Runbook does not exist, creating...")
		if err := r.Create(); err != nil {
			logger.Errorf("", "Could not create runbook: %s", err)
			return err
		}
		return nil
	}

	r.Created = rbs[0].Created

	currentTime := time.Now().UTC()
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err = collection.ReplaceOne(ctx, filter, r)

	if err != nil {
		logger.Errorf("", "Could not update runbook at %s/%s/%s: %s", r.Team, r.Project, r.Name, err.Error())
		return err
	}

	return nil
}

func (r *Runbook) Delete() error {
	filter := bson.M{"name": r.Name, "project": r.Project, "team": r.Team}

	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete runbook at %s/%s/%s: %s", r.Team, r.Project, r.Name, err.Error())
		return err
	}

	return nil
}

func DeleteRunbooks(filter bson.M) error {
	collection := mongodb.Collections[constants.MONGODB_RUNBOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete runbooks with filter %v: %s", filter, err.Error())
		return err
	}

	return nil
}

func GetRunbooks(filter bson.M) ([]*Runbook, error) {
	runbooks, err := FilterRunbooks(filter)

	if err != nil {
		logger.Errorf("", "Could not get runbooks with filter %v: %s", filter, err.Error())
		return nil, err
	}

	return runbooks, nil
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
