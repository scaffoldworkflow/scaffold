package page

import (
	"fmt"
	"net/http"
	"scaffold/manager/alert"
	"sort"
	"strings"

	"github.com/jfcarter2358/ui"
	"github.com/jfcarter2358/ui/accordion"
	"github.com/jfcarter2358/ui/breadcrumb"
	"github.com/jfcarter2358/ui/elements/br"
	"github.com/jfcarter2358/ui/elements/div"
	"github.com/jfcarter2358/ui/elements/link"
	"github.com/jfcarter2358/ui/page"
	"github.com/jfcarter2358/ui/topbar"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func AlertsSearchEndpoint(ctx *gin.Context) {
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	alerts, err := alert.GetAllAlerts()
	if err != nil {
		logger.Errorf("", "Cannot render alerts page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	filtered := []alert.Alert{}

	for _, a := range alerts {
		if strings.Contains(strings.ToLower(a.ID), strings.ToLower(query)) {
			filtered = append(filtered, *a)
		}
	}

	markdown := alertsBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func AlertsTableEndpoint(ctx *gin.Context) {
	alerts, err := alert.GetAllAlerts()
	if err != nil {
		logger.Errorf("", "Cannot render alerts page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []alert.Alert{}

	logger.Tracef("", "Got alerts: %v", alerts)

	for _, m := range alerts {
		filtered = append(filtered, *m)
	}

	logger.Tracef("", "Got filtered: %v", filtered)

	markdown := alertsBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func AlertsPageEndpoint(ctx *gin.Context) {
	markdown := alertsBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func alertsBuildPage(ctx *gin.Context) []byte {
	p := page.Page{
		ID:             "page",
		SidebarEnabled: true,
		Sidebar:        Sidebar,
		Components: []ui.Component{
			topbar.Topbar{
				Title:   "Scaffold",
				Classes: "ui-green",
				Buttons: []ui.Component{
					link.Link{
						Title:   "Logout",
						HRef:    "/auth/logout",
						Style:   "padding:12px;",
						Classes: "theme-dark",
					},
				},
				MenuClasses: "theme-light",
			},
			div.Div{
				Classes: "theme-light",
				Components: []ui.Component{
					div.Div{
						Classes: "ui-green",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Alerts",
										HRef:  "/ui/alerts",
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input search-bar theme-light" type="text"
                            name="search" placeholder="Search Alerts"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%" hx-get="/htmx/alerts/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#alerts-table-div" />`,
					},
					div.Div{
						ID:        "alerts-table-div",
						HXTrigger: "load",
						HXGet:     "/htmx/alerts/table",
						Style:     "padding:8px;",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render alerts page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func alertsBuildTable(as []alert.Alert, ctx *gin.Context) []byte {
	// token, _ := ctx.Cookie("scaffold_token")
	// u, _ := user.GetUserByLoginToken(token)

	out := ""

	categorized := map[string][]alert.Alert{}

	for _, a := range as {
		key := fmt.Sprintf("%s/%s", a.Team, a.Project)
		if val, ok := categorized[key]; ok {
			categorized[key] = append(val, a)
		} else {
			categorized[key] = []alert.Alert{a}
		}
	}

	for category, arr := range categorized {
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].Name < arr[j].Name
		})

		logger.Debugf("", "Got arr of %v in category %s", arr, category)

		rows := []ui.Component{}

		for _, a := range arr {
			logger.Tracef("", "Got a of %v", a)
			// logger.Tracef("", "Checking %v against %v", u.Groups, m.Groups)
			// isInGroup := false
			// for _, ug := range u.Groups {
			// 	if ug == "admin" {
			// 		isInGroup = true
			// 		break
			// 	}
			// 	for _, rg := range m.Groups {
			// 		if ug == rg {
			// 			isInGroup = true
			// 			break
			// 		}
			// 	}
			// 	if isInGroup {
			// 		break
			// 	}
			// }
			// if isInGroup {
			// logger.Tracef("", "Auth confirmed")

			htmlString := ""
			// logger.Debugf("", "Alert has status %s", .Status)
			if a.Enabled {
				// color := "green"
				// for _, status := range a.Statuses {
				// 	switch status {
				// 	case constants.ALERT_STATUS_ERROR:
				// 		color = "red"
				// 	case constants.ALERT_STATUS_KILLED:
				// 		color = "orange"
				// 	}
				// }
				htmlString = `<a href="/ui/alerts/` + a.ID + `" class="dark theme-base shadow-xl theme-hover-light" style="width:100%;padding:8px;display:inline-block;"><i class="fa-solid fa-circle-play ui-text-green"></i>&nbsp;&nbsp;` + a.Name + `</a>`
			} else {
				htmlString = `<a href="/ui/alerts/` + a.ID + `" class="dark theme-base shadow-xl theme-hover-light" style="width:100%;padding:8px;display:inline-block;"><i class="fa-solid fa-circle-pause ui-text-grey"></i>&nbsp;&nbsp;` + a.Name + `</a>`
			}
			logger.Debugf("", "Adding HTML string: %s", htmlString)
			rows = append(rows, ui.Raw{
				HTMLString: htmlString,
			})
			logger.Debugf("", "Rows: %v", rows)
		}
		// }
		a := accordion.Accordion{
			Classes:      "card shadow-xl theme-base theme-border-base",
			ID:           "runtime-accordion",
			ContentStyle: "padding:0px;",
			Items: []accordion.AccordionItem{
				{
					Title:      category,
					Classes:    "ui-green",
					Components: rows,
				},
			},
		}
		html, err := a.Render()
		if err != nil {
			logger.Errorf("", "Cannot render alerts card: %s", err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return []byte{}
		}
		out += html + "</br>"
	}

	return []byte(out)
}
