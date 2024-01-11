// API implements worker and manager API endpoints for Scaffold functionality
package api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"scaffold/server/config"
	"scaffold/server/user"
	"scaffold/server/utils"

	logger "github.com/jfcarter2358/go-logger"
)

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
			logger.Tracef("", "Scaffold token was empty")
			return false
		}
		usr, _ = user.GetUserByLoginToken(token)
		if usr == nil {
			logger.Tracef("", "No user exists with login token")
			return false
		}
	} else {
		logger.Tracef("", "Auth header detected")
		token = strings.Split(authString, " ")[1]

		// Is the request coming from a node itself?
		if token == config.Config.Node.PrimaryKey {
			logger.Tracef("", "Primary key detected")
			return true
		}

		// Get the user via the information
		usr, _ = user.GetUserByAPIToken(token)
		if usr == nil {
			logger.Errorf("", "Unable to get user by API token")
			return false
		}
	}

	if utils.Contains(usr.Groups, "admin") {
		logger.Tracef("", "User is part of admin group")
		return true
	}
	for _, group := range groups {
		logger.Tracef("", "Checking %v against %s", usr.Groups, group)
		if utils.Contains(usr.Groups, group) {
			return true
		}
	}

	return false
}
