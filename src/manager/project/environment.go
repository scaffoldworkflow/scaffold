package project

import (
	"fmt"
	"scaffold/manager/constants"
	"scaffold/manager/mongodb"
	"time"

	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/bson"
)

type Environment struct {
	Name        string             `json:"name" bson:"name" yaml:"name"`
	Project     string             `json:"project" bson:"project" yaml:"project"`
	Team        string             `json:"team" bson:"team" yaml:"team"`
	Permissions []string           `json:"permissions" bson:"permissions" yaml:"permissions"`
	Services    map[string]Service `json:"services" bson:"services" yaml:"services"`
	Created     string             `json:"created" bson:"created" yaml:"created"`
	Updated     string             `json:"updated" bson:"updated" yaml:"updated"`
	// Promote     Step               `json:"promote" bson:"promote" yaml:"promote"`
}

func (e *Environment) Load(tName, pName, name string) error {
	currentTime := time.Now().UTC()

	// Create empty structures if not provided to avoid nils
	if e.Permissions == nil {
		e.Permissions = make([]string, 0)
	}
	if e.Services == nil {
		e.Services = make(map[string]Service)
	}

	// Set passed fields
	e.Team = tName
	e.Project = pName
	e.Name = name

	e.Created = currentTime.Format("2006-01-02T15:04:05Z")
	e.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	// Extract environment ids
	for name, serv := range e.Services {
		if err := serv.Load(tName, pName, e.Name, name); err != nil {
			logger.Errorf("", "Unable to load service at %s/%s/%s/%s: %s", e.Team, e.Project, e.Name, name, err.Error())
			return err
		}
		if err := serv.Create(); err != nil {
			logger.Errorf("", "Unable to create environment at %s/%s/%s/%s: %s", e.Team, e.Name, e.Name, name, err)
			return err
		}
	}
	return nil
}

func (e *Environment) Create() error {
	currentTime := time.Now().UTC()
	e.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	logger.Debugf("", "Creating environment at %s/%s/%s", e.Team, e.Project, e.Name)

	es, err := GetEnvironments(bson.M{"name": e.Name, "project": e.Project, "team": e.Team})
	if err != nil {
		logger.Errorf("", "Error getting environments: %s", err.Error())
		return err
	}
	if es != nil {
		logger.Errorf("Environment already exists at %s/%s/%s", e.Team, e.Project, e.Name)
		return fmt.Errorf("environment already exists at %s/%s/%s", e.Team, e.Project, e.Name)
	}

	_, err = mongodb.Collections[constants.MONGODB_ENVIRONMENT_COLLECTION_NAME].InsertOne(mongodb.Ctx, e)
	return err
}

func (e *Environment) Update() error {
	for name, serv := range e.Services {
		serv.Team = e.Team
		serv.Project = e.Project
		serv.Environment = e.Name
		if err := (&serv).Update(); err != nil {
			logger.Errorf("", "Could not update service at %s/%s/%s/%s: %s", e.Team, e.Project, e.Name, name, err.Error())
			return err
		}
	}

	filter := bson.M{"name": e.Name, "project": e.Project, "team": e.Team}

	envs, err := GetEnvironments(filter)
	if err != nil {
		logger.Errorf("", "Could not get environments with filter %v: %s", filter, err)
		return err
	}

	if len(envs) == 0 {
		logger.Debug("", "Environment does not exist, creating...")
		if err := e.Create(); err != nil {
			logger.Errorf("", "Could not create environment: %s", err)
			return err
		}
		return nil
	}

	e.Created = envs[0].Created

	currentTime := time.Now().UTC()
	e.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_ENVIRONMENT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err = collection.ReplaceOne(ctx, filter, e)

	if err != nil {
		logger.Errorf("", "Could not update environment at %s/%s/%s: %s", e.Team, e.Project, e.Name, err.Error())
		return err
	}

	return nil
}

func (e *Environment) Delete(cascade bool) error {
	if cascade {
		cascadeFilter := bson.M{"project": e.Project, "team": e.Team, "environment": e.Name}
		DeleteServices(cascadeFilter)
	}

	filter := bson.M{"name": e.Name, "project": e.Project, "team": e.Team}

	collection := mongodb.Collections[constants.MONGODB_ENVIRONMENT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete environment at %s/%s/%s: %s", e.Team, e.Project, e.Name, err.Error())
		return err
	}

	return nil
}

func DeleteEnvironments(filter bson.M) error {
	collection := mongodb.Collections[constants.MONGODB_ENVIRONMENT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete environments with filter %v: %s", filter, err.Error())
		return err
	}

	return nil
}

func GetEnvironments(filter bson.M) ([]*Environment, error) {
	environments, err := FilterEnvironments(filter)

	if err != nil {
		logger.Errorf("", "Could not get environments with filter %v: %s", filter, err.Error())
		return nil, err
	}

	return environments, nil
}

func FilterEnvironments(filter interface{}) ([]*Environment, error) {
	// A slice of tasks for storing the decoded documents
	var environments []*Environment

	collection := mongodb.Collections[constants.MONGODB_ENVIRONMENT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return environments, err
	}

	for cur.Next(ctx) {
		var e Environment
		err := cur.Decode(&e)
		if err != nil {
			return environments, err
		}

		environments = append(environments, &e)
	}

	if err := cur.Err(); err != nil {
		return environments, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return environments, nil
}
