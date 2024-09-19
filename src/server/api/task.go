package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/utils"
	"scaffold/server/workflow"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a task
//	@description				Create a task from a JSON object
//	@tags						manager
//	@tags						task
//	@accept						json
//	@produce					json
//	@Param						task	body		task.Task	true	"Task Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task [post]
func CreateTask(ctx *gin.Context) {
	var t task.Task
	if err := ctx.ShouldBindJSON(&t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	c, err := workflow.GetWorkflowByName(t.Workflow)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	err = task.CreateTask(&t)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

//	@summary					Delete a task
//	@description				Delete a task by its name and its workflow
//	@tags						manager
//	@tags						task
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task/{task_name} [delete]
func DeleteTaskByNames(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	err := task.DeleteTaskByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Delete tasks
//	@description				Delete tasks by their workflow
//	@tags						manager
//	@tags						task
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task/{workflow_name} [delete]
func DeleteTasksByWorkflow(ctx *gin.Context) {
	cn := ctx.Param("workflow")

	err := task.DeleteTasksByWorkflow(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all tasks
//	@description				Get all tasks
//	@tags						manager
//	@tags						task
//	@produce					json
//	@success					200	{array}		task.Task
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task [get]
func GetAllTasks(ctx *gin.Context) {
	tasks, err := task.GetAllTasks()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	tasksOut := make([]task.Task, 0)
	for _, t := range tasks {
		c, err := workflow.GetWorkflowByName(t.Workflow)
		if err != nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				tasksOut = append(tasksOut, *t)
			}
			continue
		}
		tasksOut = append(tasksOut, *t)
	}

	ctx.JSON(http.StatusOK, tasksOut)
}

//	@summary					Get a task
//	@description				Get a task by its name and its workflow
//	@tags						manager
//	@tags						task
//	@produce					json
//	@success					200	{object}	task.Task
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task/{workflow_name}/{task_name} [get]
func GetTaskByNames(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	t, err := task.GetTaskByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if t == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Task %s/%s does not exist", cn, tn)})
		return
	}

	ctx.JSON(http.StatusOK, *t)
}

//	@summary					Get tasks
//	@description				Get tasks by their workflow
//	@tags						manager
//	@tags						task
//	@produce					json
//	@success					200	{array}		task.Task
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task/{workflow_name} [get]
func GetTasksByWorkflow(ctx *gin.Context) {
	cn := ctx.Param("workflow")

	t, err := task.GetTasksByWorkflow(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, t)
}

//	@summary					Update a task
//	@description				Update a task from a JSON object
//	@tags						manager
//	@tags						task
//	@accept						json
//	@produce					json
//	@Param						task	body		task.Task	true	"Task Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task/{workflow_name}/{task_name} [put]
func UpdateTaskByNames(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	var t task.Task
	if err := ctx.ShouldBindJSON(&t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := task.UpdateTaskByNames(cn, tn, &t)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Update a task
//	@description				Update a task from a JSON object
//	@tags						manager
//	@tags						task
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/task/{workflow_name}/{task_name}/enabled [put]
func ToggleTaskEnabled(ctx *gin.Context) {
	cn := ctx.Param("workflow")
	tn := ctx.Param("task")

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	t.Disabled = !t.Disabled
	if err := task.UpdateTaskByNames(cn, tn, t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	s, err := state.GetStateByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	s.Disabled = t.Disabled
	if err := state.UpdateStateByNames(cn, tn, s); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
