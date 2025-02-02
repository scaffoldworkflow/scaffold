package run

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"scaffold/manager/config"
	"scaffold/manager/constants"
	"scaffold/manager/history"
	"scaffold/manager/monitor"
	"scaffold/manager/project"
	"scaffold/manager/resource"
	"scaffold/manager/utils"
	"time"

	_ "embed"
	"text/template"

	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/bson"
)

type RunConfig struct {
	LogLevel           string `json:"log_level"`
	RunID              string
	WorkerImage        string `json:"worker_image"`
	Step               string `json:"step"`
	Context            string
	ImagePullPolicy    string `json:"image_pull_policy"`
	Language           string
	ResourceTypes      string
	Resources          string
	Script             string
	Inputs             string
	Outputs            string
	PythonRequirements string
}

func StartRun(runID string, w project.Workflow, runIdx int, context string, step string, language string) error {
	var r RunConfig
	if err := json.Unmarshal([]byte(config.Config.WorkerConfig), &r); err != nil {
		logger.Errorf("", "Unable to load worker config: %s", err.Error())
		return err
	}

	killChildren(runID, step)

	r.Step = step
	r.RunID = runID
	r.Context = base64.StdEncoding.EncodeToString([]byte(context))
	r.Language = language

	if len(w.ResourceTypes) > 0 {
		rts := make(map[string]resource.ResourceType)
		for key, val := range w.ResourceTypes {
			script := ""
			switch val.Source {
			case "file":
				script = base64.StdEncoding.EncodeToString([]byte(resource.FileScript))
			case "github_repo":
				script = base64.StdEncoding.EncodeToString([]byte(resource.GitHubRepoScript))
			default:
				// TODO: Implement http resource get
				logger.Warnf("", "Resource source %s is invalid", val.Source)
				continue
			}
			rts[key] = resource.ResourceType{
				Name:         key,
				Source:       val.Source,
				Script:       script,
				Language:     val.Language,
				Requirements: val.Requirements,
			}
		}
		rtBytes, err := json.Marshal(rts)
		if err != nil {
			logger.Errorf("", "Could not marshal resource types %v: %s", rts, err.Error())
			return err
		}
		r.ResourceTypes = string(rtBytes)
	} else {
		r.ResourceTypes = "{}"
	}
	if len(w.Resources) > 0 {
		rBytes, err := json.Marshal(w.Resources)
		if err != nil {
			logger.Errorf("", "Could not marshal resources %v: %s", w.Resources, err.Error())
			return err
		}
		r.Resources = string(rBytes)
	} else {
		r.Resources = "{}"
	}
	if len(w.Steps[step].Inputs) > 0 {
		iBytes, err := json.Marshal(w.Steps[step].Inputs)
		if err != nil {
			logger.Errorf("", "Could not marshal inputs %v: %s", w.Steps[step].Inputs, err.Error())
			return err
		}
		r.Inputs = string(iBytes)
	} else {
		r.Inputs = "[]"
	}
	if len(w.Steps[step].Outputs) > 0 {
		oBytes, err := json.Marshal(w.Steps[step].Outputs)
		if err != nil {
			logger.Errorf("", "Could not marshal outputs %v: %s", w.Steps[step].Outputs, err.Error())
			return err
		}
		r.Outputs = string(oBytes)
	} else {
		r.Outputs = "[]"
	}
	r.Script = base64.StdEncoding.EncodeToString([]byte(w.Steps[step].Run))
	r.LogLevel = w.Steps[step].LogLevel

	tmpl, err := template.New(fmt.Sprintf("k8s_job_template_%s", r.RunID)).Parse(w.JobTemplate)
	if err != nil {
		logger.Errorf("", "Cannot load job template: %s", err.Error())
		return err
	}
	var doc bytes.Buffer
	if err := tmpl.Execute(&doc, r); err != nil {
		logger.Errorf("", "Cannot render job template: %s", err.Error())
		return err
	}

	jobPath := fmt.Sprintf("/tmp/%s.yaml", r.RunID)
	logger.Debugf("", "Writing job file to %s", jobPath)
	if err := os.WriteFile(jobPath, doc.Bytes(), 0644); err != nil {
		logger.Errorf("", "Cannot write out job manifest: %s", err.Error())
		return err
	}

	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl apply -f %s", jobPath)).CombinedOutput()
	if err != nil {
		logger.Errorf("", "Error starting run %s: %s", r.RunID, err.Error())
		logger.Debugf("", "%s", string(out))
		return err
	}
	logger.Infof("", "Run %s successfully started", r.RunID)
	return nil
}

func KillRun(runID, step string) error {
	h, err := history.GetHistoryByRunID(runID)
	if err != nil {
		logger.Errorf("", "Could not get history %s to kill k8s job: %s", runID, err.Error())
		return err
	}
	s := h.GetHistoryStateByStepName(step)
	if s == nil {
		logger.Errorf("", "No step with name %s is part of run %s", step, runID)
		return fmt.Errorf("no step with name %s is part of run %s", step, runID)
	}
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl kill job scaffold-worker-%s-%s", runID, s.Step)).CombinedOutput()
	if err != nil {
		logger.Errorf("", "Error killing run %s", runID)
		logger.Debugf("", "%s", string(out))
		return err
	}
	logger.Infof("", "Run %s successfully killed", runID)
	return nil
}

// func StateChange(runID, status, stepName string) error {
// 	h, err := history.GetHistoryByRunID(runID)
// 	if err != nil {
// 		logger.Errorf("", "Could not get history corresponding to run %s", runID)
// 		return err
// 	}
// 	p, err := project.GetProjectByID(h.Project)
// 	if err != nil {
// 		logger.Errorf("", "Could not get project %s corresponding to run %s", h.Project, runID)
// 		return err
// 	}
// 	switch status {
// 	case constants.STATE_STATUS_SUCCESS:
// 		for _, st := range p.Environments[h.Environment].Services[h.Service].Workflow.Steps {
// 			shouldExecute := false
// 			for _, n := range st.DependsOn.Always {
// 				if n == stepName {
// 					if st.AutoExecute {
// 						shouldExecute = true
// 						continue
// 					}
// 				}
// 				s := h.GetHistoryStateByStepName(n)
// 				if s == nil {
// 					logger.Warnf("", "Could not get state for step %s in history %s", n, runID)
// 					continue
// 				}
// 				if s.Status != constants.STATE_STATUS_ERROR && s.Status != constants.STATE_STATUS_SUCCESS {
// 					continue
// 				}
// 			}
// 			for _, n := range st.DependsOn.Success {
// 				if n == st.Name {
// 					if st.AutoExecute {
// 						shouldExecute = true
// 						continue
// 					}
// 				}
// 				s := h.GetHistoryStateByStepName(n)
// 				if s == nil {
// 					logger.Warnf("", "Could not get state for step %s in history %s", n, runID)
// 					continue
// 				}
// 				if s.Status != constants.STATE_STATUS_SUCCESS {
// 					continue
// 				}
// 			}
// 			if shouldExecute {
// 				if err := StateChange(runID, constants.STATE_STATUS_NOT_STARTED, st.Name); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	case constants.STATE_STATUS_ERROR:
// 		for _, st := range p.Environments[h.Environment].Services[h.Service].Workflow.Steps {
// 			shouldExecute := false
// 			for _, n := range st.DependsOn.Always {
// 				if n == stepName {
// 					if st.AutoExecute {
// 						shouldExecute = true
// 						continue
// 					}
// 				}
// 				s := h.GetHistoryStateByStepName(n)
// 				if s == nil {
// 					logger.Warnf("", "Could not get state for step %s in history %s", n, runID)
// 					continue
// 				}
// 				if s.Status != constants.STATE_STATUS_ERROR && s.Status != constants.STATE_STATUS_SUCCESS {
// 					continue
// 				}
// 			}
// 			for _, n := range st.DependsOn.Success {
// 				if n == st.Name {
// 					if st.AutoExecute {
// 						shouldExecute = true
// 						continue
// 					}
// 				}
// 				s := h.GetHistoryStateByStepName(n)
// 				if s == nil {
// 					logger.Warnf("", "Could not get state for step %s in history %s", n, runID)
// 					continue
// 				}
// 				if s.Status != constants.STATE_STATUS_ERROR {
// 					continue
// 				}
// 			}
// 			if shouldExecute {
// 				if err := StateChange(runID, constants.STATE_STATUS_NOT_STARTED, st.Name); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	case constants.STATE_STATUS_NOT_STARTED:
// 		for _, st := range p.Environments[h.Environment].Services[h.Service].Workflow.Steps {
// 			for _, n := range st.DependsOn.Always {
// 				if n == st.Name {
// 					s, err := state.GetStateByNames(cn, t.Name)
// 					if err != nil {
// 						return err
// 					}
// 					s.Status = constants.STATE_STATUS_NOT_STARTED
// 					if err := state.UpdateStateByNames(cn, t.Name, s); err != nil {
// 						return err
// 					}
// 					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
// 						return err
// 					}
// 				}
// 			}
// 			for _, n := range t.DependsOn.Error {
// 				if n == tn {
// 					s, err := state.GetStateByNames(cn, t.Name)
// 					if err != nil {
// 						return err
// 					}
// 					s.Status = constants.STATE_STATUS_NOT_STARTED
// 					if err := state.UpdateStateByNames(cn, t.Name, s); err != nil {
// 						return err
// 					}
// 					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
// 						return err
// 					}
// 				}
// 			}
// 			for _, n := range t.DependsOn.Success {
// 				if n == tn {
// 					s, err := state.GetStateByNames(cn, t.Name)
// 					if err != nil {
// 						return err
// 					}
// 					s.Status = constants.STATE_STATUS_NOT_STARTED
// 					if err := state.UpdateStateByNames(cn, t.Name, s); err != nil {
// 						return err
// 					}
// 					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
// 						return err
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

func killChildren(runID, sName string) error {
	h, err := history.GetHistoryByRunID(runID)
	if err != nil {
		logger.Errorf("", "Could not get history corresponding to run %s", runID)
		return err
	}
	ps, err := project.GetProjects(bson.M{"team": h.Team, "name": h.Project})
	if err != nil {
		logger.Errorf("", "Could not get project at %s/%s corresponding to run %s", h.Team, h.Project, runID)
		return err
	}

	if ps == nil || len(ps) == 0 {
		logger.Errorf("", "Could not get project at %s/%s for run %s", h.Team, h.Project, runID)
		return err
	}

	p := ps[0]
	for name, step := range p.Environments[h.Environment].Services[h.Service].Workflow.Steps {
		if name == sName {
			continue
		}
		for _, parent := range step.DependsOn.Success {
			if parent == sName {
				killJob(runID, name)
				return killChildren(runID, name)
			}
		}
		for _, parent := range step.DependsOn.Error {
			if parent == sName {
				killJob(runID, name)
				return killChildren(runID, name)
			}
		}
		for _, parent := range step.DependsOn.Always {
			if parent == sName {
				killJob(runID, name)
				return killChildren(runID, name)
			}
		}
	}
	return nil
}

func killJob(runID, sName string) {
	jobName := fmt.Sprintf("scaffold-worker-%s-%s", runID, sName)
	_, err := exec.Command("/bin/sh", "-c", fmt.Sprintf(`kubectl patch job %s -p '{"spec":{"suspend":true}}'`, jobName)).CombinedOutput()
	if err != nil {
		logger.Warnf("", "Error killing run %s: %s", runID, err.Error())
	}
	logger.Infof("", "Job scaffold-worker-%s-%s successfully killed", runID, sName)
}

func checkDeps(step project.Step, h history.History) (bool, error) {
	for _, n := range step.DependsOn.Success {
		s := h.GetHistoryStateByStepName(n)
		if s.Status != constants.STATE_STATUS_SUCCESS {
			return false, nil
		}
	}
	for _, n := range step.DependsOn.Error {
		s := h.GetHistoryStateByStepName(n)
		if s.Status != constants.STATE_STATUS_ERROR {
			return false, nil
		}
	}
	for _, n := range step.DependsOn.Always {
		s := h.GetHistoryStateByStepName(n)
		if s.Status != constants.STATE_STATUS_SUCCESS && s.Status != constants.STATE_STATUS_ERROR {
			return false, nil
		}
	}
	return true, nil
}

func AutoTrigger() {
	for {
		for len(monitor.Finished) > 0 {
			monitor.FinishedLock.Lock()
			s := monitor.Finished[0]
			logger.Debugf("", "Doing auto trigger for step %s in run %s with status %s", s.Step, s.RunID, s.Status)
			h, err := history.GetHistoryByRunID(s.RunID)
			if err != nil {
				logger.Errorf("", "Could not get history corresponding to run %s", s.RunID)
				continue
			}
			ps, err := project.GetProjects(bson.M{"team": h.Team, "name": h.Project})
			if err != nil {
				logger.Errorf("", "Could not get project %s corresponding to run %s", h.Project, s.RunID)
				continue
			}

			if ps == nil || len(ps) == 0 {
				logger.Errorf("", "Could not get project at %s/%s for run %s", h.Team, h.Project, s.RunID)
				continue
			}

			p := ps[0]

			logger.Tracef("", "Got environment of %v", h.Environment)
			logger.Tracef("", "Got service of %v", h.Service)

			outB, _ := json.Marshal(p)
			logger.Tracef("", "Got project:\n%s", outB)

			toTrigger := []string{}

			logger.Tracef("", "Got auto trigger status %s", s.Status)

			switch s.Status {
			case constants.STATUS_TRIGGER_SUCCESS:
				logger.Tracef("", "Dropping into success case")
				logger.Tracef("", "Checking against steps %v", p.Environments[h.Environment].Services[h.Service].Workflow.Steps)
				for stName, st := range p.Environments[h.Environment].Services[h.Service].Workflow.Steps {
					logger.Tracef("", "Checking depends on status for %s: %v", stName, st.DependsOn.Success)
					logger.Tracef("", "Got auto execute of %v", st.AutoExecute)
					if utils.Contains(st.DependsOn.Success, s.Step) && st.AutoExecute {
						trigger, err := checkDeps(st, *h)
						if err != nil {
							logger.Errorf("", "Error checking dependency states: %s", err.Error())
							continue
						}
						logger.Tracef("", "Check deps returned %v", trigger)
						if trigger {
							logger.Tracef("", "Adding %s to trigger slice", stName)
							toTrigger = append(toTrigger, stName)
						}
					}
				}
			case constants.STATUS_TRIGGER_ERROR:
				logger.Tracef("", "Dropping into error case")
				logger.Tracef("", "Checking against steps %v", p.Environments[h.Environment].Services[h.Service].Workflow.Steps)
				for stName, st := range p.Environments[h.Environment].Services[h.Service].Workflow.Steps {
					logger.Tracef("", "Checking depends on status for %s: %v", stName, st.DependsOn.Error)
					logger.Tracef("", "Got auto execute of %v", st.AutoExecute)
					if utils.Contains(st.DependsOn.Error, s.Step) && st.AutoExecute {
						trigger, err := checkDeps(st, *h)
						if err != nil {
							logger.Errorf("", "Error checking dependency states: %s", err.Error())
							continue
						}
						logger.Tracef("", "Check deps returned %v", trigger)
						if trigger {
							logger.Tracef("", "Adding %s to trigger slice", stName)
							toTrigger = append(toTrigger, stName)
						}
					}
				}
			case constants.STATUS_TRIGGER_ALWAYS:
				logger.Tracef("", "Dropping into always case")
				logger.Tracef("", "Checking against steps %v", p.Environments[h.Environment].Services[h.Service].Workflow.Steps)
				for stName, st := range p.Environments[h.Environment].Services[h.Service].Workflow.Steps {
					logger.Tracef("", "Checking depends on status for %s: %v", stName, st.DependsOn.Always)
					logger.Tracef("", "Got auto execute of %v", st.AutoExecute)
					if utils.Contains(st.DependsOn.Always, s.Step) && st.AutoExecute {
						trigger, err := checkDeps(st, *h)
						if err != nil {
							logger.Errorf("", "Error checking dependency states: %s", err.Error())
							continue
						}
						logger.Tracef("", "Check deps returned %v", trigger)
						if trigger {
							logger.Tracef("", "Adding %s to trigger slice", stName)
							toTrigger = append(toTrigger, stName)
						}
					}
				}
			}
			for _, sName := range toTrigger {
				logger.Infof("", "Triggering new step: %s", sName)
				if _, ok := p.Environments[h.Environment]; !ok {
					logger.Warnf("", "No environment %s exists in project %s", h.Environment, h.Project)
					continue
				}
				if _, ok := p.Environments[h.Environment].Services[h.Service]; !ok {
					logger.Warnf("", "No service %s exists in project %s's %s environment", h.Service, h.Project, h.Environment)
					continue
				}
				w := p.Environments[h.Environment].Services[h.Service].Workflow
				idx := len(h.States)
				contextBytes, _ := json.Marshal(h.Context)
				if err := StartRun(s.RunID, w, idx, string(contextBytes), sName, w.Steps[sName].Language); err != nil {
					logger.Errorf("", "Unable to start run %s idx %d: %s", s.RunID, idx, err.Error())
					continue
				}
				logger.Tracef("", "Run %s idx %d is starting", s.RunID, idx)
				go monitor.Run(s.RunID, sName, len(h.States))
			}
			monitor.Finished = monitor.Finished[1:]
			monitor.FinishedLock.Unlock()
		}
		time.Sleep(1 * time.Second)
	}
}
