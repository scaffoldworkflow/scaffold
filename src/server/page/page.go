package page

import (
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/user"
	"time"

	"github.com/gin-gonic/gin"
)

func RedirectIndexPage(c *gin.Context) {
	c.Redirect(301, "/ui/cascades")
}

func ShowLoginPage(c *gin.Context) {
	showPage(c, "login.html", gin.H{})
}

func ShowForgotPasswordPage(c *gin.Context) {
	showPage(c, "forgot_password.html", gin.H{})
}

func ShowEmailSuccessPage(c *gin.Context) {
	showPage(c, "email_success.html", gin.H{})
}

func ShowEmailFailurePage(c *gin.Context) {
	showPage(c, "email_failure.html", gin.H{})
}

func ShowResetPasswordPage(c *gin.Context) {
	resetToken := c.Param("reset_token")
	u, _ := user.GetUserByResetToken(resetToken)

	if u == nil {
		showPage(c, "reset_password.html", gin.H{"title": "Reset Password", "Email": "N/A", "InvalidToken": "Your password reset link is invalid or expired"})
	} else {
		t, err := time.Parse("2006-01-02 15:04:05", u.ResetTokenCreated)
		if err != nil {
			showPage(c, "reset_password.html", gin.H{"title": "Reset Password", "Email": "N/A", "InvalidToken": "Your password reset link is invalid or expired"})
		} else {
			currentTime := time.Now()
			difference := currentTime.Sub(t).Hours()
			if difference > 24 {
				showPage(c, "reset_password.html", gin.H{"title": "Reset Password", "Email": "N/A", "InvalidToken": "Your password reset link is invalid or expired"})
			} else {
				showPage(c, "reset_password.html", gin.H{"title": "Reset Password", "Email": u.Email})
			}
		}
	}
}

func ShowCascadesPage(c *gin.Context) {
	cascadePointers, _ := cascade.GetAllCascades()
	cascades := make([]cascade.Cascade, len(cascadePointers))
	for idx, obj := range cascadePointers {
		cascades[idx] = *obj
	}

	showPage(c, "cascades.html", gin.H{"cascades": cascades})
}

func ShowCascadePage(c *gin.Context) {
	name := c.Param("name")
	obj, _ := cascade.GetCascadeByName(name)

	showPage(c, "cascade.html", gin.H{"cascade": *obj})
}

func ShowUsersPage(c *gin.Context) {
	token, _ := c.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	isAdmin := false
	if u.Username == config.Config.Admin.Username {
		isAdmin = true
	}

	var users []user.User
	if isAdmin {
		userPointers, _ := user.GetAllUsers()
		users = make([]user.User, len(userPointers))
		for idx, obj := range userPointers {
			users[idx] = *obj
		}
	} else {
		users = []user.User{*u}
	}

	showPage(c, "users.html", gin.H{"users": users, "is_admin": isAdmin, "admin_username": config.Config.Admin.Username})
}

func ShowUserPage(c *gin.Context) {
	username := c.Param("username")
	u, _ := user.GetUserByUsername(username)
	if u == nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{})
	}

	showPage(c, "user.html", gin.H{"user": &u})
}

func showPage(c *gin.Context, page string, header gin.H) {
	token, _ := c.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	familyName := ""
	givenName := ""
	if u != nil {
		familyName = u.FamilyName
		givenName = u.GivenName
	}

	header["family_name"] = familyName
	header["given_name"] = givenName
	header["version"] = constants.VERSION

	render(c, header, page)
}

func render(c *gin.Context, data gin.H, templateName string) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		c.XML(http.StatusOK, data["payload"])
	default:
		c.HTML(http.StatusOK, templateName, data)
	}
}
