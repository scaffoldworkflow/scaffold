package page

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/manager/api"
	"scaffold/manager/config"
	"scaffold/manager/utils"
	"strconv"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
)

func KernelExecute(ctx *gin.Context) {
	var err error
	var em api.ExecuteMessage
	if err := ctx.ShouldBindJSON(&em); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	em.BlockIdx, err = strconv.Atoi(em.BlockIdxString)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	postBody, _ := json.Marshal(em)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://localhost:%d/api/v1/kernel/%s", config.Config.Port, em.KernelID)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	var data map[string]string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &data); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
		<pre id="output_%d" style="background:#282B2E;" hx-get="/htmx/kernel/%s/%s/%d" hx-trigger="load delay:1s" hx-swap="outerHTML"></pre>
	`, em.BlockIdx, em.KernelID, data["run_id"], em.BlockIdx)))
}

func KernelBuildOutput(ctx *gin.Context) {
	kernelID := ctx.Param("kernel_id")
	runID := ctx.Param("run_id")
	blockIdx := ctx.Param("block_idx")

	logger.Debugf("", "Getting output from kernel %s with run %s on block %s", kernelID, runID, blockIdx)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://localhost:%d/api/v1/kernel/%s/%s", config.Config.Port, kernelID, runID)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Could not get kernel output: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	logger.Debugf("", "Got output from API")
	var data api.OutputMessage
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	logger.Tracef("", "Got response: %s", string(body))
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Errorf("", "Could not unmarshal body into JSON: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if data.Finished {
		logger.Tracef("", "Kernel run is finished")
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
			<pre id="output_%s" style="background:#282B2E;">%s</pre>
		`, blockIdx, data.Contents)))
	} else {
		logger.Tracef("", "Kernel run is not finished")
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
			<pre id="output_%s" style="background:#282B2E;" hx-get="/htmx/kernel/%s/%s/%s" hx-trigger="load delay:1s" hx-swap="outerHTML">%s</pre>
		`, blockIdx, kernelID, runID, blockIdx, data.Contents)))
	}
}
