package middleware

import (
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/user"
	"scaffold/server/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// This middleware ensures that a request will be aborted with an error
// if the user is not logged in
func EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		var err error

		baUsername, baPassword, hasAuth := c.Request.BasicAuth()
		if hasAuth {
			_, err := user.GetUserByUsername(baUsername)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			verified, err := user.VerifyUser(baUsername, baPassword)
			if err != nil || !verified {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			return
		}

		authString := c.Request.Header.Get("Authorization")
		if authString == "" {
			token, err = c.Cookie("scaffold_token")
			if err != nil {
				c.Redirect(307, "/ui/login")
				return
			}
		} else {
			token = strings.Split(authString, " ")[1]
		}
		if token == config.Config.Node.PrimaryKey {
			return
		}

		usr, _ := user.GetUserByAPIToken(token)
		if usr != nil {
			return
		}

		usr, _ = user.GetUserByLoginToken(token)
		if usr == nil {
			if authString == "" {
				c.Redirect(307, "/ui/login")
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func EnsureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		var err error

		authString := c.Request.Header.Get("Authorization")
		if authString == "" {
			token, err = c.Cookie("scaffold_token")
			if err != nil {
				return
			}
		} else {
			token = strings.Split(authString, " ")[1]
		}

		if token == config.Config.Node.PrimaryKey {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		usr, _ := user.GetUserByAPIToken(token)
		if usr != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		usr, _ = user.GetUserByLoginToken(token)
		if usr != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// This middleware ensures that a request will be aborted with an error
// if the user is not logged in
func EnsureCascadeGroup(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		var err error
		isUI := false
		cascadeName := c.Param(paramName)
		cs, _ := cascade.GetCascadeByName(cascadeName)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if cs == nil {
			return
		}

		authString := c.Request.Header.Get("Authorization")
		if authString == "" {
			token, err = c.Cookie("scaffold_token")
			if err != nil {
				c.Redirect(http.StatusUnauthorized, "401.html")
				return
			}
			isUI = true
		} else {
			token = strings.Split(authString, " ")[1]
		}
		if token == config.Config.Node.PrimaryKey {
			return
		}

		usr, _ := user.GetUserByAPIToken(token)
		if usr != nil {
			if utils.Contains(usr.Groups, "admin") {
				return
			}
			if cs.Groups == nil {
				return
			}
			for _, group := range cs.Groups {
				if utils.Contains(usr.Groups, group) {
					return
				}
			}
			if isUI {
				c.Redirect(http.StatusUnauthorized, "401.html")
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		usr, _ = user.GetUserByLoginToken(token)
		if usr != nil {
			if utils.Contains(usr.Groups, "admin") {
				return
			}
			for _, group := range cs.Groups {
				if utils.Contains(usr.Groups, group) {
					return
				}
			}
			if isUI {
				c.Redirect(http.StatusUnauthorized, "401.html")
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if isUI {
			c.Redirect(http.StatusUnauthorized, "401.html")
			return
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func EnsureSelf() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")

		var token string
		var err error

		baUsername, baPassword, hasAuth := c.Request.BasicAuth()
		if hasAuth {
			if baUsername != username {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			_, err := user.GetUserByUsername(baUsername)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			verified, err := user.VerifyUser(baUsername, baPassword)
			if err != nil || !verified {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			return
		}

		authString := c.Request.Header.Get("Authorization")
		if authString == "" {
			token, err = c.Cookie("scaffold_token")
			if err != nil {
				c.Redirect(307, "/ui/login")
				return
			}
		} else {
			token = strings.Split(authString, " ")[1]
		}
		if token == config.Config.Node.PrimaryKey {
			return
		}

		usr, _ := user.GetUserByAPIToken(token)
		if usr != nil {
			if usr.Username == username || StringSliceContains(usr.Groups, "admin") || StringSliceContains(usr.Roles, "admin") {
				return
			}
		}

		usr, _ = user.GetUserByLoginToken(token)
		if usr == nil {
			if authString == "" {
				c.Redirect(307, "/ui/login")
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if usr.Username == username || StringSliceContains(usr.Groups, "admin") || StringSliceContains(usr.Roles, "admin") {
			return
		}
		if authString == "" {
			c.Redirect(307, "/ui/login")
			return
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func EnsureGroupsAllowed(groups []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var usr *user.User
		authString := c.Request.Header.Get("Authorization")
		found := false
		if authString == "" {
			token, err := c.Cookie("scaffold_token")
			if err == nil {
				usr, err = user.GetUserByLoginToken(token)
				if err != nil {
					found = true
				}
			}
		} else {
			if !found {
				token := strings.Split(authString, " ")[1]
				if token == config.Config.Node.PrimaryKey {
					return
				}
				var err error
				usr, err = user.GetUserByAPIToken(token)
				if err != nil {
					c.AbortWithStatus(http.StatusUnauthorized)
				}
			}
		}
		if usr != nil {
			for _, group := range usr.Groups {
				if StringSliceContains(groups, group) {
					return
				}
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func EnsureRolesAllowed(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var usr *user.User
		var err error

		authString := c.Request.Header.Get("Authorization")
		if authString == "" {
			token, err := c.Cookie("scaffold_token")
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			usr, err = user.GetUserByLoginToken(token)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		} else {
			token := strings.Split(authString, " ")[1]
			if token == config.Config.Node.PrimaryKey {
				return
			}
			usr, err = user.GetUserByAPIToken(token)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		if usr != nil {
			for _, role := range usr.Roles {
				if StringSliceContains(roles, role) {
					return
				}
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func EnsureBasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func StringSliceContains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
