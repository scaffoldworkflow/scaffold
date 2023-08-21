package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/container"
	"scaffold/server/logger"
	"scaffold/server/manager"
	"scaffold/server/run"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/utils"
	"scaffold/server/worker"
	"strings"

	"github.com/gin-gonic/gin"
)

//	@summary					Create a run
//	@description				Create a run to be deployed to a worker
//	@tags						worker
//	@tags						run
//	@produce					json
//	@success					200 {object} object
//	@failure					500 {object} object
//	@failure					401 {object} object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/{cascade_name}/{task_name} [post]
func CreateRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	s, err := state.GetStateByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	t.RunNumber += 1

	previousName := fmt.Sprintf("SCAFFOLD_PREVIOUS-%s", t.Name)
	ps := *s
	ps.Task = previousName
	err = state.UpdateStateByNames(cn, previousName, &ps)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	s.Number += 1

	s.Status = constants.STATE_STATUS_WAITING

	obj := run.Run{
		Name:   fmt.Sprintf("%s.%s.%d", cn, tn, t.RunNumber),
		Task:   *t,
		State:  *s,
		Number: t.RunNumber,
		Groups: c.Groups,
	}
	postBody, _ := json.Marshal(obj)
	postBodyBuffer := bytes.NewBuffer(postBody)

	n, err := getAvailableNode()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s://%s:%d/api/v1/trigger", n.Protocol, n.Host, n.Port)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode >= 400 {
		utils.Error(fmt.Errorf("received trigger status code %d", resp.StatusCode), ctx, resp.StatusCode)
	}
	if _, ok := manager.InProgress[cn]; !ok {
		manager.InProgress[cn] = map[string]string{fmt.Sprintf("%s.%d", tn, t.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
		manager.ToCheck[cn] = map[string]string{fmt.Sprintf("%s.%d", tn, t.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
	} else {
		manager.InProgress[cn][fmt.Sprintf("%s.%d", tn, t.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
		manager.ToCheck[cn][fmt.Sprintf("%s.%d", tn, t.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
	}

	if err := task.UpdateTaskByNames(cn, tn, t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// if err := state.UpdateStateByNames(cn, tn, s); err != nil {
	// 	utils.Error(err, ctx, http.StatusInternalServerError)
	// 	return
	// }

	cs, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	manager.SetDependsState(cs, tn)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Create a check run
//	@description				Create a check run to be deployed to a worker
//	@tags						worker
//	@tags						run
//	@produce					json
//	@success					200 {object} object
//	@failure					500 {object} object
//	@failure					401 {object} object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/{cascade_name}/{task_name}/check [post]
func CreateCheckRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	t.Check.RunNumber += 1

	checkStateName := fmt.Sprintf("SCAFFOLD_CHECK-%s", tn)

	s, err := state.GetStateByNames(cn, checkStateName)
	if err != nil {
		logger.Infof("", "No existing state for %s/%s found", cn, checkStateName)
		s = &state.State{
			Task:     checkStateName,
			Cascade:  cn,
			Status:   constants.STATE_STATUS_WAITING,
			Started:  "",
			Finished: "",
			Output:   "",
			Number:   t.RunNumber,
			Display:  make([]map[string]interface{}, 0),
		}
		if err := state.CreateState(s); err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
	}
	s.Number = t.RunNumber

	obj := run.Run{
		Name: fmt.Sprintf("%s.%s.%d", cn, checkStateName, t.Check.RunNumber),
		Task: task.Task{
			Name:        checkStateName,
			Cascade:     t.Cascade,
			Verb:        "",
			DependsOn:   task.TaskDependsOn{},
			Image:       t.Check.Image,
			Run:         t.Check.Run,
			Store:       t.Check.Store,
			Load:        t.Check.Load,
			Env:         t.Check.Env,
			Inputs:      t.Check.Inputs,
			Updated:     t.Check.Updated,
			AutoExecute: true,
			ShouldRM:    true,
			RunNumber:   t.RunNumber,
		},
		State:  *s,
		Number: t.RunNumber,
		Groups: c.Groups,
	}
	postBody, _ := json.Marshal(obj)
	postBodyBuffer := bytes.NewBuffer(postBody)

	n, err := getAvailableNode()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s://%s:%d/api/v1/trigger", n.Protocol, n.Host, n.Port)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Errorf("", "Received trigger status code %d", resp.StatusCode)
	}
	if _, ok := manager.InProgress[cn]; !ok {
		manager.InProgress[cn] = map[string]string{fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
		manager.ToCheck[cn] = map[string]string{fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
	} else {
		manager.InProgress[cn][fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
		manager.ToCheck[cn][fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
	}

	if err := task.UpdateTaskByNames(cn, tn, t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// if err := state.UpdateStateByNames(cn, tn, s); err != nil {
	// 	utils.Error(err, ctx, http.StatusInternalServerError)
	// 	return
	// }

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Kill a run
//	@description				Instruct a manager to kill a run
//	@tags						manager
//	@tags						run
//	@success					200 {object} object
//	@failure					500 {object} object
//	@failure					401 {object} object
//	@produce 					json
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/{cascade_name}/{task_name}/{task_number} [delete]
func ManagerKillRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")
	nn := ctx.Param("number")

	key := fmt.Sprintf("%s.%s", tn, nn)

	uri := ""
	logger.Debugf("", "Looking for %s/%s.%s", cn, tn, nn)
	logger.Debugf("", "Kill in progress: %s", manager.InProgress)
	if _, ok := manager.InProgress[cn]; ok {
		if val, ok := manager.InProgress[cn][key]; ok {
			uri = val
		}
	}

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/kill/%s/%s/%s", uri, cn, tn, nn)
	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
	}
	if resp.StatusCode >= 400 {
		utils.Error(fmt.Errorf("received kill status code %d", resp.StatusCode), ctx, resp.StatusCode)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Trigger a run
//	@description				Trigger a run on a worker
//	@tags						worker
//	@tags						run
//	@accept						json
//	@Param						user	body		run.Run	true	"Run Data"
//	@success					201
//	@failure					500
//	@failure					401
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/trigger [post]
func TriggerRun(ctx *gin.Context) {
	var r run.Run
	if err := ctx.ShouldBindJSON(&r); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if r.Groups != nil {
		if !validateUserGroup(ctx, r.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
			return
		}
	}

	names := strings.Split(r.Name, ".")

	r.State = state.State{
		Task:     names[1],
		Cascade:  names[0],
		Status:   constants.STATE_STATUS_WAITING,
		Started:  "",
		Finished: "",
		Output:   "",
		Number:   r.Number,
		Display:  make([]map[string]interface{}, 0),
	}

	logger.Debugf("", "Writing new run to queue %v", r)

	worker.RunQueue = append(worker.RunQueue, r)

	ctx.Status(http.StatusCreated)
}

//	@summary					Kill a run
//	@description				Kill a run on a worker
//	@tags						worker
//	@tags						run
//	@success					200
//	@failure					500
//	@failure					401
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/kill/{cascade_name}/{task_name}/{task_number} [delete]
func KillRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")
	nn := ctx.Param("number")

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	if err := run.Kill(cn, tn, nn); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

//	@summary					Get run state
//	@description				Get state of a run being executed on a worker
//	@tags						worker
//	@tags						run
//	@tags						state
//	@produce					json
//	@success					200	{object}	state.State
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@failure					404	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/state/{cascade_name}/{task_name}/{task_number} [get]
func GetRunState(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")
	n := ctx.Param("number")

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	runName := fmt.Sprintf("%s.%s.%s", cn, tn, n)
	if container.CurrentRun.Name == runName {
		logger.Debugf("", "Run %s is currently running", runName)
		ctx.JSON(http.StatusOK, container.CurrentRun.State)
		return
	}
	for _, r := range worker.RunQueue {
		if r.Name == runName {
			logger.Debugf("", "Run %s is waiting in queue", runName)
			ctx.JSON(http.StatusOK, r.State)
			return
		}
	}
	if r, ok := container.CompletedRuns[runName]; ok {
		logger.Debugf("", "Run %s is completed", runName)
		ctx.JSON(http.StatusOK, r.State)
		delete(container.CompletedRuns, runName)
		return
	}
	ctx.Status(http.StatusNotFound)
}
