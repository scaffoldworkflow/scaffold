package page

import (
	"fmt"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/monitor"
	"strings"

	"github.com/jfcarter2358/ui"
	"github.com/jfcarter2358/ui/breadcrumb"
	"github.com/jfcarter2358/ui/elements/br"
	"github.com/jfcarter2358/ui/elements/div"
	"github.com/jfcarter2358/ui/elements/h1"
	"github.com/jfcarter2358/ui/elements/link"
	"github.com/jfcarter2358/ui/page"
	"github.com/jfcarter2358/ui/sidebar"
	"github.com/jfcarter2358/ui/topbar"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func MonitorPageEndpoint(ctx *gin.Context) {
	markdown := monitorBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func monitorBuildPage(ctx *gin.Context) []byte {
	id := ctx.Param("id")

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
					Title: "Monitors",
					HRef:  "/ui/monitors",
				},
				link.Link{
					Title: "Runbooks",
					HRef:  "/ui/runbooks",
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
				Components: []ui.Component{
					link.Link{
						Title:   "Logout",
						HRef:    "/auth/logout",
						Style:   "padding:12px;",
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
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Monitors",
										HRef:  "/ui/monitors",
									},
									link.Link{
										Title: id,
										HRef:  fmt.Sprintf("/ui/monitors/%s", id),
									},
								},
								Style: "margin-left:16px;",
							},
							div.Div{
								ID:        "monitor_enabled_icon",
								Classes:   "theme-base rounded-md",
								Style:     "padding-left:12px;margin-right:64px;width:40px;float:right;margin-top:-34px;height:32px;padding-top:4px;",
								HXGet:     fmt.Sprintf("/htmx/monitor/enabled/%s", id),
								HXTrigger: "load",
							},
							div.Div{
								ID:        "monitor_status_icon",
								Classes:   "theme-base rounded-md",
								Style:     "padding-left:12px;margin-right:16px;width:40px;float:right;margin-top:-34px;height:32px;padding-top:4px;",
								HXGet:     fmt.Sprintf("/htmx/monitor/status/%s", id),
								HXTrigger: "load, every 2s",
							},
						},
					},
					div.Div{
						ID:      "monitor-metadata-header",
						Classes: "ui-blue",
						Style:   "padding:8px;",
						Components: []ui.Component{
							h1.H1{
								Contents: "Metadata",
							},
						},
					},
					div.Div{
						ID:        "monitor-metadata-div",
						HXTrigger: "load",
						HXGet:     fmt.Sprintf("/htmx/monitor/metadata/%s", id),
					},
					div.Div{
						ID:      "monitor-requirements-header",
						Classes: "ui-blue",
						Style:   "padding:8px;",
						Components: []ui.Component{
							h1.H1{
								Contents: "Requirements",
							},
						},
					},
					div.Div{
						ID:        "monitor-requirements-div",
						HXTrigger: "load",
						HXGet:     fmt.Sprintf("/htmx/monitor/requirements/%s", id),
					},
					div.Div{
						ID:      "monitor-Contents-header",
						Classes: "ui-blue",
						Style:   "padding:8px;",
						Components: []ui.Component{
							h1.H1{
								Contents: "Contents",
							},
						},
					},
					div.Div{
						ID:        "monitor-contents-div",
						HXTrigger: "load",
						HXGet:     fmt.Sprintf("/htmx/monitor/contents/%s", id),
					},
					div.Div{
						ID:      "monitor-alerts-header",
						Classes: "ui-blue",
						Style:   "padding:8px;",
						Components: []ui.Component{
							h1.H1{
								Contents: "Alerts",
							},
						},
					},
					div.Div{
						ID:        "monitor-alerts-div",
						HXTrigger: "load, every 2s",
						HXGet:     fmt.Sprintf("/htmx/monitor/alerts/%s", id),
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
			ui.Raw{
				HTMLString: `
					<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/styles/obsidian.min.css">
					<script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/highlight.min.js"></script>
					<script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/languages/python.min.js"></script>
					<script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/languages/markdown.min.js"></script>
					<script src="https://cdn.jsdelivr.net/gh/WebCoder49/code-input@2.3/code-input.min.js"></script>
					<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/WebCoder49/code-input@2.3/code-input.min.css">
					<script src="https://cdn.jsdelivr.net/gh/WebCoder49/code-input@2.3/plugins/indent.min.js"></script>
					<script src="https://unpkg.com/htmx.org/dist/ext/json-enc.js"></script>
					<script>
						codeInput.registerTemplate("syntax-highlighted", 
							codeInput.templates.hljs(
								hljs, 
								[
									// You can add or remove plugins in this list from https://github.com/WebCoder49/code-input/blob/main/plugins/README.md.
									// All plugins used must be imported above.
									new codeInput.plugins.Indent(true, 4) // Allow Tab-key indentation, with 4 spaces indentation
								]
							)
						);
						// Register templates with different names here, if needed.
					</script>
					<style>
						code-input {
							width: 100%;
							margin-top: 4px;
							margin-left: 0px;
							--padding: 20px;
						}
					</style>
					`,
			},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render run page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func MonitorMetadataEndpoint(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		logger.Errorf("", "Cannot render monitor metadata endpoint: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	markdown := monitorBuildMetadata(*m, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func MonitorRequirementsEndpoint(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		logger.Errorf("", "Cannot render monitor requirements endpoint: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	markdown := monitorBuildRequirements(*m, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func MonitorContentsEndpoint(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		logger.Errorf("", "Cannot render monitor contents endpoint: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	markdown := monitorBuildContents(*m, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func MonitorEnabledEndpoint(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		logger.Errorf("", "Cannot render monitor enabled endpoint: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	markdown := monitorBuildEnabled(*m, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func MonitorStatusEndpoint(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		logger.Errorf("", "Cannot render monitor status endpoint: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	markdown := monitorBuildStatus(*m, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func MonitorAlertsEndpoint(ctx *gin.Context) {
	id := ctx.Param("id")

	m, err := monitor.GetMonitorByID(id)
	if err != nil {
		logger.Errorf("", "Cannot render monitor status endpoint: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	markdown := monitorBuildAlerts(*m, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func monitorBuildEnabled(m monitor.Monitor, ctx *gin.Context) []byte {
	icon := div.Div{}
	if m.Enabled {
		icon.Components = []ui.Component{
			ui.Raw{
				HTMLString: `<i class="fa-solid fa-toggle-on"></i>`,
			},
		}
	} else {
		icon.Components = []ui.Component{
			ui.Raw{
				HTMLString: `<i class="fa-solid fa-toggle-off"></i>`,
			},
		}
	}
	html, err := icon.Render()
	if err != nil {
		logger.Errorf("", "Cannot render monitor enabled: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func monitorBuildStatus(m monitor.Monitor, ctx *gin.Context) []byte {
	icon := div.Div{}
	switch m.Status {
	case constants.MONITOR_STATUS_RUNNING:
		icon.Components = []ui.Component{
			ui.Raw{
				HTMLString: `<i class="fa-solid fa-circle-check ui-text-green"></i>`,
			},
		}
	case constants.MONITOR_STATUS_STOPPED:
		icon.Components = []ui.Component{
			ui.Raw{
				HTMLString: `<i class="fa-solid fa-circle-exclamation ui-text-red"></i>`,
			},
		}
	case constants.MONITOR_STATUS_ALERT:
		icon.Components = []ui.Component{
			ui.Raw{
				HTMLString: `<i class="fa-solid fa-circle-pause ui-text-yellow"></i>`,
			},
		}
	}
	html, err := icon.Render()
	if err != nil {
		logger.Errorf("", "Cannot render monitor enabled: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func monitorBuildMetadata(m monitor.Monitor, ctx *gin.Context) []byte {
	p := ui.Raw{
		HTMLString: fmt.Sprintf(
			`<pre style="background:#282B2E">ID: 
Workflow: %s
Created: %s
Updated: %s
Kind: %s
</pre>`,
			m.Workflow, m.Created, m.Updated, m.Kind),
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render monitor metadata: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func monitorBuildRequirements(m monitor.Monitor, ctx *gin.Context) []byte {
	p := ui.Raw{
		HTMLString: fmt.Sprintf(`<pre style="background:#282B2E">%s</pre>`, m.Requirements),
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render monitor requirements: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func monitorBuildContents(m monitor.Monitor, ctx *gin.Context) []byte {
	return []byte(fmt.Sprintf(`
<div class="theme-base">
	<code-input
		id="monitor-code"
		template="syntax-highlighted"
		language="python"
		disabled
	>%s</code-input>
</div>
`, m.Contents))
}

func monitorBuildAlerts(m monitor.Monitor, ctx *gin.Context) []byte {
	p := ui.Raw{
		HTMLString: fmt.Sprintf(
			`<pre style="background:#282B2E">%s</pre>`,
			strings.Join(m.Alerts, "\n")),
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render monitor metadata: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
