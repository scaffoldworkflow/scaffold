package api

import (
	"net/http"
	"scaffold/manager/project"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// @summary					Create a workflow
// @description				Create a workflow from a JSON object
// @tags						manager
// @tags						workflow
// @accept						json
// @produce					json
// @Param						workflow	body		project.Project	true	"Project Data"
// @success					201			{object}	object
// @failure					500			{object}	object
// @failure					401			{object}	object
// @securityDefinitions.apiKey	token
// @in							header
// @name						Authorization
// @security					X-Scaffold-API
// @router						/api/v1/workflow [post]
func CreateProject(ctx *gin.Context) {
	var p project.Project
	if err := ctx.ShouldBindJSON(&p); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := (&p).Create(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

// @summary					Delete a workflow
// @description				Delete a workflow by its id
// @tags						manager
// @tags						workflow
// @produce					json
// @success					200	{object}	object
// @failure					500	{object}	object
// @failure					401	{object}	object
// @securityDefinitions.apiKey	token
// @in							header
// @name						Authorization
// @security					X-Scaffold-API
// @router						/api/v1/workflow/{workflow_id} [delete]
func DeleteProjects(ctx *gin.Context) {
	filter := buildQuery(ctx)

	if err := project.DeleteProjects(filter); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// @summary					Get all workflows
// @description				Get all workflows
// @tags						manager
// @tags						workflow
// @produce					json
// @success					200	{array}		project.Project
// @failure					500	{object}	object
// @failure					401	{object}	object
// @securityDefinitions.apiKey	token
// @in							header
// @id						Authorization
// @security					X-Scaffold-API
// @router						/api/v1/workflow [get]
func GetProjects(ctx *gin.Context) {
	filter := buildQuery(ctx)

	projects, err := project.GetProjects(filter)

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
	projectsOut := make([]project.Project, 0)
	for _, p := range projects {
		projectsOut = append(projectsOut, *p)
	}

	ctx.JSON(http.StatusOK, projectsOut)
}

// @summary					Update a workflow
// @description				Update a workflow from a JSON object
// @tags						manager
// @tags						workflow
// @accept						json
// @produce					json
// @Param						workflow	body		project.Project	true	"Project Data"
// @success					201			{object}	object
// @failure					500			{object}	object
// @failure					401			{object}	object
// @securityDefinitions.apiKey	token
// @in							header
// @name						Authorization
// @security					X-Scaffold-API
// @router						/api/v1/workflow/{workflow_name} [put]
func UpdateProject(ctx *gin.Context) {
	var p project.Project
	if err := ctx.ShouldBindJSON(&p); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := (&p).Update(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
