package api

import (
	"errors"
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/input"
	"scaffold/server/manager"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a input
//	@description				Create a input from a JSON object
//	@tags						manager
//	@tags						input
//	@accept						json
//	@produce					json
//	@Param						input	body		input.Input	true	"Input Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input [post]
func CreateInput(ctx *gin.Context) {
	var i input.Input
	if err := ctx.ShouldBindJSON(&i); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	c, err := cascade.GetCascadeByName(i.Cascade)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	err = input.CreateInput(&i)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

//	@summary					Delete a input
//	@description				Delete a input by its name and its cascade
//	@tags						manager
//	@tags						input
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input/{input_name} [delete]
func DeleteInputByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	n := ctx.Param("name")

	err := input.DeleteInputByNames(cn, n)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Delete inputs
//	@description				Delete inputs by their cascade
//	@tags						manager
//	@tags						input
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input/{cascade_name} [delete]
func DeleteInputsByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	err := input.DeleteInputsByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all inputs
//	@description				Get all inputs
//	@tags						manager
//	@tags						input
//	@produce					json
//	@success					200	{array}		input.Input
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input [get]
func GetAllInputs(ctx *gin.Context) {
	inputs, err := input.GetAllInputs()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	inputsOut := make([]input.Input, 0)
	for _, i := range inputs {
		c, err := cascade.GetCascadeByName(i.Cascade)
		if err == nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				inputsOut = append(inputsOut, *i)
			}
			continue
		}
		inputsOut = append(inputsOut, *i)
	}

	ctx.JSON(http.StatusOK, inputsOut)
}

//	@summary					Get a input
//	@description				Get a input by its name and its cascade
//	@tags						manager
//	@tags						input
//	@produce					json
//	@success					200	{object}	input.Input
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input/{cascade_name}/{input_name} [get]
func GetInputByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	n := ctx.Param("name")

	i, err := input.GetInputByNames(cn, n)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if i == nil {
		ctx.JSON(http.StatusOK, input.Input{})
		return
	}

	ctx.JSON(http.StatusOK, *i)
}

//	@summary					Get inputs
//	@description				Get inputs by their cascade
//	@tags						manager
//	@tags						input
//	@produce					json
//	@success					200	{array}		input.Input
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input/{cascade_name} [get]
func GetInputsByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	i, err := input.GetInputsByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, i)
}

//	@summary					Update a input
//	@description				Update a input from a JSON object
//	@tags						manager
//	@tags						input
//	@accept						json
//	@produce					json
//	@Param						input	body		input.Input	true	"Input Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input/{cascade_name}/{input_name} [put]
func UpdateInputByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	n := ctx.Param("name")

	var i input.Input
	if err := ctx.ShouldBindJSON(&i); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := input.UpdateInputByNames(cn, n, &i)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Trigger update to dependent tasks
//	@description				Trigger updates of states for dependent tasks
//	@tags						manager
//	@tags						input
//	@produce					json
//	@success					200		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/input/{cascade_name}/update [post]
func UpdateInputDependenciesByName(ctx *gin.Context) {
	name := ctx.Param("cascade")

	var changed []string
	if err := ctx.ShouldBindJSON(&changed); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := manager.InputChangeStateChange(name, changed)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
