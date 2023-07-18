package run

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/job"
	"scaffold/server/step"
	"time"

	"github.com/google/uuid"
	"github.com/jfcarter2358/ceresdb-go/connection"
)

const JOB_STATUS_RUNNING = "running"
const JOB_STATUS_FAILED = "failed"
const JOB_STATUS_SUCCESS = "success"

type StepStatus struct {
	Status   string `json:"status"`
	Started  string `json:"started"`
	Finished string `json:"finished"`
	Error    string `json:"error"`
	Output   string `json:"output"`
}

type JobStatus struct {
	Steps    map[string]StepStatus `json:"steps"`
	Status   string                `json:"status"`
	Started  string                `json:"started"`
	Finished string                `json:"finished"`
	Error    string                `json:"error"`
}

type Run struct {
	ID       string               `json:"id"`
	Pipeline string               `json:"pipeline"`
	Jobs     map[string]JobStatus `json:"jobs"`
	Status   string               `json:"status"`
	Started  string               `json:"started"`
	Finished string               `json:"finished"`
	Error    string               `json:"error"`
}

func CreateRun(newRun Run) (string, string) {
	currentTime := time.Now().UTC()
	newRun.Started = currentTime.Format("2006-01-02T15:04:05Z")
	newRun.ID = uuid.NewString()
	newRun.Jobs = make(map[string]JobStatus)

	queryData, _ := json.Marshal(&newRun)
	queryString := fmt.Sprintf("post record %v.runs %v", config.Config.DB.Name, string(queryData))
	_, err := connection.Query(queryString)

	if err != nil {
		return err.Error(), ""
	}

	go startRun(newRun)

	return constants.STATUS_CREATED, newRun.ID
}

func DeleteRunByID(id string) string {
	ids := []string{id}

	queryData, _ := json.Marshal(&ids)
	queryString := fmt.Sprintf("delete record %v.runs %v", config.Config.DB.Name, string(queryData))
	_, err := connection.Query(queryString)

	if err != nil {
		return err.Error()
	}

	return constants.STATUS_OK
}

func GetAllRuns(filter, limit, count, orderasc, orderdsc string) (string, []Run) {
	queryString := fmt.Sprintf("get record %v.runs", config.Config.DB.Name)

	if filter != constants.REQUEST_DEFAULT_FILTER {
		queryString += fmt.Sprintf(" | filter %v", filter)
	}
	if limit != constants.REQUEST_DEFAULT_LIMIT {
		queryString += fmt.Sprintf(" | limit %v", limit)
	}
	if count != constants.REQUEST_DEFAULT_COUNT {
		queryString += " | count"
	}
	if orderasc != constants.REQUEST_DEFAULT_ORDERDSC {
		queryString += fmt.Sprintf(" | orderasc %v", orderasc)
	}
	if orderdsc != constants.REQUEST_DEFAULT_COUNT {
		queryString += fmt.Sprintf(" | orderdsc %v", orderdsc)
	}

	data, err := connection.Query(queryString)
	if err != nil {
		return err.Error(), []Run{}
	}
	marshalled, _ := json.Marshal(data)
	var output []Run
	json.Unmarshal(marshalled, &output)

	return constants.STATUS_FOUND, output
}

func GetRunByID(runID string) (string, Run) {
	queryString := fmt.Sprintf("get record %v.runs | filter id = \"%s\"", config.Config.DB.Name, runID)

	data, err := connection.Query(queryString)
	if err != nil {
		return err.Error(), Run{}
	}
	marshalled, _ := json.Marshal(data)
	var output []Run
	json.Unmarshal(marshalled, &output)

	if len(output) == 0 {
		return constants.STATUS_NOT_FOUND, Run{}
	}

	return constants.STATUS_OK, output[0]
}

func UpdateRunByID(runID string, newRun Run) string {
	queryString := fmt.Sprintf("get record %v.runs | filter id = \"%s\"", config.Config.DB.Name, runID)
	datum, err := connection.Query(queryString)
	if err != nil {
		return err.Error()
	}

	if len(datum) == 0 {
		return constants.STATUS_NOT_FOUND
	}

	if newRun.Pipeline != "" {
		datum[0]["pipeline"] = newRun.Pipeline
	}
	if newRun.Status != "" {
		datum[0]["status"] = newRun.Status
	}
	if newRun.Started != "" {
		datum[0]["started"] = newRun.Started
	}
	if newRun.Finished != "" {
		datum[0]["finished"] = newRun.Finished
	}
	if len(newRun.Jobs) > 0 {
		datum[0]["jobs"] = newRun.Jobs
	}

	queryData, _ := json.Marshal(&datum)
	queryString = fmt.Sprintf("put record %v.runs %v", config.Config.DB.Name, string(queryData))
	_, err = connection.Query(queryString)
	if err != nil {
		return err.Error()
	}

	return constants.STATUS_OK
}

func startRun(r Run) {
	status, p := pipeline.GetPipelineByID(r.Pipeline)
	if status != constants.STATUS_OK {
		r.Error = fmt.Sprintf("count not find pipeline with id %s", r.Pipeline)
		r.Status = constants.RUN_STATUS_ERROR
		UpdateRunByID(r.ID, r)
		return
	}

	r.Status = constants.RUN_STATUS_RUNNING

	for _, jid := range p.Jobs {
		_, j := job.GetJobByID(jid)

		currentTime := time.Now().UTC()

		jobStatus := JobStatus{
			Status:  constants.JOB_STATUS_RUNNING,
			Steps:   make(map[string]StepStatus),
			Started: currentTime.Format("2006-01-02T15:04:05Z"),
		}
		r.Jobs[j.Name] = jobStatus
		UpdateRunByID(r.ID, r)

		for _, sid := range j.Steps {
			_, s := step.GetStepByID(sid)

			currentTime := time.Now().UTC()

			stepStatus := StepStatus{
				Started: currentTime.Format("2006-01-02T15:04:05Z"),
				Status:  constants.STEP_STATUS_RUNNING,
			}
			jobStatus.Steps[s.Name] = stepStatus
			r.Jobs[j.Name] = jobStatus
			UpdateRunByID(r.ID, r)

			switch s.Type {
			case constants.STEP_TYPE_DOCKER:
				err := os.MkdirAll(fmt.Sprintf("/tmp/run-%s", r.ID), 0755)
				if err != nil {
					r.Error = err.Error()
					r.Status = constants.RUN_STATUS_ERROR
					r.Jobs = map[string]JobStatus{}
					currentTime := time.Now().UTC()
					r.Finished = currentTime.Format("2006-01-02T15:04:05Z")
					UpdateRunByID(r.ID, r)
					return
				}

			case constants.STEP_TYPE_LITERAL:
				dirPath := fmt.Sprintf("/tmp/run-%s-%s", r.ID, s.ID)
				scriptPath := dirPath + "/.run.sh"
				outputPath := dirPath + "/.output"
				exitCodePath := dirPath + "/.exitcode"

				err := os.MkdirAll(dirPath, 0755)
				if err != nil {
					r.Error = err.Error()
					r.Status = constants.RUN_STATUS_ERROR
					r.Jobs = map[string]JobStatus{}
					currentTime := time.Now().UTC()
					r.Finished = currentTime.Format("2006-01-02T15:04:05Z")
					UpdateRunByID(r.ID, r)
					return
				}

				data := []byte(s.Contents)
				err = os.WriteFile(scriptPath, data, 0777)
				if err != nil {
					r.Error = err.Error()
					r.Status = constants.RUN_STATUS_ERROR
					r.Jobs = map[string]JobStatus{}
					currentTime := time.Now().UTC()
					r.Finished = currentTime.Format("2006-01-02T15:04:05Z")
					UpdateRunByID(r.ID, r)
					return
				}

				go exec.Command("bash", "-c", fmt.Sprintf("%s 2>&1 > %s; echo \"$?\" > %s", scriptPath, outputPath, exitCodePath)).Run()

				for {
					if _, err := os.Stat(exitCodePath); err == nil {
						rc, _ := os.ReadFile(exitCodePath)
						rcString := string(rc)
						if rcString[:len(rcString)-1] != "0" {
							stepStatus.Status = constants.STEP_STATUS_ERROR
							break
						}
						stepStatus.Status = constants.STEP_STATUS_SUCCESS
						break
					}
					out, _ := os.ReadFile(outputPath)
					stepStatus.Output = string(out)
					jobStatus.Steps[s.Name] = stepStatus
					r.Jobs[j.Name] = jobStatus
					UpdateRunByID(r.ID, r)

					time.Sleep(1 * time.Millisecond)
				}

				currentTime := time.Now().UTC()
				finishTime := currentTime.Format("2006-01-02T15:04:05Z")

				out, _ := os.ReadFile(outputPath)
				stepStatus.Output = string(out)
				stepStatus.Finished = finishTime
				jobStatus.Steps[s.Name] = stepStatus
				r.Jobs[j.Name] = jobStatus
			}
		}
		currentTime = time.Now().UTC()
		jobStatus.Finished = currentTime.Format("2006-01-02T15:04:05Z")
		didSucceed := true
		for _, step := range jobStatus.Steps {
			if step.Status != constants.STEP_STATUS_SUCCESS {
				jobStatus.Status = constants.JOB_STATUS_FAILED
				r.Status = constants.RUN_STATUS_FAILED
				didSucceed = false
			}
		}
		if didSucceed {
			jobStatus.Status = constants.JOB_STATUS_SUCCESS
			r.Status = constants.RUN_STATUS_SUCCESS
		}
		r.Jobs[j.Name] = jobStatus
		UpdateRunByID(r.ID, r)

		if r.Status == constants.RUN_STATUS_FAILED {
			break
		}
	}
}
