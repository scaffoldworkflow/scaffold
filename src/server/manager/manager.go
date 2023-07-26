package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/mongodb"
	"scaffold/server/state"
	"scaffold/server/user"
	"scaffold/server/utils"
	"time"

	"github.com/gin-gonic/gin"
)

var InProgress = map[string]map[string]string{}

func Run() {
	mongodb.InitCollections()
	filestore.InitBucket()

	health.IsHealthy = true

	user.VerifyAdmin()
	auth.Nodes = make([]auth.NodeObject, 0)

	health.IsReady = true

	InProgress = make(map[string]map[string]string)

	for {
		newNodes := []auth.NodeObject{}
		for _, n := range auth.Nodes {
			queryURL := fmt.Sprintf("http://%s:%d/health/healthy", n.Host, n.Port)
			resp, err := http.Get(queryURL)
			if err != nil || resp.StatusCode >= 400 {
				continue
			}
			newNodes = append(newNodes, n)
		}
		auth.Nodes = newNodes

		cascades, err := cascade.GetAllCascades()
		if err == nil {
			for _, c := range cascades {
				if taskMap, ok := InProgress[c.Name]; ok {
					for _, t := range c.Tasks {
						if hostPort, ok := taskMap[t.Name]; ok {
							httpClient := &http.Client{}
							requestURL := fmt.Sprintf("http://%s/api/v1/state/%s/%s", hostPort, c.Name, t.Name)
							req, _ := http.NewRequest("GET", requestURL, nil)
							req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
							req.Header.Set("Content-Type", "application/json")
							resp, err := httpClient.Do(req)

							if err != nil {
								fmt.Printf("Error: %s", err.Error())
								continue
							}
							if resp.StatusCode == http.StatusOK {
								//Read the response body
								body, err := ioutil.ReadAll(resp.Body)
								if err != nil {
									fmt.Printf("Error: %s", err.Error())
									continue
								}
								var s state.State
								var temp map[string]map[string]interface{}
								json.Unmarshal(body, &temp)

								tempBytes, _ := json.Marshal(temp["state"])
								json.Unmarshal(tempBytes, &s)

								state.UpdateStateByNames(c.Name, t.Name, &s)

								if s.Status == constants.STATE_STATUS_SUCCESS {
									triggerDepends(c, t.Name)
									delete(InProgress[c.Name], t.Name)
								}

								resp.Body.Close()
							}
						}
					}
				}
			}
		}
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
	}
}

func triggerDepends(c *cascade.Cascade, tn string) {
	success := []string{}
	states, _ := state.GetStatesByCascade(c.Name)
	for _, s := range states {
		if s.Status == constants.STATE_STATUS_SUCCESS {
			success = append(success, s.Task)
		}
	}
	for _, t := range c.Tasks {
		if utils.Contains(t.DependsOn, tn) {
			shouldTrigger := true
			for _, n := range t.DependsOn {
				if !utils.Contains(success, n) {
					shouldTrigger = false
					break
				}
			}
			if !shouldTrigger {
				continue
			}
			httpClient := &http.Client{}
			requestURL := fmt.Sprintf("http://localhost:%d/api/v1/run/%s/%s", config.Config.HTTPPort, c.Name, t.Name)
			req, _ := http.NewRequest("POST", requestURL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
			req.Header.Set("Content-Type", "application/json")
			resp, err := httpClient.Do(req)
			if err != nil {
				fmt.Printf("ERROR: %s", err.Error())
				continue
			}
			if resp.StatusCode >= 400 {
				panic(fmt.Sprintf("Received trigger status code %d", resp.StatusCode))
			}
		}
	}
}

func GetStatus(ctx *gin.Context) {
	var nodes []string
	for _, node := range auth.Nodes {
		nodes = append(nodes, node.Host)
	}

	ctx.JSON(http.StatusOK, gin.H{"nodes": nodes})
}
