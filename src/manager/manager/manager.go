package manager

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"scaffold/manager/auth"
	"scaffold/manager/config"
	"scaffold/manager/constants"
	"scaffold/manager/health"
	"scaffold/manager/monitor"
	"scaffold/manager/release"
	"scaffold/manager/run"
	"scaffold/manager/user"
	"time"

	logger "github.com/jfcarter2358/go-logger"
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
	health.IsHealthy = true

	if err := user.VerifyAdmin(); err != nil {
		logger.Fatalf("", "Unable to create admin user: %s", err.Error())
	}
	auth.Nodes = make(map[string]auth.NodeObject)

	health.IsReady = true

	go healthCheck()
	go run.AutoTrigger()
	go release.PromoteAutoTrigger()
	go PruneJobs()
}

func healthCheck() {
	for {
		for key, n := range auth.Nodes {
			// if n.Ping > config.Config.HeartbeatBackoff {
			// 	ss, err := state.GetStatesByWorker(n.Name)
			// 	if err != nil {
			// 		logger.Errorf("", "Unable to get states by worker: %s", n.Name)
			// 	}
			// 	for _, s := range ss {
			// 		switch s.Status {
			// 		case constants.STATE_STATUS_RUNNING:
			// 			DoKill(s.Workflow, s.Task)
			// 		case constants.STATE_STATUS_WAITING:
			// 			DoKill(s.Workflow, s.Task)
			// 		}
			// 	}
			// }
			n.Ping += 1
			auth.NodeLock.Lock()
			auth.Nodes[key] = n
			auth.NodeLock.Unlock()
		}
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Millisecond)
	}
}

func GetStatus() (bool, []UINode) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Fatalf("", "Got error getting DNS: %s", err)
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

func PruneJobs() {
	for {
		// TODO: Make this sleep configurable
		time.Sleep(1 * time.Second)
		logger.Debugf("", "Getting jobs in namespace %s", config.Config.K8sNamespace)
		jobOut, jobErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl get job -n %s -o json", config.Config.K8sNamespace)).CombinedOutput()
		if jobErr != nil {
			logger.Errorf("", "Error getting jobs: %s", jobErr.Error())
			logger.Debugf("", "%s", string(jobOut))
			continue
		}
		var js monitor.Jobs
		if err := json.Unmarshal(jobOut, &js); err != nil {
			logger.Errorf("", "Error loading job JSON: %s", err.Error())
			logger.Debugf("", "%s", jobOut)
			continue
		}

		now := time.Now()

		for _, j := range js.Items {
			logger.Tracef("", "Completion time for job %s: %s", j.Metadata.Name, j.Status.CompletionTime)
			if j.Status.CompletionTime != "" {
				finished, err := time.Parse(time.RFC3339, j.Status.CompletionTime)
				if err != nil {
					logger.Errorf("", "Cannot parse job completion time %s: %s", j.Status.CompletionTime, err)
				}
				diff := now.Sub(finished)
				if diff.Seconds() > float64(config.Config.RunPruneDuration) {
					logger.Debugf("", "Deleting job %s in namespace %s", j.Metadata.Name, config.Config.K8sNamespace)
					deleteOut, deleteErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl delete job -n %s %s", config.Config.K8sNamespace, j.Metadata.Name)).CombinedOutput()
					if deleteErr != nil {
						logger.Errorf("", "Error deleting job %s: %s", j.Metadata.Name, deleteErr)
						logger.Debugf("", "%s", string(deleteOut))
						continue
					}
				}
				continue
			}
			if len(j.Status.Conditions) > 0 {
				removed := false
				for _, c := range j.Status.Conditions {
					if c.Type == "Failed" {

						logger.Debugf("", "Found failed job %s", j.Metadata.Name)
						finished, err := time.Parse(time.RFC3339, c.LastTransitionTime)
						if err != nil {
							logger.Errorf("", "Cannot parse job last transition time %s: %s", c.LastTransitionTime, err)
						}

						diff := now.Sub(finished)
						if diff.Seconds() > float64(config.Config.RunPruneDuration) {
							logger.Debugf("", "Deleting job %s in namespace %s", j.Metadata.Name, config.Config.K8sNamespace)
							deleteOut, deleteErr := exec.Command("/bin/sh", "-c", fmt.Sprintf("kubectl delete job -n %s %s", config.Config.K8sNamespace, j.Metadata.Name)).CombinedOutput()
							if deleteErr != nil {
								logger.Errorf("", "Error deleting job %s: %s", j.Metadata.Name, deleteErr)
								logger.Debugf("", "%s", string(deleteOut))
								continue
							}
							removed = true
						}
						break
					}
				}
				if removed {
					continue
				}
			}
			logger.Debugf("", "Job %s has not yet completed", j.Metadata.Name)
		}
	}
}
