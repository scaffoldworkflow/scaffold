package api

import (
	"net/http"
	"scaffold/manager/project"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateEnvironment(ctx *gin.Context) {
	var e project.Environment
	if err := ctx.ShouldBindJSON(&e); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := (&e).Create(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func DeleteEnvironments(ctx *gin.Context) {
	filter := buildQuery(ctx)

	if err := project.DeleteEnvironments(filter); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetEnvironments(ctx *gin.Context) {
	filter := buildQuery(ctx)

	environments, err := project.GetEnvironments(filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	environmentsOut := make([]project.Environment, 0)
	for _, e := range environments {
		environmentsOut = append(environmentsOut, *e)
	}

	ctx.JSON(http.StatusOK, environmentsOut)
}

func UpdateEnvironment(ctx *gin.Context) {
	var e project.Environment
	if err := ctx.ShouldBindJSON(&e); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := (&e).Update(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
