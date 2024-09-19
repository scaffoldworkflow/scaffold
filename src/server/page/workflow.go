package page

import (
	"fmt"
	"net/http"
	"scaffold/server/ui"
	"scaffold/server/ui/breadcrumb"
	"scaffold/server/ui/button"
	"scaffold/server/ui/card"
	"scaffold/server/ui/collapse"
	"scaffold/server/ui/elements/br"
	"scaffold/server/ui/elements/div"
	"scaffold/server/ui/elements/h1"
	"scaffold/server/ui/elements/link"
	"scaffold/server/ui/elements/pre"
	"scaffold/server/ui/modal"
	"scaffold/server/ui/page"
	"scaffold/server/ui/sidebar"
	"scaffold/server/ui/topbar"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func WorkflowPageEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	markdown := workflowBuildPage(workflowName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildPage(workflowName string, ctx *gin.Context) []byte {
	wName := ctx.Param("name")
	p := page.Page{
		ID:             "page",
		SidebarEnabled: true,
		Sidebar: sidebar.Sidebar{
			ID:      "sidebar",
			Classes: "theme-light",
			Components: []ui.Component{
				h1.H1{
					Contents: "Scaffold",
					Classes:  "scaffold-green",
				},
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
				Classes: "theme-dark rounded-md",
				Components: []ui.Component{
					div.Div{
						Classes: "scaffold-green rounded-md",
						Style:   "padding-bottom:1px;",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Workflows",
										HRef:  "/ui/workflows",
									},
									link.Link{
										Title: wName,
										HRef:  fmt.Sprintf("/ui/workflows/%s", wName),
									},
								},
								Style: "margin-left:16px;",
							},
							button.Button{
								ID:      "auto_execute_button",
								OnClick: "auto_execute_modal.showModal()",
								Title:   `Auto Execute&nbsp;<i class="fa-solid fa-play"></i>`,
								Style:   "float:right;display:inline-block;margin-right:8px;margin-top:-32px;margin-bottom:8px;",
								Classes: "theme-base",
							},
							button.Button{
								ID:      "inputs_button",
								OnClick: "inputs_modal.showModal()",
								Title:   `Inputs&nbsp;<i class="fa-solid fa-pencil"></i>`,
								Style:   "float:right;display:inline-block;margin-right:8px;margin-top:-32px;margin-bottom:8px;",
								Classes: "theme-base",
							},
							// button.Button{
							// 	ID:      "file_button",
							// 	OnClick: "file_modal.showModal()",
							// 	Title:   `Files&nbsp;<i class="fa-solid fa-file"></i>`,
							// 	Style:   "float:right;display:inline-block;margin-right:8px;margin-top:-32px;margin-bottom:8px;",
							// 	Classes: "theme-base",
							// },
							ui.Raw{
								HTMLString: `<input id="search" class="w3-input w3-round search-bar dark theme-light" type="text" name="search" placeholder="Search Tasks" style="margin-left:5%;margin-top:8px;margin-bottom:8px;width:90%;" oninput="render()">`,
							},
						},
					},
					card.Card{
						ID:      "workflow-card",
						Classes: "theme-light",
						Style:   "width:100%;padding:0px;height:100%;",
					},
					br.BR{},
					modal.Modal{
						ID: "auto_execute_modal",
						Components: []ui.Component{
							h1.H1{
								Contents: `<h1 id="current-auto-execute-header" class="text-3xl" style="float:left;padding-top:8px;">Auto Execute</h1>`,
								Classes:  "scaffold-green rounded-md",
								Style:    "width:100%;",
							},
							br.BR{},
							br.BR{},
							br.BR{},
							div.Div{
								ID: "auto-execute-div",
							},
						},
					},
					modal.Modal{
						ID: "inputs_modal",
						Components: []ui.Component{
							h1.H1{
								Contents: `
								<h1 id="current-inputs-header" class="text-3xl" style="float:left;padding-top:8px;">Inputs</h1>
								<span onclick="saveInputs()" class="w3-button dark theme-light w3-round" style="margin-top:5px;float:right;margin-right:8px;"><i class="fa-solid fa-floppy-disk" id="save-icon"></i></span>
								`,
								Classes: "scaffold-green rounded-md",
								Style:   "width:100%;",
							},
							br.BR{},
							br.BR{},
							br.BR{},
							div.Div{
								ID: "current-input-div",
							},
						},
					},
					modal.Modal{
						ID: "state_modal",
						Components: []ui.Component{
							h1.H1{
								Contents: `
								<h1 id="current-state-header" class="text-3xl" style="float:left;padding-top:8px;"></h1>
								<span onclick="killRun()" class="w3-button dark theme-light w3-round" style="margin-top:5px;float:right;margin-right:8px;"><i class="fa-solid fa-stop"></i></span>
								<span onclick="triggerRun()" class="w3-button dark theme-light w3-round" style="margin-top:5px;float:right;margin-right:8px;"><i class="fa-solid fa-play"></i></span>
								<span onclick="toggleDisable()" class="w3-button dark theme-light w3-round" style="margin-top:5px;float:right;margin-right:8px;"><i class="fa-solid fa-toggle-on" id="toggle-icon"></i></span>
								`,
								Classes: "scaffold-green rounded-md",
								Style:   "width:100%;",
							},
							br.BR{},
							br.BR{},
							br.BR{},
							collapse.Collapse{
								TitleID:      "state-status-collapse",
								Classes:      "theme-light",
								TitleClasses: "scaffold-green",
								Title:        "Status",
								Components: []ui.Component{
									br.BR{},
									div.Div{
										Components: []ui.Component{
											ui.Raw{
												HTMLString: `<span id="state-status"></span>`,
											},
											br.BR{},
											ui.Raw{
												HTMLString: `<span id="state-started"></span>`,
											},
											br.BR{},
											ui.Raw{
												HTMLString: `<span id="state-finished"></span>`,
											},
											br.BR{},
										},
									},
								},
							},
							br.BR{},
							collapse.Collapse{
								TitleID:      "state-context-collapse",
								Classes:      "theme-light",
								TitleClasses: "scaffold-green",
								Title:        "Context",
								Components: []ui.Component{
									br.BR{},
									div.Div{
										ID: "state-context",
									},
								},
							},
							br.BR{},
							collapse.Collapse{
								TitleID:      "state-display-collapse",
								Classes:      "theme-light",
								TitleClasses: "scaffold-green",
								Title:        "Display",
								Components: []ui.Component{
									div.Div{
										ID: "state-current-display-data",
									},
								},
							},
							br.BR{},
							collapse.Collapse{
								TitleID:      "state-output-collapse",
								Classes:      "theme-light",
								TitleClasses: "scaffold-green",
								Title:        "Output",
								Components: []ui.Component{
									pre.Pre{
										ID:    "state-output",
										Style: "font-family:monospace;overflow-x:scroll",
									},
								},
							},
						},
						BoxClasses: "max-w-none w-4/5",
					},
					ui.Raw{
						HTMLString: `
						<script src="/static/js/jquery-ui.min.js"></script>
						<script src="/static/js/dagre.min.js"></script>
						<script src="/static/js/workflow.js"></script>
						`,
					},
				},
				Style: "margin:64px;height:200%;padding-bottom:150px;",
			},
			br.BR{},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render workflow page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
