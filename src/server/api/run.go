package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/history"
	"scaffold/server/manager"
	"scaffold/server/msg"
	"scaffold/server/rabbitmq"
	"scaffold/server/run"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/utils"
	"scaffold/server/workflow"

	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
)

//	@summary					Kill a run
//	@description				Instruct a manager to kill a run
//	@tags						manager
//	@tags						run
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@produce					json
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/{workflow_name}/{task_name}/{task_number} [delete]
func ManagerKillRun(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	logger.Infof("", "Triggering run kill for %s.%s", cn, tn)
	manager.DoKill(cn, tn)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
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
//	@router						/api/v1/kill/{workflow_name}/{task_name} [delete]
func KillRun(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	c, err := workflow.GetWorkflowByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if t.Kind == constants.TASK_KIND_CONTAINER {
		if err := run.ContainerKill(cn, tn); err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
	} else if t.Kind == constants.TASK_KIND_LOCAL {
		if err := run.LocalKill(cn, tn); err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

//	@summary					Create a run
//	@description				Create a run from a workflow and task
//	@tags						manager
//	@tags						run
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/{workflow_name}/{task_name} [post]
func CreateRun(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	c, err := workflow.GetWorkflowByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	s, err := state.GetStateByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if t.Disabled {
		utils.Error(fmt.Errorf("task %s is disabled", tn), ctx, http.StatusServiceUnavailable)
		return
	}

	runID := uuid.New().String()

	m := msg.TriggerMsg{
		Task:     tn,
		Workflow: cn,
		Action:   constants.ACTION_TRIGGER,
		Groups:   c.Groups,
		Number:   t.RunNumber + 1,
		RunID:    runID,
		Context:  s.Context,
	}

	h := history.History{
		RunID:    runID,
		States:   make([]state.State, 0),
		Workflow: cn,
	}

	if err := history.CreateHistory(&h); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	logger.Infof("", "Creating run with message %v", m)
	rabbitmq.ManagerPublish(m)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get run status
//	@description				Get status of a run by ID
//	@tags						manager
//	@tags						run
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/{run_id} [get]
func GetRunStatus(ctx *gin.Context) {
	runID := ctx.Param("runID")
	h, err := history.GetHistoryByRunID(runID)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	running := false
	errored := false
	waiting := false
	killed := false
	success := false

	s := h.States[len(h.States)-1]
	t := s.Task

	ss, err := state.GetStateByNames(h.Workflow, t)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	switch ss.Status {
	case constants.STATE_STATUS_RUNNING:
		running = true
	case constants.STATE_STATUS_ERROR:
		errored = true
	case constants.STATE_STATUS_WAITING, constants.STATE_STATUS_NOT_STARTED:
		waiting = true
	case constants.STATE_STATUS_KILLED:
		killed = true
	case constants.STATE_STATUS_SUCCESS:
		success = true
	}

	ctx.JSON(http.StatusOK, gin.H{
		"running": running,
		"errored": errored,
		"waiting": waiting,
		"killed":  killed,
		"success": success,
		"task":    t,
	})
}
