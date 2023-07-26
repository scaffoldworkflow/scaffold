
var CANVAS_ID
var CARD_ID
var flowchartUpdateIntervalMilliSeconds = "50"

var state_colors = {
    "not_started": "scaffold-charcoal",
    "success": "scaffold-green",
    "error": "scaffold-red",
    "running": "scaffold-blue",
    "waiting": "scaffold-yellow"
}
color_keys = ["not_started", "success", "error", "running", "waiting"]

function add_node_children(els, name, depends) {
    for (let idx = 0; idx < els.length; idx++) {
        if (depends.includes(els[idx].name)) {
            els[idx].children.push({name: name, children: []})
            return {elements: els, was_found: true}
        }
    
        let data = add_node_children(els[idx].children, name, depends)
        els[idx].children = data.elements
        found = data.was_found
        if (found) {
            return {elements: els, was_found: true}
        }
    }
    return {elements: els, was_found: false}
}

function build_node_structure(structure) {
    let els = []
    let to_add = {}

    for (let [key, val] of Object.entries(structure)) {
        to_add[key] = []
        for (let idx = 0; idx < val.length; idx++) {
            to_add[key].push(val[idx])
        }
    }

    let found = true
    while (found) {
        found = false
        let to_remove = []
        for (let [key, val] of Object.entries(to_add)) {
            if (val.length == 0) {
                found = true
                let data = add_node_children(els, key, structure[key])
                els = data.elements
                let was_found = data.was_found
                if (!was_found) {
                    els.push({name: key, children: []})
                }
                for (let [key2, val2] of Object.entries(to_add)) {
                    var idx = val2.indexOf(key);
                    if (idx !== -1) {
                        to_add[key2].splice(idx, 1);
                    }
                }
                to_remove.push(key)
            }
        }
        for (let idx = 0; idx < to_remove.length; idx++) {
            delete to_add[to_remove[idx]]
        }
    }
    return els
}

function get_node_width(name) {
    html = `<div hidden
            class="w3-round w3-border"
            style="height:${h}px;position:absolute;text-align:center;vertical-align:middle;line-height:${h}px;z-index:998;"
            id="placeholder">${name}</div>`
    $("#cascade-card").append(html)
    width = $("#placeholder").width() + (m* 2)
    $("#placeholder").remove()
    return width
}

function get_x_y(elements, x, y) {
    let out = {}
    let width = 0
    for (let idx = 0; idx < elements.length; idx++) {
        let el = elements[idx]
        let node_width = get_node_width(el.name)

        out[el.name] = {name: el.name, x: x, y: y, w: node_width, hw: node_width / 2}
        width += node_width

        data = get_x_y(el.children, x, y + h + p)
        let to_add = data.positions
        let width_modifier = data.width

        if (width_modifier > width) {
            x += width_modifier + p
        } else {
            x += width + p
        }

        for (let [key, val] of Object.entries(to_add)) {
            out[key] = val
        }
    }
    return {positions: out, width: width}
}

function getLinkCoordsFromName(n, nd) {
    for (let i = 0; i < nd.length; i++) {
        if (nd[i].name == n) {
            return {x: nd[i].x, y: nd[i].y}
        }
    }
    return {x: 0, y: 0}
}

function getStatusFromName(n, states) {
    if (states != undefined) {
        for (let i = 0; i < states.length; i++) {
            if (states[i].task== n) {
                return states[i].status
            }
        }
    }
    return "not_started"
}

function checkIfDependsOn(n, tasks) {
    for (let i = 0; i < tasks.length; i++) {
        if (tasks[i].name == n) {
            if (tasks[i].depends_on == null) {
                return false
            }
            return true
        }
    }
    return false
}

function checkIfDependedOn(n, tasks) {
    for (let i = 0; i < tasks.length; i++) {
        if (tasks[i].depends_on != null) {
            if (tasks[i].depends_on.includes(n)) {
                return true
            }
        }
    }
    return false
}

function setupNodes() {
    for (let idx = 0; idx < tasks.length; idx++) {
        if (tasks[idx].depends_on != null) {
            // Create links as needed
            for (let jdx = 0; jdx < tasks[idx].depends_on.length; jdx++) {
                link_data.push({
                    from: tasks[idx].depends_on[jdx],
                    to: tasks[idx].name
                })
                console.log("Depends on link")
            }
            structure[tasks[idx].name] = [...tasks[idx].depends_on]
        } else {
            structure[tasks[idx].name] = []
        }
    }
    
    elements = build_node_structure(structure)

    let data = get_x_y(elements, 500, p)
    positions = data.positions

    for (let [_, val] of Object.entries(positions)) {
        node_data.push(val)
    }
}

function renderNodes(cardID) {
    $(`#${cardID}`).find('div').remove()

    do {
        
    } while (tasks == null || states == null)

    for (let i = 0; i < node_data.length; i++) {
        current_status = getStatusFromName(node_data[i].name, states)
        color = state_colors[current_status]
        dependedOn = checkIfDependedOn(node_data[i].name, tasks)
        dependsOn = checkIfDependsOn(node_data[i].name, tasks)
        html = `<div 
            class="w3-round w3-border theme-border-light ${color} light"
            style="width:${node_data[i].w}px;height:${h}px;top:${node_data[i].y-hh}px;left:${node_data[i].x-node_data[i].hw}px;position:absolute;text-align:center;vertical-align:middle;line-height:${h}px;z-index:998;"
            id="${node_data[i].name}"
            ondblclick="changeStateName('${node_data[i].name}')">
            ${node_data[i].name}`
        if (dependsOn) {
            html += `<div class="arrow-down" style="top:-10px;left:${node_data[i].hw - 10}px;position:absolute;">`
        }
        if (dependedOn) {
            if (dependsOn) {
                html += `<div class="arrow-down" style="top:${h}px;left:-10px;position:absolute;">`
            } else {
                html += `<div class="arrow-down" style="top:${h}px;left:${node_data[i].hw - 10}px;position:absolute;">`
            }
        }
        html += '</div>'
        $(`#${cardID}`).append(html)
        $(`#${node_data[i].name}`).draggable()
    }
}

function updateNodeColors() {
    for (let i = 0; i < node_data.length; i++) {
        current_status = getStatusFromName(node_data[i].name, states)
        color = state_colors[current_status]

        for (let j = 0; j < color_keys.length; j++) {
            $(`#${node_data[i].name}`).removeClass(state_colors[color_keys[j]])
        }
        $(`#${node_data[i].name}`).addClass(color)
    }
}

function renderLines(canvasID) {
    canvas = document.getElementById(canvasID);
    ctx = canvas.getContext("2d");
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.lineWidth = 4;
    ctx.lineCap = "round"

    ctx.beginPath();
    
    for (let i = 0; i < link_data.length; i++) {
        s = getLinkCoordsFromName(link_data[i].from, node_data)
        s.y += hh + 10
        e = getLinkCoordsFromName(link_data[i].to, node_data)
        e.y -= hh + 10
        delta = 10
        c_1 = {x: s.x, y: s.y + delta}
        c_2 = {x: e.x, y: e.y - delta}
        
        ctx.moveTo(e.x, e.y);
        ctx.bezierCurveTo(c_1.x, c_1.y, c_2.x, c_2.y, s.x, s.y);    
    }
    ctx.stroke()
    ctx.closePath()
}

function updateFlowchart(canvasID, cardID) {
    for (let i = 0; i < node_data.length; i++) {
        pos = $(`#${node_data[i].name}`).position()
        if (pos != undefined) {
            node_data[i].x = pos.left + node_data[i].hw
            node_data[i].y = pos.top + hh
        }
    }

    updateNodeColors()
    renderLines(canvasID, cardID)
}

function sleep(milliseconds) {
    const date = Date.now();
    let currentDate = null;
    do {
        currentDate = Date.now();
    } while (currentDate - date < milliseconds);
}

function initFlowchart(canvasID, cardID) {
    // Set canvas ratio to match screen
    let parent_width = $(`#${cardID}`).width()
    $(`#${cardID}`).append(`<canvas 
            style="position:absolute;top:0px;left:0px;height:1000px;width:${parent_width}px;z-index:1;"
            height="1000px"
            width="${parent_width}px"
            id="${canvasID}"
            draggable="true">
    </canvas>`)
    $(`#${canvasID}`).width(`${parent_width}px`)
    $(`#${canvasID}`).css("width", `${parent_width}px`)

    CANVAS_ID = canvasID
    CARD_ID = cardID

    setupNodes()
    renderNodes(cardID)
    renderLines(canvasID)

    $("#spinner").css("display", "none")
    $("#page-darken").css("opacity", "0")

    setInterval(function() {
        updateFlowchart(canvasID, cardID)
    }, flowchartUpdateIntervalMilliSeconds);
}
