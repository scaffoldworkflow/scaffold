package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/health"
	"scaffold/server/msg"
	"scaffold/server/run"
	"scaffold/server/state"
	"scaffold/server/task"
	"time"

	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"
)

var RunQueue []run.Run
var startTime int64

var JoinKey = ""
var PrimaryKey = ""
var ID = ""
var isRunning = false

func Run() {
	startTime = time.Now().UTC().Unix()

	ID = uuid.New().String()

	go EnsureManagerConnection()

	health.IsHealthy = true
	if config.Config.RestartPeriod > 0 {
		for {
			if !isRunning {
				now := time.Now().UTC().Unix()
				if now-startTime > int64(config.Config.RestartPeriod) {
					os.Exit(0)
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
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

func QueueDataReceive(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	isRunning = true

	var m msg.TriggerMsg

	if err := json.Unmarshal(data, &m); err != nil {
		logger.Errorf("", "Error processing queue message: %s", err.Error())
		isRunning = false
		return err
	}

	logger.Debugf("", "Action object: %v", m)

	switch m.Action {
	case constants.ACTION_TRIGGER:
		t, err := task.GetTaskByNames(m.Workflow, m.Task)
		if err != nil {
			logger.Errorf("", "Error getting task %s.%s: %s", m.Workflow, m.Task, err.Error())
			isRunning = false
			return err
		}

		r := run.Run{
			Name:   uuid.New().String(),
			Task:   *t,
			Number: m.Number,
			Groups: m.Groups,
			RunID:  m.RunID,
			State: state.State{
				Task:     m.Task,
				Workflow: m.Workflow,
				Status:   constants.STATE_STATUS_WAITING,
				Started:  "",
				Finished: "",
				Output:   "",
				Number:   m.Number,
				Worker:   ID,
				Display:  make([]map[string]interface{}, 0),
				Context:  m.Context,
			},
			Worker:  ID,
			Context: m.Context,
		}

		if t.Kind == constants.TASK_KIND_CONTAINER {
			// run.ContainerKill(m.Workflow, m.Task)

			for {
				s, err := state.GetStateByNames(m.Workflow, m.Task)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_RUNNING {
					break
				}
				time.Sleep(time.Duration(config.Config.CheckInterval) * time.Millisecond)
			}

			shouldRestart, _ := run.StartContainerRun(&r)
			for shouldRestart {
				shouldRestart, _ = run.StartContainerRun(&r)
			}
		}
		if t.Kind == constants.TASK_KIND_LOCAL {
			// run.LocalKill(m.Workflow, m.Task)

			for {
				s, err := state.GetStateByNames(m.Workflow, m.Task)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_RUNNING {
					break
				}
				time.Sleep(time.Duration(config.Config.CheckInterval) * time.Millisecond)
			}

			shouldRestart, _ := run.StartLocalRun(&r)
			for shouldRestart {
				shouldRestart, _ = run.StartLocalRun(&r)
			}
		}

		logger.Debugf("", "Run finished")
	}

	isRunning = false
	return nil
}
