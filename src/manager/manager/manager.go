package manager

import (
	"log"
	"net"
	"scaffold/manager/auth"
	"scaffold/manager/config"
	"scaffold/manager/constants"
	"scaffold/manager/health"
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
