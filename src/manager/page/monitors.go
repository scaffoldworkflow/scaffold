package page

// import (
// 	"net/http"
// 	"scaffold/manager/constants"
// 	"scaffold/manager/monitor"
// 	"scaffold/manager/user"
// 	"sort"
// 	"strings"

// 	"github.com/jfcarter2358/ui"
// 	"github.com/jfcarter2358/ui/accordion"
// 	"github.com/jfcarter2358/ui/breadcrumb"
// 	"github.com/jfcarter2358/ui/elements/br"
// 	"github.com/jfcarter2358/ui/elements/div"
// 	"github.com/jfcarter2358/ui/elements/link"
// 	"github.com/jfcarter2358/ui/page"
// 	"github.com/jfcarter2358/ui/topbar"

// 	_ "embed"

// 	"github.com/gin-gonic/gin"
// 	logger "github.com/jfcarter2358/go-logger"
// )

// func MonitorsSearchEndpoint(ctx *gin.Context) {
// 	searchTerm, ok := ctx.GetQuery("search")
// 	if !ok {
// 		ctx.Status(http.StatusBadRequest)
// 		return
// 	}
// 	query := strings.TrimSpace(searchTerm)

// 	monitors, err := monitor.GetAllMonitors()
// 	if err != nil {
// 		logger.Errorf("", "Cannot render monitors page: %s", err.Error())
// 		ctx.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}

// 	filtered := []monitor.Monitor{}

// 	for _, m := range monitors {
// 		if strings.Contains(strings.ToLower(m.ID), strings.ToLower(query)) {
// 			filtered = append(filtered, *m)
// 		}
// 	}

// 	markdown := monitorsBuildTable(filtered, ctx)

// 	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
// }

// func MonitorsTableEndpoint(ctx *gin.Context) {
// 	monitors, err := monitor.GetAllMonitors()
// 	if err != nil {
// 		logger.Errorf("", "Cannot render monitors page: %s", err.Error())
// 		ctx.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}
// 	filtered := []monitor.Monitor{}

// 	logger.Tracef("", "Got monitors: %v", monitors)

// 	for _, m := range monitors {
// 		filtered = append(filtered, *m)
// 	}

// 	logger.Tracef("", "Got filtered: %v", filtered)

// 	markdown := monitorsBuildTable(filtered, ctx)

// 	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
// }

// func MonitorsPageEndpoint(ctx *gin.Context) {
// 	markdown := monitorsBuildPage(ctx)
// 	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
// }

// func monitorsBuildPage(ctx *gin.Context) []byte {
// 	p := page.Page{
// 		ID:             "page",
// 		SidebarEnabled: true,
// 		Sidebar:        Sidebar,
// 		Components: []ui.Component{
// 			topbar.Topbar{
// 				Title:   "Scaffold",
// 				Classes: "ui-green",
// 				Buttons: []ui.Component{
// 					link.Link{
// 						Title:   "Logout",
// 						HRef:    "/auth/logout",
// 						Style:   "padding:12px;",
// 						Classes: "theme-dark",
// 					},
// 				},
// 				MenuClasses: "theme-light",
// 			},
// 			div.Div{
// 				Classes: "theme-light",
// 				Components: []ui.Component{
// 					div.Div{
// 						Classes: "ui-green",
// 						Components: []ui.Component{
// 							breadcrumb.Breadcrumb{
// 								Components: []ui.Component{
// 									link.Link{
// 										Title: "Monitors",
// 										HRef:  "/ui/monitors",
// 									},
// 								},
// 								Style: "margin-left:16px;",
// 							},
// 						},
// 					},
// 					ui.Raw{
// 						HTMLString: `<input id="search" class="w3-input  search-bar theme-light" type="text"
//                             name="search" placeholder="Search Monitors"
//                             style="margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%" hx-get="/htmx/monitors/search"
//                             hx-trigger="keyup changed delay:250ms" hx-target="#monitors-table-div" />`,
// 					},
// 					div.Div{
// 						ID:        "monitors-table-div",
// 						HXTrigger: "load",
// 						HXGet:     "/htmx/monitors/table",
// 						Style:     "padding:8px;",
// 					},
// 				},
// 				Style: "margin:64px;",
// 			},
// 			br.BR{},
// 		},
// 	}
// 	html, err := p.Render()
// 	if err != nil {
// 		logger.Errorf("", "Cannot render monitors page: %s", err.Error())
// 		ctx.AbortWithStatus(http.StatusInternalServerError)
// 		return []byte{}
// 	}
// 	return []byte(html)
// }

// func monitorsBuildTable(ms []monitor.Monitor, ctx *gin.Context) []byte {
// 	token, _ := ctx.Cookie("scaffold_token")
// 	u, _ := user.GetUserByLoginToken(token)

// 	out := ""

// 	categorized := map[string][]monitor.Monitor{}

// 	for _, m := range ms {
// 		if val, ok := categorized[m.Workflow]; ok {
// 			categorized[m.Workflow] = append(val, m)
// 		} else {
// 			categorized[m.Workflow] = []monitor.Monitor{m}
// 		}
// 	}

// 	for category, arr := range categorized {
// 		sort.Slice(arr, func(i, j int) bool {
// 			return arr[i].Name < arr[j].Name
// 		})

// 		logger.Debugf("", "Got arr of %v in category %s", arr, category)

// 		rows := []ui.Component{}

// 		for _, m := range arr {
// 			logger.Tracef("", "Got m of %v", m)
// 			logger.Tracef("", "Checking %v against %v", u.Groups, m.Groups)
// 			isInGroup := false
// 			for _, ug := range u.Groups {
// 				if ug == "admin" {
// 					isInGroup = true
// 					break
// 				}
// 				for _, rg := range m.Groups {
// 					if ug == rg {
// 						isInGroup = true
// 						break
// 					}
// 				}
// 				if isInGroup {
// 					break
// 				}
// 			}
// 			if isInGroup {
// 				logger.Tracef("", "Auth confirmed")
// 				status := m.Status
// 				htmlString := ""
// 				logger.Debugf("", "Monitor has status %s", m.Status)
// 				switch status {
// 				case constants.MONITOR_STATUS_RUNNING:
// 					htmlString = `<a href="/ui/monitors/` + m.ID + `" class="dark theme-base shadow-xl theme-hover-light" style="width:100%;padding:8px;display:inline-block;"><i class="fa-solid fa-circle-check ui-text-green"></i>&nbsp;&nbsp;` + m.Name + `</a>`
// 				case constants.MONITOR_STATUS_ALERT:
// 					htmlString = `<a href="/ui/monitors/` + m.ID + `" class="dark theme-base shadow-xl theme-hover-light" style="width:100%;padding:8px;display:inline-block;"><i class="fa-solid fa-circle-exclamation ui-text-red"></i>&nbsp;&nbsp;` + m.Name + `</a>`
// 				case constants.MONITOR_STATUS_STOPPED:
// 					htmlString = `<a href="/ui/monitors/` + m.ID + `" class="dark theme-base shadow-xl theme-hover-light" style="width:100%;padding:16px;display:inline-block;"><i class="fa-solid fa-circle-pause ui-text-yellow"></i>&nbsp;&nbsp;` + m.Name + `</a>`
// 				}
// 				logger.Debugf("", "Adding HTML string: %s", htmlString)
// 				rows = append(rows, ui.Raw{
// 					HTMLString: htmlString,
// 				})
// 				logger.Debugf("", "Rows: %v", rows)
// 			}
// 		}
// 		a := accordion.Accordion{
// 			Classes:      "card shadow-xl theme-base theme-border-base",
// 			ID:           "runtime-accordion",
// 			ContentStyle: "padding:0px;",
// 			Items: []accordion.AccordionItem{
// 				{
// 					Title:      category,
// 					Classes:    "ui-green",
// 					Components: rows,
// 				},
// 			},
// 		}
// 		html, err := a.Render()
// 		if err != nil {
// 			logger.Errorf("", "Cannot render monitors card: %s", err.Error())
// 			ctx.AbortWithStatus(http.StatusInternalServerError)
// 			return []byte{}
// 		}
// 		out += html + "</br>"
// 	}

// 	return []byte(out)
// }
