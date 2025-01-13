package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/monitor"
	"scaffold/server/utils"
	"scaffold/server/workflow"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateMonitor(ctx *gin.Context) {
	var m monitor.Monitor
	if err := ctx.ShouldBindJSON(&m); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	c, err := workflow.GetWorkflowByName(m.Workflow)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	err = monitor.CreateMonitor(&m)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func GetAllMonitors(ctx *gin.Context) {
	monitors, err := monitor.GetAllMonitors()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	monitorsOut := make([]monitor.Monitor, 0)
	for _, m := range monitors {
		c, err := workflow.GetWorkflowByName(m.Workflow)
		if err != nil || c == nil {
			logger.Warnf("", "Monitor workflow %s does not exist", c.Name)
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				monitorsOut = append(monitorsOut, *m)
			}
			continue
		}
		monitorsOut = append(monitorsOut, *m)
	}

	ctx.JSON(http.StatusOK, monitorsOut)
}

func GetMonitorByID(ctx *gin.Context) {
	id := ctx.Param("id")
	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &m)
}

func GetMonitorsByWorkflow(ctx *gin.Context) {
	wn := ctx.Param("workflow")
	ms, err := monitor.GetMonitorsByWorkflow(wn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &ms)
}

func UpdateMonitorByID(ctx *gin.Context) {
	id := ctx.Param("id")
	var m monitor.Monitor
	if err := ctx.ShouldBindJSON(&m); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := monitor.UpdateMonitorByID(id, &m)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func DeleteMonitorByID(ctx *gin.Context) {
	id := ctx.Param("id")

	err := monitor.DeleteMonitorByID(id)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func DeleteMonitorsByWorkflow(ctx *gin.Context) {
	wn := ctx.Param("workflow")

	err := monitor.DeleteMonitorsByWorkflow(wn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func ToggleMonitor(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		utils.Error(fmt.Errorf("no monitor exists with ID %s", id), ctx, http.StatusNotFound)
		return
	}
	if err := m.Toggle(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func StartMonitor(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		utils.Error(fmt.Errorf("no monitor exists with ID %s", id), ctx, http.StatusNotFound)
		return
	}
	if m.Status == constants.MONITOR_STATUS_RUNNING || m.Status == constants.MONITOR_STATUS_ALERT {
		utils.Error(fmt.Errorf("monitor %s is already running", id), ctx, http.StatusConflict)
		return
	}
	if err := m.Start(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func StopMonitor(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		utils.Error(fmt.Errorf("no monitor exists with ID %s", id), ctx, http.StatusNotFound)
		return
	}
	if m.Status == constants.MONITOR_STATUS_STOPPED {
		utils.Error(fmt.Errorf("monitor %s is already stopped", id), ctx, http.StatusConflict)
		return
	}
	if err := m.Stop(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
