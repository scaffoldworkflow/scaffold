package api

import (
	"net/http"
	"scaffold/manager/history"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// @summary					Get a history
// @description				Get a run history by a run ID
// @tags						manager
// @tags						history
// @produce					json
// @success					200	{object}	history.History
// @failure					500	{object}	object
// @failure					401	{object}	object
// @securityDefinitions.apiKey	token
// @in							header
// @name						Authorization
// @security					X-Scaffold-API
// @router						/api/v1/history/{run_id} [get]
func GetHistory(ctx *gin.Context) {
	runID := ctx.Param("runID")
	h, err := history.GetHistoryByRunID(runID)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &h)
}

func GetAllHistories(ctx *gin.Context) {
	hs, err := history.GetAllHistories()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	historiesOut := make([]history.History, 0)
	for _, h := range hs {
		historiesOut = append(historiesOut, *h)
	}

	ctx.JSON(http.StatusOK, &historiesOut)
}

// func CreateHistory(ctx *gin.Context) {
// 	var h history.History
// 	if err := ctx.ShouldBindJSON(&h); err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}

// 	if err := (&h).Create(); err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}

// 	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
// }

// func DeleteHistories(ctx *gin.Context) {
// 	filter := buildQuery(ctx)

// 	if err := history.DeleteHistories(filter); err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
// }

func GetHistories(ctx *gin.Context) {
	filter := buildQuery(ctx)

	histories, err := history.GetHistories(filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	historiesOut := make([]history.History, 0)
	for _, h := range histories {
		historiesOut = append(historiesOut, *h)
	}

	ctx.JSON(http.StatusOK, historiesOut)
}

// func UpdateHistory(ctx *gin.Context) {
// 	var h history.History
// 	if err := ctx.ShouldBindJSON(&h); err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}

// 	if err := (&h).Update(); err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
// }
