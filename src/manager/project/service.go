package project

import (
	"fmt"
	"scaffold/manager/constants"
	"scaffold/manager/mongodb"
	"time"

	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/bson"
)

type Service struct {
	Name        string   `json:"name" bson:"name" yaml:"name"`
	Team        string   `json:"team" bson:"team" yaml:"team"`
	Project     string   `json:"project" bson:"project" yaml:"project"`
	Environment string   `json:"environment" bson:"environment" yaml:"environment"`
	Permissions []string `json:"permissions" bson:"permissions" yaml:"permissions"`
	GitHubRepo  string   `json:"github_repo" bson:"github_repo" yaml:"github_repo"`
	Workflow    Workflow `json:"workflow" bson:"workflow" yaml:"workflow"`
	WorkflowArn string   `json:"workflow_arn" yaml:"workflow_arn" yaml:"workflow_arn"`
	Created     string   `json:"created" bson:"created" yaml:"created"`
	Updated     string   `json:"updated" bson:"updated" yaml:"updated"`
}

func (s *Service) Load(tName, pName, eName, name string) error {
	currentTime := time.Now().UTC()

	// Create empty structures if not provided to avoid nils
	if s.Permissions == nil {
		s.Permissions = make([]string, 0)
	}

	// Set passed fields
	s.Team = tName
	s.Project = pName
	s.Environment = eName
	s.Name = name

	s.Created = currentTime.Format("2006-01-02T15:04:05Z")
	s.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	if err := (&s.Workflow).Load(); err != nil {
		logger.Errorf("", "Unable to load workflow at %s/%s/%s/%s: %s", s.Team, s.Project, s.Name, name, err.Error())
		return err
	}
	return nil
}

func (s *Service) Create() error {
	currentTime := time.Now().UTC()
	s.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	logger.Debugf("", "Creating service at %s/%s/%s/%s", s.Team, s.Project, s.Environment, s.Name)

	ss, err := GetEnvironments(bson.M{"name": s.Name, "project": s.Project, "team": s.Team, "environment": s.Environment})
	if err != nil {
		logger.Errorf("Error getting services: %s", err.Error())
		return err
	}
	if ss != nil {
		logger.Errorf("Service already exists at %s/%s/%s/%s", s.Team, s.Project, s.Environment, s.Name)
		return fmt.Errorf("service already exists at %s/%s/%s/%s", s.Team, s.Project, s.Environment, s.Name)
	}

	_, err = mongodb.Collections[constants.MONGODB_SERVICE_COLLECTION_NAME].InsertOne(mongodb.Ctx, s)
	return err
}

func (s *Service) Update() error {

	if err := (&s.Workflow).Update(); err != nil {
		logger.Errorf("", "Could not update workflow at %s/%s/%s/%s: %s", s.Team, s.Project, s.Environment, s.Name, err.Error())
		return err
	}

	filter := bson.M{"name": s.Name, "project": s.Project, "team": s.Team, "environment": s.Environment}

	currentTime := time.Now().UTC()
	s.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_SERVICE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.ReplaceOne(ctx, filter, s)

	if err != nil {
		logger.Errorf("", "Could not update service at %s/%s/%s/%s: %s", s.Team, s.Project, s.Environment, s.Name, err.Error())
		return err
	}

	if result.ModifiedCount == 0 {
		logger.Debug("", "Service does not exist, creating...")
		if err := s.Create(); err != nil {
			logger.Errorf("", "Could not create service")
			return err
		}
	}

	return nil
}

func (s *Service) Delete() error {

	filter := bson.M{"name": s.Name, "project": s.Project, "team": s.Team, "environment": s.Environment}

	collection := mongodb.Collections[constants.MONGODB_SERVICE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete environment at %s/%s/%s/%s: %s", s.Team, s.Project, s.Environment, s.Name, err.Error())
		return err
	}

	return nil
}

func DeleteServices(filter bson.M) error {
	collection := mongodb.Collections[constants.MONGODB_SERVICE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete services with filter %v: %s", filter, err.Error())
		return err
	}

	return nil
}

func GetServices(filter bson.M) ([]*Service, error) {
	services, err := FilterServices(filter)

	if err != nil {
		logger.Errorf("", "Could not get services with filter %v: %s", filter, err.Error())
		return nil, err
	}

	return services, nil
}

func FilterServices(filter interface{}) ([]*Service, error) {
	// A slice of tasks for storing the decoded documents
	var services []*Service

	collection := mongodb.Collections[constants.MONGODB_SERVICE_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return services, err
	}

	for cur.Next(ctx) {
		var s Service
		err := cur.Decode(&s)
		if err != nil {
			return services, err
		}

		services = append(services, &s)
	}

	if err := cur.Err(); err != nil {
		return services, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return services, nil
}
