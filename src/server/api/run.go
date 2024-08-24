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
//	@success					200 {object} object
//	@failure					500 {object} object
//	@failure					401 {object} object
//	@produce 					json
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/{workflow_name}/{task_name}/{task_number} [delete]
func ManagerKillRun(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")
	// nn, err := strconv.Atoi(ctx.Param("number"))
	// if err != nil {
	// 	utils.Error(err, ctx, http.StatusInternalServerError)
	// }

	// logger.Debugf("", "Looking for %s/%s.%d", cn, tn, nn)
	// s, err := state.GetStateByNamesNumber(cn, tn, nn)
	// if err != nil {
	// 	utils.Error(err, ctx, http.StatusInternalServerError)
	// }

	// nd := auth.Nodes[s.Worker]

	// httpClient := http.Client{}
	// requestURL := fmt.Sprintf("%s://%s:%d/api/v1/kill/%s/%s/%d", nd.Protocol, nd.Host, nd.Port, cn, tn, nn)
	// req, _ := http.NewRequest("DELETE", requestURL, nil)
	// req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	// resp, err := httpClient.Do(req)
	// if err != nil {
	// 	utils.Error(err, ctx, http.StatusInternalServerError)
	// }
	// if resp.StatusCode >= 400 {
	// 	utils.Error(fmt.Errorf("received kill status code %d", resp.StatusCode), ctx, resp.StatusCode)
	// 	return
	// }

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
//	@success					201			{object}	object
//	@failure					500			{object}	object
//	@failure					401			{object}	object
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
	// bulwark.QueuePush(bulwark.WorkerClient, m)
	rabbitmq.ManagerPublish(m)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
