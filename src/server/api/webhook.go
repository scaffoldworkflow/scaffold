package api

import (
	"fmt"
	"net/http"
	"scaffold/server/bulwark"
	"scaffold/server/cascade"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/input"
	"scaffold/server/msg"
	"scaffold/server/task"
	"scaffold/server/utils"
	"scaffold/server/webhook"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a webhook
//	@description				Create a webhook from a JSON object
//	@tags						manager
//	@tags						webhook
//	@accept						json
//	@produce					json
//	@Param						webhook	body		webhook.Webhook	true	"Webhook Data"
//	@success					201			{object}	object
//	@failure					500			{object}	object
//	@failure					401			{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/webhook [post]
func CreateWebhook(ctx *gin.Context) {
	var w webhook.Webhook
	if err := ctx.ShouldBindJSON(&w); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := webhook.CreateWebhook(&w); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

//	@summary					Get a webhook
//	@description				Get a webhook by its ID
//	@tags						manager
//	@tags						webhook
//	@produce					json
//	@success					200	{object}	webhook.Webhook
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/webhook/{webhook_id} [get]
func GetWebhookByID(ctx *gin.Context) {
	id := ctx.Param("id")

	w, err := webhook.GetWebhookByID(id)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if w == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Webhook %s does not exist", id)})
		return
	}

	ctx.JSON(http.StatusOK, *w)
}

//	@summary					Delete a webhook
//	@description				Delete a webhook by its ID
//	@tags						manager
//	@tags						webhook
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/webhook/{webhook_id} [delete]
func DeleteWebhookByID(ctx *gin.Context) {
	id := ctx.Param("id")

	err := webhook.DeleteWebhookByID(id)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all webhooks
//	@description				Get all webhooks
//	@tags						manager
//	@tags						webhook
//	@produce					json
//	@success					200	{array}		state.State
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/webhook [get]
func GetAllWebhooks(ctx *gin.Context) {
	webhooks, err := webhook.GetAllWebhooks()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, webhooks)
}

//	@summary					Update a webhook
//	@description				Update a webhook from a JSON object
//	@tags						manager
//	@tags						webhook
//	@accept						json
//	@produce					json
//	@Param						state	body		state.State	true	"State Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/webhook/{webhook_id} [put]
func UpdateWebhooksByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var w webhook.Webhook
	if err := ctx.ShouldBindJSON(&w); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := webhook.UpdateWebhooksByID(id, &w)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

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
//	@router						/api/v1/webhook/{cascade_name}/{webhook_id} [post]
func TriggerWebhookByID(ctx *gin.Context) {
	id := ctx.Param("id")

	w, err := webhook.GetWebhookByID(id)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	var data map[string]string
	if err := ctx.ShouldBindJSON(&data); err != nil {
		logger.Warnf("", "No input data found for trigger on webhook %s", id)

	} else {

		d, err := datastore.GetDataStoreByCascade(w.Cascade)
		if err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}

		for name, datum := range data {
			for envName := range d.Env {
				if envName != name {
					continue
				}
				d.Env[name] = datum
			}
		}

		inputs := []input.Input{}
		err = datastore.UpdateDataStoreByCascade(w.Cascade, d, inputs)
		if err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
	}

	c, err := cascade.GetCascadeByName(w.Cascade)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	t, err := task.GetTaskByNames(w.Cascade, w.Entrypoint)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if t.Disabled {
		utils.Error(fmt.Errorf("task %s is disabled", w.Entrypoint), ctx, http.StatusServiceUnavailable)
		return
	}

	m := msg.TriggerMsg{
		Task:    w.Entrypoint,
		Cascade: w.Cascade,
		Action:  constants.ACTION_TRIGGER,
		Groups:  c.Groups,
		Number:  t.RunNumber + 1,
	}

	logger.Infof("", "Creating run with message %v", m)
	bulwark.QueuePush(bulwark.WorkerClient, m)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
