package middleware

import (
	"net/http"
	"scaffold/manager/config"
	"scaffold/manager/user"
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

func EnsureLoggedInAPI() gin.HandlerFunc {
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
				c.AbortWithStatus(http.StatusForbidden)
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
				c.AbortWithStatus(http.StatusForbidden)
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

func StringSliceContains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
