package project

import (
	"fmt"
	"scaffold/manager/constants"
	"scaffold/manager/mongodb"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"go.mongodb.org/mongo-driver/bson"
)

type Project struct {
	Permissions  []string               `json:"permissions" bson:"permissions" yaml:"permissions"`
	Name         string                 `json:"name" bson:"name" yaml:"name"`
	Environments map[string]Environment `json:"envs" bson:"envs" yaml:"envs"`
	Created      string                 `json:"created" bson:"created" yaml:"created"`
	Updated      string                 `json:"updated" bson:"updated" yaml:"updated"`
	Team         string                 `json:"team" bson:"team" yaml:"team"`
}

func (p *Project) Load() error {
	currentTime := time.Now().UTC()

	// Check for required fields
	if p.Name == "" {
		logger.Errorf("", "Project is missing required field 'name'")
		return fmt.Errorf("project is missing required field 'name'")
	}
	if p.Team == "" {
		logger.Errorf("", "Project is missing required field 'team'")
		return fmt.Errorf("project is missing required field 'team'")
	}

	// Create empty structures if not provided to avoid nils
	if p.Permissions == nil {
		p.Permissions = make([]string, 0)
	}
	if p.Environments == nil {
		p.Environments = make(map[string]Environment)
	}

	// Set generated fields
	p.Created = currentTime.Format("2006-01-02T15:04:05Z")
	p.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	// Extract environment ids
	for name, env := range p.Environments {
		if err := env.Load(p.Team, p.Name, name); err != nil {
			logger.Errorf("", "Unable to load environment at %s/%s/%s: %s", p.Team, p.Name, name, err)
			return err
		}
		if err := env.Create(); err != nil {
			logger.Errorf("", "Unable to create environment at %s/%s/%s: %s", p.Team, p.Name, name, err)
			return err
		}
	}
	return nil
}

func (p *Project) Create() error {
	currentTime := time.Now().UTC()
	p.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	logger.Debugf("", "Creating project for %s on team %s", p.Name, p.Team)

	ps, err := GetProjects(bson.M{"name": p.Name, "team": p.Team})
	if err != nil {
		logger.Errorf("", "Error getting projects: %s", err)
		return err
	}
	if ps != nil {
		logger.Errorf("Project already exists at %s/%s", p.Team, p.Name)
		return fmt.Errorf("project already exists at %s/%s", p.Team, p.Name)
	}

	if err := DeleteEnvironments(bson.M{"project": p.Name, "team": p.Team}); err != nil {
		logger.Warnf("", "Error nuking environments for %s/%s", p.Team, p.Name)
	}
	if err := DeleteServices(bson.M{"project": p.Name, "team": p.Team}); err != nil {
		logger.Warnf("", "Error nuking services for %s/%s", p.Team, p.Name)
	}

	if err := p.Load(); err != nil {
		logger.Errorf("", "Unable to load project %s/%s: %s", p.Team, p.Name, err)
		return err
	}

	_, err = mongodb.Collections[constants.MONGODB_PROJECT_COLLECTION_NAME].InsertOne(mongodb.Ctx, p)
	return err
}

func (p *Project) Update() error {
	for name, env := range p.Environments {
		env.Project = p.Name
		env.Team = p.Team
		if err := (&env).Update(); err != nil {
			logger.Errorf("", "Could not update environment at %s/%s/%s: %s", p.Team, p.Name, name, err.Error())
			return err
		}
	}

	filter := bson.M{"name": p.Name, "team": p.Team}

	projs, err := GetProjects(filter)
	if err != nil {
		logger.Errorf("", "Could not get projects with filter %v: %s", filter, err)
		return err
	}

	if len(projs) == 0 {
		logger.Debug("", "Project does not exist, creating...")
		if err := p.Create(); err != nil {
			logger.Errorf("", "Could not create project: %s", err)
			return err
		}
		return nil
	}

	p.Created = projs[0].Created

	currentTime := time.Now().UTC()
	p.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_PROJECT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err = collection.ReplaceOne(ctx, filter, p)

	if err != nil {
		logger.Errorf("", "Could not update project at %s/%s: %s", p.Team, p.Name, err.Error())
		return err
	}

	return nil
}

func (p *Project) Delete(cascade bool) error {
	if cascade {
		cascadeFilter := bson.M{"name": p.Name, "team": p.Team}
		DeleteServices(cascadeFilter)
		DeleteEnvironments(cascadeFilter)
	}

	filter := bson.M{"name": p.Name, "team": p.Team}

	collection := mongodb.Collections[constants.MONGODB_PROJECT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete project at %s/%s: %s", p.Team, p.Name, err.Error())
		return err
	}

	return nil
}

func DeleteProjects(filter bson.M) error {
	collection := mongodb.Collections[constants.MONGODB_PROJECT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not delete projects with filter %v: %s", filter, err.Error())
		return err
	}

	return nil
}

func GetProjects(filter bson.M) ([]*Project, error) {

	projects, err := FilterProjects(filter)

	if err != nil {
		logger.Errorf("", "Could not get projects with filter %v: %s", filter, err.Error())
		return nil, err
	}

	return projects, nil
}

func FilterProjects(filter interface{}) ([]*Project, error) {
	// A slice of tasks for storing the decoded documents
	var projects []*Project

	collection := mongodb.Collections[constants.MONGODB_PROJECT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return projects, err
	}

	for cur.Next(ctx) {
		var p Project
		err := cur.Decode(&p)
		if err != nil {
			return projects, err
		}

		projects = append(projects, &p)
	}

	if err := cur.Err(); err != nil {
		return projects, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return projects, nil
}
