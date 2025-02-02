package release

import (
	"fmt"
	"scaffold/client/logger"
	"scaffold/manager/constants"
	"scaffold/manager/mongodb"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type Asset struct {
	Contents string
	URL      string
	Name     string
}

type Release struct {
	ID       string   `json:"id" bson:"id"`
	Runs     []string `json:"runs" bson:"runs"`
	Status   string   `json:"status" bson:"status"`
	Assets   []Asset  `json:"assets" bson:"assets"`
	Project  string   `json:"project" bson:"project"`
	Team     string   `json:"team" bson:"team"`
	Created  string   `json:"created" bson:"created"`
	Updated  string   `json:"updated" bson:"updated"`
	Services []string `json:"services" bson:"services"`
}

func (r *Release) Load() error {
	currentTime := time.Now().UTC()
	r.ID = uuid.NewString()

	// Check for required fields
	if r.Project == "" {
		logger.Errorf("", "Release is missing required field 'project'")
		return fmt.Errorf("release is missing required field 'project'")
	}
	if r.Team == "" {
		logger.Errorf("", "Release is missing required field 'team'")
		return fmt.Errorf("release is missing required field 'team'")
	}

	// Create empty structures if not provided to avoid nils
	if r.Services == nil {
		r.Services = make([]string, 0)
	}
	if r.Environments == nil {
		r.Environments = make([]string, 0)
	}

	// Set generated fields
	r.Created = currentTime.Format("2006-01-02T15:04:05Z")
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	return nil
}

func (r *Release) Create() error {
	currentTime := time.Now().UTC()
	r.Updated = currentTime.Format("2007-01-02T15:04:05Z")

	logger.Debugf("", "Creating release %s at %s/%s", r.ID, r.Team, r.Project)

	_, err := mongodb.Collections[constants.MONGODB_RELEASE_COLLECTION_NAME].InsertOne(mongodb.Ctx, r)
	return err
}

func (r *Release) Update() error {
	filter := bson.M{"id": r.ID}

	currentTime := time.Now().UTC()
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_RELEASE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.ReplaceOne(ctx, filter, r)

	if err != nil {
		logger.Errorf("", "Could not update release at %s/%s/%s: %s", r.Team, r.Project, r.ID, err.Error())
		return err
	}

	if result.ModifiedCount == 0 {
		logger.Debug("", "Release does not exist, creating...")
		if err := r.Create(); err != nil {
			logger.Errorf("", "Could not create project")
			return err
		}
	}

	return nil
}

func (r *Release) Delete() error {
	filter := bson.M{"id": r.ID}

	collection := mongodb.Collections[constants.MONGODB_RELEASE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete release at %s/%s/%s: %s", r.Team, r.Project, r.ID, err.Error())
		return err
	}

	return nil
}

func DeleteReleases(filter bson.M) error {
	collection := mongodb.Collections[constants.MONGODB_RELEASE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete releases with filter %v: %s", filter, err.Error())
		return err
	}

	return nil
}

func GetReleases(filter bson.M) ([]*Release, error) {

	releases, err := FilterReleases(filter)

	if err != nil {
		logger.Errorf("", "Could not get releases with filter %v: %s", filter, err.Error())
		return nil, err
	}

	return releases, nil
}

func FilterReleases(filter interface{}) ([]*Release, error) {
	// A slice of tasks for storing the decoded documents
	var releases []*Release

	collection := mongodb.Collections[constants.MONGODB_RELEASE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return releases, err
	}

	for cur.Next(ctx) {
		var r Release
		err := cur.Decode(&r)
		if err != nil {
			return releases, err
		}

		releases = append(releases, &r)
	}

	if err := cur.Err(); err != nil {
		return releases, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return releases, nil
}
