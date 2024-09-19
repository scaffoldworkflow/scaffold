const Workflow = class {
    constructor(canvas_id, parent_id, canvas_style, node_style, node_z_index, class_string, tasks, pin_colors) {
        this.padding = 50
        this.initial_width = 200

        this.pin_colors = pin_colors
        this.tasks = tasks
        this.structure = {}
        this.nodes = {}
        this.links = []
        this.width = ""
        this.height = ""
        this.canvas_style = `position:absolute;top:0px;left:0px;z-index:1;`
        this.node_z_index = 995
        this.margin = 10
        this.input_color = "#888888"
        this.class_string = class_string
        this.shadow_color = "#000000"
        this.max_deflection = 10
        this.widths = []
        this.heights = []
        this.active = {
            "foo2": [
                "pin 1"
            ] 
        }
        this.canvas_id = canvas_id

        if (canvas_style != "") {
            this.canvas_style = canvas_style
        }
        if (node_style != "") {
            this.node_style = node_style
        }
        if (node_z_index != "") {
            this.node_z_index = node_z_index
        }

        this.parent = $(`#${parent_id}`)
        this.height = `${this.parent.height()}px;`
        this.width = `${this.parent.width()}px;`

        
        
        $(this.parent).append(`<canvas 
                style=""${this.canvas_style}width:${this.width};"
                height="${this.height}"
                width="${this.width}"
                id="${canvas_id}"
                draggable="true">
        </canvas>`)

        this.canvas = $(`#${canvas_id}`)
        $(this.canvas).width(`${this.width}px`)
        $(this.canvas).css("width", `${this.width}px`)
        $(this.canvas).css("height", `${this.height}px`)

        this.SetupNodes()
        this.RenderNodes()
        this.RenderLines()
    }

    GetNodeSize(key) {
        let bg_color = this.tasks[key].title.background_color
        let fg_color = this.tasks[key].title.foreground_color
        let extra_icons = this.tasks[key].extra_icons

        let rows = ''
        let outputs = []
        if (this.tasks[key].out != null && this.tasks[key].out != undefined) {
            outputs = Object.keys(this.tasks[key].out)
        }
        
        let html = `<div class="w3-round-large w3-border w3-card w3-border ${this.class_string}" style="position:fixed;z-index:995;" id="placeholder">
                    <header class="w3-container w3-round-large" style="background-color:${bg_color}">
                        <h3 style="color:${fg_color}">${key}${extra_icons}</h3>
                    </header>
                </div>`

        $(this.parent).append(html)
        let width = $("#placeholder").width() + (this.margin * 2)
        let height = $("#placeholder").height() + (this.margin * 2)
        $("#placeholder").remove()
        return {width: width, height: height}
    }

    SetupNodes() {
        // Set an object for the graph label
        var g = new dagre.graphlib.Graph();
        
        g.setGraph({});

        // Default to assigning a new object as a label for each new edge.
        g.setDefaultEdgeLabel(function () {
            return {};
        });

        for (let [name, task] of Object.entries(this.tasks)) {
            let size = this.GetNodeSize(name)
            g.setNode(name, { label: name, width: size.width, height: size.height });

            if (task.out != null && task.out != undefined) {
                for (let [pin, nodes] of Object.entries(task.out)) {
                    for (let node of nodes) {
                        this.links.push({
                            from: name,
                            to: node,
                            pin: pin,
                            color: this.pin_colors[pin],
                        })
                        g.setEdge(name, node);
                    }
                }
            }
        }

        // Create layout
        dagre.layout(g, {align: "UL"});

        g.nodes().forEach(function (v) {
            let n = g.node(v)
            this.nodes[v] = {name: v, x: n.x, y: n.y, width: n.width, height: n.height}
        }, this);
    }

    RenderNodes() {
        $(this.parent).find('div').remove();

        for (let [key, val] of Object.entries(this.nodes)) {
            let bg_color = this.tasks[key].title.background_color
            let fg_color = this.tasks[key].title.foreground_color

            let outputs = []
            if (this.tasks[key].out != null && this.tasks[key].out != undefined) {
                outputs = Object.keys(this.tasks[key].out)
            }

            let func_def = ""
            if (this.tasks[key].func != undefined) {
                func_def = `ondblclick="${this.tasks[key].func}('${key}')"`
            }

            let html = `<div class="w3-round-large w3-border w3-card w3-border ${this.class_string}" style="position:absolute;z-index:995;left:${val.x}px;top:${val.y}px;" id="${key}" ${func_def}>
                        <header class="w3-container w3-round-large" style="background-color:${bg_color}" id="${key}-header">
                            <h3 id="${key}-header-text" style="color:${fg_color}">${this.tasks[key].title.text}</h3>
                        </header>
                    </div>`

            $(this.parent).append(html)
            $(`#${key}`).draggable()
        }
    }

    GetPinCoords(node_name, pin_name) {
        let offset_y = this.nodes[node_name].y
        offset_y += $(`#${node_name}-header`).height() / 2

        let offset_x = this.nodes[node_name].x 
        offset_x += $(`#${node_name}`).width() / 2

        let x = offset_x
        let y = offset_y
        return {x: x, y: y}
    }

    DrawCurve(ctx, points, tension) {
        ctx.beginPath();
        ctx.moveTo(points[0].x, points[0].y);

        var t = (tension != null) ? tension : 1;
        for (var i = 0; i < points.length - 1; i++) {
            var p0 = (i > 0) ? points[i - 1] : points[0];
            var p1 = points[i];
            var p2 = points[i + 1];
            var p3 = (i != points.length - 2) ? points[i + 2] : p2;

            var cp1x = p1.x + (p2.x - p0.x) / 6 * t;
            var cp1y = p1.y + (p2.y - p0.y) / 6 * t;

            var cp2x = p2.x - (p3.x - p1.x) / 6 * t;
            var cp2y = p2.y - (p3.y - p1.y) / 6 * t;

            ctx.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, p2.x, p2.y);
        }
        ctx.stroke();
        ctx.closePath()
    }

    RenderLines() {
        let canvas = document.getElementById(this.canvas_id);
        let ctx = canvas.getContext("2d");
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        
        for (let link of this.links) {
            let start = this.GetPinCoords(link.from, link.pin)
            let end = this.GetPinCoords(link.to, 'Input')

            let delta_x = end.x - start.x
            let delta_y = end.y - start.y

            let control_1 = {x: start.x + (3 * delta_x / 16), y: start.y + (delta_y / 8)}
            let control_2 = {x: end.x - (3 * delta_x / 16), y: end.y - (delta_y / 8)}

            let delta_control_1 = control_1.y - start.y
            if (delta_control_1 < 0 && -delta_control_1 > this.max_deflection) {
                control_1.y = start.y - this.max_deflection
            } else if (delta_control_1 > 0 && delta_control_1 > this.max_deflection) {
                control_1.y = start.y + this.max_deflection
            }

            let delta_control_2 = control_2.y - end.y
            if (delta_control_2 < 0 && -delta_control_2 > this.max_deflection) {
                control_2.y = end.y - this.max_deflection
            } else if (delta_control_2 > 0 && delta_control_2 > this.max_deflection) {
                control_2.y = end.y + this.max_deflection
            }

            let midpoint = {x: start.x + delta_x / 2, y: start.y + delta_y / 2}

            ctx.lineCap = "round"
            ctx.lineWidth = 10;
            ctx.strokeStyle = link.color
            this.DrawCurve(ctx, [start, control_1, midpoint, control_2, end])

            ctx.beginPath();
            ctx.strokeStyle = link.color
            ctx.fillRect(start.x - 4, start.y - 4, 8, 8)
            ctx.strokeStyle = this.input_color
            ctx.fillRect(end.x - 4, end.y - 4, 8, 8)
            ctx.stroke();
            ctx.closePath()
        }
    }

    UpdateWorkflow() {
        for (let [key, val] of Object.entries(this.nodes)) {
            let pos = $(`#${key}`).position()
            if (pos != undefined) {
                this.nodes[key].x = pos.left
                this.nodes[key].y = pos.top
            }
        }

        this.RenderLines()
    }
}

var healthIcons = {
	'healthy': 'fa-circle-check',
	'degraded': 'fa-circle-exclamation',
	'unhealthy': 'fa-circle-xmark',
	'unknown': 'fa-circle-question'
}

var healthColors = {
	'healthy': 'scaffold-text-green',
	'degraded': 'scaffold-text-yellow',
	'unhealthy': 'scaffold-text-red',
	'unknown': 'scaffold-text-charcoal'
}

var healthText = {
	'healthy': 'Up',
	'degraded': 'Degraded',
	'unhealthy': 'Down',
	'unknown': 'Unknown'
}

var icons = {
	'healthy': 'fa-circle-check',
	'warn': 'fa-circle-exclamation',
	'error': 'fa-circle-xmark',
	'deploying': 'fa-spinner fa-pulse',
	'unknown': 'fa-circle-question',
	'not-deployed': 'fa-circle'
}

var colors = {
	'healthy': 'green',
	'warn': 'yellow',
	'error': 'red',
	'deploying': 'grey',
	'unknown': 'orange',
	'not-deployed': 'grey'
}

var CurrentStateName

var Checksums = new Map()

var datastore=null
var inputs=null

function updateStateStatus(force) {
    // let ids = ["workflow-state-header", "workflow-output-header", "workflow-status-header", "workflow-code-header", "workflow-context-header"]
    let ids = ["state-status-collapse", "state-context-collapse", "state-display-collapse", "state-output-collapse"]
    if (CurrentStateName != "" && states != undefined) {
        for (let state of states) {
            if (state.task == CurrentStateName) {
                let color = state_colors[state.status]
                let text_color = state_text_colors[state.status]
                for (let id of ids) {
                    for (let color of color_keys) {
                        $(`#${id}`).removeClass(state_colors[color])
                    }
                }
                for (let id of ids) {
                    $(`#${id}`).addClass(color)
                }
                $("#state-run").text(`Run: ${state.number}`)
                $("#state-name").text(state.task)
                $("#state-status").text(`Status: ${state.status}`)
                $("#state-started").text(`Started: ${state.started}`)
                $("#state-finished").text(`Finished: ${state.finished}`)
                
                $("#toggle-icon").removeClass("fa-toggle-off");
                $("#toggle-icon").removeClass("fa-toggle-on");
                if (state.disabled) {
                    $("#toggle-icon").addClass("fa-toggle-off");
                } else {
                    $("#toggle-icon").addClass("fa-toggle-on");
                }

                // if (Checksums.has(state.task)) {
                //     if (Checksums.get(state.task) != state.output_checksum) {
                $("#state-output").text(state.output)
                $(`#state-context`).empty();
                $(`#state-context`).append(buildContextTable(state.context, color, text_color))
                buildDisplay(state.display, "current", color, text_color)
                        // Checksums.set(state.task, s.output_checksum)
                //     }
                // } else {
                //     $("#state-output").text(state.output)
                //     $(`#state-context`).empty();
                //     $(`#state-context`).append(buildContextTable(state.context, color, text_color))
                //     buildDisplay(state.display, "current", color, text_color)
                //     Checksums.set(state.task, s.output_checksum)
                // }
                continue
            }
        }
    }
}

function buildContextTable(context, color, text_color) {
    // create the card
    let output = `<div class="w3-border w3-card theme-light theme-border-light w3-round">`
    output += `
        <header class="w3-container ${color}" id="workflow-context-header">
            <h4>Context Values</h4>
        </header>
    `

    // create the table
    output += `<table class="w3-table dark w3-border theme-border-light theme-table-striped">`
    
    // header
    output += `<tr>`
    output += `
        <th class="table-title w3-medium ${text_color}">
            <span class="table-title-text">Name</span>
        </th>
        <th class="table-title w3-medium ${text_color}">
            <span class="table-title-text">Value</span>
        </th>
    `
    output += `</tr>`

    // add table data
    for (let key in context) {
        output += `<tr>`
        output += `
            <td>
                ${key}
            </td>
            <td>
                ${context[key]}
            </td>
        `
        output += `</tr>`
    }
    
    // close out the table
    output += `</table>`

    // close out the card
    output += `
        </div>
        <br>
    `

    return output
}

function buildDisplayTable(cd, color, text_color) {
    // create the card
    let output = `<div class="w3-border w3-card theme-light theme-border-light w3-round">`
    if (cd.name != null && cd.name != undefined && cd.name != "") {
        output += `
            <header class="w3-container ${color}">
                <h4>${cd.name}</h4>
            </header>
        `
    }

    // create the table
    output += `<table class="w3-table dark w3-border theme-border-light theme-table-striped">`
    
    // header
    output += `<tr>`
    for (let i = 0; i < cd.header.length; i++) {
        let value = cd.header[i]
        output += `
            <th class="table-title w3-medium ${text_color}">
                <span class="table-title-text">${value}</span>
            </th>
        `
    }
    output += `</tr>`

    // add table data
    for (let i = 0; i < cd.data.length; i++) {
        output += `<tr>`
        row = cd.data[i]
        for (let j = 0; j < row.length; j++) {
            let value = row[j]
            output += `
                <td>
                    ${value}
                </td>
            `
        }
        output += `</tr>`
    }
    
    // close out the table
    output += `</table>`

    // close out the card
    output += `
        </div>
        <br>
    `

    return output
}

function buildDisplayPre(cd, color) {
    // create the card
    let output = `<div class="w3-border w3-card theme-light theme-border-light w3-round">`
    if (cd.name != null && cd.name != undefined && cd.name != "") {
        output += `
            <header class="w3-container ${color}">
                <h4>${cd.name}</h4>
            </header>
        `
    }
    output += `
        <div class="w3-container">
    `

    // add pre data
    output += `
        <pre style="font-family:monospace;">
            ${cd.data}
        </pre>
    `

    // close out the card
    output += `
        </div>
        </div>
        <br>
    `

    return output
}

function buildDisplayValue(cd, color) {
    // create the card
    let output = `<div class="w3-border w3-card theme-light theme-border-light w3-round">`
    if (cd.name != null && cd.name != undefined && cd.name != "") {
        output += `
            <header class="w3-container ${color}">
                <h4>${cd.name}</h4>
            </header>
        `
    }
    output += `
        <div class="w3-container">
    `

    // add pre data
    output += `
        <p>
            ${cd.data}
        </p>
    `

    // close out the card
    output += `
        </div>
        </div>
        <br>
    `

    return output
}

function buildDisplay(display, item, color, text_color) {
    // if (display == null || display == undefined || display.length == 0) {
    //     $(`#state-display-collapse`).css("display", "none")
    //     return
    // }

    $(`#state-${item}-display-data`).empty();
    output = ""
    for (let i = 0; i < display.length; i++) {
        let cd = display[i];
        let local_color = color;
        if (cd.color != "" && cd.color != undefined && cd.color != null) {
            local_color = cd.color;
        }
        if (cd.kind == "table") {
            $(`#state-${item}-display-data`).append(buildDisplayTable(cd, local_color, text_color))
        } else if (cd.type == "pre") {
            $(`#state-${item}-display-data`).append(buildDisplayPre(cd, local_color))
        } else {
            $(`#state-${item}-display-data`).append(buildDisplayValue(cd, local_color))
        }
    }
    $(`#state-display-collapse`).css("display", "block")
}

function killRun() {
    let parts = window.location.href.split('/')
    let c = parts[parts.length - 1]

    let t = CurrentStateName

    if (c == "" || t == "") {
        $("#error-container").text(`Invalid task information, workflow: ${c}, task: ${t}`)
        openModal('error-modal')
        return
    }

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: `/api/v1/run/${c}/${t}`,
        type: "DELETE",
        contentType: "application/json",
        success: function (result) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
        },
        error: function (response) {
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            openModal('error-modal')
        }
    });
}

function toggleDisable() {
    let parts = window.location.href.split('/')
    let c = parts[parts.length - 1]

    let t = CurrentStateName

    if (c == "" || t == "") {
        $("#error-container").text(`Invalid task information, workflow: ${c}, task: ${t}`)
        openModal('error-modal')
        return
    }

    $.ajax({
        url: `/api/v1/task/${c}/${t}/enabled`,
        type: "PUT",
        contentType: "application/json",
        success: function (result) {
        },
        error: function (response) {
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            openModal('error-modal')
        }
    });
}

function changeStateName(name) {
    CurrentStateName = name
    $("#current-state-header").html(CurrentStateName)
    updateStateStatus(true)
    state_modal.showModal()
}



function openInput() {
    getDataStore()
    toggleCurrentInput()
}

function getDataStore() {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/datastore/" + workflowName,
        type: "GET",
        success: function (result) {
            datastore = result
            getInputs(true, datastore)
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    });
}

function getInputs(trigger, datastore) {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/input/" + workflowName,
        type: "GET",
        success: function (result) {
            inputs = result
            if (trigger) {
                loadInputData(inputs)
            }
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    });
}

function loadInputData(inputs) {
    console.log("Loading input data!")
    $("#current-input-div").empty()

    for (let idx = 0; idx < inputs.length; idx++) {
        let i = inputs[idx]
        console.log(`Got input with name ${i.name}`)
        let value = datastore.env[i.name]
        let html = `<div class="w3-bar-item theme-base w3-border-bottom theme-border-light">
            <b>${i.description}</b>
        </div>
        <div class="w3-bar-item theme-light w3-border-bottom theme-border-light">
            <input
                class="w3-input theme-light"
                type="${i.type}"
                id="${i.name}"
            >
        </div>`
        $("#current-input-div").append(html)
        $(`#${i.name}`).val(value)
    }
}

function changeIcon() {
    if ($("#save-icon").hasClass("fa-floppy-disk")) {
        $("#save-icon").removeClass("fa-floppy-disk")
        $("#save-icon").addClass("fa-check")
    } else {
        $("#save-icon").removeClass("fa-check")
        $("#save-icon").addClass("fa-floppy-disk")
    }
}

function saveInputs() {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")
    
    for (let i = 0; i < inputs.length; i++) {
        value = $(`#${inputs[i].name}`).val()
        datastore.env[inputs[i].name] = value
    }

    $.ajax({
        type: "PUT",
        url: `/api/v1/datastore/${workflowName}`,
        contentType: "application/json",
        dataType: "json",
        data: JSON.stringify(datastore),
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            changeIcon()
            setInterval(changeIcon, 1500)
        },
        error: function(response) {
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            openModal('error-modal')
        }
    });
}

function toggleAccordion(id) {
  var x = document.getElementById(id);
  if (x.className.indexOf("w3-show") == -1) {
    x.className += " w3-show";
    $(`#${id}-icon`).removeClass("fa-caret-down");
    $(`#${id}-icon`).addClass("fa-caret-up");
  } else {
    x.className = x.className.replace(" w3-show", "");
    $(`#${id}-icon`).removeClass("fa-caret-up");
    $(`#${id}-icon`).addClass("fa-caret-down");
  }
}


stateIntervalMilliSeconds = 500
workflowIntervalMilliSeconds = 50

var tasks = {}
var states = []
var datastore = {}
var link_data = []
var node_data = []
var structure = {}
var elements = []
var positions = {}
var rawTasks = []

right_panel_width = 500

var workflow

pin_colors = {
    "Success": "#A3BE8C",
    "Error": "#BF616A",
    "Always": "#5E81AC",
}

var state_colors = {
    "not_started": "scaffold-charcoal",
    "success": "scaffold-green",
    "error": "scaffold-red",
    "running": "scaffold-blue",
    "waiting": "scaffold-yellow",
    "killed": "scaffold-orange"
}

var state_icons = {
    "not_started": '<i class="w3-medium fa-regular fa-circle"></i>',
    "success": '<i class="w3-medium fa-solid fa-circle-check"></i>',
    "error": '<i class="w3-medium fa-solid fa-circle-exclamation"></i>',
    "running": '<i class="w3-medium fa-sharp fa-solid fa-spinner fa-spin"></i>',
    "waiting": '<i class="w3-medium fa-solid fa-clock"></i>',
    "killed": '<i class="w3-medium fa-solid fa-skull"></i>'
}

var state_colors_hex = {
    "not_started": "#373F51",
    "success": "#A3BE8C",
    "error": "#BF616A",
    "running": "#5E81AC",
    "waiting": "#EBCB8B",
    "killed": "#D08770"
}

var state_text_colors = {
    "not_started": "scaffold-text-charcoal",
    "success": "scaffold-text-green",
    "error": "scaffold-text-red",
    "running": "scaffold-text-blue",
    "waiting": "scaffold-text-yellow",
    "killed": "scaffold-text-orange"
}

color_keys = ["not_started", "success", "error", "running", "waiting", "killed"]

var hidden = []
var disabled = []

function render() {
    let prefix = $("#search").val();
    prefix = prefix.toLowerCase();

    if (prefix == "") {
        hidden = []
    } else {
        for (let [key, task] of Object.entries(tasks)) {
            if (key.toLowerCase().indexOf(prefix) == -1) {
                hidden.push(key)
            }
        }
    }
    updateNodes()
}


function toggleCurrentState() {
    let sidebar = document.getElementById("current-state")
    if (sidebar.className.indexOf("show") == -1) {
        // Close input
        let input = document.getElementById("current-input")
        input.classList.remove("show");
        $("#current-input").css("left", `calc(100%)`)
        // Close legend
        let legend = document.getElementById("current-legend")
        legend.classList.remove("show");
        $("#current-legend").css("left", `calc(100%)`)
        // Show state
        updateStateStatus(true)
        sidebar.classList.add("show");
        sidebar.classList.remove("right-slide-out-500");
        sidebar.classList.add("right-slide-in-500")
        $("#current-state").css("left", `calc(100% - ${right_panel_width}px)`)
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("right-slide-in-500");
        sidebar.classList.add("right-slide-out-500")
        $("#current-state").css("left", `calc(100%)`)
    }
}

function updateDatastore() {
    $("#current-state-div").empty()

    for (let [key, value] of Object.entries(datastore.env)) {
        shouldShow = true
        for (let idx = 0; idx < inputs.length; idx++) {
            let i = inputs[idx]
            if (i.name == key) {
                if (i.type == "password") {
                    shouldShow = false
                    break
                }
            }
        }
        if (!shouldShow) {
            continue
        }
        html = `<div class="w3-bar-item theme-base w3-border-bottom theme-border-light">
            <b>${key}</b>
        </div>
        <div class="w3-bar-item theme-light w3-border-bottom theme-border-base">
            ${value}
        </div>`
        $("#current-state-div").append(html)
    }
}

function getDatastore(shouldUpdate) {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/datastore/" + workflowName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            datastore = result
            if (shouldUpdate) {
                updateDatastore()
            } else {
                getInputs(true, datastore)
            }
            
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    });
}

function goToTop() {
    // $('body').scrollTop(0);
    // This prevents the page from scrolling down to where it was previously.
    if ('scrollRestoration' in history) {
        history.scrollRestoration = 'manual';
    }
    // This is needed if the user scrolls down during page load and you want to make sure the page is scrolled to the top once it's fully loaded. This has Cross-browser support.
    window.scrollTo(0,0);
}

function toggleCheckbox(taskName) {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]
    data = rawTasks[taskName]

    if (document.getElementById(`${taskName}-checkbox`).checked) {
        tasks[taskName].auto_execute = true
        data.auto_execute = true
    } else {
        tasks[taskName].auto_execute = false
        data.auto_execute = false
    }

    tasks[taskName].extra_icons = ''

    if (tasks[taskName].cron != "" && tasks[taskName].cron != null && tasks[taskName].cron != undefined) {
        tasks[taskName].extra_icons = '<i class="fa-solid fa-clock w3-medium" style="float:right;margin-right:4px;margin-left:4px;"></i>'
    }
    if (tasks[taskName].auto_execute) {
        tasks[taskName].extra_icons += '<i class="fa-solid fa-forward w3-medium" style="float:right;margin-right:4px;margin-left:4px;"></i>'
    }

    rawTasks[taskName] = data

    $.ajax({
        url: "/api/v1/task/" + workflowName + "/" + taskName,
        type: "PUT",
        contentType: "application/json",
        dataType: "json",
        data: JSON.stringify(data),
        success: function (result) {
            console.log("Task updated")
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    })
}

function getTasks() {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/task/" + workflowName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            rawTasks = {}
            autoExecuteContents = ""
            for (let task of result) {
                rawTasks[task.name] = task
                autoExecuteContents += `<label class="label cursor-pointer">`
                if (task.auto_execute) {
                    autoExecuteContents += `<input id="${task.name}-checkbox" type="checkbox" class="checkbox" onclick="toggleCheckbox('${task.name}')" checked="checked">`
                } else {
                    autoExecuteContents += `<input id="${task.name}-checkbox" type="checkbox" class="checkbox" onclick="toggleCheckbox('${task.name}')">`
                }
                // autoExecuteContents += '<span class="checkmark"></span>'
                autoExecuteContents += `&nbsp;<span class="label-text">${task.name}</span></label>`
                let status = getStatusFromName(task.name)
                let extra_icons = ""
                if (task.auto_execute) {
                    extra_icons += '<i class="fa-solid fa-forward w3-medium" style="float:right;margin-right:4px;margin-left:4px;"></i>'
                }
                if (task.cron != "") {
                    extra_icons += '<i class="fa-solid fa-clock w3-medium" style="float:right;margin-right:4px;margin-left:4px;"></i>'
                }

                tasks[task.name] = {
                    "title": {
                        "background_color": state_colors_hex[status],
                        "foreground_color": "#ffffff",
                        "text": `${state_icons[status]}&nbsp;&nbsp;${task.name}${extra_icons}`
                    },
                    "out": {},
                    "func": "changeStateName",
                    "disabled": task.disabled,
                    "extra_icons": extra_icons,
                    "parents": []
                }
            }
            $("#auto-execute-div").html(autoExecuteContents)
            for (let task of result) {
                if (task.depends_on.success != null && task.depends_on.success != undefined && task.depends_on.success.length > 0) {
                    for (let name of task.depends_on.success) {
                        console.log(name)
                        tasks[task.name].parents.push(name)
                        if (tasks[name].out['Success'] !== undefined) {
                            tasks[name].out['Success'].push(task.name)
                        } else {
                            tasks[name].out['Success'] = [task.name]
                        }
                    }
                }
                if (task.depends_on.error != null && task.depends_on.error != undefined && task.depends_on.error.length > 0) {
                    for (let name of task.depends_on.error) {
                        tasks[task.name].parents.push(name)
                        if (tasks[name].out['Error'] !== undefined) {
                            tasks[name].out['Error'].push(task.name)
                        } else {
                            tasks[name].out['Error'] = [task.name]
                        }
                    }
                }
                if (task.depends_on.always != null && task.depends_on.always != undefined && task.depends_on.always.length > 0) {
                    for (let name of task.depends_on.always) {
                        tasks[task.name].parents.push(name)
                        if (tasks[name].out['Always'] !== undefined) {
                            tasks[name].out['Always'].push(task.name)
                        } else {
                            tasks[name].out['Always'] = [task.name]
                        }
                    }
                }
            }
            getStates(true)
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    });
}

function getStates(shouldInit) {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/state/" + workflowName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            states = result 
            console.log(states)
            updateNodes()
            if (shouldInit) {
                workflow = new Workflow("workflow-canvas", "workflow-card", "", "", 995, "theme-light", tasks, pin_colors)
                setInterval(function() {
                    workflow.UpdateWorkflow()
                }, workflowIntervalMilliSeconds);
                setInterval(function() {
                    updateStateStatus(false)
                }, stateIntervalMilliSeconds)
            }
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    });
}

function getStatusFromName(n, states) {
    if (states != undefined) {
        for (let state of states) {
            if (state.task == n) {
                return state.status
            }
        }
    }
    return "not_started"
}

function triggerRun() {
    let parts = window.location.href.split('/')
    let workflowName = parts[parts.length - 1]
    let taskName = CurrentStateName
    
    if (taskName != "") {
        $.ajax({
        url: "/api/v1/run/" + workflowName + "/" + taskName,
        type: "POST",
        success: function (result) {
            console.log("Run triggered")
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    });
    }
}

function updateNodes() {
    disabled = []
    for (let state of states) {
        if (state.disabled) {
            disabled.push(state.task)
        }
    }
    for (let [key, task] of Object.entries(tasks)) {
        $(`#${key}`).css("filter", `brightness(100%)`)

        if (hidden.includes(key)) {
            $(`#${key}`).css("filter", `brightness(66%)`)
        }

        if (disabled.includes(key)) {
            $(`#${key}`).css("filter", `brightness(33%)`)
        }
    }
}

function updateStates() {
    if (states != undefined) {
        for (let state of states) {
            if (state.task.startsWith('SCAFFOLD_CHECK') || state.task.startsWith('SCAFFOLD_PREVIOUS')) {
                continue
            }
            let color = state_colors_hex[state.status]
            let icon = state_icons[state.status]
            let extra_icons = tasks[state.task].extra_icons
            let current_color = $(`#${state.task}-header`).css('background-color')
            if (color == current_color) {
                continue
            }
            $(`#${state.task}-header`).css('background-color', color)
            $(`#${state.task}-header-text`).html(`${icon}&nbsp;&nbsp;${state.task}${extra_icons}` )
        }
    }
}

function toggleAutoExecuteMenu() {
    var x = document.getElementById("auto-execute-menu");
    if (x.className.indexOf("w3-show") == -1) { 
        x.className += " w3-show";
    } else {
        x.className = x.className.replace(" w3-show", "");
    }
}

getDatastore(false)
getTasks()
setInterval(function() {
    getStates(false)
}, stateIntervalMilliSeconds)

setInterval(function() {
    getDatastore(true)
}, stateIntervalMilliSeconds)

setInterval(function() {
    updateStates()
}, workflowIntervalMilliSeconds)
