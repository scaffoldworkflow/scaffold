package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/cmd"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/container"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/logger"
	"scaffold/server/mongodb"
	"scaffold/server/run"
	"scaffold/server/state"
	"scaffold/server/task"
	"strings"
	"time"
)

var RunQueue []run.Run

var JoinKey = ""
var PrimaryKey = ""

func Run() {
	mongodb.InitCollections()
	filestore.InitBucket()
	container.CompletedRuns = make(map[string]run.Run)
	StartWebsocketServer()

	health.IsHealthy = true

	logger.Infof("", "Trying to join manager")
	for {
		err := JoinManager()
		if err == nil {
			logger.Successf("", "Successfully joined manager")
			break
		}
		logger.Errorf("", "Unable to reach manager: %s", err.Error())
		logger.Debugf("", "Trying manager again in 5 seconds")
		time.Sleep(5 * time.Second)
	}

	health.IsReady = true
	health.IsAvailable = true

	go container.PruneContainers()

	PollQueue()
}

func JoinManager() error {
	JoinKey = config.Config.Node.JoinKey
	PrimaryKey = config.Config.Node.PrimaryKey

	obj := auth.NodeJoinObject{
		Name:     config.Config.Host,
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

func EnsureManagerConnection() {
	for {
		if err := CheckManagerHealth(); err != nil {
			logger.Errorf("", "Unable to reach manager: %s", err.Error())
			err = JoinManager()
			if err != nil {
				logger.Successf("", "Successfully joined manager")
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func PollQueue() {
	container.InitContainers()
	for {
		if !health.IsAvailable {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if len(RunQueue) > 0 {
			health.IsAvailable = false
			container.CurrentRun, RunQueue = RunQueue[0], RunQueue[1:]
			container.CurrentName = container.CurrentRun.Name
			shouldRestart, _ := run.StartRun(&container.CurrentRun)
			if shouldRestart {
				logger.Debugf("", "Should restart is true")
				RunQueue = append([]run.Run{container.CurrentRun}, RunQueue...)
			} else {
				logger.Debugf("", "Current run name: %s", container.CurrentRun.Name)
				c := container.CurrentRun
				t := c.Task
				s := c.State
				p := c.Previous
				container.CompletedRuns[container.CurrentRun.Name] = run.Run{
					Name: c.Name,
					Task: task.Task{
						Name:        t.Name,
						Cascade:     t.Cascade,
						Verb:        t.Verb,
						DependsOn:   t.DependsOn,
						Image:       t.Image,
						Run:         t.Run,
						Store:       t.Store,
						Load:        t.Load,
						Env:         t.Env,
						Inputs:      t.Inputs,
						Updated:     t.Updated,
						Check:       t.Check,
						RunNumber:   t.RunNumber,
						AutoExecute: t.AutoExecute,
					},
					State: state.State{
						Task:     s.Task,
						Cascade:  s.Cascade,
						Status:   s.Status,
						Started:  s.Started,
						Finished: s.Finished,
						Output:   s.Output,
						Number:   t.RunNumber,
						Display:  s.Display,
					},
					Previous: state.State{
						Task:     p.Task,
						Cascade:  p.Cascade,
						Status:   p.Status,
						Started:  p.Started,
						Finished: p.Finished,
						Output:   p.Output,
						Number:   p.Number,
						Display:  p.Display,
					},
					Number: container.CurrentRun.Number,
					Groups: c.Groups,
				}
				parts := strings.Split(c.Name, ".")
				nameParts := strings.Split(parts[1], "-")
				if !strings.HasSuffix(nameParts[0], "CHECK") {
					logger.Debugf("", "Adding groups %v to LastGroups", c.Groups)
					container.LastRun = append(container.LastRun, container.CurrentRun.Name)
					container.LastImage = append(container.LastImage, container.CurrentRun.Task.Image)
					container.LastGroups = append(container.LastGroups, c.Groups)
				}
			}
			logger.Debugf("", "current run: %v", container.CurrentRun)
			container.CurrentRun = run.Run{}
			health.IsAvailable = true
		}
	}
}

func StartWebsocketServer() {
	logger.Info("", "Starting websocket application")
	//Open a goroutine execution start program
	// go socket.Manager.Start()
	go cmd.StartWSServer()
}
