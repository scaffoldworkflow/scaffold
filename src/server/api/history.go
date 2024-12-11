package api

import (
	"net/http"
	"scaffold/server/history"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
)

//	@summary					Get a history
//	@description				Get a run history by a run ID
//	@tags						manager
//	@tags						history
//	@produce					json
//	@success					200	{object}	history.History
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/history/{run_id} [get]
func GetHistory(ctx *gin.Context) {
	runID := ctx.Param("runID")
	h, err := history.GetHistoryByRunID(runID)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &h)
}
