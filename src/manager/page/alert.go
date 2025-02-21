package page

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/manager/alert"
	"scaffold/manager/api"
	"scaffold/manager/config"
	"scaffold/manager/utils"

	"github.com/jfcarter2358/ui"
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

func AlertPageEndpoint(ctx *gin.Context) {
	markdown := alertBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func AlertExecute(ctx *gin.Context) {
	id := ctx.Param("id")

	a, err := alert.GetAlertByID(id)
	if err != nil {
		logger.Errorf("", "Could not find alert %s to run", id)
		utils.Error(err, ctx, http.StatusNotFound)
		return
	}

	kernelID, err := a.SetupKernel()
	if err != nil {
		logger.Errorf("", "Could not setup kernel for alert %s", id)
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	em := api.ExecuteMessage{
		Contents: a.Script,
		KernelID: kernelID,
		Language: a.Language,
	}

	postBody, _ := json.Marshal(em)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://localhost:%d/api/v1/kernel/%s", config.Config.Port, kernelID)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Kernel trigger for id %s has failed with error %s", id, err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("", "Encountered error reading body: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
	}
	var data map[string]string
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		logger.Fatalf("", "Encountered error unmarshalling status JSON: %s", err.Error())
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	runID := data["run_id"]

	ki := a.Kernels[kernelID]
	ki.RunID = runID
	a.Kernels[kernelID] = ki

	alert.UpdateAlertByID(id, a)
	go a.Check(kernelID)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`<div hx-get="/htmx/alert/%s/output/%s" hx-trigger="load delay:1s" hx-swap="outerHTML"><div
class="ui-blue"
style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
>Kernel ID: %s</div>
<pre style="background:#282B2E">Setting up kernel, please wait...</pre><br></div>`, id, kernelID, kernelID)))
}

func AlertBuildOutput(ctx *gin.Context) {
	alertID := ctx.Param("alert_id")
	kernelID := ctx.Param("kernel_id")

	a, err := alert.GetAlertByID(alertID)
	if err != nil {
		logger.Errorf("", "Cannot get alert %s to check kernel output", alertID)
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ki := a.Kernels[kernelID]

	if ki.Finished {
		logger.Tracef("", "Kernel run is finished")
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`<div><div
class="ui-blue"
style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
>Kernel ID: %s</div>
	<pre style="background:#282B2E;">Setting up kernel, please wait...
%s</pre><br></div>`, kernelID, ki.Output)))
	} else {
		logger.Tracef("", "Kernel run is not finished")
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`<div hx-get="/htmx/alert/%s/output/%s" hx-trigger="load delay:1s" hx-swap="outerHTML"><div
class="ui-blue"
style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
>Kernel ID: %s</div>
<pre style="background:#282B2E">Setting up kernel, please wait...
%s</pre><br></div>`, alertID, kernelID, kernelID, ki.Output)))
	}
}

func alertBuildCode(a alert.Alert) string {
	return fmt.Sprintf(`
	<button
		class="ui-green"
		style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
		hx-post="/htmx/alert/%s/execute"
		hx-swap="beforeend"
		hx-target="#display_div"
	>
		<i class="fa-solid fa-play"></i>
	</button>
	<code-input
		id="code_block"
		template="syntax-highlighted"
		language="python"
		disabled
	>%s</code-input>
	<div
		class="ui-green"
		style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
		id="outputs_header"
	>Outputs:</div>
`, a.ID, a.Script)
}

func alertBuildPage(ctx *gin.Context) []byte {
	alertID := ctx.Param("id")

	a, err := alert.GetAlertByID(alertID)
	if err != nil {
		logger.Errorf("", "Cannot get alert with ID %s: %s", alertID, err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}

	if a == nil {
		logger.Errorf("", "Cannot get alert with ID %s: %s", alertID, err.Error())
		ctx.AbortWithStatus(http.StatusNotFound)
		return []byte{}
	}

	outputs := []ui.Component{
		ui.Raw{
			HTMLString: alertBuildCode(*a),
		},
	}
	for kernelID, info := range a.Kernels {
		outputs = append(outputs, ui.Raw{
			HTMLString: fmt.Sprintf(`<div hx-get="/htmx/alert/%s/output/%s" hx-trigger="load delay:1s" hx-swap="outerHTML"><div
class="ui-blue"
style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
>Kernel ID: %s</div>
<pre style="background:#282B2E">Setting up kernel, please wait...
%s</pre><br></div>`, alertID, kernelID, kernelID, info.Output),
		})
	}
	p := page.Page{
		ID:             "page",
		SidebarEnabled: true,
		Sidebar:        Sidebar,
		Components: []ui.Component{
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
									link.Link{
										Title: alertID,
										HRef:  fmt.Sprintf("/ui/alerts/%s", alertID),
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					div.Div{
						ID:         "display_div",
						Style:      "padding:8px;margin:8px;",
						Classes:    "theme-base",
						Components: outputs,
					},
					br.BR{},
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
