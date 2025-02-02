package api

import (
	"net/http"
	"scaffold/manager/release"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
func GetRelease(ctx *gin.Context) {
	runID := ctx.Param("runID")
	r, err := release.GetReleases(bson.M{"id": runID})
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &r)
}

func GetAllReleases(ctx *gin.Context) {
	rs, err := release.GetReleases(bson.M{})
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	releasesOut := make([]release.Release, 0)
	for _, r := range rs {
		releasesOut = append(releasesOut, *r)
	}

	ctx.JSON(http.StatusOK, &releasesOut)
}

type ReleaseDef struct {
	Context map[string]interface{} `json:"context"`
	Assets []release.Asset `json:"assets"`
}

func Create(ctx *gin.Context) {
	var data ReleaseDef
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	params := ctx.Request.URL.Query()
	r := &release.Release{
		Assets: data.Assets
		Project      string   `json:"project" bson:"project"`
		Team         string   `json:"team" bson:"team"`
		Created      string   `json:"created" bson:"created"`
		Updated      string   `json:"updated" bson:"updated"`
		Services     []string `json:"services" bson:"services"`
	}
}
