package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/history"
	"scaffold/server/msg"
	"scaffold/server/rabbitmq"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/utils"
	"scaffold/server/workflow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"
)

//	@summary					Trigger a webhook
//	@description				Trigger a webhook with optional input data
//	@tags						manager
//	@tags						webhook
//	@accept						json
//	@produce					json
//	@Param						data	body		object	true	"Webhook input Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/webhook/{workflow_name}/{webhook_id} [post]
func TriggerWebhookByID(ctx *gin.Context) {
	wName := ctx.Param("workflow")
	tName := ctx.Param("task")

	var data map[string]string
	if err := ctx.ShouldBindJSON(&data); err != nil {
		logger.Warnf("", "No input data found for trigger on webhook %s", wName)
	}

	w, err := workflow.GetWorkflowByName(wName)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if w.Groups != nil {
		if !validateUserGroup(ctx, w.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	t, err := task.GetTaskByNames(wName, tName)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if t.Disabled {
		utils.Error(fmt.Errorf("task %s is disabled", t.Name), ctx, http.StatusServiceUnavailable)
		return
	}

	runID := uuid.New().String()

	m := msg.TriggerMsg{
		Task:     tName,
		Workflow: wName,
		Action:   constants.ACTION_TRIGGER,
		Groups:   w.Groups,
		Number:   t.RunNumber + 1,
		RunID:    runID,
		Context:  data,
	}

	h := history.History{
		RunID:    runID,
		States:   make([]state.State, 0),
		Workflow: wName,
	}

	if err := history.CreateHistory(&h); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	logger.Infof("", "Creating run with message %v", m)
	rabbitmq.ManagerPublish(m)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK", "run_id": runID})
}
