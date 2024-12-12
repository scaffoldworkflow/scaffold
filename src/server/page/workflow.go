package page

import (
	"fmt"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/state"
	"scaffold/server/utils"

	"github.com/jfcarter2358/ui"
	"github.com/jfcarter2358/ui/breadcrumb"
	"github.com/jfcarter2358/ui/button"
	"github.com/jfcarter2358/ui/card"
	"github.com/jfcarter2358/ui/collapse"
	"github.com/jfcarter2358/ui/elements/br"
	"github.com/jfcarter2358/ui/elements/div"
	"github.com/jfcarter2358/ui/elements/h1"
	"github.com/jfcarter2358/ui/elements/link"
	"github.com/jfcarter2358/ui/elements/pre"
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

type OutputData struct {
	Count int `json:"count"`
}

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
					Classes:  "ui-green",
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
				Classes: "theme-dark rounded-md",
				Components: []ui.Component{
					div.Div{
						Classes: "ui-green rounded-md",
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
								Classes:  "ui-green rounded-md",
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
								Classes: "ui-green rounded-md",
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
					// modal.Modal{
					// 	ID:         "state_modal",
					// 	Components: []ui.Component{},
					// 	BoxClasses: "max-w-none w-4/5",
					// },
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
								Classes: fmt.Sprintf("%s rounded-md", "ui-green"),
								Style:   "width:100%;",
							},
							br.BR{},
							br.BR{},
							br.BR{},
							collapse.Collapse{
								TitleID:      "state-status-collapse",
								Classes:      "theme-light",
								TitleClasses: "ui-green",
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
								TitleClasses: "ui-green",
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
								TitleClasses: "ui-green",
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
								TitleClasses: "ui-green",
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
						<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.5.1/jquery.min.js"></script>
						<script src="https://malsup.github.io/jquery.form.js"></script>
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

func WorkflowModalEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	taskName := ctx.Param("task")
	markdown := workflowBuildModal(workflowName, taskName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildModal(workflowName, taskName string, ctx *gin.Context) []byte {
	s, err := state.GetStateByNames(workflowName, taskName)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return []byte(fmt.Sprintf("Unable to access state for task %s in workflow %s", taskName, workflowName))
	}
	color := getStateColor(*s)
	d := modal.Modal{
		ID: "state_modal",
		Components: []ui.Component{
			h1.H1{
				Contents: `
					<h1 id="current-state-header" class="text-3xl" style="float:left;padding-top:8px;"></h1>
					<span onclick="killRun()" class="w3-button dark theme-light w3-round" style="margin-top:5px;float:right;margin-right:8px;"><i class="fa-solid fa-stop"></i></span>
					<span onclick="triggerRun()" class="w3-button dark theme-light w3-round" style="margin-top:5px;float:right;margin-right:8px;"><i class="fa-solid fa-play"></i></span>
					<span onclick="toggleDisable()" class="w3-button dark theme-light w3-round" style="margin-top:5px;float:right;margin-right:8px;"><i class="fa-solid fa-toggle-on" id="toggle-icon"></i></span>
				`,
				Classes: fmt.Sprintf("%s rounded-md", color),
				Style:   "width:100%;",
			},
			br.BR{},
			br.BR{},
			br.BR{},
			collapse.Collapse{
				TitleID:      "state-status-collapse",
				Classes:      "theme-light",
				TitleClasses: color,
				Title:        "Status",
				Components: []ui.Component{
					br.BR{},
					div.Div{
						Components: []ui.Component{
							ui.Raw{
								HTMLString: fmt.Sprintf(`<span id="state-status" hx-get="/htmx/workflow/%s/status/%s" hx-trigger="load, every 2s"></span>`, workflowName, taskName),
							},
							br.BR{},
							ui.Raw{
								HTMLString: fmt.Sprintf(`<span id="state-started" hx-get="/htmx/workflow/%s/started/%s" hx-trigger="load, every 2s"></span>`, workflowName, taskName),
							},
							br.BR{},
							ui.Raw{
								HTMLString: fmt.Sprintf(`<span id="state-finished" hx-get="/htmx/workflow/%s/finished/%s" hx-trigger="load, every 2s"></span>`, workflowName, taskName),
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
				TitleClasses: color,
				Title:        "Context",
				Components: []ui.Component{
					br.BR{},
					div.Div{
						ID:        "state-context",
						HXTrigger: "load",
						HXGet:     fmt.Sprintf("/htmx/workflow/%s/context/%s", workflowName, taskName),
					},
				},
			},
			br.BR{},
			collapse.Collapse{
				TitleID:      "state-display-collapse",
				Classes:      "theme-light",
				TitleClasses: color,
				Title:        "Display",
				Components: []ui.Component{
					div.Div{
						ID:        "state-current-display-data",
						HXTrigger: "load, every 2s",
						HXGet:     fmt.Sprintf("/htmx/workflow/%s/display/%s", workflowName, taskName),
					},
				},
			},
			br.BR{},
			collapse.Collapse{
				TitleID:      "state-output-collapse",
				Classes:      "theme-light",
				TitleClasses: color,
				Title:        "Output",
				Components: []ui.Component{
					ui.Raw{
						HTMLString: `
						<pre 
							id="state-output"
							style="font-family:monospace;overflow-x:scroll"
							hx-trigger="load, every 2s"
							hx-post="/htmx/workflow/%s/output/%s",
							hx-vals='js:{
								"count": jQuery("#state-output").html().length
							}'
							hx-swap="beforeend"
							</pre>
							`,
					},
				},
			},
		},
		BoxClasses: "max-w-none w-4/5",
	}
	html, err := d.Render()
	if err != nil {
		logger.Errorf("", "Cannot render dashboard table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func WorkflowStatusEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	taskName := ctx.Param("task")
	markdown := workflowBuildStatus(workflowName, taskName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildStatus(workflowName, taskName string, ctx *gin.Context) []byte {
	s, err := state.GetStateByNames(workflowName, taskName)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return []byte(fmt.Sprintf("Unable to access state for task %s in workflow %s", taskName, workflowName))
	}
	return []byte(s.Status)
}

func WorkflowStartedEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	taskName := ctx.Param("task")
	markdown := workflowBuildStarted(workflowName, taskName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildStarted(workflowName, taskName string, ctx *gin.Context) []byte {
	s, err := state.GetStateByNames(workflowName, taskName)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return []byte(fmt.Sprintf("Unable to access state for task %s in workflow %s", taskName, workflowName))
	}
	return []byte(s.Started)
}

func WorkflowFinishedEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	taskName := ctx.Param("task")
	markdown := workflowBuildFinished(workflowName, taskName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildFinished(workflowName, taskName string, ctx *gin.Context) []byte {
	s, err := state.GetStateByNames(workflowName, taskName)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return []byte(fmt.Sprintf("Unable to access state for task %s in workflow %s", taskName, workflowName))
	}
	return []byte(s.Finished)
}

func WorkflowContextEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	taskName := ctx.Param("task")
	markdown := workflowBuildContext(workflowName, taskName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildContext(workflowName, taskName string, ctx *gin.Context) []byte {
	s, err := state.GetStateByNames(workflowName, taskName)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return []byte(fmt.Sprintf("Unable to access state for task %s in workflow %s", taskName, workflowName))
	}
	return []byte(s.Finished)
}

func WorkflowDisplayEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	taskName := ctx.Param("task")
	markdown := workflowBuildDisplay(workflowName, taskName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildDisplay(workflowName, taskName string, ctx *gin.Context) []byte {
	s, err := state.GetStateByNames(workflowName, taskName)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return []byte(fmt.Sprintf("Unable to access state for task %s in workflow %s", taskName, workflowName))
	}
	html := ""
	for _, currentDisplay := range s.Display {
		color := getStateColor(*s)
		textColor := getStateTextColor(*s)
		html += buildDisplayTable(currentDisplay, color, textColor, ctx)
	}
	return []byte(html)
}

func buildDisplayTable(displayMap map[string]interface{}, color, textColor string, ctx *gin.Context) string {
	headers := []header.Header{}
	rows := [][]cell.Cell{}

	for _, h := range displayMap["header"].([]interface{}) {
		headers = append(headers, header.Header{
			Contents: h.(string),
			Classes:  "text-lg",
		})
	}

	for _, data := range displayMap["data"].([]interface{}) {
		row := []cell.Cell{}
		for _, datum := range data.([]interface{}) {
			row = append(row, cell.Cell{
				Contents: datum.(string),
			})
		}
		rows = append(rows, row)
	}

	c := card.Card{
		Classes: "theme-light theme-border-light",
		Components: []ui.Component{
			ui.Raw{
				HTMLString: fmt.Sprintf(`
            <header class="w3-container %s">
                <h4>%s</h4>
            </header>
        `, color, displayMap["name"].(string)),
			},
			table.Table{
				ID:            "workflows_table",
				Headers:       headers,
				Rows:          rows,
				Classes:       "theme-light",
				Style:         "width:100%;",
				HeaderClasses: "rounded-md ui-green",
			},
		},
	}

	html, err := c.Render()
	if err != nil {
		logger.Errorf("", "Cannot render display table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return ""
	}
	return html
}

func WorkflowOutputEndpoint(ctx *gin.Context) {
	workflowName := ctx.Param("name")
	taskName := ctx.Param("task")
	markdown := workflowBuildOutput(workflowName, taskName, ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowBuildOutput(workflowName, taskName string, ctx *gin.Context) []byte {
	s, err := state.GetStateByNames(workflowName, taskName)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
		return []byte(fmt.Sprintf("Unable to access state for task %s in workflow %s", taskName, workflowName))
	}

	var d OutputData
	if err := ctx.ShouldBindJSON(&d); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return []byte{}
	}
	output := s.Output
	output = output[d.Count:]
	return []byte(output)
}

func getStateColor(s state.State) string {
	switch s.Status {
	case constants.STATE_STATUS_ERROR:
		return fmt.Sprintf("ui-%s", constants.UI_COLORS[constants.NODE_ERROR])
	case constants.STATE_STATUS_KILLED:
		return fmt.Sprintf("ui-%s", constants.UI_COLORS[constants.NODE_KILLED])
	case constants.STATE_STATUS_NOT_STARTED:
		return fmt.Sprintf("ui-%s", constants.UI_COLORS[constants.NODE_NOT_DEPLOYED])
	case constants.STATE_STATUS_RUNNING:
		return fmt.Sprintf("ui-%s", constants.UI_COLORS[constants.NODE_RUNNING])
	case constants.STATE_STATUS_SUCCESS:
		return fmt.Sprintf("ui-%s", constants.UI_COLORS[constants.NODE_SUCCESS])
	case constants.STATE_STATUS_WAITING:
		return fmt.Sprintf("ui-%s", constants.UI_COLORS[constants.NODE_WAITING])
	}
	return fmt.Sprintf("ui-%s", constants.UI_COLORS[constants.NODE_UNKNOWN])
}

func getStateTextColor(s state.State) string {
	switch s.Status {
	case constants.STATE_STATUS_ERROR:
		return fmt.Sprintf("ui-text-%s", constants.UI_COLORS[constants.NODE_ERROR])
	case constants.STATE_STATUS_KILLED:
		return fmt.Sprintf("ui-text-%s", constants.UI_COLORS[constants.NODE_KILLED])
	case constants.STATE_STATUS_NOT_STARTED:
		return fmt.Sprintf("ui-text-%s", constants.UI_COLORS[constants.NODE_NOT_DEPLOYED])
	case constants.STATE_STATUS_RUNNING:
		return fmt.Sprintf("ui-text-%s", constants.UI_COLORS[constants.NODE_RUNNING])
	case constants.STATE_STATUS_SUCCESS:
		return fmt.Sprintf("ui-text-%s", constants.UI_COLORS[constants.NODE_SUCCESS])
	case constants.STATE_STATUS_WAITING:
		return fmt.Sprintf("ui-text-%s", constants.UI_COLORS[constants.NODE_WAITING])
	}
	return fmt.Sprintf("ui-text-%s", constants.UI_COLORS[constants.NODE_UNKNOWN])
}
