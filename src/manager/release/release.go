package release

import (
	"encoding/json"
	"fmt"
	"scaffold/client/logger"
	"scaffold/manager/constants"
	"scaffold/manager/history"
	"scaffold/manager/mongodb"
	"scaffold/manager/monitor"
	"scaffold/manager/project"
	"scaffold/manager/run"
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
	ID               string                     `json:"id" bson:"id"`
	Runs             []string                   `json:"runs" bson:"runs"`
	Histories        map[string]history.History `json:"histories"`
	Status           string                     `json:"status" bson:"status"`
	Assets           []Asset                    `json:"assets" bson:"assets"`
	Project          string                     `json:"project" bson:"project"`
	Team             string                     `json:"team" bson:"team"`
	Created          string                     `json:"created" bson:"created"`
	Updated          string                     `json:"updated" bson:"updated"`
	PromoteRuns      []string                   `json:"promote_runs" bson:"promote_runs"`
	PromoteHistories map[string]history.History `json:"promote_histories" bson:"promote_histories"`
	Services         []string                   `json:"services" bson:"services"`
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
	r.Assets = make([]Asset, 0)
	r.Runs = make([]string, 0)
	r.Histories = make(map[string]history.History)
	r.PromoteRuns = make([]string, 0)
	r.PromoteHistories = make(map[string]history.History)

	// Set generated fields
	r.Created = currentTime.Format("2006-01-02T15:04:05Z")
	r.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	return nil
}

func (r *Release) Hydrate() error {
	return nil
}

/*
OBJECT FUNCTIONALITY
*/

type PromoteMessage struct {
	Context            map[string]interface{} `json:"context"`
	RunB64             string                 `json:"run_b64"`
	Language           string                 `json:"language"`
	PythonRequirements string                 `json:"python_requirements"`
	Project            string
	Team               string
	Environment        string `json:"environment"`
}

func (r *Release) Promote(pm PromoteMessage) (string, error) {
	runID := uuid.NewString()
	h := &history.History{
		RunID:       runID,
		States:      make([]history.State, 0),
		Project:     pm.Project,
		Team:        pm.Team,
		Environment: pm.Environment,
	}
	if err := history.CreateHistory(h); err != nil {
		logger.Errorf("", "Cannot create history: %s", err.Error())
		return "", err
	}

	contextBytes, err := json.Marshal(pm.Context)
	if err != nil {
		logger.Errorf("", "Unable to marshal context JSON: %s", err.Error())
		return "", err
	}

	if err := run.StartPromoteRun(runID, string(contextBytes), pm.RunB64, pm.Language); err != nil {
		logger.Errorf("", "Unable to start run: %s", runID)
		return "", err
	}

	go monitor.PromoteRun(runID, "promote", 0)

	return runID, nil
}

// TODO: Better handle errors here, should probabl remove them from the slice so we don't keep trying to process them over and over
func PromoteAutoTrigger() {
	for {
		for len(monitor.PromoteFinished) > 0 {
			monitor.PromoteFinishedLock.Lock()
			s := monitor.PromoteFinished[0]
			logger.Debugf("", "Doing promote auto trigger for step %s in run %s with status %s", s.Step, s.RunID, s.Status)
			h, err := history.GetHistoryByRunID(s.RunID)
			if err != nil {
				logger.Errorf("", "Could not get history corresponding to run %s", s.RunID)
				continue
			}

			ps, err := project.GetProjects(bson.M{"team": h.Team, "name": h.Project})
			if err != nil {
				logger.Errorf("", "Could not get project %s/%s corresponding to run %s", h.Team, h.Project, s.RunID)
				continue
			}

			if len(ps) == 0 {
				logger.Errorf("", "Could not get project at %s/%s for run %s", h.Team, h.Project, s.RunID)
				continue
			}

			p := ps[0]

			rs, err := GetReleases(bson.M{"id": h.ReleaseID})
			if err != nil {
				logger.Errorf("Could not get release %s: %s", h.ReleaseID, err.Error())
				continue
			}

			if len(rs) == 0 {
				logger.Errorf("", "Could not get release %s", h.ReleaseID)
				continue
			}

			r := rs[0]

			logger.Infof("", "Triggering new run for environment %s", h.Environment)
			if _, ok := p.Environments[h.Environment]; !ok {
				logger.Warnf("", "No environment %s exists in project %s", h.Environment, h.Project)
				continue
			}
			for _, svc := range r.Services {
				if _, ok := p.Environments[h.Environment].Services[svc]; !ok {
					logger.Warnf("", "No service %s exists in project %s/%s's %s environment", svc, h.Team, h.Project, h.Environment)
					continue
				}

				w := p.Environments[h.Environment].Services[svc].Workflow
				sName := ""
				for k, _ := range w.Steps {
					sName = k
					break
				}
				contextBytes, _ := json.Marshal(h.Context)
				if err := run.StartRun(s.RunID, w, 0, string(contextBytes), sName, w.Steps[sName].Language); err != nil {
					logger.Errorf("", "Unable to start run %s idx %d: %s", s.RunID, 0, err.Error())
					continue
				}
				logger.Tracef("", "Run %s idx %d is starting", s.RunID, 0)
				go monitor.Run(s.RunID, sName, len(h.States))
				r.Runs = append(r.Runs, s.RunID)
				if err := r.Update(); err != nil {
					logger.Errorf("", "Unable to update release %s: %s", r.ID, err)
					continue
				}
			}
			monitor.PromoteFinished = monitor.PromoteFinished[1:]
			monitor.PromoteFinishedLock.Unlock()
		}
		time.Sleep(1 * time.Second)
	}
}

/*
CRUD OPERATIONS
*/

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
