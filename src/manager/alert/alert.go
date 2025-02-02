package alert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/manager/config"
	"scaffold/manager/constants"
	"scaffold/manager/kernel"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"scaffold/manager/mongodb"

	_ "embed"

	logger "github.com/jfcarter2358/go-logger"
)

type Alert struct {
	Workflow     string                `json:"workflow" bson:"workflow" yaml:"workflow"`
	Created      string                `json:"created" bson:"created" yaml:"created"`
	Updated      string                `json:"updated" bson:"updated" yaml:"updated"`
	ID           string                `json:"id" bson:"id" yaml:"id"`
	Name         string                `json:"name" bson:"name" yaml:"name"`
	Script       string                `json:"script" bson:"script" yaml:"script"`
	Kernels      map[string]KernelInfo `json:"kernels" bson:"kernels" yaml:"kernels"`
	Requirements string                `json:"requirements" bson:"requirements" yaml:"requirements"`
	Language     string                `json:"language" bson:"language" yaml:"language"`
	Enabled      bool                  `json:"enabled" bson:"enabled" yaml:"enabled"`
}

type KernelInfo struct {
	RunID    string `json:"run_id" bson:"run_id" yaml:"run_id"`
	Output   string `json:"output" bson:"output" yaml:"output"`
	Status   string `json:"status" bson:"status" yaml:"status"`
	Finished bool   `json:"finished" bson:"finished" yaml:"finished"`
}

type OutputMessage struct {
	Contents string `json:"contents"`
	Finished bool   `json:"finished"`
}

func CreateAlert(a *Alert) error {
	currentTime := time.Now().UTC()
	a.Created = currentTime.Format("2006-01-02T15:04:05Z")
	a.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	a.Kernels = make(map[string]KernelInfo)

	logger.Debugf("", "Creating alert %s for workflow %s", a.ID, a.Workflow)

	_, err := mongodb.Collections[constants.MONGODB_ALERT_COLLECTION_NAME].InsertOne(mongodb.Ctx, a)
	return err
}

func DeleteAlertByID(id string) error {
	filter := bson.M{"id": id}

	logger.Debugf("", "Deleting alert with ID %s", id)

	as, err := FilterAlerts(filter)
	if err != nil {
		logger.Errorf("", "Could not get alerts to kill Kernels")
		return err
	}

	logger.Debugf("", "Got %d alerts with ID %s", len(as), id)

	for _, a := range as {
		if err := a.Stop(); err != nil {
			logger.Errorf("", "Unable to stop alert %s", a.ID)
			return err
		}
	}

	collection := mongodb.Collections[constants.MONGODB_ALERT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not perform delete_many on alerts")
		return err
	}

	logger.Debugf("", "Deleted %d alerts with ID %s", result.DeletedCount, id)

	if result.DeletedCount == 0 {
		return fmt.Errorf("no alert found with ID %s", id)
	}

	return nil
}

func DeleteAlertsByWorkflow(workflow string) error {
	filter := bson.M{"workflow": workflow}

	as, err := FilterAlerts(filter)
	if err != nil {
		logger.Errorf("", "Could not get alerts to kill Kernels")
		return err
	}

	for _, a := range as {
		if err := a.Stop(); err != nil {
			logger.Errorf("", "Unable to stop alert %s", a.ID)
			return err
		}
	}

	collection := mongodb.Collections[constants.MONGODB_ALERT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		logger.Errorf("", "Could not perform delete_many on alerts")
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no alerts found with workflow %s", workflow)
	}

	return nil
}

func GetAllAlerts() ([]*Alert, error) {
	filter := bson.D{{}}

	alerts, err := FilterAlerts(filter)

	return alerts, err
}

func GetAlertByID(id string) (*Alert, error) {
	filter := bson.M{"id": id}

	alerts, err := FilterAlerts(filter)

	if err != nil {
		return nil, err
	}

	if len(alerts) == 0 {
		return nil, nil
	}

	if len(alerts) > 1 {
		logger.Errorf("", "%v", alerts)
		return nil, fmt.Errorf("multiple alerts found with ID %s", id)
	}

	return alerts[0], nil
}

func GetAlertsByWorkflow(workflow string) ([]*Alert, error) {
	filter := bson.M{"workflow": workflow}

	alerts, err := FilterAlerts(filter)

	if err != nil {
		return nil, err
	}

	return alerts, nil
}

func UpdateAlertByID(id string, a *Alert) error {
	filter := bson.M{"id": id}

	currentTime := time.Now().UTC()
	a.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_ALERT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.ReplaceOne(ctx, filter, a)

	if err != nil {
		logger.Errorf("", "Encountered error in update: %s", err.Error())
		// if result.ModifiedCount == 0 {
		// 	logger.Debugf("", "Alert %s does not exist, creating...", a.ID)
		// 	return CreateAlert(a)
		// }
		return err
	}

	if result.ModifiedCount == 0 {
		logger.Debugf("", "Alert %s does not exist, creating...", a.ID)
		return CreateAlert(a)
	}

	logger.Infof("", "Updated alert with ID %s", id)

	return err
}

func FilterAlerts(filter interface{}) ([]*Alert, error) {
	// A slice of tasks for storing the decoded documents
	var alerts []*Alert

	collection := mongodb.Collections[constants.MONGODB_ALERT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return alerts, err
	}

	for cur.Next(ctx) {
		var a Alert
		err := cur.Decode(&a)
		if err != nil {
			return alerts, err
		}

		alerts = append(alerts, &a)
	}

	if err := cur.Err(); err != nil {
		return alerts, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return alerts, nil
}

func (a *Alert) SetupKernel() (string, error) {
	k := &kernel.Kernel{
		Requirements: a.Requirements,
	}
	if err := k.Spawn(); err != nil {
		logger.Errorf("", "Unable to spawn kernel for alert %s", a.ID)
		return "", err
	}
	logger.Debugf("", "Successfully spawned kernel with ID %s", k.ID)
	max_iterations := 100 // this equates to 10 seconds of waiting for the kernel to start
	iteration := 0
	for !k.Ready {
		time.Sleep(100 * time.Millisecond)
		if iteration == max_iterations {
			logger.Errorf("", "Kernel health check timed out with ID %s", k.ID)
			return "", fmt.Errorf("kernel health check timed out with ID %s", k.ID)
		}
		iteration += 1
	}
	kernel.Kernels[k.ID] = k
	a.Kernels[k.ID] = KernelInfo{}
	UpdateAlertByID(a.ID, a)
	return k.ID, nil
}

func (a *Alert) Run() (string, error) {
	// if !a.Enabled {
	// 	logger.Errorf("", "Alert %s is not enabled", a.ID)
	// 	return "", fmt.Errorf("alert %s is not enabled", a.ID)
	// }
	// k := &kernel.Kernel{
	// 	Requirements: a.Requirements,
	// }
	// if err := k.Spawn(); err != nil {
	// 	return "", err
	// }
	// logger.Debugf("", "Successfully spawned alert kernel with ID %s", k.ID)
	// a.Kernels = append(a.Kernels, k.ID)
	// max_iterations := 100 // this equates to 10 seconds of waiting for the kernel to start
	// iteration := 0
	// for !k.Ready {
	// 	time.Sleep(100 * time.Millisecond)
	// 	if iteration == max_iterations {
	// 		return "", fmt.Errorf("alert kernel health check timed out with ID %s", k.ID)
	// 	}
	// 	iteration += 1
	// }

	// kernel.Kernels[k.ID] = k
	// UpdateAlertByID(a.ID, a)
	// kernelID, err := a.setupKernel()
	// if err != nil {
	// 	logger.Errorf("", "Cannot setup kernel for alert %s", a.ID)
	// 	return "", err
	// }
	// go a.Check(kernelID)
	// return kernelID, nil
	return "", nil
}

func getKernelOutput(kernelID string, kernelInfo KernelInfo) (string, bool, error) {
	logger.Debugf("", "Getting output from kernel %s with run %s", kernelID, kernelInfo.RunID)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://localhost:%d/api/v1/kernel/%s/%s", config.Config.Port, kernelID, kernelInfo.RunID)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Could not get kernel output: %s", err.Error())
		return "", false, err
	}

	logger.Debugf("", "Got output from API")
	var data OutputMessage
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("", "Cannot read kernel output response: %s", err.Error())
		return "", false, err
	}
	logger.Tracef("", "Got response: %s", string(body))
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Errorf("", "Could not unmarshal body into JSON: %s", err.Error())
		return "", false, err
	}
	return data.Contents, data.Finished, nil
}

func (a *Alert) Check(kernelID string) {
	// k := kernel.Kernels[id]
	// defer k.Kill()
	// a.Statuses[id] = constants.ALERT_STATUS_RUNNING
	// k.Channel <- fmt.Sprintf("marathon::lang_%s\n%s", a.Language, a.Script)
	// for k.Running {
	// 	logger.Debugf("", "Kernel is still running...")
	// 	output := k.GetOutput()
	// 	a.Outputs[id] = output
	// 	UpdateAlertByID(id, a)
	// 	time.Sleep(1 * time.Second)
	// }
	// if k.Errored {
	// 	logger.Errorf("", "Kernel %s errored out", id)
	// 	a.Statuses[id] = constants.ALERT_STATUS_ERROR
	// }
	// if k.Killed {
	// 	logger.Errorf("", "Kernel %s was killed", id)
	// 	a.Statuses[id] = constants.ALERT_STATUS_KILLED
	// }
	// output := k.GetOutput()
	// a.Outputs[id] = output
	// UpdateAlertByID(id, a)
	// delete(kernel.Kernels, id)
	// utils.Remove(a.Kernels, id)

	finished := false
	output := ""
	var err error

	for !finished {

		ki := a.Kernels[kernelID]

		output, finished, err = getKernelOutput(kernelID, ki)
		if err != nil {
			logger.Errorf("", "Error getting kernel output: %s", err.Error())
			continue
		}

		ki.Output = output
		ki.Finished = finished
		a.Kernels[kernelID] = ki

		UpdateAlertByID(a.ID, a)

		time.Sleep(1 * time.Second)
	}
	// TODO: Delete kernel after it's done
}

func (a *Alert) Stop() error {
	for kernelID := range a.Kernels {
		if err := kernel.Kernels[kernelID].Kill(); err != nil {
			return err
		}
		delete(kernel.Kernels, kernelID)
		// utils.Remove(a.Kernels, id)
	}
	return nil
}
