package api

import (
	"net/http"
	"scaffold/manager/history"
	"scaffold/manager/monitor"
	"scaffold/manager/project"
	"scaffold/manager/run"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/bson"
)

type ExecutionMessage struct {
	Name    string `json:"name"`
	Context string `json:"context"`
}

func KillRun(ctx *gin.Context) {
	runID := ctx.Param("run_id")
	step := ctx.Param("step")
	if err := run.KillRun(runID, step); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func StartRun(ctx *gin.Context) {
	tName := ctx.Param("team")
	pName := ctx.Param("project")
	eName := ctx.Param("environment")
	sName := ctx.Param("step")
	svcName := ctx.Param("service")
	runID := uuid.New().String()

	ps, err := project.GetProjects(bson.M{"name": pName, "team": tName})
	if err != nil {
		logger.Errorf("", "Count not get project at %s/%s: %s", tName, pName, err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if len(ps) == 0 {
		logger.Errorf("", "Could not get project at %s/%s for run %s", tName, pName, runID)
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	p := ps[0]

	if _, ok := p.Environments[eName]; !ok {
		logger.Errorf("", "No environment exists at %s/%s/%s", tName, pName, eName)
		utils.Error(err, ctx, http.StatusNotFound)
		return
	}
	if _, ok := p.Environments[eName].Services[svcName]; !ok {
		logger.Errorf("", "No service exists at %s/%s/%s/%s", tName, pName, eName, svcName)
		utils.Error(err, ctx, http.StatusNotFound)
		return
	}

	w := p.Environments[eName].Services[svcName].Workflow
	h := &history.History{
		RunID:       runID,
		States:      make([]history.State, 0),
		Project:     pName,
		Team:        tName,
		Environment: eName,
		Service:     svcName,
		Workflow:    w,
	}
	if err := history.CreateHistory(h); err != nil {
		logger.Errorf("", "Cannot create history: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := run.StartRun(runID, w, 0, "", sName, w.Steps[sName].Language); err != nil {
		logger.Errorf("", "Unable to start run: %s", runID)
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	go monitor.Run(runID, sName, 0)

	ctx.JSON(http.StatusCreated, gin.H{"run_id": runID})

}

// func GetRunNext(ctx *gin.Context) {
// 	runID := ctx.Param("run_id")
// 	stepName, context, err := history.GetNextStep(runID)
// 	if err != nil {
// 		logger.Errorf("", "Could not get next step for history %s: %s", runID, err.Error())
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}
// 	if stepName == "" {
// 		ctx.Status(http.StatusNoContent)
// 		return
// 	}
// 	em := ExecutionMessage{
// 		Name:    stepName,
// 		Context: context,
// 	}
// 	ctx.JSON(http.StatusOK, em)
// }
