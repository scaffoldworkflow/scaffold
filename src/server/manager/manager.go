package manager

import (
	"fmt"
	"net/http"
	"scaffold/server/api"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/mongodb"
	"scaffold/server/user"
	"time"
)

func Run() {
	mongodb.InitCollections()

	api.IsHealthy = true

	user.VerifyAdmin()
	auth.Nodes = make([]auth.NodeObject, 0)

	api.IsReady = true

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
		time.Sleep(time.Duration(config.Config.HeartbeatInterval) * time.Second)
	}
}
