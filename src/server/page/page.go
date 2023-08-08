package page

import (
	"encoding/json"
	"fmt"
	"net/http"
	"scaffold/server/auth"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/logger"
	"scaffold/server/user"
	"scaffold/server/utils"
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
	token, _ := c.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	if u == nil {
		c.Redirect(http.StatusTemporaryRedirect, "401.html")
		return
	}

	cascadePointers, _ := cascade.GetAllCascades()
	cascades := make([]map[string]interface{}, len(cascadePointers))
	for idx, obj := range cascadePointers {
		objBytes, _ := json.Marshal(obj)
		var data map[string]interface{}
		json.Unmarshal(objBytes, &data)
		isInGroup := false
		for _, ug := range u.Groups {
			if ug == "admin" {
				isInGroup = true
				break
			}
			for _, cg := range obj.Groups {
				if ug == cg {
					isInGroup = true
					break
				}
			}
			if isInGroup {
				break
			}
		}
		data["in_group"] = isInGroup
		cascades[idx] = data
	}

	showPage(c, "cascades.html", gin.H{"cascades": cascades})
}

func ShowCascadePage(c *gin.Context) {
	token, _ := c.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	if u == nil {
		c.Redirect(http.StatusTemporaryRedirect, "401.html")
		return
	}

	name := c.Param("name")
	obj, err := cascade.GetCascadeByName(name)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "500.html")
		return
	}

	isInGroup := false
	for _, ug := range u.Groups {
		fmt.Printf("Checking user group %s\n", ug)
		if ug == "admin" {
			isInGroup = true
			break
		}
		for _, cg := range obj.Groups {
			fmt.Printf("Checking cascade group: %s\n", cg)
			if ug == cg {
				fmt.Println("In group!")
				isInGroup = true
				break
			}
		}
		if isInGroup {
			break
		}
	}

	if !isInGroup {
		c.Redirect(http.StatusTemporaryRedirect, "401.html")
		return
	}

	showPage(c, "cascade.html", gin.H{"cascade": *obj})
}

func ShowFilesPage(c *gin.Context) {
	token, _ := c.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	if u == nil {
		c.Redirect(http.StatusTemporaryRedirect, "401.html")
		return
	}

	objects := []filestore.ObjectMetadata{}
	fileMetadata, err := filestore.ListObjects()
	if err != nil {
		logger.Errorf("", "Unable to get filestore objects: %s", err.Error())
		utils.DynamicAPIResponse(c, "500.html", http.StatusInternalServerError, gin.H{})
	}

	cascadeList := []string{}
	inGroups := map[string]bool{}

	datastores, _ := datastore.GetAllDataStores()
	cascades, _ := cascade.GetAllCascades()
	for idx, d := range datastores {
		cascadeList = append(cascadeList, d.Name)

		for _, f := range d.Files {
			path := fmt.Sprintf("%s/%s", d.Name, f)
			fm := filestore.ObjectMetadata{
				Name:     f,
				Cascade:  d.Name,
				Modified: fileMetadata[path].Modified,
			}
			objects = append(objects, fm)
		}
		isInGroup := false
		for _, ug := range u.Groups {
			if ug == "admin" {
				isInGroup = true
				break
			}
			for _, cg := range cascades[idx].Groups {
				if ug == cg {
					isInGroup = true
					break
				}
			}
			if isInGroup {
				break
			}
		}
		inGroups[d.Name] = isInGroup
	}

	showPage(c, "files.html", gin.H{"objects": objects, "cascades": cascadeList, "in_groups": inGroups})
}

func ShowUsersPage(c *gin.Context) {
	token, _ := c.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	if u == nil {
		c.Redirect(http.StatusTemporaryRedirect, "401.html")
		return
	}

	isAdmin := false
	if utils.Contains(u.Groups, "admin") || utils.Contains(u.Roles, "admin") {
		isAdmin = true
	}

	groups, _ := auth.GetAllGroups()
	roles := auth.GetAllRoles()

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

	showPage(c, "users.html", gin.H{"users": users, "is_admin": isAdmin, "admin_username": config.Config.Admin.Username, "groups": groups, "roles": roles})
}

func ShowUserPage(c *gin.Context) {
	username := c.Param("username")
	u, _ := user.GetUserByUsername(username)
	if u == nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{})
	}

	groupObj := make([]map[string]string, len(u.Groups))
	for idx, val := range u.Groups {
		groupObj[idx] = map[string]string{
			"value": val,
		}
	}
	groupJSON, _ := json.Marshal(groupObj)

	roleObj := make([]map[string]string, len(u.Roles))
	for idx, val := range u.Roles {
		roleObj[idx] = map[string]string{
			"value": val,
		}
	}
	roleJSON, _ := json.Marshal(roleObj)

	logger.Debugf("", "group json: %s", string(groupJSON))
	logger.Debugf("", "role json: %s", string(roleJSON))

	showPage(c, "user.html", gin.H{"user": &u, "role_tag_json": string(roleJSON), "group_tag_json": string(groupJSON)})
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
