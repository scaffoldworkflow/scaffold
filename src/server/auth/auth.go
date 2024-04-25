package auth

import (
	"crypto/tls"
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/user"
	"scaffold/server/utils"
	"strings"
	"sync"
	"time"

	logger "github.com/jfcarter2358/go-logger"

	"encoding/base64"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

type PasswordResetObject struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Email           string `json:"email"`
}

type NodeJoinObject struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	WSPort   int    `json:"ws_port"`
	Protocol string `json:"protocol"`
	JoinKey  string `json:"join_key"`
	Version  string `json:"version"`
}

type NodeObject struct {
	Name      string `json:"name" bson:"name"`
	Host      string `json:"host" bson:"host"`
	Port      int    `json:"port" bson:"port"`
	WSPort    int    `json:"ws_port" bson:"ws_port"`
	Protocol  string `json:"protocol"`
	Healthy   bool   `json:"healthy" bson:"healthy"`
	Available bool   `json:"available" bson:"available"`
	Version   string `json:"version" bson:"version"`
	Ping      int    `json:"ping" bson:"ping"`
}

var Nodes = make(map[string]NodeObject)
var LastScheduledIdx = 0
var NodeLock = &sync.RWMutex{}

func PerformLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	rememberMe := c.PostForm("remember_me")

	if username == "" || password == "" {
		if val, ok := c.Request.Header["Authorization"]; ok {
			authToken := strings.Split(val[0], " ")[1]
			authTokenBytes, _ := base64.StdEncoding.DecodeString(authToken)
			authData := strings.Split(string(authTokenBytes), ":")

			username = authData[0]
			password = authData[1]

			valid, _ := user.VerifyUser(username, password)
			if valid {
				token := utils.GenerateToken(32)
				c.SetCookie("scaffold_token", token, 3600, "", "", false, false)
				u, _ := user.GetUserByUsername(username)
				hashedToken, err := HashAndSalt([]byte(token))
				if err != nil {
					c.HTML(http.StatusBadRequest, "login.html", gin.H{
						"ErrorTitle":   "Login Failed",
						"ErrorMessage": err.Error()})
				}
				u.LoginToken = hashedToken
				user.UpdateUserByUsername(username, u)
				return
			}
		}
	} else {
		valid, err := user.VerifyUser(username, password)
		if valid {
			token := utils.GenerateToken(32)
			if rememberMe == "on" {
				c.SetCookie("scaffold_token", token, 604800, "", "", false, false)
			} else {
				c.SetCookie("scaffold_token", token, 3600, "", "", false, false)
			}

			u, _ := user.GetUserByUsername(username)
			hashedToken, err := HashAndSalt([]byte(token))
			if err != nil {
				c.HTML(http.StatusBadRequest, "login.html", gin.H{
					"ErrorTitle":   "Login Failed",
					"ErrorMessage": err.Error()})
			}
			u.LoginToken = hashedToken
			user.UpdateUserByUsername(username, u)

			c.Redirect(302, "/")
			return
		} else {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": err.Error()})
		}
	}
	c.AbortWithStatus(http.StatusUnauthorized)
}

func PerformLogout(c *gin.Context) {
	token, err := c.Cookie("scaffold_token")

	if err != nil {
		u, err := user.GetUserByLoginToken(token)
		if err != nil {
			u.LoginToken = ""
			user.UpdateUserByUsername(u.Username, u)
		}
	}
	c.SetCookie("scaffold_token", "", -1, "", "", false, true)

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func RequestPasswordReset(c *gin.Context) {
	email := c.PostForm("email")

	u, _ := user.GetUserByEmail(email)

	if u != nil {
		token := utils.GenerateToken(32)
		currentTime := time.Now()

		hashedToken, err := HashAndSalt([]byte(token))
		if err != nil {
			panic(err)
		}
		u.ResetToken = hashedToken
		u.ResetTokenCreated = currentTime.Format("2006-01-02 15:04:05")
		user.UpdateUserByUsername(u.Username, u)

		m := gomail.NewMessage()

		// Set E-Mail sender
		m.SetHeader("From", config.Config.Reset.Email)

		// Set E-Mail receivers
		m.SetHeader("To", email)

		// Set E-Mail subject
		m.SetHeader("Subject", "Scaffold Password Reset")

		// Set E-Mail body. You can set plain text or html with text/html
		m.SetBody("text/html", "<p>To reset your Scaffold account password, click on the following link or paste it into your browser:</p><br><a href=\""+config.Config.BaseURL+"/ui/reset_password/"+token+"\">"+config.Config.BaseURL+"/ui/reset_password/"+token+"</a><br>This link will expire after 24 hours.")

		// Settings for SMTP server
		d := gomail.NewDialer(config.Config.Reset.Host, config.Config.Reset.Port, config.Config.Reset.Email, config.Config.Reset.Password)

		// This is only needed when SSL/TLS certificate is not valid on server.
		// In production this should be set to false.
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		// Now send E-Mail
		if err := d.DialAndSend(m); err != nil {
			logger.Fatal("", err.Error())
		}
		c.Redirect(302, "/ui/email_success")
	} else {
		c.Redirect(302, "/ui/email_failure")
	}
}

func DoPasswordReset(c *gin.Context) {
	r := PasswordResetObject{}

	err := c.ShouldBindJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if r.Password != r.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
	}
	u, err := user.GetUserByEmail(r.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u.Password, err = HashAndSalt([]byte(r.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	user.UpdateUserByUsername(u.Username, u)
	c.JSON(http.StatusOK, gin.H{})
}

func HashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", nil
	}
	return string(hash), nil
}

func JoinNode(ctx *gin.Context) {
	var n NodeJoinObject
	if err := ctx.ShouldBindJSON(&n); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if n.JoinKey == config.Config.Node.JoinKey {
		ipAddr := ctx.ClientIP()
		logger.Debugf("", "Joining node %s, %d, %d", ipAddr, n.Port, n.WSPort)
		if nd, ok := Nodes[n.Name]; ok {
			nd.Ping = 0
			Nodes[n.Name] = nd
			ctx.Status(http.StatusOK)
			return
		}

		Nodes[n.Name] = NodeObject{
			Name:     n.Name,
			Host:     ipAddr,
			Port:     n.Port,
			WSPort:   n.WSPort,
			Healthy:  true,
			Version:  n.Version,
			Protocol: n.Protocol,
			Ping:     0,
		}
		ctx.Status(http.StatusOK)
		return
	}

	ctx.Status(http.StatusUnauthorized)
}

func GetAllGroups() ([]string, error) {
	groups := []string{}

	cascades, err := cascade.GetAllCascades()
	if err != nil {
		return []string{}, err
	} else {
		for _, c := range cascades {
			for _, group := range c.Groups {
				if !utils.Contains(groups, group) {
					groups = append(groups, group)
				}
			}
		}
	}
	users, err := user.GetAllUsers()
	if err != nil {
		return []string{}, err
	} else {
		for _, u := range users {
			for _, group := range u.Groups {
				if !utils.Contains(groups, group) {
					groups = append(groups, group)
				}
			}
		}
	}

	return groups, nil
}

func GetAllRoles() []string {
	return []string{"read", "write", "admin"}
}
