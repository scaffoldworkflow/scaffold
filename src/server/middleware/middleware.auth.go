package middleware

import (
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/user"
	"strings"

	"github.com/gin-gonic/gin"
)

// This middleware ensures that a request will be aborted with an error
// if the user is not logged in
func EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		var err error

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

func EnsureSelf() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")

		var token string
		var err error

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

func EnsureGroupsAllowedByCascade() gin.HandlerFunc {
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

		name := c.Param("cascade")
		cs, err := cascade.GetCascadeByName(name)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}

		if len(cs.Groups) == 0 {
			return
		}

		if usr != nil {
			for _, group := range usr.Groups {
				if StringSliceContains(cs.Groups, group) {
					return
				}
			}
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
			for _, role := range usr.Roles {
				if StringSliceContains(roles, role) {
					return
				}
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
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
