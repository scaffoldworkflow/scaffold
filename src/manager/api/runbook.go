package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/manager/kernel"
	"scaffold/manager/runbook"
	"scaffold/manager/utils"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RunbookKernelSetup(ctx *gin.Context) {
	runbookID := ctx.Param("runbook_id")
	r, err := runbook.GetRunbookByID(runbookID)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	k := &kernel.Kernel{
		Requirements: r.Requirements,
	}
	if err := k.Spawn(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	logger.Debugf("", "Successfully spawned kernel with ID %s", k.ID)
	max_iterations := 100 // this equates to 10 seconds of waiting for the kernel to start
	iteration := 0
	for !k.Ready {
		time.Sleep(100 * time.Millisecond)
		if iteration == max_iterations {
			utils.Error(fmt.Errorf("kernel health check timed out with ID %s", k.ID), ctx, http.StatusInternalServerError)
			return
		}
		iteration += 1
	}
	kernel.Kernels[k.ID] = k
	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/ui/runbooks/%s/%s", runbookID, k.ID))
}

//	@summary					Create a runbook
//	@description				Create a runbook from a JSON object
//	@tags						manager
//	@tags						runbook
//	@accept						json
//	@produce					json
//	@Param						runbook	body		runbook.Runbook	true	"Task Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/runbook [post]
func CreateRunbook(ctx *gin.Context) {
	var r runbook.Runbook
	if err := ctx.ShouldBindJSON(&r); err != nil {
		logger.Errorf("", "Error binding runbook JSON: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if r.Groups != nil {
		if !validateUserGroup(ctx, r.Groups) {
			logger.Warnf("", "User does not have group access")
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	if err := runbook.CreateRunbook(&r); err != nil {
		logger.Errorf("", "Runbook create failed: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	logger.Debugf("", "Runbook successfully created")
	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

//	@summary					Delete runbooks
//	@description				Delete runbooks by their ID
//	@tags						manager
//	@tags						runbook
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/runbook/{id} [delete]
func DeleteRunbookByID(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := runbook.DeleteRunbookByID(id); err != nil {
		logger.Errorf("", "Failed to delete runbook %s: %s", id, err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	logger.Debugf("", "Runbook %s deleted", id)
	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all runbooks
//	@description				Get all runbooks
//	@tags						manager
//	@tags						runbook
//	@produce					json
//	@success					200	{array}		runbook.Runbook
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/runbook [get]
func GetAllRunbooks(ctx *gin.Context) {
	runbooks, err := runbook.GetAllRunbooks()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warnf("", "No runbooks returned")
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		logger.Errorf("", "Could not get runbooks: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	runbooksOut := make([]runbook.Runbook, 0)
	for _, r := range runbooks {
		if r.Groups != nil {
			if validateUserGroup(ctx, r.Groups) {
				runbooksOut = append(runbooksOut, *r)
			}
			continue
		}
		runbooksOut = append(runbooksOut, *r)
	}

	logger.Debugf("", "Got %d runbooks", len(runbooksOut))

	ctx.JSON(http.StatusOK, runbooksOut)
}

//	@summary					Get runbooks
//	@description				Get runbooks by their ID
//	@tags						manager
//	@tags						runbook
//	@produce					json
//	@success					200	{array}		runbook.Runbook
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/runbook/{id} [get]
func GetRunbookByID(ctx *gin.Context) {
	id := ctx.Param("id")

	r, err := runbook.GetRunbookByID(id)
	if err != nil {
		logger.Errorf("", "Unable to get runbook with ID %s", id)
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	logger.Debugf("", "Successfully returned runbook %s", id)

	ctx.JSON(http.StatusOK, r)
}

//	@summary					Update a runbook
//	@description				Update a runbook from a JSON object
//	@tags						manager
//	@tags						runbook
//	@accept						json
//	@produce					json
//	@Param						runbook	body		runbook.Runbook	true	"Task Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/runbook/{id} [put]
func UpdateRunbookByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var r runbook.Runbook
	if err := ctx.ShouldBindJSON(&r); err != nil {
		logger.Errorf("", "Unable to bind JSON on runbook update: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	r.ID = id

	if err := runbook.UpdateRunbookByID(&r); err != nil {
		logger.Errorf("", "Unable to update runbook %s: %s", id, err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	logger.Debugf("", "Runbook %s updated", id)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
