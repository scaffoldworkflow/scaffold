package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/bulwark"
	"scaffold/server/cmd"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/container"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/mongodb"
	"scaffold/server/msg"
	"scaffold/server/run"
	"scaffold/server/state"
	"scaffold/server/task"
	"time"

	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"
)

var RunQueue []run.Run

var JoinKey = ""
var PrimaryKey = ""
var ID = ""
var isRunning = false
var currentTask = ""
var currentCascade = ""

func Run() {
	ID = uuid.New().String()

	mongodb.InitCollections()
	filestore.InitBucket()
	container.CompletedRuns = make(map[string]run.Run)
	StartWebsocketServer()

	go EnsureManagerConnection()

	health.IsHealthy = true

	go container.PruneContainers()

	go bulwark.RunManager(nil)
	go bulwark.RunWorker(QueueDataReceive)
	// go bulwark.RunBuffer(BufferDataReceive)

	go bufferCheck()
	queueCheck()
}

func JoinManager() error {
	JoinKey = config.Config.Node.JoinKey
	PrimaryKey = config.Config.Node.PrimaryKey

	obj := auth.NodeJoinObject{
		Name:     ID,
		Host:     config.Config.Host,
		Port:     config.Config.Port,
		WSPort:   config.Config.WSPort,
		Protocol: config.Config.Protocol,
		JoinKey:  JoinKey,
		Version:  constants.VERSION,
	}
	postBody, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s://%s:%d/auth/join", config.Config.Node.ManagerProtocol, config.Config.Node.ManagerHost, config.Config.Node.ManagerPort)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("received join status code %d", resp.StatusCode)
	}
	return nil
}

func CheckManagerHealth() error {
	queryURL := fmt.Sprintf("%s://%s:%d/health/ready", config.Config.Node.ManagerProtocol, config.Config.Node.ManagerHost, config.Config.Node.ManagerPort)
	resp, err := http.Get(queryURL)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("unable to reach manager, response code: %d", resp.StatusCode)
	}
	return nil
}

func DoPing() int {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s://%s:%d/health/ping/%s", config.Config.Node.ManagerProtocol, config.Config.Node.ManagerHost, config.Config.Node.ManagerPort, ID)
	req, _ := http.NewRequest("POST", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", PrimaryKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "manager ping returned error %s", err.Error())
		return -1
	}
	if resp.StatusCode >= 400 {
		return resp.StatusCode
	}
	return 0
}

func EnsureManagerConnection() {
	err := JoinManager()
	health.IsReady = false
	for err != nil {
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
		err = JoinManager()
	}
	health.IsReady = true
	for {
		rc := DoPing()
		if rc == http.StatusUnauthorized {
			err := JoinManager()
			health.IsReady = false
			for err != nil {
				health.IsReady = true
				time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
				err = JoinManager()
			}
		}
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
	}
}

func QueueDataReceive(endpoint, data string) error {
	isRunning = true
	logger.Debugf("", "Got queue pop data: %s", data)

	if len(data) == 0 {
		isRunning = false
		return nil
	}

	var m msg.TriggerMsg
	// bytes, err := json.Marshal([]byte(data))
	// if err != nil {
	// 	logger.Errorf("", "Error processing queue message: %s", err.Error())
	// 	return err
	// }
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		logger.Errorf("", "Error processing queue message: %s", err.Error())
		isRunning = false
		return err
	}

	logger.Debugf("", "Action object: %v", m)

	switch m.Action {
	case constants.ACTION_TRIGGER:
		t, err := task.GetTaskByNames(m.Cascade, m.Task)
		if err != nil {
			logger.Errorf("", "Error getting task %s.%s: %s", m.Cascade, m.Task, err.Error())
			isRunning = false
			return err
		}

		r := run.Run{
			Name:   uuid.New().String(),
			Task:   *t,
			Number: m.Number,
			Groups: m.Groups,
			State: state.State{
				Task:     m.Task,
				Cascade:  m.Cascade,
				Status:   constants.STATE_STATUS_WAITING,
				Started:  "",
				Finished: "",
				Output:   "",
				Number:   m.Number,
				Worker:   ID,
				Display:  make([]map[string]interface{}, 0),
			},
			Worker: ID,
		}

		currentTask = m.Task
		currentCascade = m.Cascade

		run.Kill(bulwark.ManagerClient, []string{fmt.Sprintf("%s.%s", m.Cascade, m.Task)})

		shouldRestart, _ := run.StartRun(bulwark.ManagerClient, &r)
		for shouldRestart {
			shouldRestart, _ = run.StartRun(bulwark.ManagerClient, &r)
		}

		logger.Debugf("", "Run finished")
		container.LastRun = append(container.LastRun, container.CurrentRun.Name)
		container.LastImage = append(container.LastImage, container.CurrentRun.Task.Image)
		container.LastGroups = append(container.LastGroups, m.Groups)

		currentTask = ""
		currentCascade = ""
	}

	isRunning = false
	return nil
}

func queueCheck() {
	for {
		time.Sleep(time.Duration(config.Config.BulwarkCheckInterval) * time.Millisecond)
		if health.IsReady {
			logger.Tracef("", "Sleeping...")
			time.Sleep(time.Duration(config.Config.BulwarkCheckInterval) * time.Millisecond)
			if !isRunning {
				logger.Debugf("", "Worker checking queue")
				bulwark.QueuePop(bulwark.WorkerClient)
			}
		}
	}
}

// func BufferDataReceive(endpoint, data string) error {
// 	if len(data) == 0 {
// 		return nil
// 	}
// 	logger.Debugf("", "Got buffer data %s", data)
// 	var runNames []string
// 	if err := json.Unmarshal([]byte(data), &runNames); err != nil {
// 		logger.Errorf("", "Unable to marshal JSON: %s", err.Error())
// 		return err
// 	}

// 	if err := run.Kill(bulwark.ManagerClient, runNames); err != nil {
// 		logger.Errorf("", "Encountered error trying to kill runs: %s", err.Error())
// 	}
// 	return nil
// }

func bufferCheck() {
	for {
		time.Sleep(time.Duration(config.Config.BulwarkCheckInterval) * time.Millisecond)
		if health.IsReady {
			logger.Tracef("", "Sleeping...")
			time.Sleep(time.Duration(config.Config.BulwarkCheckInterval) * time.Millisecond)
			logger.Debugf("", "Worker checking buffer")
			// bulwark.BufferGet(bulwark.BufferClient)
			logger.Errorf("", "Current run: %s.%s", currentCascade, currentTask)
			if currentTask != "" && currentCascade != "" {
				s, err := state.GetStateByNames(currentCascade, currentTask)
				if err != nil {
					logger.Errorf("", "Error getting run state: %s", err.Error())
					continue
				}
				logger.Debugf("", "Current run state: %s, %v", s.Status, s.Killed)
				if s.Killed {
					if err := run.Kill(bulwark.ManagerClient, []string{fmt.Sprintf("%s-%s", currentCascade, currentTask)}); err != nil {
						logger.Errorf("", "Error killing run: %s", err.Error())
					}
				}
			}
		}
	}
}

func StartWebsocketServer() {
	logger.Info("", "Starting websocket application")
	//Open a goroutine execution start program
	// go socket.Manager.Start()
	go cmd.StartWSServer()
}
