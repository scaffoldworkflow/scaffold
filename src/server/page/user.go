package page

import (
	"fmt"
	"net/http"
	"scaffold/server/ui"
	"scaffold/server/ui/breadcrumb"
	"scaffold/server/ui/button"
	"scaffold/server/ui/card"
	"scaffold/server/ui/elements/br"
	"scaffold/server/ui/elements/div"
	"scaffold/server/ui/elements/h1"
	"scaffold/server/ui/elements/link"
	"scaffold/server/ui/modal"
	"scaffold/server/ui/page"
	"scaffold/server/ui/sidebar"
	"scaffold/server/ui/table"
	"scaffold/server/ui/table/cell"
	"scaffold/server/ui/table/header"
	"scaffold/server/ui/topbar"
	"scaffold/server/user"
	"strings"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func UserPageEndpoint(ctx *gin.Context) {
	markdown := userBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func userBuildPage(ctx *gin.Context) []byte {
	username := ctx.Param("username")
	u, err := user.GetUserByUsername(username)
	if err != nil {
		logger.Errorf("", "Cannot render user page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}

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
				Classes: "scaffold-green",
				Components: []ui.Component{
					button.Button{
						ID:      "generate_api_token_button",
						OnClick: "api_token_modal.showModal()",
						Title:   `Generate API Token`,
						Classes: "theme-base rounded-md",
						Style:   "margin-right:16px;",
					},
					button.Button{
						ID:      "save_user_button",
						OnClick: "saveUser()",
						Title:   `Save User`,
						Classes: "theme-base rounded-md",
						Style:   "margin-right:16px;",
					},
				},
				Buttons: []ui.Component{
					link.Link{
						Title:   "Logout",
						HRef:    "/auth/logout",
						Classes: "rounded-md",
					},
				},
				MenuClasses: "theme-light",
			},
			div.Div{
				Classes: "theme-light rounded-md",
				Components: []ui.Component{
					div.Div{
						Classes: "scaffold-green rounded-md",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Users",
										HRef:  "/ui/users",
										Style: "margin-top:8px;",
									},
									link.Link{
										Title: username,
										HRef:  fmt.Sprintf("/ui/users/%s", username),
										Style: "margin-top:8px;",
									},
								},
								Style: "margin-left:16px;height:56px;",
							},
						},
					},
					br.BR{},
					ui.Raw{
						HTMLString: `<label style="padding-left:32px;padding-right:32px;">Username</label>`,
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input style="padding-left:32px;padding-right:32px;" class="w3-input w3-round theme-light" type="text" id="user-add-username" value="%s" disabled>`, username),
					},
					ui.Raw{
						HTMLString: `<label style="padding-left:32px;padding-right:32px;">Password</label>`,
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input style="padding-left:32px;padding-right:32px;" class="w3-input w3-round theme-light" type="password" id="user-add-password" value="%s">`, u.Password),
					},
					ui.Raw{
						HTMLString: `<label style="padding-left:32px;padding-right:32px;">Given Name</label>`,
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input style="padding-left:32px;padding-right:32px;" class="w3-input w3-round theme-light" type="text" id="user-add-given-name" value="%s">`, u.GivenName),
					},
					ui.Raw{
						HTMLString: `<label style="padding-left:32px;padding-right:32px;">Family Name</label>`,
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input style="padding-left:32px;padding-right:32px;" class="w3-input w3-round theme-light" type="text" id="user-add-family-name" value="%s">`, u.FamilyName),
					},
					ui.Raw{
						HTMLString: `<label style="padding-left:32px;padding-right:32px;">Email</label>`,
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input style="padding-left:32px;padding-right:32px;" class="w3-input w3-round dark theme-light" type="text" id="user-add-email" value="%s">`, u.Email),
					},
					ui.Raw{
						HTMLString: `<label style="padding-left:32px;padding-right:32px;">Groups</label>`,
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input style="padding-left:32px;padding-right:32px;" class="w3-input w3-round theme-light" type="text" id="group-tags" value="%s">`, strings.Join(u.Groups, ",")),
					},
					ui.Raw{
						HTMLString: `<label style="padding-left:32px;padding-right:32px;">Roles</label>`,
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input style="padding-left:32px;padding-right:32px;" class="w3-input w3-round theme-light" type="text" id="role-tags" value="%s">`, strings.Join(u.Roles, ",")),
					},
					br.BR{},
					h1.H1{
						Contents: `
						<h1 id="api-token-header" class="text-xl" style="float:left;padding-top:8px;padding-left:32px;">API Tokens</h1>
						`,
						Classes: "scaffold-green rounded-md",
						Style:   "width:100%;",
					},
					br.BR{},
					br.BR{},
					table.Table{
						ID: "api-token-table",
						Headers: []header.Header{
							{
								Classes:  "text-lg",
								Contents: "Name",
							},
							{
								Classes:  "text-lg",
								Contents: "Created",
							},
							{
								Contents: "",
								Classes:  "text-lg",
							},
						},
						Rows: func(u user.User) [][]cell.Cell {
							output := make([][]cell.Cell, 0)
							for _, token := range u.APITokens {
								r := []cell.Cell{
									{
										Contents: token.Name,
									},
									{
										Contents: token.Created,
									},
									{
										Contents: fmt.Sprintf(`<div class="icon"><i class="fa-solid fa-trash-can w3-large pointer-cursor" onclick="revokeAPIToken('%s')"></i></div>`, token.Name),
									},
								}
								output = append(output, r)
							}
							return output
						}(*u),
					},
					br.BR{},
				},
				Style: "margin:64px;",
			},
			br.BR{},
			ui.Raw{
				HTMLString: `
				<script src="/static/js/jquery-ui.min.js"></script>
				<script src="/static/js/user.js"></script>
				<script src="https://malsup.github.io/jquery.form.js"></script> 
				`,
			},
			modal.Modal{
				ID: "api_token_modal",
				Components: []ui.Component{
					h1.H1{
						Contents: `
						Generate API Token
						`,
						Classes: "scaffold-green rounded-md text-3xl",
						Style:   "width:100%;",
					},
					br.BR{},
					br.BR{},
					br.BR{},
					ui.Raw{
						HTMLString: `<label>Name</label>`,
					},
					ui.Raw{
						HTMLString: `<input class="w3-input w3-round theme-light" type="text" id="user-generate-api-token-name">`,
					},
					br.BR{},
					card.Card{
						Classes: "scaffold-green",
						Components: []ui.Component{
							div.Div{
								ID: "api-token-field",
							},
						},
					},
					br.BR{},
					br.BR{},
					button.Button{
						ID:      "do_generate_api_token_button",
						OnClick: "generateAPIToken()",
						Title:   `Generate`,
						Style:   "",
						Classes: "theme-base",
					},
				},
				Classes: "",
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
