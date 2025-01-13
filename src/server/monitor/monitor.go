package monitor

import (
	"encoding/json"
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/runbook/kernel"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"scaffold/server/mongodb"

	_ "embed"

	logger "github.com/jfcarter2358/go-logger"
)

//go:embed k8s_monitor.py
var k8sMonitorScript string

//go:embed file_monitor.py
var fileMonitorScript string

type Monitor struct {
	Name         string                 `json:"name" bson:"name" yaml:"name"`
	ID           string                 `json:"id" bson:"id" yaml:"id"`
	Workflow     string                 `json:"workflow" bson:"workflow" yaml:"workflow"`
	Created      string                 `json:"created" bson:"created" yaml:"created"`
	Updated      string                 `json:"updated" bson:"updated" yaml:"updated"`
	Contents     string                 `json:"contents" bson:"contents" yaml:"contents"`
	Arguments    map[string]interface{} `json:"arguments" bson:"arguments" yaml:"arguments"`
	Groups       []string               `json:"groups" bson:"groups" yaml:"groups"`
	Kind         string                 `json:"kind" bson:"kind" yaml:"kind"`
	Status       string                 `json:"status" bson:"status" yaml:"status"`
	Requirements string                 `json:"requirements" bson:"requirements" yaml:"requirements"`
	Enabled      bool                   `json:"enabled" bson:"enabled" yaml:"enabled"`
	Alerts       []string               `json:"alerts" bson:"alerts" yaml:"alerts"`
}

var Kernels map[string]*kernel.Kernel

func InitMonitors() error {
	monitors, err := GetAllMonitors()
	if err != nil {
		logger.Errorf("", "Unable to get monitors on initialization: %s", err.Error())
		return err
	}
	for _, m := range monitors {
		if m.Enabled {
			if err := m.Start(); err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateMonitor(m *Monitor) error {
	currentTime := time.Now().UTC()
	m.Created = currentTime.Format("2006-01-02T15:04:05Z")
	m.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	switch m.Kind {
	case constants.MONITOR_KIND_K8S:
		m.Contents = k8sMonitorScript
	case constants.MONITOR_KIND_FILE:
		m.Contents = fileMonitorScript
	}

	logger.Debugf("", "Creating monitor %s for workflow %s", m.ID, m.Workflow)

	if m.Enabled {
		m.Status = constants.MONITOR_STATUS_RUNNING
		if err := m.Start(); err != nil {
			logger.Errorf("", "Cannot start kernel for monitor %s", m.ID)
			return err
		}
	} else {
		m.Status = constants.MONITOR_STATUS_STOPPED
	}
	_, err := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME].InsertOne(mongodb.Ctx, m)
	return err
}

func DeleteMonitorByID(id string) error {
	filter := bson.M{"id": id}

	logger.Debugf("", "Deleting monitor with ID %s", id)

	ms, err := FilterMonitors(filter)
	if err != nil {
		logger.Errorf("", "Could not get monitors to kill Kernels")
		return err
	}

	logger.Debugf("", "Got %d monitors with ID %s", len(ms), id)

	for _, m := range ms {
		if err := m.Stop(); err != nil {
			logger.Errorf("", "Unable to stop monitor %s", m.ID)
		}
		delete(Kernels, m.ID)
	}

	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	logger.Debugf("", "Deleted %d monitors with ID %s", result.DeletedCount, id)

	if result.DeletedCount == 0 {
		return fmt.Errorf("no monitor found with ID %s", id)
	}

	return nil
}

func DeleteMonitorsByWorkflow(workflow string) error {
	filter := bson.M{"workflow": workflow}

	ms, err := FilterMonitors(filter)
	if err != nil {
		logger.Errorf("", "Could not get monitors to kill Kernels")
		return err
	}

	for _, m := range ms {
		if err := m.Stop(); err != nil {
			logger.Errorf("", "Unable to stop monitor %s", m.ID)
		}
		delete(Kernels, m.ID)
	}

	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no monitors found with workflow %s", workflow)
	}

	return nil
}

func GetAllMonitors() ([]*Monitor, error) {
	filter := bson.D{{}}

	monitors, err := FilterMonitors(filter)

	return monitors, err
}

func GetMonitorByID(id string) (*Monitor, error) {
	filter := bson.M{"id": id}

	monitors, err := FilterMonitors(filter)

	if err != nil {
		return nil, err
	}

	if len(monitors) == 0 {
		return nil, nil
	}

	if len(monitors) > 1 {
		logger.Errorf("", "%v", monitors)
		return nil, fmt.Errorf("multiple monitors found with ID %s", id)
	}

	return monitors[0], nil
}

func GetMonitorsByWorkflow(workflow string) ([]*Monitor, error) {
	filter := bson.M{"workflow": workflow}

	monitors, err := FilterMonitors(filter)

	if err != nil {
		return nil, err
	}

	return monitors, nil
}

func UpdateMonitorByID(id string, m *Monitor) error {
	filter := bson.M{"id": id}

	currentTime := time.Now().UTC()
	m.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
	ctx := mongodb.Ctx

	// opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, m)

	if err != nil {
		logger.Errorf("", "Encountered error in update: %s", err.Error())
		if result.ModifiedCount == 0 {
			logger.Debugf("", "Monitor %s does not exist, creating...", m.ID)
			return CreateMonitor(m)
		}
		return err
	}

	if result.ModifiedCount == 0 {
		logger.Debugf("", "Monitor %s does not exist, creating...", m.ID)
		return CreateMonitor(m)
	}

	logger.Infof("", "Updated monitor with ID %s", id)

	return err
}

func FilterMonitors(filter interface{}) ([]*Monitor, error) {
	// A slice of tasks for storing the decoded documents
	var monitors []*Monitor

	collection := mongodb.Collections[constants.MONGODB_MONITOR_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return monitors, err
	}

	for cur.Next(ctx) {
		var m Monitor
		err := cur.Decode(&m)
		if err != nil {
			return monitors, err
		}

		monitors = append(monitors, &m)
	}

	if err := cur.Err(); err != nil {
		return monitors, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return monitors, nil
}

func (m *Monitor) Start() error {
	k := &kernel.Kernel{
		Requirements: m.Requirements,
	}
	if err := k.Spawn(); err != nil {
		return err
	}
	logger.Debugf("", "Successfully spawned monitor kernel with ID %s", k.ID)
	max_iterations := 100 // this equates to 10 seconds of waiting for the kernel to start
	iteration := 0
	for !k.Ready {
		time.Sleep(100 * time.Millisecond)
		if iteration == max_iterations {
			return fmt.Errorf("monitor kernel health check timed out with ID %s", k.ID)
		}
		iteration += 1
	}

	arguments, err := json.Marshal(m.Arguments)
	if err != nil {
		logger.Errorf("", "Could not marshal arguments into JSON: %s", err.Error())
		return err
	}

	k.Channel <- fmt.Sprintf("config = %s\n\n%s", string(arguments), m.Contents)
	Kernels[m.ID] = k
	m.Status = constants.MONITOR_STATUS_RUNNING
	go m.Check()
	return nil
}

func (m *Monitor) Toggle() error {
	switch m.Status {
	case constants.MONITOR_STATUS_RUNNING:
		logger.Debugf("", "Toggling monitor %s to stopped", m.ID)
		if err := m.Stop(); err != nil {
			logger.Errorf("", "Error starting monitor %s: %s", m.ID, err.Error())
			return err
		}
	case constants.MONITOR_STATUS_STOPPED:
		logger.Debugf("", "Toggling monitor %s to running", m.ID)
		if err := m.Start(); err != nil {
			logger.Errorf("", "Error stopping monitor %s: %s", m.ID, err.Error())
		}
	}
	return nil
}

func (m *Monitor) Stop() error {
	m.Status = constants.MONITOR_STATUS_STOPPED

	if err := Kernels[m.ID].Kill(); err != nil {
		logger.Errorf("", "Error killing monitor %s's kernel %s: %s", m.ID, Kernels[m.ID].ID, err.Error())
		return err
	}
	return nil
}

func (m *Monitor) Check() {
	for m.Status == constants.MONITOR_STATUS_RUNNING || m.Status == constants.MONITOR_STATUS_ALERT {
		output := Kernels[m.ID].GetOutput()
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "error::") {
				m.Status = constants.MONITOR_STATUS_ALERT
				// create and send alert here
			}
		}
		time.Sleep(1 * time.Second)
	}
}
