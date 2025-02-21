package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/manager/alert"
	"scaffold/manager/config"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: Update alert API to new standard with query params

func CreateAlert(ctx *gin.Context) {
	var a alert.Alert
	if err := ctx.ShouldBindJSON(&a); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// c, err := workflow.GetWorkflowByName(a.Workflow)
	// if err != nil {
	// 	utils.Error(err, ctx, http.StatusNotFound)
	// }
	// if c.Groups != nil {
	// 	if !validateUserGroup(ctx, c.Groups) {
	// 		utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
	// 	}
	// }

	if err := alert.CreateAlert(&a); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func GetAllAlerts(ctx *gin.Context) {
	alerts, err := alert.GetAllAlerts()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	alertsOut := make([]alert.Alert, 0)
	for _, m := range alerts {
		// c, err := workflow.GetWorkflowByName(m.Workflow)
		// if err != nil || c == nil {
		// 	logger.Warnf("", "Alert workflow %s does not exist", c.Name)
		// 	continue
		// }
		// if c.Groups != nil {
		// 	if validateUserGroup(ctx, c.Groups) {
		// 		alertsOut = append(alertsOut, *m)
		// 	}
		// 	continue
		// }
		alertsOut = append(alertsOut, *m)
	}

	ctx.JSON(http.StatusOK, alertsOut)
}

func GetAlertByID(ctx *gin.Context) {
	id := ctx.Param("id")
	m, err := alert.GetAlertByID(id)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &m)
}

func GetAlertsByWorkflow(ctx *gin.Context) {
	wn := ctx.Param("workflow")
	as, err := alert.GetAlertsByWorkflow(wn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &as)
}

func UpdateAlertByID(ctx *gin.Context) {
	id := ctx.Param("id")
	var a alert.Alert
	if err := ctx.ShouldBindJSON(&a); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := alert.UpdateAlertByID(id, &a)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func DeleteAlertByID(ctx *gin.Context) {
	id := ctx.Param("id")

	err := alert.DeleteAlertByID(id)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// func DeleteAlertsByWorkflow(ctx *gin.Context) {
// 	wn := ctx.Param("workflow")

// 	err := alert.DeleteAlertsByWorkflow(wn)

// 	if err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
// }

func StartAlert(ctx *gin.Context) {
	id := ctx.Param("id")

	a, err := alert.GetAlertByID(id)
	if err != nil {
		logger.Errorf("", "Could not find alert %s to run", id)
		utils.Error(err, ctx, http.StatusNotFound)
		return
	}

	kernelID, err := a.SetupKernel()
	if err != nil {
		logger.Errorf("", "Could not setup kernel for alert %s", id)
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	em := ExecuteMessage{
		Contents: a.Script,
		KernelID: kernelID,
		Language: a.Language,
	}

	postBody, _ := json.Marshal(em)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://localhost:%d/api/v1/kernel/%s", config.Config.Port, kernelID)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Kernel trigger for id %s has failed with error %s", id, err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("", "Encountered error reading body: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
	}
	var data map[string]string
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		logger.Fatalf("", "Encountered error unmarshalling status JSON: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	runID := data["run_id"]

	ki := a.Kernels[kernelID]
	ki.RunID = runID
	a.Kernels[kernelID] = ki

	alert.UpdateAlertByID(id, a)
	go a.Check(kernelID)

	ctx.JSON(http.StatusOK, gin.H{"kernel_id": kernelID, "run_id": runID})
}

func StopAlert(ctx *gin.Context) {
	id := ctx.Param("id")

	a, err := alert.GetAlertByID(id)
	if err != nil {
		utils.Error(fmt.Errorf("no alert exists with ID %s", id), ctx, http.StatusNotFound)
		return
	}
	if err := a.Stop(); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
