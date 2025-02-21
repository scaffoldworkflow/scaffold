package monitor

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"scaffold/manager/config"
	"scaffold/manager/constants"
	"scaffold/manager/history"
	"scaffold/manager/utils"
	"strings"
	"sync"
	"time"

	logger "github.com/jfcarter2358/go-logger"
)

var Finished = make([]history.State, 0)
var FinishedLock = &sync.RWMutex{}
var PromoteFinished = make([]history.State, 0)
var PromoteFinishedLock = &sync.Mutex{}

type Metadata struct {
	Labels map[string]string `json:"labels"`
	Name   string            `json:"name"`
}

type Condition struct {
	Type               string `json:"type"`
	LastTransitionTime string `json:"lastTransitionTime"`
}

type Status struct {
	CompletionTime string      `json:"completionTime"`
	StartTime      string      `json:"startTime"`
	Ready          int         `json:"ready"`
	Succeeded      int         `json:"succeeded"`
	Failed         int         `json:"failed"`
	Active         int         `json:"active"`
	Conditions     []Condition `json:"conditions"`
	Phase          string      `json:"phase"`
}

type Spec struct {
	Suspend bool `json:"suspend"`
}

type Job struct {
	Spec     Spec     `json:"spec"`
	Status   Status   `json:"status"`
	Metadata Metadata `json:"metadata"`
}

type Jobs struct {
	Items []Job `json:"items"`
}

type Pod struct {
	Metadata Metadata `json:"metadata"`
	Status   Status   `json:"status"`
}

type Pods struct {
	Pods []Pod `json:"items"`
}

func MonitorDrift() error {
	hs, err := history.GetAllHistories()
	if err != nil {
		logger.Errorf("", "Unable to get histories to reconcile monitor drift: %s", err)
		return err
	}

	for _, h := range hs {
		for _, s := range h.States {
			if s.Status == constants.STATE_STATUS_RUNNING || s.Status == constants.STATE_STATUS_WAITING {
				if s.Step == "promote" {
					go PromoteRun(h.RunID, "promote", s.Idx)
				} else {
					go Run(h.RunID, s.Step, s.Idx)
				}
			}
		}
	}
	return nil
}

func Run(runID, step string, runIdx int) {
	previousCount := 0
	errorRetry := 10
	waitRetry := 10
	count := 0
	s := &history.State{
		Step:     step,
		Status:   "",
		Started:  "",
		Finished: "",
		Output:   "",
		Idx:      runIdx,
		RunID:    runID,
	}
	if err := history.UpdateState(runID, runIdx, *s); err != nil {
		logger.Errorf("", "Unable to add state to history %s: %s", runID, err.Error())
		return
	}
	podName := ""
	// defer run.StateChange(runID, s.Status, step)
	for {
		// TODO: Make this sleep configurable
		time.Sleep(1 * time.Second)
		if count == errorRetry {
			logger.Errorf("", "Error retry limit reached, killing monitor %s", runID)
			return
		}
		logger.Debugf("", "Getting job status for monitor %s", runID)
		jobOut, jobErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl get job -n %s scaffold-worker-%s-%s -o json", config.Config.K8sNamespace, runID, step)).CombinedOutput()
		if jobErr != nil {
			logger.Errorf("", "Error getting job status for %s: %s", runID, jobErr.Error())
			logger.Debugf("", "%s", string(jobOut))
			count += 1
			continue
		}
		var j Job
		if err := json.Unmarshal(jobOut, &j); err != nil {
			logger.Errorf("", "Error loading job JSON: %s", err.Error())
			count += 1
			continue
		}

		if podName == "" {
			waiting := true
			waitCount := 0
			var ps Pods
			for waiting {
				if count == waitRetry {
					logger.Errorf("", "Wait retry limit reached, killing monitor %s", runID)
					return
				}
				podOut, podErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl get pod -n %s -l controller-uid=%s -o json", config.Config.K8sNamespace, j.Metadata.Labels["controller-uid"])).CombinedOutput()
				if podErr != nil {
					logger.Errorf("", "Error getting pods for %s job: %s", runID, podErr.Error())
					logger.Debugf("", "%s", string(podOut))
					waitCount += 1
					continue
				}
				if err := json.Unmarshal(podOut, &ps); err != nil {
					logger.Errorf("", "Error loading pods JSON: %s", err.Error())
					waitCount += 1
					continue
				}
				if len(ps.Pods) == 0 {
					logger.Errorf("", "No pods found for job %s", runID)
					s.Status = constants.STATE_STATUS_ERROR
					s.Finished = j.Status.CompletionTime
					history.UpdateState(runID, runIdx, *s)
					FinishedLock.Lock()
					Finished = append(Finished, *s)
					FinishedLock.Unlock()
					return
				}
				podName = ps.Pods[0].Metadata.Name
				if utils.Contains([]string{"Running", "Succeeded", "Failed", "Completed"}, ps.Pods[0].Status.Phase) {
					logger.Infof("", "Pod %s has transitioned phase to %s", podName, ps.Pods[0].Status.Phase)
					waiting = false
					continue
				}
				if ps.Pods[0].Status.Phase != "Running" {
					logger.Debugf("", "Waiting for pod %s to be in Running state", podName)
					logger.Tracef("", "Got pod information: %v", ps.Pods[0])
					waitCount += 1
					time.Sleep(1 * time.Second)
				}
			}
		}

		logOut, logErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl logs -n %s %s", config.Config.K8sNamespace, podName)).CombinedOutput()
		if logErr != nil {
			logger.Errorf("", "Error getting pod logs for job %s: %s", runID, logErr.Error())
			logger.Debugf("", "%s", string(logOut))
			count += 1
			continue
		}
		count = 0
		lines := strings.Split(string(logOut), "\n")
		context := s.ParseLogs(lines, previousCount)
		history.UpdateContext(runID, context)
		s.Status = constants.STATE_STATUS_RUNNING
		s.Started = j.Status.StartTime
		previousCount = len(lines)

		logger.Tracef("", "Got job status %v", j)
		logger.Tracef("", "Got job conditions %v", j.Status.Conditions)
		if j.Spec.Suspend {
			logger.Tracef("", "Job killed")
			s.Status = constants.STATE_STATUS_KILLED
			s.Finished = j.Status.CompletionTime
			FinishedLock.Lock()
			Finished = append(Finished, *s)
			FinishedLock.Unlock()
			history.UpdateState(runID, runIdx, *s)
			return
		}
		if len(j.Status.Conditions) > 0 {
			logger.Tracef("", "Got condition '%s'", j.Status.Conditions[0].Type)
			switch j.Status.Conditions[0].Type {
			case "Failed":
				logger.Tracef("", "Dropped into failed")
				s.Status = constants.STATE_STATUS_ERROR
				s.Finished = j.Status.CompletionTime
				history.UpdateState(runID, runIdx, *s)
				FinishedLock.Lock()
				Finished = append(Finished, *s)
				FinishedLock.Unlock()
				return
			case "Complete":
				logger.Tracef("", "Dropped into complete")
				s.Status = constants.STATE_STATUS_SUCCESS
				s.Finished = j.Status.CompletionTime
				logger.Tracef("", "Updating state with %v", s)
				history.UpdateState(runID, runIdx, *s)
				FinishedLock.Lock()
				Finished = append(Finished, *s)
				FinishedLock.Unlock()
				return
			default:
				logger.Warnf("", "Invalid state status of %s", j.Status.Conditions[0].Type)
			}
		}
		history.UpdateState(runID, runIdx, *s)
	}
}

// TODO: Update this so we pass in the finished slice and the lock maybe to remove code duplication?
func PromoteRun(runID, step string, runIdx int) {
	previousCount := 0
	errorRetry := 10
	waitRetry := 10
	count := 0
	s := &history.State{
		Step:     step,
		Status:   "",
		Started:  "",
		Finished: "",
		Output:   "",
		Idx:      runIdx,
		RunID:    runID,
	}
	if err := history.UpdateState(runID, runIdx, *s); err != nil {
		logger.Errorf("", "Unable to add state to history %s: %s", runID, err.Error())
		return
	}
	podName := ""
	for {
		// TODO: Make this sleep configurable
		time.Sleep(1 * time.Second)
		if count == errorRetry {
			logger.Errorf("", "Error retry limit reached, killing monitor %s", runID)
			return
		}
		logger.Debugf("", "Getting job status for monitor %s", runID)
		jobOut, jobErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl get job -n %s scaffold-worker-%s-%s -o json", config.Config.K8sNamespace, runID, step)).CombinedOutput()
		if jobErr != nil {
			logger.Errorf("", "Error getting job status for %s: %s", runID, jobErr.Error())
			logger.Debugf("", "%s", string(jobOut))
			count += 1
			continue
		}
		var j Job
		if err := json.Unmarshal(jobOut, &j); err != nil {
			logger.Errorf("", "Error loading job JSON: %s", err.Error())
			count += 1
			continue
		}

		if podName == "" {
			waiting := true
			waitCount := 0
			var ps Pods
			for waiting {
				if count == waitRetry {
					logger.Errorf("", "Wait retry limit reached, killing monitor %s", runID)
					return
				}
				podOut, podErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl get pod -n %s -l controller-uid=%s -o json", config.Config.K8sNamespace, j.Metadata.Labels["controller-uid"])).CombinedOutput()
				if podErr != nil {
					logger.Errorf("", "Error getting pods for %s job: %s", runID, podErr.Error())
					logger.Debugf("", "%s", string(podOut))
					waitCount += 1
					continue
				}
				if err := json.Unmarshal(podOut, &ps); err != nil {
					logger.Errorf("", "Error loading pods JSON: %s", err.Error())
					waitCount += 1
					continue
				}
				if len(ps.Pods) == 0 {
					logger.Errorf("", "No pods found for job %s", runID)
					s.Status = constants.STATE_STATUS_ERROR
					s.Finished = j.Status.CompletionTime
					PromoteFinishedLock.Lock()
					PromoteFinished = append(Finished, *s)
					PromoteFinishedLock.Unlock()
					history.UpdateState(runID, runIdx, *s)
					return
				}
				podName = ps.Pods[0].Metadata.Name
				if utils.Contains([]string{"Running", "Succeeded", "Failed", "Completed"}, ps.Pods[0].Status.Phase) {
					logger.Infof("", "Pod %s has transitioned phase to %s", podName, ps.Pods[0].Status.Phase)
					waiting = false
					continue
				}
				if ps.Pods[0].Status.Phase != "Running" {
					logger.Debugf("", "Waiting for pod %s to be in Running state", podName)
					logger.Tracef("", "Got pod information: %v", ps.Pods[0])
					waitCount += 1
					time.Sleep(1 * time.Second)
				}
			}
		}

		logOut, logErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl logs -n %s %s", config.Config.K8sNamespace, podName)).CombinedOutput()
		if logErr != nil {
			logger.Errorf("", "Error getting pod logs for job %s: %s", runID, logErr.Error())
			logger.Debugf("", "%s", string(logOut))
			count += 1
			continue
		}
		count = 0
		lines := strings.Split(string(logOut), "\n")
		context := s.ParseLogs(lines, previousCount)
		history.UpdateContext(runID, context)
		s.Status = constants.STATE_STATUS_RUNNING
		s.Started = j.Status.StartTime
		previousCount = len(lines)

		logger.Tracef("", "Got job status %v", j)
		logger.Tracef("", "Got job conditions %v", j.Status.Conditions)
		if j.Spec.Suspend {
			logger.Tracef("", "Job killed")
			s.Status = constants.STATE_STATUS_KILLED
			s.Finished = j.Status.CompletionTime
			history.UpdateState(runID, runIdx, *s)
			PromoteFinishedLock.Lock()
			PromoteFinished = append(PromoteFinished, *s)
			PromoteFinishedLock.Unlock()
			return
		}
		if len(j.Status.Conditions) > 0 {
			logger.Tracef("", "Got condition '%s'", j.Status.Conditions[0].Type)
			switch j.Status.Conditions[0].Type {
			case "Failed":
				logger.Tracef("", "Dropped into failed")
				s.Status = constants.STATE_STATUS_ERROR
				s.Finished = j.Status.CompletionTime
				history.UpdateState(runID, runIdx, *s)
				PromoteFinishedLock.Lock()
				PromoteFinished = append(PromoteFinished, *s)
				PromoteFinishedLock.Unlock()
				return
			case "Complete":
				logger.Tracef("", "Dropped into complete")
				s.Status = constants.STATE_STATUS_SUCCESS
				s.Finished = j.Status.CompletionTime
				logger.Tracef("", "Updating state with %v", s)
				history.UpdateState(runID, runIdx, *s)
				PromoteFinishedLock.Lock()
				PromoteFinished = append(PromoteFinished, *s)
				PromoteFinishedLock.Unlock()
				return
			default:
				logger.Warnf("", "Invalid state status of %s", j.Status.Conditions[0].Type)
			}
		}

		logger.Tracef("", "No conditions found")
		history.UpdateState(runID, runIdx, *s)
	}
}

// import (
// 	"encoding/json"
// 	"fmt"
// 	"scaffold/manager/constants"
// 	"scaffold/manager/kernel"
// 	"strings"
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson"

// 	"scaffold/manager/mongodb"

// 	_ "embed"

// 	logger "github.com/jfcarter2358/go-logger"
// )

// //go:embed k8s_monitor.py
// var k8sMonitorScript string

// //go:embed file_monitor.py
// var fileMonitorScript string

// type Monitor struct {
// 	Name         string                 `json:"name" bson:"name" yaml:"name"`
// 	ID           string                 `json:"id" bson:"id" yaml:"id"`
// 	Workflow     string                 `json:"workflow" bson:"workflow" yaml:"workflow"`
// 	Created      string                 `json:"created" bson:"created" yaml:"created"`
// 	Updated      string                 `json:"updated" bson:"updated" yaml:"updated"`
// 	Contents     string                 `json:"contents" bson:"contents" yaml:"contents"`
// 	Arguments    map[string]interface{} `json:"arguments" bson:"arguments" yaml:"arguments"`
// 	Groups       []string               `json:"groups" bson:"groups" yaml:"groups"`
// 	Kind         string                 `json:"kind" bson:"kind" yaml:"kind"`
// 	Status       string                 `json:"status" bson:"status" yaml:"status"`
// 	Requirements string                 `json:"requirements" bson:"requirements" yaml:"requirements"`
// 	Enabled      bool                   `json:"enabled" bson:"enabled" yaml:"enabled"`
// 	Alerts       []string               `json:"alerts" bson:"alerts" yaml:"alerts"`
// }

// var Kernels map[string]*kernel.Kernel

// func InitMonitors() error {
// 	monitors, err := GetAllMonitors()
// 	if err != nil {
// 		logger.Errorf("", "Unable to get monitors on initialization: %s", err.Error())
// 		return err
// 	}
// 	for _, m := range monitors {
// 		if m.Enabled {
// 			if err := m.Start(); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func CreateMonitor(m *Monitor) error {
// 	currentTime := time.Now().UTC()
// 	m.Created = currentTime.Format("2006-01-02T15:04:05Z")
// 	m.Updated = currentTime.Format("2006-01-02T15:04:05Z")

// 	switch m.Kind {
// 	case constants.MONITOR_KIND_K8S:
// 		m.Contents = k8sMonitorScript
// 	case constants.MONITOR_KIND_FILE:
// 		m.Contents = fileMonitorScript
// 	}

// 	logger.Debugf("", "Creating monitor %s for workflow %s", m.ID, m.Workflow)

// 	if m.Enabled {
// 		m.Status = constants.MONITOR_STATUS_RUNNING
// 		if err := m.Start(); err != nil {
// 			logger.Errorf("", "Cannot start kernel for monitor %s", m.ID)
// 			return err
// 		}
// 	} else {
// 		m.Status = constants.MONITOR_STATUS_STOPPED
// 	}
// 	_, err := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME].InsertOne(mongodb.Ctx, m)
// 	return err
// }

// func DeleteMonitorByID(id string) error {
// 	filter := bson.M{"id": id}

// 	logger.Debugf("", "Deleting monitor with ID %s", id)

// 	ms, err := FilterMonitors(filter)
// 	if err != nil {
// 		logger.Errorf("", "Could not get monitors to kill Kernels")
// 		return err
// 	}

// 	logger.Debugf("", "Got %d monitors with ID %s", len(ms), id)

// 	for _, m := range ms {
// 		if err := m.Stop(); err != nil {
// 			logger.Errorf("", "Unable to stop monitor %s", m.ID)
// 		}
// 		delete(Kernels, m.ID)
// 	}

// 	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
// 	ctx := mongodb.Ctx

// 	result, err := collection.DeleteMany(ctx, filter)

// 	if err != nil {
// 		return err
// 	}

// 	logger.Debugf("", "Deleted %d monitors with ID %s", result.DeletedCount, id)

// 	if result.DeletedCount == 0 {
// 		return fmt.Errorf("no monitor found with ID %s", id)
// 	}

// 	return nil
// }

// func DeleteMonitorsByWorkflow(workflow string) error {
// 	filter := bson.M{"workflow": workflow}

// 	ms, err := FilterMonitors(filter)
// 	if err != nil {
// 		logger.Errorf("", "Could not get monitors to kill Kernels")
// 		return err
// 	}

// 	for _, m := range ms {
// 		if err := m.Stop(); err != nil {
// 			logger.Errorf("", "Unable to stop monitor %s", m.ID)
// 		}
// 		delete(Kernels, m.ID)
// 	}

// 	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
// 	ctx := mongodb.Ctx

// 	result, err := collection.DeleteMany(ctx, filter)

// 	if err != nil {
// 		return err
// 	}

// 	if result.DeletedCount == 0 {
// 		return fmt.Errorf("no monitors found with workflow %s", workflow)
// 	}

// 	return nil
// }

// func GetAllMonitors() ([]*Monitor, error) {
// 	filter := bson.D{}

// 	monitors, err := FilterMonitors(filter)

// 	return monitors, err
// }

// func GetMonitorByID(id string) (*Monitor, error) {
// 	filter := bson.M{"id": id}

// 	monitors, err := FilterMonitors(filter)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(monitors) == 0 {
// 		return nil, nil
// 	}

// 	if len(monitors) > 1 {
// 		logger.Errorf("", "%v", monitors)
// 		return nil, fmt.Errorf("multiple monitors found with ID %s", id)
// 	}

// 	return monitors[0], nil
// }

// func GetMonitorsByWorkflow(workflow string) ([]*Monitor, error) {
// 	filter := bson.M{"workflow": workflow}

// 	monitors, err := FilterMonitors(filter)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return monitors, nil
// }

// func UpdateMonitorByID(id string, m *Monitor) error {
// 	filter := bson.M{"id": id}

// 	currentTime := time.Now().UTC()
// 	m.Updated = currentTime.Format("2006-01-02T15:04:05Z")

// 	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
// 	ctx := mongodb.Ctx

// 	// opts := options.Replace().SetUpsert(true)

// 	result, err := collection.ReplaceOne(ctx, filter, m)

// 	if err != nil {
// 		logger.Errorf("", "Encountered error in update: %s", err.Error())
// 		// if result.ModifiedCount == 0 {
// 		// 	logger.Debugf("", "Monitor %s does not exist, creating...", m.ID)
// 		// 	return CreateMonitor(m)
// 		// }
// 		return err
// 	}

// 	if result.ModifiedCount == 0 {
// 		logger.Debugf("", "Monitor %s does not exist, creating...", m.ID)
// 		return CreateMonitor(m)
// 	}

// 	logger.Infof("", "Updated monitor with ID %s", id)

// 	return err
// }

// func FilterMonitors(filter interface{}) ([]*Monitor, error) {
// 	// A slice of tasks for storing the decoded documents
// 	var monitors []*Monitor

// 	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
// 	ctx := mongodb.Ctx

// 	cur, err := collection.Find(ctx, filter)
// 	if err != nil {
// 		return monitors, err
// 	}

// 	for cur.Next(ctx) {
// 		var m Monitor
// 		err := cur.Decode(&m)
// 		if err != nil {
// 			return monitors, err
// 		}

// 		monitors = append(monitors, &m)
// 	}

// 	if err := cur.Err(); err != nil {
// 		return monitors, err
// 	}

// 	// once exhausted, close the cursor
// 	cur.Close(ctx)

// 	return monitors, nil
// }

// func (m *Monitor) Start() error {
// 	k := &kernel.Kernel{
// 		Requirements: m.Requirements,
// 	}
// 	if err := k.Spawn(); err != nil {
// 		return err
// 	}
// 	logger.Debugf("", "Successfully spawned monitor kernel with ID %s", k.ID)
// 	max_iterations := 100 // this equates to 10 seconds of waiting for the kernel to start
// 	iteration := 0
// 	for !k.Ready {
// 		time.Sleep(100 * time.Millisecond)
// 		if iteration == max_iterations {
// 			return fmt.Errorf("monitor kernel health check timed out with ID %s", k.ID)
// 		}
// 		iteration += 1
// 	}

// 	arguments, err := json.Marshal(m.Arguments)
// 	if err != nil {
// 		logger.Errorf("", "Could not marshal arguments into JSON: %s", err.Error())
// 		return err
// 	}

// 	k.Channel <- fmt.Sprintf("config = %s\n\n%s", string(arguments), m.Contents)
// 	Kernels[m.ID] = k
// 	m.Status = constants.MONITOR_STATUS_RUNNING
// 	go m.Check()
// 	return nil
// }

// func (m *Monitor) Toggle() error {
// 	switch m.Status {
// 	case constants.MONITOR_STATUS_RUNNING:
// 		logger.Debugf("", "Toggling monitor %s to stopped", m.ID)
// 		if err := m.Stop(); err != nil {
// 			logger.Errorf("", "Error starting monitor %s: %s", m.ID, err.Error())
// 			return err
// 		}
// 	case constants.MONITOR_STATUS_STOPPED:
// 		logger.Debugf("", "Toggling monitor %s to running", m.ID)
// 		if err := m.Start(); err != nil {
// 			logger.Errorf("", "Error stopping monitor %s: %s", m.ID, err.Error())
// 		}
// 	}
// 	return nil
// }

// func (m *Monitor) Stop() error {
// 	m.Status = constants.MONITOR_STATUS_STOPPED

// 	if err := Kernels[m.ID].Kill(); err != nil {
// 		logger.Errorf("", "Error killing monitor %s's kernel %s: %s", m.ID, Kernels[m.ID].ID, err.Error())
// 		return err
// 	}
// 	return nil
// }

// func (m *Monitor) Check() {
// 	for m.Status == constants.MONITOR_STATUS_RUNNING || m.Status == constants.MONITOR_STATUS_ALERT {
// 		output := Kernels[m.ID].GetOutput()
// 		lines := strings.Split(output, "\n")
// 		for _, line := range lines {
// 			line = strings.TrimSpace(line)
// 			if strings.HasPrefix(line, "error::") {
// 				m.Status = constants.MONITOR_STATUS_ALERT
// 				// create and send alert here
// 			}
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// }
