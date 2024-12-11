package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/constants"
	scron "scaffold/server/cron"
	"scaffold/server/health"
	"scaffold/server/history"
	"scaffold/server/msg"
	"scaffold/server/proxy"
	"scaffold/server/rabbitmq"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/user"
	"scaffold/server/utils"
	"scaffold/server/workflow"
	"time"

	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"

	"github.com/gorilla/mux"
)

var toKill []string

type UINode struct {
	Status  string
	Name    string
	IP      string
	Version string
	Color   string
	Text    string
	Icon    string
}

func Run() {
	// r := http.NewServeMux()
	r := mux.NewRouter()
	// mux.Handle("/ws", websocket.Handler(run))
	r.HandleFunc("/{host}/{port}/{workflow}/{run}/{version}",
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

	ws, err := workflow.GetAllWorkflows()
	if err != nil {
		panic(err)
	}
	workflow.SetCache(ws)

	scron.Start()
}

func QueueDataReceive(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var m msg.RunMsg
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		logger.Errorf("", "Error processing queue message: %s", err.Error())
		return err
	}
	switch m.Status {
	case constants.STATE_STATUS_SUCCESS:
		logger.Debugf("", "Task %s has completed with status success", m.Task)
		if err := history.AddStateToHistory(m.RunID, m.State); err != nil {
			logger.Errorf("", "Error updating history: %s", err.Error())
			return err
		}
		stateChange(m.Workflow, m.Task, constants.STATE_STATUS_SUCCESS, m.Context, m.RunID)
		autoTrigger(m.Workflow, m.Task, constants.STATUS_TRIGGER_SUCCESS, m.Context, m.RunID)
		autoTrigger(m.Workflow, m.Task, constants.STATUS_TRIGGER_ALWAYS, m.Context, m.RunID)
	case constants.STATE_STATUS_ERROR:
		logger.Debugf("", "Task %s has completed with status error", m.Task)
		if err := history.AddStateToHistory(m.RunID, m.State); err != nil {
			logger.Errorf("", "Error updating history: %s", err.Error())
			return err
		}
		stateChange(m.Workflow, m.Task, constants.STATE_STATUS_ERROR, m.Context, m.RunID)
		autoTrigger(m.Workflow, m.Task, constants.STATUS_TRIGGER_ERROR, m.Context, m.RunID)
		autoTrigger(m.Workflow, m.Task, constants.STATUS_TRIGGER_ALWAYS, m.Context, m.RunID)
	case constants.STATE_STATUS_KILLED:
		logger.Debugf("", "Task %s has completed with status killed", m.Task)
		if err := history.AddStateToHistory(m.RunID, m.State); err != nil {
			logger.Errorf("", "Error updating history: %s", err.Error())
			return err
		}
		id := fmt.Sprintf("%s-%s", m.Workflow, m.Task)
		toKill = utils.Remove(toKill, id)
		if err := rabbitmq.KillPublish(map[string]string{"id": id}); err != nil {
			logger.Errorf("", "Error publishing kill id: %s", err.Error())
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
	// 	stateChange(m.Workflow, m.Task, constants.STATUS_TRIGGER_SUCCESS)
	// 	stateChange(m.Workflow, m.Task, constants.STATUS_TRIGGER_ALWAYS)
	// } else if m.Status == constants.STATE_STATUS_ERROR {
	// 	logger.Debugf("", "Task %s has completed with error", m.Task)
	// 	stateChange(m.Workflow, m.Task, constants.STATUS_TRIGGER_ERROR)
	// 	stateChange(m.Workflow, m.Task, constants.STATUS_TRIGGER_ALWAYS)
	// }
	return nil
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
						DoKill(s.Workflow, s.Task)
					case constants.STATE_STATUS_WAITING:
						DoKill(s.Workflow, s.Task)
					}
				}
			}
			n.Ping += 1
			auth.NodeLock.Lock()
			auth.Nodes[key] = n
			auth.NodeLock.Unlock()
		}
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
	}
}

func stateChange(cn, tn, status string, context map[string]string, runID string) error {
	ss, err := state.GetStateByNames(cn, tn)
	if err != nil {
		logger.Errorf("", "Cannot get state for %s", cn)
		return err
	}
	ss.Context = utils.MergeDict(ss.Context, context)
	if err := state.UpdateStateByNames(cn, tn, ss); err != nil {
		return err
	}
	ts, err := task.GetTasksByWorkflow(cn)
	if err != nil {
		logger.Errorf("", "Cannot change state for %s", cn)
		return err
	}
	switch status {
	case constants.STATE_STATUS_SUCCESS:
		for _, t := range ts {
			shouldExecute := false
			sss, err := state.GetStateByNames(cn, t.Name)
			if err != nil {
				return err
			}
			for _, n := range t.DependsOn.Always {

				if n == tn {
					sss.Context = utils.MergeDict(sss.Context, context)
					if err := state.UpdateStateByNames(cn, t.Name, sss); err != nil {
						return err
					}
					if t.AutoExecute {
						shouldExecute = true
						continue
					}
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_ERROR && s.Status != constants.STATE_STATUS_SUCCESS {
					continue
				}
			}
			for _, n := range t.DependsOn.Success {
				if n == tn {
					sss.Context = utils.MergeDict(sss.Context, context)
					if err := state.UpdateStateByNames(cn, t.Name, sss); err != nil {
						return err
					}
					if t.AutoExecute {
						shouldExecute = true
						continue
					}
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s == nil {
					continue
				}
				if s.Status != constants.STATE_STATUS_SUCCESS {
					continue
				}
			}
			if shouldExecute {
				if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
					return err
				}
			}
		}
	case constants.STATE_STATUS_ERROR:
		for _, t := range ts {
			shouldExecute := false
			sss, err := state.GetStateByNames(cn, t.Name)
			if err != nil {
				return err
			}
			for _, n := range t.DependsOn.Always {

				if n == tn {
					sss.Context = utils.MergeDict(sss.Context, context)
					if err := state.UpdateStateByNames(cn, t.Name, sss); err != nil {
						return err
					}
					if t.AutoExecute {
						shouldExecute = true
						continue
					}
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s.Status != constants.STATE_STATUS_ERROR && s.Status != constants.STATE_STATUS_SUCCESS {
					continue
				}
			}
			for _, n := range t.DependsOn.Error {
				if n == tn {
					sss.Context = utils.MergeDict(sss.Context, context)
					if err := state.UpdateStateByNames(cn, t.Name, sss); err != nil {
						return err
					}
					if t.AutoExecute {
						shouldExecute = true
						continue
					}
				}
				s, err := state.GetStateByNames(cn, n)
				if err != nil {
					return err
				}
				if s == nil {
					continue
				}
				if s.Status != constants.STATE_STATUS_ERROR {
					continue
				}
			}
			if shouldExecute {
				if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
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
					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
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
					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
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
					if err := stateChange(cn, t.Name, constants.STATE_STATUS_NOT_STARTED, context, runID); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func checkDeps(cn string, t *task.Task) (bool, error) {
	for _, n := range t.DependsOn.Success {
		s, err := state.GetStateByNames(cn, n)
		if err != nil {
			return false, err
		}
		if s.Status != constants.STATE_STATUS_SUCCESS {
			return false, nil
		}
	}
	for _, n := range t.DependsOn.Error {
		s, err := state.GetStateByNames(cn, n)
		if err != nil {
			return false, err
		}
		if s.Status != constants.STATE_STATUS_ERROR {
			return false, nil
		}
	}
	for _, n := range t.DependsOn.Always {
		s, err := state.GetStateByNames(cn, n)
		if err != nil {
			return false, err
		}
		if s.Status != constants.STATE_STATUS_SUCCESS && s.Status != constants.STATE_STATUS_ERROR {
			return false, nil
		}
	}
	return true, nil
}

func autoTrigger(cn, tn, status string, context map[string]string, runID string) error {
	logger.Debugf("", "Doing auto trigger for %s %s with status %s", cn, tn, status)
	ts, err := task.GetTasksByWorkflow(cn)
	if err != nil {
		logger.Errorf("", "Cannot perform auto trigger for %s", cn)
		return err
	}
	toTrigger := []string{}

	switch status {
	case constants.STATUS_TRIGGER_SUCCESS:
		for _, t := range ts {
			if utils.Contains(t.DependsOn.Success, tn) && t.AutoExecute {
				trigger, err := checkDeps(cn, t)
				if err != nil {
					logger.Errorf("", "Error checking dependency states: %s", err.Error())
					return err
				}
				if trigger {
					toTrigger = append(toTrigger, t.Name)
				}
			}
		}
	case constants.STATUS_TRIGGER_ERROR:
		for _, t := range ts {
			if utils.Contains(t.DependsOn.Error, tn) && t.AutoExecute {
				trigger, err := checkDeps(cn, t)
				if err != nil {
					logger.Errorf("", "Error checking dependency states: %s", err.Error())
					return err
				}
				if trigger {
					toTrigger = append(toTrigger, t.Name)
				}
			}
		}
	case constants.STATUS_TRIGGER_ALWAYS:
		for _, t := range ts {
			if utils.Contains(t.DependsOn.Always, tn) && t.AutoExecute {
				trigger, err := checkDeps(cn, t)
				if err != nil {
					logger.Errorf("", "Error checking dependency states: %s", err.Error())
					return err
				}
				if trigger {
					toTrigger = append(toTrigger, t.Name)
				}
			}
		}
	}
	for _, t := range toTrigger {
		if err := DoTrigger(cn, t, context, runID); err != nil {
			return err
		}
	}
	return nil
}

func DoTrigger(wn, tn string, context map[string]string, runID string) error {
	if runID == "" {
		runID = uuid.New().String()
		h := history.History{
			RunID:    runID,
			States:   make([]state.State, 0),
			Workflow: wn,
		}

		if err := history.CreateHistory(&h); err != nil {
			return err
		}
	}
	c, err := workflow.GetWorkflowByName(wn)
	if err != nil {
		return err
	}

	t, err := task.GetTaskByNames(wn, tn)
	if err != nil {
		return err
	}

	if t.Disabled {
		return nil
	}

	for _, s := range t.DependsOn.Success {
		ss, err := state.GetStateByNames(wn, s)
		if err != nil {
			return err
		}
		if ss.Status != constants.STATE_STATUS_SUCCESS {
			return nil
		}
	}
	for _, s := range t.DependsOn.Error {
		ss, err := state.GetStateByNames(wn, s)
		if err != nil {
			return err
		}
		if ss.Status != constants.STATE_STATUS_SUCCESS {
			return nil
		}
	}
	// for _, s := range t.DependsOn.Always {
	// 	ss, err := state.GetStateByNames(cn, s)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if ss.Status == constants.STATE_STATUS_NOT_STARTED {
	// 		return nil
	// 	}
	// }

	s, err := state.GetStateByNames(wn, tn)
	if err != nil {
		return err
	}
	s.Status = constants.STATE_STATUS_WAITING
	if err := state.UpdateStateByNames(wn, tn, s); err != nil {
		return err
	}
	if err := history.AddStateToHistory(runID, *s); err != nil {
		logger.Errorf("", "Error updating history: %s", err.Error())
		return err
	}

	// if err := DoKill(cn, tn); err != nil {
	// 	return err
	// }

	m := msg.TriggerMsg{
		Task:     tn,
		Workflow: wn,
		Action:   constants.ACTION_TRIGGER,
		Groups:   c.Groups,
		Number:   t.RunNumber + 1,
		RunID:    runID,
		Context:  context,
	}

	logger.Infof("", "Triggering run with message %v", m)
	return rabbitmq.ManagerPublish(m)
	// return bulwark.QueuePush(bulwark.WorkerClient, m)
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
			logger.Errorf("", "Encountered error killing run: %v", err)
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
	c, err := workflow.GetWorkflowByName(name)
	if err != nil {
		return err
	}
	for _, t := range c.Tasks {
		for _, i := range changed {
			if utils.Contains(utils.Keys(t.Inputs), i) {
				stateChange(name, t.Name, constants.STATE_STATUS_NOT_STARTED, map[string]string{}, "")
				// state.CopyStatesByNames(name, t.Name, fmt.Sprintf("SCAFFOLD_PREVIOUS-%s", t.Name))
				state.ClearStateByNames(name, t.Name, t.RunNumber)
				break
			}
		}
	}
	return nil
}

func GetStatus() (bool, []UINode) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddress := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddress.IP.String()

	nodes := make([]UINode, 0)
	managerStatus := "healthy"
	if !health.IsHealthy {
		managerStatus = "degraded"
	}
	toRemove := []string{}
	downCount := 0
	n := UINode{
		Name:    config.Config.Host,
		IP:      ip,
		Status:  constants.NODE_HEALTHY,
		Version: constants.VERSION,
		Color:   constants.UI_HEALTH_COLORS[managerStatus],
		Text:    constants.UI_HEALTH_TEXT[managerStatus],
		Icon:    constants.UI_HEALTH_ICONS[managerStatus],
	}
	nodes = append(nodes, n)
	for id, node := range auth.Nodes {
		if node.Ping < config.Config.PingHealthyThreshold {
			status := constants.NODE_HEALTHY
			n := UINode{
				Name:    node.Name,
				IP:      node.Host,
				Status:  constants.NODE_HEALTHY,
				Version: node.Version,
				Color:   constants.UI_HEALTH_COLORS[status],
				Text:    constants.UI_HEALTH_TEXT[status],
				Icon:    constants.UI_HEALTH_ICONS[status],
			}
			nodes = append(nodes, n)
			continue
		}
		if node.Ping < config.Config.PingUnknownThreshold {
			status := constants.NODE_UNKNOWN
			n := UINode{
				Name:    node.Name,
				IP:      node.Host,
				Status:  constants.NODE_HEALTHY,
				Version: node.Version,
				Color:   constants.UI_HEALTH_COLORS[status],
				Text:    constants.UI_HEALTH_TEXT[status],
				Icon:    constants.UI_HEALTH_ICONS[status],
			}
			nodes = append(nodes, n)
			downCount += 1
			continue
		}
		status := constants.NODE_UNHEALTHY
		n := UINode{
			Name:    node.Name,
			IP:      node.Host,
			Status:  constants.NODE_HEALTHY,
			Version: node.Version,
			Color:   constants.UI_HEALTH_COLORS[status],
			Text:    constants.UI_HEALTH_TEXT[status],
			Icon:    constants.UI_HEALTH_ICONS[status],
		}
		nodes = append(nodes, n)
		if node.Ping > config.Config.PingDownThreshold {
			toRemove = append(toRemove, id)
		}
		downCount += 1
	}

	auth.NodeLock.Lock()
	for _, id := range toRemove {
		delete(auth.Nodes, id)
	}
	auth.NodeLock.Unlock()

	return downCount == 0, nodes
}
