package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/mongodb"
	"scaffold/server/run"
	"time"
)

var RunQueue []run.Run
var CompletedRuns []run.Run
var CurrentRun run.Run
var JoinKey = ""
var PrimaryKey = ""

func Run() {
	mongodb.InitCollections()
	filestore.InitBucket()

	health.IsHealthy = true

	JoinKey = config.Config.Node.JoinKey
	PrimaryKey = config.Config.Node.PrimaryKey

	obj := auth.NodeJoinObject{
		Name:    config.Config.HTTPHost,
		Host:    config.Config.HTTPHost,
		Port:    config.Config.HTTPPort,
		JoinKey: JoinKey,
	}
	postBody, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://%s:%d/auth/join", config.Config.Node.ManagerHost, config.Config.Node.ManagerPort)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode >= 400 {
		panic(fmt.Sprintf("Received join status code %d", resp.StatusCode))
	}

	health.IsReady = true
	health.IsAvailable = true

	PollQueue()
}

func PollQueue() {
	for {
		if !health.IsAvailable {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if len(RunQueue) > 0 {
			health.IsAvailable = false
			CurrentRun, RunQueue = RunQueue[0], RunQueue[1:]
			run.StartRun(&CurrentRun)
			CompletedRuns = append(CompletedRuns, CurrentRun)
			health.IsAvailable = true
		}
	}
}
