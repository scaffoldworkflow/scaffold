package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/server/auth"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
)

//	@summary					Get available containers
//	@description				List containers available for execution
//	@tags						worker
//	@produce					json
//	@success					200	{array}		string
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/health/available [get]
// func GetAvailableContainers(ctx *gin.Context) {
// 	output := []string{}

// 	for idx, groups := range container.LastGroups {
// 		if validateUserGroup(ctx, groups) {
// 			output = append(output, container.LastRun[idx])
// 		}
// 	}

// 	ctx.JSON(http.StatusOK, container.LastRun)
// }

//	@summary					Get all containers
//	@description				List all containers
//	@tags						worker
//	@produce					json
//	@success					200	{array}		string
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/run/containers [get]
func GetAllContainers(ctx *gin.Context) {
	available := map[string][]string{}
	for _, n := range auth.Nodes {
		httpClient := http.Client{}
		requestURL := fmt.Sprintf("%s://%s:%d/api/v1/available", n.Protocol, n.Host, n.Port)
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Set("Authorization", ctx.Request.Header.Get("X-Scaffold-API"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)

		if err != nil {
			logger.Errorf("", "Error getting available containers %s", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			//Read the response body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Errorf("", "Error reading body: %s", err.Error())
				continue
			}
			var data []string
			json.Unmarshal(body, &data)

			if len(data) > 0 {
				available[fmt.Sprintf("%s:%d", n.Host, n.WSPort)] = data
			}
			resp.Body.Close()
		}
	}
	ctx.JSON(http.StatusOK, available)
}
