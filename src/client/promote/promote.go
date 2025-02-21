package promote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/client/auth"
	"time"

	logger "github.com/jfcarter2358/go-logger"
)

func DoPromote(profile, task, data, context string, follow bool) {

	logger.Debugf("", "Trigger task")
	p := auth.ReadProfile(profile)

	logger.Debugf("", "Getting context")
	if context == "" {
		context = p.Workflow
	}

	uri := fmt.Sprintf("%s://%s:%s/api/v1/webhook/%s/%s", p.Protocol, p.Host, p.Port, context, task)

	var inputData map[string]interface{}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &inputData); err != nil {
			logger.Fatalf("", "Cannot unmarshal data to send to trigger: %s", err.Error())
		}
	} else {
		inputData = make(map[string]interface{})
	}

	doTrigger(p, context, uri, task, inputData, follow)
}

type Status struct {
	Task         string `json:"task"`
	Running      bool   `json:"running"`
	Errored      bool   `json:"errored"`
	Waiting      bool   `json:"waiting"`
	Killed       bool   `json:"killed"`
	Success      bool   `json:"success"`
	Unknown      bool   `json:"unknown"`
	NotStarted   bool   `json:"not_started"`
	CurrentState string
}

func getRunStatus(p auth.ProfileObj, context, runID string) Status {
	requestURL := fmt.Sprintf("%s://%s:%s/api/v1/run/%s", p.Protocol, p.Host, p.Port, runID)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "POST request failed with error: %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "POST request failed with status code %v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("", "Encountered error reading body: %s", err.Error())
	}
	var s Status
	logger.Debugf("", "%s", string([]byte(body)))
	err = json.Unmarshal([]byte(body), &s)
	if err != nil {
		logger.Fatalf("", "Encountered error unmarshalling status JSON: %s", err.Error())
	}
	if s.Running {
		s.CurrentState = "running"
	} else if s.Errored {
		s.CurrentState = "errored"
	} else if s.Waiting {
		s.CurrentState = "waiting"
	} else if s.Killed {
		s.CurrentState = "killed"
	} else if s.Success {
		s.CurrentState = "success"
	} else if s.NotStarted {
		s.CurrentState = "not_started"
	} else {
		s.Unknown = true
		s.CurrentState = "unknown"
	}
	return s
}

type History struct {
	States []State `json:"states" bson:"states" yaml:"states"`
}

type State struct {
	Task   string `json:"task" bson:"task" yaml:"task"`
	Output string `json:"output" bson:"output" yaml:"output"`
}

func getLogs(p auth.ProfileObj, context, runID, task string) {
	requestURL := fmt.Sprintf("%s://%s:%s/api/v1/history/%s", p.Protocol, p.Host, p.Port, runID)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "POST request failed with error: %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "POST request failed with status code %v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("", "Encountered error reading body: %s", err.Error())
	}
	var h History
	err = json.Unmarshal([]byte(body), &h)
	if err != nil {
		logger.Fatalf("", "Encountered error unmarshalling history JSON: %s", err.Error())
	}
	for _, s := range h.States {
		if s.Task == task {
			logger.Debugf("", "Found history state with task %s", task)
			logger.Errorf("", "Dumping logs...")
			fmt.Println(s.Output)
		}
	}
}

type WebhookResponse struct {
	Message string `json:"message"`
	RunID   string `json:"run_id"`
}

func doTrigger(p auth.ProfileObj, context, requestURL, task string, data map[string]interface{}, follow bool) {
	logger.Infof("", "Triggering workflow task at %s/%s", context, task)
	postBody, _ := json.Marshal(data)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "POST request failed with error: %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "POST request failed with status code %v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("", "Encountered error reading body: %s", err.Error())
	}
	var output map[string]string
	err = json.Unmarshal([]byte(body), &output)
	if err != nil {
		logger.Fatalf("", "Encountered error unmarshalling webhook JSON: %s", err.Error())
	}
	if output["message"] != "OK" {
		logger.Fatalf("", "Webhook reported status %s", output["message"])
	}
	runID := output["run_id"]
	if !follow {
		logger.Successf("", "Triggered workflow with run ID %s", runID)
		return
	}
	running := true
	lastStatus := "not_started"
	lastTask := task
	counter := 0
	maxChecks := 10
	sleepInterval := 100
	for running {
		s := getRunStatus(p, context, runID)
		if err != nil {
			logger.Fatalf("", "Error getting run status: %s", err.Error())
		}
		if lastTask != s.Task {
			if lastTask != "" {
				logger.Successf("", "Task %s exited with status 'success'", lastTask)
			}
			if s.Task != "" {
				logger.Infof("", "Task %s transitioned from status 'not_started' -> '%s'", s.Task, s.CurrentState)
				if s.Errored {
					logger.Errorf("", "Task %s exited with status 'errored'", s.Task)
				}
			}
			counter = 0
			lastTask = s.Task
			lastStatus = s.CurrentState
			time.Sleep(time.Duration(sleepInterval) * time.Millisecond)
			continue
		} else {
			if !s.Running && !s.Waiting && !s.NotStarted {
				counter += 1
			}
			if counter > maxChecks {
				if s.Errored || s.Killed || s.Unknown {
					logger.Errorf("", "Workflow failed")
					getLogs(p, context, runID, s.Task)
					return
				}
				logger.Successf("", "Workflow succeeded!")
				return
			}
		}

		if s.CurrentState != lastStatus {
			if s.Errored {
				logger.Errorf("", "Task %s exited with status 'errored'", s.Task)
				lastTask = s.Task
				lastStatus = s.CurrentState
				time.Sleep(time.Duration(sleepInterval) * time.Millisecond)
				continue
			}
			logger.Infof("", "Task %s transitioned from status '%s' -> '%s'", s.Task, lastStatus, s.CurrentState)
		}
		lastTask = s.Task
		lastStatus = s.CurrentState
		time.Sleep(time.Duration(sleepInterval) * time.Millisecond)
	}
}
