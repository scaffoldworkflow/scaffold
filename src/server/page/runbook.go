package page

import (
	"fmt"
	"net/http"
	"scaffold/server/runbook"

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

func RunbookPageEndpoint(ctx *gin.Context) {
	markdown := runbookBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func runbookBuildBlock(b runbook.Block, idx int, kernelID string) string {
	switch b.Type {
	case "code":
		return fmt.Sprintf(`
<div style="padding:8px;margin:8px;" class="theme-base">
	<button
		class="rounded-sm ui-green"
		style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
		hx-post="/htmx/kernel/execute"
		hx-vals='js:{
			"contents": document.getElementById("code_%d").value,
			"block_idx_str": "%d",
			"kernel_id": "%s"
		}'
		hx-ext="json-enc"
		hx-target="#output_%d"
		hx-swap="outerHTML"
	>
		<i class="fa-solid fa-play"></i>
	</button>
	<code-input
		id="code_%d"
		template="syntax-highlighted"
		language="python"
	>%s</code-input>
	<div
		class="ui-text-green"
		style="padding-left:16px;padding-right:16px;padding-top:4px;padding-bottom:4px;margin-left:0px;"
	>Output:</div>
	<pre id="output_%d" style="background:#282B2E"></pre>
</div>
`, idx, idx, kernelID, idx, idx, b.Contents, idx)
	case "markdown":
		return fmt.Sprintf(`
<div style="padding:8px;margin:8px;" class="theme-base">
	<zero-md
		hx-trigger="dblclick"
		hx-get="/htmx/runbook/block/editor"
		hx-vals='js:{
			"contents": document.getElementById("markown_%d").value
		}'
		hx-ext="json-enc"
		hx-swap="outerHTML"
		margin="16px;"
	>
		<template>
			<style>
				:host { display: block; position: relative; contain: content; }
				:host([hidden]) { display: none; }
				pre {
					background: #282B2E;
				}
			</style>
			<!-- KaTeX styles (needed for math) -->
			<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0/dist/katex.min.css" />
		</template>
		<script type="text/markdown" id="markdown_%d">%s</script>
		<style>
			zero-md {
				display: block;
				position: relative;
			}
		</style>
	</zero-md>
</div>
`, idx, idx, b.Contents)
	}
	return ""
}

func runbookBuildPage(ctx *gin.Context) []byte {
	runbookID := ctx.Param("runbook_id")
	kernelID := ctx.Param("kernel_id")

	r, err := runbook.GetRunbookByID(runbookID)
	if err != nil {
		logger.Errorf("", "Cannot get runbook with ID %s: %s", runbookID, err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}

	blockElements := []ui.Component{}
	for idx, b := range r.Blocks {
		blockElements = append(blockElements, ui.Raw{
			HTMLString: runbookBuildBlock(b, idx, kernelID),
		})
	}
	blockElements = append(blockElements, ui.Raw{
		HTMLString: "</br>",
	})
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
<script type="module" src="https://cdn.jsdelivr.net/npm/zero-md@3?register"></script>
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
						Classes: "theme-dark rounded-md",
					},
				},
				MenuClasses: "theme-light",
			},
			div.Div{
				Classes: "theme-light rounded-md",
				Components: append([]ui.Component{
					div.Div{
						Classes: "ui-green rounded-md",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Runbooks",
										HRef:  "/ui/runbooks",
									},
									link.Link{
										Title: runbookID,
										HRef:  fmt.Sprintf("/ui/runbooks/%s/%s", runbookID, kernelID),
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
				},
					blockElements...),
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
