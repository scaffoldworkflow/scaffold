package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/state"
	"scaffold/server/utils"
	"scaffold/server/workflow"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a state
//	@description				Create a state from a JSON object
//	@tags						manager
//	@tags						state
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
//	@router						/api/v1/state [post]
func CreateState(ctx *gin.Context) {
	var s state.State
	if err := ctx.ShouldBindJSON(&s); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	c, err := workflow.GetWorkflowByName(s.Workflow)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	err = state.CreateState(&s)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

//	@summary					Delete a state
//	@description				Delete a state by its name and its workflow
//	@tags						manager
//	@tags						state
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/state/{workflow_name}/{state_name} [delete]
func DeleteStateByNames(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	err := state.DeleteStateByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Delete states
//	@description				Delete states by their workflow
//	@tags						manager
//	@tags						state
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/state/{workflow_name} [delete]
func DeleteStatesByWorkflow(ctx *gin.Context) {
	cn := ctx.Param("workflow")

	err := state.DeleteStatesByWorkflow(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all states
//	@description				Get all states
//	@tags						manager
//	@tags						state
//	@produce					json
//	@success					200	{array}		state.State
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/state [get]
func GetAllStates(ctx *gin.Context) {
	states, err := state.GetAllStates()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	statesOut := make([]state.State, 0)
	for _, s := range states {
		c, err := workflow.GetWorkflowByName(s.Workflow)
		if err != nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				statesOut = append(statesOut, *s)
			}
			continue
		}
		statesOut = append(statesOut, *s)
	}

	ctx.JSON(http.StatusOK, statesOut)
}

//	@summary					Get a state
//	@description				Get a state by its name and its workflow
//	@tags						manager
//	@tags						state
//	@produce					json
//	@success					200	{object}	state.State
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/state/{workflow_name}/{state_name} [get]
func GetStateByNames(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	s, err := state.GetStateByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if s == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("State %s/%s does not exist", cn, tn)})
		return
	}

	ctx.JSON(http.StatusOK, *s)
}

//	@summary					Get states
//	@description				Get states by their workflow
//	@tags						manager
//	@tags						state
//	@produce					json
//	@success					200	{array}		state.State
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/state/{workflow_name} [get]
func GetStatesByWorkflow(ctx *gin.Context) {
	cn := ctx.Param("workflow")

	s, err := state.GetStatesByWorkflow(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, s)
}

//	@summary					Update a state
//	@description				Update a state from a JSON object
//	@tags						manager
//	@tags						state
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
//	@router						/api/v1/state/{workflow_name}/{state_name} [put]
func UpdateStateByNames(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	var s state.State
	if err := ctx.ShouldBindJSON(&s); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := state.UpdateStateByNames(cn, tn, &s)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
