package api

import (
	"net/http"
	"scaffold/manager/history"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
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

// func GetHistories(ctx *gin.Context) {
// 	filter := buildQuery(ctx)

// 	histories, err := history.GetHistories(filter)

// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			ctx.JSON(http.StatusNoContent, []interface{}{})
// 			return
// 		}
// 		logger.Errorf("", "Got error trying to get histories with filter %s: %s", filter, err)
// 		ctx.AbortWithError(http.StatusInternalServerError, err)
// 		return
// 	}

// 	// Need to copy each history from pointer to value since pointers are returned
// 	// weirdly (I think at least)
// 	releasesOut := make([]history.History, len(histories))
// 	for _, h := range histories {
// 		historiesOut = append(releasesOut, *r)
// 	}

// 	ctx.JSON(http.StatusOK, releasesOut)
// }
