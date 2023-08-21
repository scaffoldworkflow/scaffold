// API implements worker and manager API endpoints for Scaffold functionality
package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/logger"
	"scaffold/server/user"
	"scaffold/server/utils"
)

func getAvailableNode() (*auth.NodeObject, error) {
	if len(auth.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes to schedule runs on")
	}
	nodeIdx := auth.LastScheduledIdx + 1

	for idx, n := range auth.Nodes {
		queryURL := fmt.Sprintf("%s://%s:%d/health/available", n.Protocol, n.Host, n.Port)
		resp, err := http.Get(queryURL)
		if err != nil || resp.StatusCode >= 400 {
			continue
		}
		nodeIdx = idx
		break
	}
	if nodeIdx >= len(auth.Nodes) {
		nodeIdx = 0
	}
	auth.LastScheduledIdx = nodeIdx

	return &auth.Nodes[nodeIdx], nil
}

func validateUserGroup(ctx *gin.Context, groups []string) bool {
	var token string
	var err error
	var usr *user.User

	logger.Infof("", "Validating user against groups: %v", groups)

	if len(groups) == 0 {
		return true
	}

	// Check if we have an auth header
	authString := ctx.Request.Header.Get("Authorization")
	if authString == "" {
		// Check if the request is coming from a logged in UI user
		token, err = ctx.Cookie("scaffold_token")
		if err != nil {
			return false
		}
		usr, _ = user.GetUserByLoginToken(token)
		if usr == nil {
			return false
		}
	} else {
		token = strings.Split(authString, " ")[1]
	}

	// Is the request coming from a node itself?
	if token == config.Config.Node.PrimaryKey {
		return true
	}

	// Get the user via the information
	usr, _ = user.GetUserByAPIToken(token)
	if usr == nil {
		return false
	}

	if utils.Contains(usr.Groups, "admin") {
		return true
	}
	for _, group := range groups {
		if utils.Contains(usr.Groups, group) {
			return true
		}
	}

	return false
}
