package page

import (
	"fmt"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/history"
	"scaffold/server/ui"
	"scaffold/server/ui/breadcrumb"
	"scaffold/server/ui/elements/br"
	"scaffold/server/ui/elements/div"
	"scaffold/server/ui/elements/link"
	"scaffold/server/ui/elements/pre"
	"scaffold/server/ui/page"
	"scaffold/server/ui/sidebar"
	"scaffold/server/ui/timeline"
	"scaffold/server/ui/timeline/item"
	"scaffold/server/ui/topbar"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
	"gopkg.in/yaml.v3"
)

func HistoryStateEndpoint(ctx *gin.Context) {
	markdown := historyBuildState(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func HistoryTimelineEndpoint(ctx *gin.Context) {
	runID := ctx.Param("run_id")
	h, err := history.GetHistoryByRunID(runID)
	if err != nil {
		logger.Errorf("", "Cannot render run timeline page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	markdown := historyBuildTimeline(*h, runID, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func HistoryPageEndpoint(ctx *gin.Context) {
	markdown := historyBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func historyBuildState(ctx *gin.Context) []byte {
	runID := ctx.Param("run_id")
	stateName := ctx.Param("state_name")

	h, err := history.GetHistoryByRunID(runID)
	if err != nil {
		if err != nil {
			logger.Errorf("", "Cannot render run state page: %s", err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return []byte{}
		}
	}

	for _, s := range h.States {
		if s.Task == stateName {
			bytes, err := yaml.Marshal(s)
			if err != nil {
				logger.Errorf("", "Cannot marshal state to yaml: %s", err.Error())
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return []byte{}
			}
			return bytes
		}
	}

	logger.Errorf("", "No state exists with name %s in run %s: %s", stateName, runID, err.Error())
	ctx.AbortWithStatus(http.StatusInternalServerError)
	return []byte{}
}

func historyBuildPage(ctx *gin.Context) []byte {
	runID := ctx.Param("run_id")
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
						Classes: "scaffold-green rounded-md",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Runs",
										HRef:  "/ui/runs",
									},
									link.Link{
										Title: runID,
										HRef:  fmt.Sprintf("/ui/runs/%s", runID),
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					div.Div{
						ID:        "history-state-div",
						HXTrigger: "load, every 2s",
						HXGet:     fmt.Sprintf("/htmx/runs/timeline/%s", runID),
						Style:     "padding-left:64px;",
					},
					ui.Raw{
						HTMLString: "<hr>",
					},
					pre.Pre{
						ID:    "history-state-pre",
						Style: "margin-bottom:64px;padding:32px;",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
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

func historyBuildTimeline(h history.History, runID string, ctx *gin.Context) []byte {
	t := timeline.Timeline{
		ID:    "history_timeline",
		Items: make([]item.Item, 0),
		Style: "margin-bottom:64px;",
	}

	for idx, s := range h.States {
		t.Items = append(t.Items, item.Item{
			ID:          fmt.Sprintf("item-%s-%s", s.Workflow, s.Task),
			IsFirst:     idx == 0,
			IsLast:      idx == len(h.States)-1,
			BoxContents: s.Task,
			IconClasses: func(status string) string {
				switch status {
				case constants.STATE_STATUS_RUNNING:
					return "fa-solid fa-spinner fa-pause scaffold-text-blue"
				case constants.STATE_STATUS_ERROR:
					return "fa-solid fa-circle-xmark scaffold-text-red"
				case constants.STATE_STATUS_KILLED:
					return "fa-solid fa-skull scaffold-text-orange"
				case constants.STATE_STATUS_NOT_STARTED:
					return "fa-regular fa-circle scaffold-text-charcoal"
				case constants.STATE_STATUS_WAITING:
					return "fa-solid fa-clock scaffold-text-yellow"
				case constants.STATE_STATUS_SUCCESS:
					return "fa-solid fa-circle-check scaffold-text-green"
				}
				return "fa-solid fa-circle-question scaffold-text-charcoal"
			}(s.Status),
			LineColor: func(status string) string {
				switch status {
				case constants.STATE_STATUS_RUNNING:
					return "scaffold-blue"
				case constants.STATE_STATUS_ERROR:
					return "scaffold-red"
				case constants.STATE_STATUS_KILLED:
					return "scaffold-orange"
				case constants.STATE_STATUS_NOT_STARTED:
					return "scaffold-charcoal"
				case constants.STATE_STATUS_WAITING:
					return "scaffold-yellow"
				case constants.STATE_STATUS_SUCCESS:
					return "scaffold-green"
				}
				return "scaffold-charcoal"
			}(s.Status),
			HXTrigger: "click",
			HXGet:     fmt.Sprintf("/htmx/runs/timeline/%s/status/%s", runID, s.Task),
			HXTarget:  "#history-state-pre",
		})
	}

	html, err := t.Render()
	if err != nil {
		logger.Errorf("", "Cannot render run timeline: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}