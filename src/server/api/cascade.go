package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/utils"
	"scaffold/server/workflow"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a workflow
//	@description				Create a workflow from a JSON object
//	@tags						manager
//	@tags						workflow
//	@accept						json
//	@produce					json
//	@Param						workflow	body		workflow.Workflow	true	"Workflow Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/workflow [post]
func CreateWorkflow(ctx *gin.Context) {
	var c workflow.Workflow
	if err := ctx.ShouldBindJSON(&c); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	err := workflow.CreateWorkflow(&c)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

//	@summary					Delete a workflow
//	@description				Delete a workflow by its name
//	@tags						manager
//	@tags						workflow
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/workflow/{workflow_name} [delete]
func DeleteWorkflowByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := workflow.DeleteWorkflowByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all workflows
//	@description				Get all workflows
//	@tags						manager
//	@tags						workflow
//	@produce					json
//	@success					200	{array}		workflow.Workflow
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/workflow [get]
func GetAllWorkflows(ctx *gin.Context) {
	workflows, err := workflow.GetAllWorkflows()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// Need to copy each workflow from pointer to value since pointers are returned
	// weirdly (I think at least)
	workflowsOut := make([]workflow.Workflow, 0)
	for _, c := range workflows {
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				workflowsOut = append(workflowsOut, *c)
			}
			continue
		}
		workflowsOut = append(workflowsOut, *c)
	}

	ctx.JSON(http.StatusOK, workflowsOut)
}

//	@summary					Get a workflow
//	@description				Get a workflow by its name
//	@tags						manager
//	@tags						workflow
//	@produce					json
//	@success					200	{object}	workflow.Workflow
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/workflow/{workflow_name} [get]
func GetWorkflowByName(ctx *gin.Context) {
	name := ctx.Param("name")

	c, err := workflow.GetWorkflowByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if c == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Workflow %s does not exist", name)})
		return
	}

	ctx.JSON(http.StatusOK, *c)
}

//	@summary					Update a workflow
//	@description				Update a workflow from a JSON object
//	@tags						manager
//	@tags						workflow
//	@accept						json
//	@produce					json
//	@Param						workflow	body		workflow.Workflow	true	"Workflow Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/workflow/{workflow_name} [put]
func UpdateWorkflowByName(ctx *gin.Context) {
	name := ctx.Param("name")

	var c workflow.Workflow
	if err := ctx.ShouldBindJSON(&c); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := workflow.UpdateWorkflowByName(name, &c)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
