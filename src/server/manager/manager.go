package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/bulwark"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	scron "scaffold/server/cron"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/mongodb"
	"scaffold/server/msg"
	"scaffold/server/proxy"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/user"
	"scaffold/server/utils"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

var toKill []string

func Run() {
	mongodb.InitCollections()
	filestore.InitBucket()
	bulwark.QueueCreate(config.Config.ManagerQueueName)
	bulwark.QueueCreate(config.Config.WorkerQueueName)
	bulwark.BufferCreate(config.Config.KillBufferName)

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

	toKill = make([]string, 0)

	health.IsHealthy = true

	if err := user.VerifyAdmin(); err != nil {
		logger.Fatalf("", "Unable to create admin user: %s", err.Error())
	}
	auth.Nodes = make(map[string]auth.NodeObject)

	health.IsReady = true

	go healthCheck()

	go bulwark.RunManager(QueueDataReceive)
	go bulwark.RunWorker(nil)
	go bulwark.RunBuffer(BufferDataReceive)

	go scron.Start()

	queueCheck()
}

func QueueDataReceive(endpoint, data string) error {
	if len(data) == 0 {
		return nil
	}
	var m msg.RunMsg
	// bytes, err := json.Marshal([]byte(data))
	// if err != nil {
	// 	logger.Errorf("", "Error processing queue message: %s", err.Error())
	// 	return err
	// }
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		logger.Errorf("", "Error processing queue message: %s", err.Error())
		return err
	}
	switch m.Status {
	case constants.STATE_STATUS_SUCCESS:
		logger.Debugf("", "Task %s has completed with status success", m.Task)
		stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_SUCCESS)
		stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_ALWAYS)
	case constants.STATE_STATUS_ERROR:
		logger.Debugf("", "Task %s has completed with status error", m.Task)
		stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_ERROR)
		stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_ALWAYS)
	case constants.STATE_STATUS_KILLED:
		logger.Debugf("", "Task %s has completed with status killed", m.Task)
		id := fmt.Sprintf("%s-%s", m.Cascade, m.Task)
		toKill = utils.Remove(toKill, id)
		if err := bulwark.BufferSet(bulwark.BufferClient, toKill); err != nil {
			logger.Errorf("", "Encountered error while updating buffer: %s", err.Error())
			return err
		}
	}
	return nil
}

func BufferDataReceive(endpoint, data string) error {
	// if len(data) == 0 {
	// 	return nil
	// }
	// var m msg.RunMsg
	// if err := json.Unmarshal([]byte(data), &m); err != nil {
	// 	logger.Errorf("", "Error processing buffer message: %s", err.Error())
	// 	return err
	// }

	// if m.Status == constants.STATE_STATUS_SUCCESS {
	// 	logger.Debugf("", "Task %s has completed with success", m.Task)
	// 	stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_SUCCESS)
	// 	stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_ALWAYS)
	// } else if m.Status == constants.STATE_STATUS_ERROR {
	// 	logger.Debugf("", "Task %s has completed with error", m.Task)
	// 	stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_ERROR)
	// 	stateChange(m.Cascade, m.Task, constants.STATUS_TRIGGER_ALWAYS)
	// }
	return nil
}

func queueCheck() {
	for {
		logger.Tracef("", "Sleeping...")
		time.Sleep(time.Duration(config.Config.BulwarkCheckInterval) * time.Millisecond)
		logger.Debugf("", "Worker manager queue")
		bulwark.QueuePop(bulwark.ManagerClient)
		// for _, id := range buffers {
		// 	bulwark.BufferClient.Endpoint = fmt.Sprintf("%s/%s", bconst.ENDPOINT_TYPE_BUFFER, id)
		// 	bulwark.BufferGet(bulwark.BufferClient)
		// }
	}
}

func healthCheck() {
	for {
		for key, n := range auth.Nodes {
			if n.Ping > config.Config.HeartbeatBackoff {
				ss, err := state.GetStatesByWorker(n.Name)
				if err != nil {
					logger.Errorf("", "Unable to get states by worker: %s", n.Name)
				}
				for _, s := range ss {
					switch s.Status {
					case constants.STATE_STATUS_RUNNING:
						DoKill(s.Cascade, s.Task)
					case constants.STATE_STATUS_WAITING:
						DoKill(s.Cascade, s.Task)
					}
				}
			}
			n.Ping += 1
			for auth.NodeLock {
				time.Sleep(250 * time.Millisecond)
			}
			auth.NodeLock = true
			auth.Nodes[key] = n
			auth.NodeLock = false
		}
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
	}
}

func stateChange(cn, tn, status string) error {
	ts, err := task.GetTasksByCascade(cn)
	if err != nil {
		logger.Errorf("", "Cannot change state for %s", cn)
		return err
	}
	switch status {
	case constants.STATE_STATUS_SUCCESS:
		for _, t := range ts {
			shouldExecute := false
			for _, n := range t.DependsOn.Always {
				if n == tn && t.AutoExecute {
					shouldExecute = true
					continue
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_ERROR && s.Status != constants.STATE_STATUS_SUCCESS {
					return nil
				}
			}
			for _, n := range t.DependsOn.Success {
				if n == tn && t.AutoExecute {
					shouldExecute = true
					continue
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_SUCCESS {
					return nil
				}
			}
			if shouldExecute {
				if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED); err != nil {
					return err
				}
				if err := DoTrigger(cn, t.Name); err != nil {
					return err
				}
			}
		}
	case constants.STATE_STATUS_ERROR:
		for _, t := range ts {
			shouldExecute := false
			for _, n := range t.DependsOn.Always {
				if n == tn && t.AutoExecute {
					shouldExecute = true
					continue
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_ERROR && s.Status != constants.STATE_STATUS_SUCCESS {
					return nil
				}
			}
			for _, n := range t.DependsOn.Error {
				if n == tn && t.AutoExecute {
					shouldExecute = true
					continue
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_ERROR {
					return nil
				}
			}
			if shouldExecute {
				if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED); err != nil {
					return err
				}
				if err := DoTrigger(cn, t.Name); err != nil {
					return err
				}
			}
		}
	case constants.STATE_STATUS_NOT_STARTED:
		for _, t := range ts {
			for _, n := range t.DependsOn.Always {
				if n == tn {
					s, err := state.GetStateByNames(cn, t.Name)
					if err != nil {
						return err
					}
					s.Status = constants.STATE_STATUS_NOT_STARTED
					if err := state.UpdateStateByNames(cn, t.Name, s); err != nil {
						return err
					}
					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED); err != nil {
						return err
					}
				}
			}
			for _, n := range t.DependsOn.Error {
				if n == tn {
					s, err := state.GetStateByNames(cn, t.Name)
					if err != nil {
						return err
					}
					s.Status = constants.STATE_STATUS_NOT_STARTED
					if err := state.UpdateStateByNames(cn, t.Name, s); err != nil {
						return err
					}
					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED); err != nil {
						return err
					}
				}
			}
			for _, n := range t.DependsOn.Success {
				if n == tn {
					s, err := state.GetStateByNames(cn, t.Name)
					if err != nil {
						return err
					}
					s.Status = constants.STATE_STATUS_NOT_STARTED
					if err := state.UpdateStateByNames(cn, t.Name, s); err != nil {
						return err
					}
					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func DoTrigger(cn, tn string) error {
	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		return err
	}

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		return err
	}

	if t.Disabled {
		return nil
	}

	s, err := state.GetStateByNames(cn, tn)
	if err != nil {
		return err
	}
	s.Status = constants.STATE_STATUS_WAITING
	if err := state.UpdateStateByNames(cn, tn, s); err != nil {
		return err
	}

	if err := DoKill(cn, tn); err != nil {
		return err
	}

	m := msg.TriggerMsg{
		Task:    tn,
		Cascade: cn,
		Action:  constants.ACTION_TRIGGER,
		Groups:  c.Groups,
		Number:  t.RunNumber + 1,
	}

	logger.Infof("", "Triggering run with message %v", m)
	return bulwark.QueuePush(bulwark.WorkerClient, m)
}

func DoKill(cn, tn string) error {
	// id := fmt.Sprintf("%s-%s", cn, tn)
	// toKill = append(toKill, id)
	// toKill = utils.RemoveDuplicateValues(toKill)
	// return bulwark.BufferSet(bulwark.BufferClient, toKill)
	logger.Tracef("", "Killing run %s.%s", cn, tn)
	// return stateChange(cn, tn, constants.STATE_STATUS_KILLED)
	// return state.UpdateStateKilledByNames(cn, tn, true)

	for _, node := range auth.Nodes {
		uri := fmt.Sprintf("%s://%s:%d", node.Protocol, node.Host, node.Port)
		httpClient := &http.Client{}
		requestURL := fmt.Sprintf("%s/api/v1/run/%s/%s", uri, cn, tn)
		req, _ := http.NewRequest("DELETE", requestURL, nil)
		req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			logger.Fatalf("", "Encountered error killing run: %v", err)
			return err
		}
		if resp.StatusCode >= 400 {
			logger.Debugf("", "Got status code %d when trying to kill run", resp.StatusCode)
			return fmt.Errorf("got status code %d when trying to kill run", resp.StatusCode)
		}
		logger.Debugf("", "Run kill successfully triggered at %s", uri)
	}
	return nil
}

func InputChangeStateChange(name string, changed []string) error {
	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		return err
	}
	for _, t := range c.Tasks {
		for _, i := range changed {
			if utils.Contains(utils.Keys(t.Inputs), i) {
				stateChange(name, t.Name, constants.STATE_STATUS_NOT_STARTED)
				// state.CopyStatesByNames(name, t.Name, fmt.Sprintf("SCAFFOLD_PREVIOUS-%s", t.Name))
				state.ClearStateByNames(name, t.Name, t.RunNumber)
				break
			}
		}
	}
	return nil
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
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddress := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddress.IP.String()

	nodes := make([]map[string]string, 0)
	managerStatus := "healthy"
	if !health.IsHealthy {
		managerStatus = "degraded"
	}
	nodes = append(nodes, map[string]string{"name": config.Config.Host, "ip": ip, "status": managerStatus, "version": constants.VERSION})
	for _, node := range auth.Nodes {
		if node.Ping < config.Config.PingHealthyThreshold {
			nodes = append(nodes, map[string]string{"name": node.Name, "ip": node.Host, "status": "healthy", "version": node.Version})
			continue
		}
		if node.Ping < config.Config.PingUnknownThreshold {
			nodes = append(nodes, map[string]string{"name": node.Name, "ip": node.Host, "status": "unknown", "version": node.Version})
			continue
		}
		nodes = append(nodes, map[string]string{"name": node.Name, "ip": node.Host, "status": "unhealthy", "version": node.Version})
	}

	ctx.JSON(http.StatusOK, gin.H{"nodes": nodes})
}
