package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/logger"
	"scaffold/server/mongodb"
	"scaffold/server/proxy"
	"scaffold/server/state"
	"scaffold/server/user"
	"scaffold/server/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

var InProgress = map[string]map[string]string{}
var ToCheck = map[string]map[string]string{}

func Run() {
	mongodb.InitCollections()
	filestore.InitBucket()

	// r := http.NewServeMux()
	r := mux.NewRouter()
	// mux.Handle("/ws", websocket.Handler(run))
	r.HandleFunc("/{host}/{port}/{cascade}/{run}/{version}",
		func(w http.ResponseWriter, req *http.Request) {
			proxy.NewProxy().ServeHTTP(w, req)
		})

	// http.Handle("/api/someAPI", apiHandler)
	// go http.ListenAndServe(fmt.Sprintf(":%d", config.Config.WSPort), proxy.NewProxy())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.WSPort),
		Handler: r,
	}

	go func() {
		log.Printf("Running reverse proxy at %s://0.0.0.0:%d\n", config.Config.Protocol, config.Config.WSPort)

		if config.Config.TLSEnabled {
			if serverErr := server.ListenAndServeTLS(config.Config.TLSCrtPath, config.Config.TLSKeyPath); serverErr != nil {
				logger.Fatalf("", "Error running websocket server: %s", serverErr)
			}
		} else {
			if serverErr := server.ListenAndServe(); serverErr != nil {
				logger.Fatalf("", "Error running websocket server: %s", serverErr)
			}
		}
	}()

	health.IsHealthy = true

	user.VerifyAdmin()
	auth.Nodes = make([]auth.NodeObject, 0)
	auth.UnknownNodes = make(map[string]auth.DegradedNodeObject)
	auth.UnhealthyNodes = make(map[string]auth.DegradedNodeObject)

	health.IsReady = true

	InProgress = make(map[string]map[string]string)
	ToCheck = make(map[string]map[string]string)

	for {
		newNodes := []auth.NodeObject{}
		for _, n := range auth.Nodes {
			queryURL := fmt.Sprintf("%s://%s:%d/health/healthy", n.Protocol, n.Host, n.Port)
			logger.Debugf("", "Querying %s", queryURL)
			resp, err := http.Get(queryURL)
			if err != nil {
				auth.UnknownNodes[n.Host] = auth.DegradedNodeObject{
					Node:  n,
					Count: 1,
				}
				continue
			}
			logger.Debugf("", "Got response code %d", resp.StatusCode)
			if resp.StatusCode >= 400 {
				auth.UnhealthyNodes[n.Host] = auth.DegradedNodeObject{
					Node:  n,
					Count: 1,
				}
				continue
			}
			newNodes = append(newNodes, n)
		}
		// Check unhealthy nodes for status change
		toDelete := []string{}
		for _, n := range auth.UnhealthyNodes {
			queryURL := fmt.Sprintf("%s://%s:%d/health/healthy", n.Node.Protocol, n.Node.Host, n.Node.Port)
			logger.Debugf("", "Querying %s", queryURL)
			resp, err := http.Get(queryURL)
			if err != nil {
				logger.Errorf("", "Got error %s when trying to contact unhealthy node %s", err, n.Node.Host)
				auth.UnknownNodes[n.Node.Host] = auth.DegradedNodeObject{
					Node:  n.Node,
					Count: 1,
				}
				toDelete = append(toDelete, n.Node.Host)
				continue
			}
			if resp.StatusCode >= 400 {
				logger.Errorf("", "Got response code %d when trying to contact unhealthy node %s", resp.StatusCode, n.Node.Host)
				continue
			}
			newNodes = append(newNodes, n.Node)
			toDelete = append(toDelete, n.Node.Host)
		}

		for _, key := range toDelete {
			delete(auth.UnhealthyNodes, key)
		}

		// Check unknown nodes for status change
		toDelete = []string{}
		for _, n := range auth.UnknownNodes {
			queryURL := fmt.Sprintf("%s://%s:%d/health/healthy", n.Node.Protocol, n.Node.Host, n.Node.Port)
			logger.Debugf("", "Querying %s", queryURL)
			resp, err := http.Get(queryURL)
			if err != nil {
				logger.Errorf("", "Got error %s when trying to contact unknown node %s", err, n.Node.Host)
				val := auth.UnknownNodes[n.Node.Host]
				val.Count += 1
				if val.Count >= config.Config.HeartbeatBackoff {
					toDelete = append(toDelete, n.Node.Host)
				}
				auth.UnknownNodes[n.Node.Host] = val
				continue
			}
			if resp.StatusCode >= 400 {
				logger.Errorf("", "Got response code %d when trying to contact unknown node %s", resp.StatusCode, n.Node.Host)
				auth.UnhealthyNodes[n.Node.Host] = auth.DegradedNodeObject{
					Node:  n.Node,
					Count: 1,
				}
				toDelete = append(toDelete, n.Node.Host)
				continue
			}
			newNodes = append(newNodes, n.Node)
			toDelete = append(toDelete, n.Node.Host)
		}

		for _, key := range toDelete {
			delete(auth.UnknownNodes, key)
		}

		auth.Nodes = newNodes

		cascades, err := cascade.GetAllCascades()
		if err == nil {
			for _, c := range cascades {
				taskMap := map[string]string{}
				if _, ok := ToCheck[c.Name]; ok {
					taskMap = ToCheck[c.Name]
				}
				for _, t := range c.Tasks {
					for key := range taskMap {
						if strings.HasPrefix(key, t.Name) {
							parts := strings.Split(key, ".")
							hostPort := taskMap[key]
							httpClient := http.Client{}
							requestURL := fmt.Sprintf("%s/api/v1/state/%s/%s/%s", hostPort, c.Name, t.Name, parts[1])
							req, _ := http.NewRequest("GET", requestURL, nil)
							req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
							req.Header.Set("Content-Type", "application/json")
							resp, err := httpClient.Do(req)

							if err != nil {
								logger.Errorf("", "Error getting run state %s.%s: %s", c.Name, t.Name, err.Error())
								continue
							}
							if resp.StatusCode == http.StatusOK {
								//Read the response body
								body, err := ioutil.ReadAll(resp.Body)
								if err != nil {
									logger.Errorf("", "Error reading body: %s", err.Error())
									continue
								}
								var s state.State
								var temp map[string]interface{}
								json.Unmarshal(body, &temp)

								tempBytes, _ := json.Marshal(temp)
								json.Unmarshal(tempBytes, &s)

								state.UpdateStateByNames(c.Name, t.Name, &s)

								logger.Tracef("", "Got state %s", s.Status)

								if s.Status == constants.STATE_STATUS_SUCCESS {
									logger.Debugf("", "Task %s has completed with success, removing from InProgress", key)
									triggerDepends(c, t.Name, constants.STATUS_TRIGGER_SUCCESS)
									triggerDepends(c, t.Name, constants.STATUS_TRIGGER_ALWAYS)
									delete(InProgress[c.Name], key)
									delete(ToCheck[c.Name], key)
								} else if s.Status == constants.STATE_STATUS_ERROR {
									logger.Debugf("", "Task %s has completed with error, removing from InProgress", key)
									triggerDepends(c, t.Name, constants.STATUS_TRIGGER_ERROR)
									triggerDepends(c, t.Name, constants.STATUS_TRIGGER_ALWAYS)
									delete(ToCheck[c.Name], key)
								}

								resp.Body.Close()
							}
						}
					}
					checkStateName := fmt.Sprintf("SCAFFOLD_CHECK-%s", t.Name)
					for key := range taskMap {
						if strings.HasPrefix(key, checkStateName) {
							parts := strings.Split(key, ".")
							hostPort := taskMap[key]
							httpClient := http.Client{}
							requestURL := fmt.Sprintf("%s/api/v1/state/%s/%s/%s", hostPort, c.Name, checkStateName, parts[1])
							req, _ := http.NewRequest("GET", requestURL, nil)
							req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
							req.Header.Set("Content-Type", "application/json")
							resp, err := httpClient.Do(req)

							if err != nil {
								logger.Errorf("", "Error getting run state %s.%s: %s", c.Name, checkStateName, err.Error())
								continue
							}
							if resp.StatusCode == http.StatusOK {
								//Read the response body
								body, err := ioutil.ReadAll(resp.Body)
								if err != nil {
									logger.Errorf("", "Error reading body: %s", err.Error())
									continue
								}
								var s state.State
								var temp map[string]interface{}
								json.Unmarshal(body, &temp)

								logger.Debugf("", "Raw worker state: %v", temp)

								tempBytes, _ := json.Marshal(temp)
								json.Unmarshal(tempBytes, &s)

								logger.Debugf("", "Got state from worker: %v", &s)

								state.UpdateStateByNames(c.Name, checkStateName, &s)

								if s.Status == constants.STATE_STATUS_SUCCESS {
									logger.Debugf("", "Removing state %s from InProgress", checkStateName)
									delete(InProgress[c.Name], key)
									delete(ToCheck[c.Name], key)
								}
								if s.Status == constants.STATE_STATUS_ERROR {
									ss, err := state.GetStateByNames(c.Name, t.Name)
									if err != nil {
										logger.Errorf("", "Issue getting state %s %s: %s", c.Name, t.Name, err.Error())
										continue
									}
									ss.Status = constants.STATE_STATUS_ERROR
									state.UpdateStateByNames(c.Name, t.Name, ss)
									triggerDepends(c, t.Name, constants.STATUS_TRIGGER_ERROR)
									triggerDepends(c, t.Name, constants.STATUS_TRIGGER_ALWAYS)
									delete(InProgress[c.Name], key)
									delete(ToCheck[c.Name], key)
								}

								resp.Body.Close()
							}
						}
					}
					if t.Check.Interval > 0 {
						s, err := state.GetStateByNames(c.Name, t.Name)
						if err != nil {
							logger.Errorf("", "Issue getting state %s %s: %s", c.Name, t.Name, err.Error())
							continue
						}

						if s.Status != constants.STATE_STATUS_SUCCESS {
							continue
						}

						if s.Finished == "" {
							continue
						}
						currentTime := time.Now().UTC()
						logger.Debugf("", "State finished: %s", s.Finished)
						lastRun, _ := time.Parse("2006-01-02T15:04:05Z MST", fmt.Sprintf("%s UTC", s.Finished))

						cs, err := state.GetStateByNames(c.Name, checkStateName)
						if err == nil {
							logger.Debugf("", "Found state for %s", checkStateName)
							if cs.Finished != "" {
								lastRun, _ = time.Parse("2006-01-02T15:04:05Z MST", fmt.Sprintf("%s UTC", cs.Finished))
								logger.Debugf("", "Check state finished: %s", cs.Finished)
								logger.Debugf("", "Check run is of status %s", cs.Status)
							}
							if cs.Status == constants.STATE_STATUS_RUNNING || cs.Status == constants.STATE_STATUS_WAITING {
								logger.Debugf("", "Check run is of status %s", cs.Status)
								continue
							}
						} else {
							logger.Warnf("", "No state found for %s", checkStateName)
						}

						diff := int(currentTime.Sub(lastRun).Seconds())

						if diff < t.Check.Interval {
							continue
						}

						logger.Debugf("", "Triggering check run for %s %s", c.Name, t.Name)
						httpClient := http.Client{}
						requestURL := fmt.Sprintf("%s://localhost:%d/api/v1/run/%s/%s/check", config.Config.Protocol, config.Config.Port, c.Name, t.Name)
						req, _ := http.NewRequest("POST", requestURL, nil)
						req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
						req.Header.Set("Content-Type", "application/json")
						resp, err := httpClient.Do(req)
						if err != nil {
							logger.Error("", err.Error())
							continue
						}
						if resp.StatusCode >= 400 {
							logger.Errorf("", "Received trigger status code %d", resp.StatusCode)
						}
					}
				}
			}
		}
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
	}
}

func triggerDepends(c *cascade.Cascade, tn, status string) {
	toTrigger := []string{}
	states, _ := state.GetStatesByCascade(c.Name)
	switch status {
	case constants.STATUS_TRIGGER_SUCCESS:
		for _, s := range states {
			if s.Status == constants.STATE_STATUS_SUCCESS {
				toTrigger = append(toTrigger, s.Task)
			}
		}
	case constants.STATUS_TRIGGER_ERROR:
		for _, s := range states {
			if s.Status == constants.STATE_STATUS_ERROR {
				toTrigger = append(toTrigger, s.Task)
			}
		}
	case constants.STATUS_TRIGGER_ALWAYS:
		for _, s := range states {
			if s.Status == constants.STATE_STATUS_ERROR || s.Status == constants.STATE_STATUS_SUCCESS {
				toTrigger = append(toTrigger, s.Task)
			}
		}
	}
	for _, t := range c.Tasks {
		dependsOn := make([]string, 0)
		switch status {
		case constants.STATUS_TRIGGER_SUCCESS:
			if t.DependsOn.Success == nil {
				continue
			}
			dependsOn = append(dependsOn, t.DependsOn.Success...)
		case constants.STATUS_TRIGGER_ERROR:
			if t.DependsOn.Error == nil {
				continue
			}
			dependsOn = append(dependsOn, t.DependsOn.Error...)
		case constants.STATUS_TRIGGER_ALWAYS:
			if t.DependsOn.Always == nil {
				continue
			}
			dependsOn = append(dependsOn, t.DependsOn.Always...)
		}
		if utils.Contains(dependsOn, tn) {
			shouldTrigger := true
			for _, n := range dependsOn {
				if !utils.Contains(toTrigger, n) {
					shouldTrigger = false
					break
				}
			}
			if !shouldTrigger || !t.AutoExecute {
				continue
			}
			httpClient := http.Client{}
			requestURL := fmt.Sprintf("%s://localhost:%d/api/v1/run/%s/%s", config.Config.Protocol, config.Config.Port, c.Name, t.Name)
			req, _ := http.NewRequest("POST", requestURL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
			req.Header.Set("Content-Type", "application/json")
			resp, err := httpClient.Do(req)
			if err != nil {
				logger.Errorf("", "Depends run trigger error: %s", err.Error())
				continue
			}
			if resp.StatusCode >= 400 {
				logger.Errorf("", "Received trigger status code %d", resp.StatusCode)
				panic(fmt.Sprintf("Received trigger status code %d", resp.StatusCode))
			}
		}
	}
}

func InputChangeStateChange(name string, changed []string) error {
	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		return err
	}
	for _, i := range changed {
		for _, t := range c.Tasks {
			ss, err := state.GetStateByNames(c.Name, t.Name)
			if err == nil {
				previousName := fmt.Sprintf("SCAFFOLD_PREVIOUS-%s", t.Name)
				ss.Task = previousName
				state.UpdateStateByNames(c.Name, previousName, ss)
			}
			s := &state.State{
				Task:     t.Name,
				Cascade:  c.Name,
				Status:   constants.STATE_STATUS_NOT_STARTED,
				Started:  "",
				Finished: "",
				Output:   "",
				Number:   t.RunNumber,
				Display:  make([]map[string]interface{}, 0),
			}
			if err := state.UpdateStateByNames(c.Name, s.Task, s); err != nil {
				logger.Errorf("", "Cannot update state %s %s: %s", c.Name, t.Name, err.Error())
				continue
			}
			if utils.Contains(utils.Keys(t.Inputs), i) {
				SetDependsState(c, t.Name)
			}
		}
	}
	return nil
}

func SetDependsState(c *cascade.Cascade, tn string) {
	for _, t := range c.Tasks {
		dependsOn := make([]string, 0)
		if t.DependsOn.Success != nil {
			dependsOn = append(dependsOn, t.DependsOn.Success...)
		}
		if t.DependsOn.Error != nil {
			dependsOn = append(dependsOn, t.DependsOn.Error...)
		}
		if t.DependsOn.Always != nil {
			dependsOn = append(dependsOn, t.DependsOn.Always...)
		}
		if utils.Contains(dependsOn, tn) {
			if !utils.Contains(dependsOn, tn) {
				continue
			}
			ss, err := state.GetStateByNames(c.Name, t.Name)
			if err == nil {
				previousName := fmt.Sprintf("SCAFFOLD_PREVIOUS-%s", t.Name)
				ss.Task = previousName
				state.UpdateStateByNames(c.Name, previousName, ss)
			}
			s := &state.State{
				Task:     t.Name,
				Cascade:  c.Name,
				Status:   constants.STATE_STATUS_NOT_STARTED,
				Started:  "",
				Finished: "",
				Output:   "",
				Number:   t.RunNumber,
				Display:  make([]map[string]interface{}, 0),
			}
			if err := state.UpdateStateByNames(c.Name, s.Task, s); err != nil {
				logger.Errorf("", "Cannot update state %s %s: %s", c.Name, s.Task, err.Error())
				continue
			}
			SetDependsState(c, s.Task)
		}
	}
}

//
//	@Summary		Get status of all nodes
//	@Description	Get status from all nodes
//	@tags			manager
//	@tags			health
//	@accept			json
//	@produce		json
//	@Success		200	{object}	object
//	@Router			/health/status [get]
func GetStatus(ctx *gin.Context) {
	nodes := make([]map[string]string, 0)
	managerStatus := "healthy"
	if !health.IsHealthy {
		managerStatus = "degraded"
	}
	nodes = append(nodes, map[string]string{"name": config.Config.Host, "status": managerStatus, "version": constants.VERSION})
	for _, node := range auth.Nodes {
		nodes = append(nodes, map[string]string{"name": node.Host, "status": "healthy", "version": node.Version})
	}
	for key, node := range auth.UnknownNodes {
		nodes = append(nodes, map[string]string{"name": key, "status": "unknown", "version": node.Node.Version})
	}
	for key, node := range auth.UnhealthyNodes {
		nodes = append(nodes, map[string]string{"name": key, "status": "degraded", "version": node.Node.Version})
	}

	ctx.JSON(http.StatusOK, gin.H{"nodes": nodes})
}
