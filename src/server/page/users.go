package page

import (
	"net/http"
	"scaffold/server/user"
	"scaffold/server/utils"
	"sort"
	"strings"

	"github.com/jfcarter2358/ui"
	"github.com/jfcarter2358/ui/breadcrumb"
	"github.com/jfcarter2358/ui/button"
	"github.com/jfcarter2358/ui/elements/br"
	"github.com/jfcarter2358/ui/elements/div"
	"github.com/jfcarter2358/ui/elements/h1"
	"github.com/jfcarter2358/ui/elements/input"
	"github.com/jfcarter2358/ui/elements/label"
	"github.com/jfcarter2358/ui/elements/link"
	"github.com/jfcarter2358/ui/modal"
	"github.com/jfcarter2358/ui/page"
	"github.com/jfcarter2358/ui/sidebar"
	"github.com/jfcarter2358/ui/table"
	"github.com/jfcarter2358/ui/table/cell"
	"github.com/jfcarter2358/ui/table/header"
	"github.com/jfcarter2358/ui/topbar"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func UsersSearchEndpoint(ctx *gin.Context) {
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	users, err := user.GetAllUsers()
	if err != nil {
		logger.Errorf("", "Cannot render users page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	filtered := []user.User{}

	for _, u := range users {
		if strings.Contains(strings.ToLower(u.FamilyName), strings.ToLower(query)) || strings.Contains(strings.ToLower(u.GivenName), strings.ToLower(query)) {
			filtered = append(filtered, *u)
		}
	}

	markdown := usersBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func UsersTableEndpoint(ctx *gin.Context) {
	users, err := user.GetAllUsers()
	if err != nil {
		logger.Errorf("", "Cannot render users page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []user.User{}

	for _, u := range users {
		filtered = append(filtered, *u)
	}

	markdown := usersBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func UsersPageEndpoint(ctx *gin.Context) {
	markdown := usersBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func usersBuildPage(ctx *gin.Context) []byte {
	p := page.Page{
		ID:             "page",
		SidebarEnabled: true,
		Sidebar: sidebar.Sidebar{
			ID:      "sidebar",
			Classes: "theme-light",
			Components: []ui.Component{
				link.Link{
					Title: "Dashboard",
					HRef:  "/ui/dashboard",
				},
				link.Link{
					Title: "Runs",
					HRef:  "/ui/runs",
				},
				link.Link{
					Title: "Users",
					HRef:  "/ui/users",
				},
				link.Link{
					Title: "Workflows",
					HRef:  "/ui/workflows",
				},
			},
		},
		Components: []ui.Component{
			topbar.Topbar{
				Title:   "Scaffold",
				Classes: "ui-green",
				Buttons: []ui.Component{
					link.Link{
						Title:   "Logout",
						HRef:    "/auth/logout",
						Style:   "passing:12px;",
						Classes: "theme-dark rounded-md",
					},
				},
				MenuClasses: "theme-light",
			},
			div.Div{
				Classes: "theme-light rounded-md",
				Components: []ui.Component{
					div.Div{
						Classes: "ui-green rounded-md",
						Style:   "height:64px;",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Users",
										HRef:  "/ui/users",
									},
								},
								Style: "margin-left:16px;",
							},
							button.Button{
								ID:      "inputs_button",
								OnClick: "create_user_modal.showModal()",
								Title:   `Create User&nbsp;<i class="fa-solid fa-user-plus"></i>`,
								Style:   "float:right;display:inline-block;margin-right:8px;margin-top:-32px;margin-bottom:8px;",
								Classes: "theme-base",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input w3-round search-bar theme-light" type="text"
                            name="search" placeholder="Search Users"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%" hx-get="/htmx/users/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#users-table-div" />`,
					},
					div.Div{
						ID:        "users-table-div",
						HXTrigger: "load",
						HXGet:     "/htmx/users/table",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
			modal.Modal{
				ID: "create_user_modal",
				Components: []ui.Component{
					h1.H1{
						Contents: `<h1 id="create-user-header" class="text-3xl" style="float:left;padding-top:8px;">Create User</h1>`,
						Classes:  "ui-green rounded-md",
						Style:    "width:100%;",
					},
					br.BR{},
					br.BR{},
					label.Label{
						Contents: "Username",
					},
					input.Input{
						Classes: "w3-input w3-round theme-light",
						Type:    "text",
						ID:      "users-add-username",
					},
					label.Label{
						Contents: "Password",
					},
					input.Input{
						Classes: "w3-input w3-round theme-light",
						Type:    "password",
						ID:      "users-add-password",
					},
					label.Label{
						Contents: "Given Name",
					},
					input.Input{
						Classes: "w3-input w3-round theme-light",
						Type:    "text",
						ID:      "users-add-given-name",
					},
					label.Label{
						Contents: "Family Name",
					},
					input.Input{
						Classes: "w3-input w3-round theme-light",
						Type:    "text",
						ID:      "users-add-family-name",
					},
					label.Label{
						Contents: "Email",
					},
					input.Input{
						Classes: "w3-input w3-round theme-light",
						Type:    "text",
						ID:      "users-add-email",
					},
					label.Label{
						Contents: "Groups",
					},
					input.Input{
						Classes: "w3-input w3-round theme-light",
						Type:    "text",
						ID:      "users-add-groups",
					},
					label.Label{
						Contents: "Roles",
					},
					input.Input{
						Classes: "w3-input w3-round theme-light",
						Type:    "text",
						ID:      "users-add-roles",
					},
					button.Button{
						ID:      "do_add_user_button",
						OnClick: "addUser()",
						Title:   `Add User`,
						Style:   "",
						Classes: "theme-base",
					},
				},
			},
			ui.Raw{
				HTMLString: `
					<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.5.1/jquery.min.js"></script>
					<script src="/static/js/users.js"></script>
					`,
			},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render users page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func usersBuildTable(us []user.User, ctx *gin.Context) []byte {
	sort.Slice(us, func(i, j int) bool {
		return us[i].GivenName < us[j].GivenName
	})

	t := table.Table{
		ID: "users_table",
		Headers: []header.Header{
			{
				Contents: "Username",
				Classes:  "text-lg",
			},
			{
				Contents: "Given Name",
				Classes:  "text-lg",
			},
			{
				Contents: "Family Name",
				Classes:  "text-lg",
			},
			{
				Contents: "Email",
				Classes:  "text-lg",
			},
			{
				Contents: "Groups",
				Classes:  "text-lg",
			},
			{
				Contents: "Roles",
				Classes:  "text-lg",
			},
			{
				Contents: "Created",
				Classes:  "text-lg",
			},
			{
				Contents: "Updated",
				Classes:  "text-lg",
			},
			{
				Contents: "",
				Classes:  "text-lg",
			},
			{
				Contents: "",
				Classes:  "text-lg",
			},
		},
		Rows:          make([][]cell.Cell, 0),
		Classes:       "theme-light",
		Style:         "width:100%;",
		HeaderClasses: "rounded-md ui-green",
	}

	token, _ := ctx.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	isAdmin := false
	if utils.Contains(u.Groups, "admin") || utils.Contains(u.Roles, "admin") {
		isAdmin = true
	}

	if isAdmin {
		for _, uu := range us {
			groups := strings.Join(uu.Groups, ",")
			roles := strings.Join(uu.Roles, ",")
			r := []cell.Cell{
				{
					Contents: uu.Username,
				},
				{
					Contents: uu.GivenName,
				},
				{
					Contents: uu.FamilyName,
				},
				{
					Contents: uu.Email,
				},
				{
					Contents: groups,
				},
				{
					Contents: roles,
				},
				{
					Contents: uu.Created,
				},
				{
					Contents: uu.Updated,
				},
				{
					Contents: `<a href="/ui/users/` + uu.Username + `" class="table-link-link w3-right-align dark theme-text"
                    style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-link"></i>
                </a>`,
				},
				{
					Contents: `<div class="table-link-link w3-right-align dark theme-text" style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-trash" style="cursor:pointer;" onclick="deleteUser('` + uu.Username + `')"></i>
                </div>`,
				},
			}
			t.Rows = append(t.Rows, r)
		}
	} else {
		groups := strings.Join(u.Groups, ",")
		roles := strings.Join(u.Roles, ",")
		r := []cell.Cell{
			{
				Contents: u.Username,
			},
			{
				Contents: u.GivenName,
			},
			{
				Contents: u.FamilyName,
			},
			{
				Contents: u.Email,
			},
			{
				Contents: groups,
			},
			{
				Contents: roles,
			},
			{
				Contents: u.Created,
			},
			{
				Contents: u.Updated,
			},
			{
				Contents: `<a href="/ui/users/` + u.Username + `" class="table-link-link w3-right-align dark theme-text"
                    style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-link"></i>
                </a>`,
			},
			{
				Contents: `<div class="table-link-link w3-right-align dark theme-text" style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-trash" style="cursor:pointer;" onclick="deleteUser('` + u.Username + `')"></i>
                </div>`,
			},
		}
		t.Rows = append(t.Rows, r)
	}

	html, err := t.Render()
	if err != nil {
		logger.Errorf("", "Cannot render users table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
