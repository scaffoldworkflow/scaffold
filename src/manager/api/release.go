package api

import (
	"fmt"
	"net/http"
	"scaffold/manager/release"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReleaseDef struct {
	Services []string               `json:"services"`
	Context  map[string]interface{} `json:"context"`
	Assets   []release.Asset        `json:"assets"`
}

func CreateRelease(ctx *gin.Context) {
	var data ReleaseDef
	if err := ctx.ShouldBindJSON(&data); err != nil {
		logger.Errorf("", "Could not bind release JSON: %s", err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	t := ctx.Param("team")
	p := ctx.Param("project")

	r := &release.Release{
		Assets:   data.Assets,
		Project:  p,
		Team:     t,
		Services: data.Services,
	}
	if err := r.Create(); err != nil {
		logger.Errorf("", "Could not create release for %s/%s", t, p)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": r.ID})
}

func DeleteReleases(ctx *gin.Context) {
	filter := buildQuery(ctx)

	if err := release.DeleteReleases(filter); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetReleases(ctx *gin.Context) {
	filter := buildQuery(ctx)

	releases, err := release.GetReleases(filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		logger.Errorf("", "Got error trying to get releases with filter %s: %s", filter, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Need to copy each workflow from pointer to value since pointers are returned
	// weirdly (I think at least)
	releasesOut := make([]release.Release, len(releases))
	for _, r := range releases {
		releasesOut = append(releasesOut, *r)
	}

	ctx.JSON(http.StatusOK, releasesOut)
}

func UpdateRelease(ctx *gin.Context) {
	var r release.Release
	if err := ctx.ShouldBindJSON(&r); err != nil {
		logger.Errorf("", "Could not bind release JSON: %s", err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := (&r).Update(); err != nil {
		logger.Errorf("", "Could not update release: %s", err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func PromoteRelease(ctx *gin.Context) {
	rID := ctx.Param("release_id")
	rs, err := release.GetReleases(bson.M{"id": rID})
	if err != nil {
		logger.Errorf("", "Could not get release for ID %s: %s", rID, err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if len(rs) == 0 {
		logger.Warnf("", "Could not get release for ID %s: not found", rID)
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("not found"))
	}
	r := rs[0]

	var pm release.PromoteMessage
	if err := ctx.ShouldBindJSON(&pm); err != nil {
		logger.Errorf("", "Could not bind promotion JSON: %s", err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	runID, err := r.Promote(pm)
	if err != nil {
		logger.Errorf("", "Unable to promote release %s to environment %s: %s", r.ID, pm.Environment, err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.PromoteRuns = append(r.PromoteRuns, runID)

	if err := r.Update(); err != nil {
		logger.Errorf("", "Could not update release %s: %s", r.ID, err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"run_id": runID})
}
