
var CANVAS_ID
var CARD_ID
var flowchartUpdateIntervalMilliSeconds = "50"

var state_colors = {
    "not_started": "scaffold-yellow",
    "success": "scaffold-green",
    "error": "scaffold-red",
    "running": "scaffold-blue"
}

function sort_nodes(s, n, m, t) {
    for (let i = 0; i < s.length; i++) {
        var temp_idx = []
        var temp_name = []
        for (let j = 0; j < n.length; j++) {
            if (t[n[j]].depends_on.includes(s[i])) {
                temp_idx.push(n[j])
                temp_name.push(t[n[j]].name)
                n.splice(j, 1)
                break
            }
        }
        for (let j = 0; j < n.length; j++) {
            temp_idx.push(n[j])
            temp_name.push(t[n[j]].name)
            n.splice(j, 1)
        }
        n = [...temp_idx]
        m = [...temp_name]
    }
    return {idx: n, name: m}
}

function getLinkCoordsFromName(n, nd) {
    for (let i = 0; i < nd.length; i++) {
        if (nd[i].name == n) {
            return {x: nd[i].x, y: nd[i].y}
        }
    }
    return {x: 0, y: 0}
}

function getStatusFromName(n, tasks) {
    for (let i = 0; i < tasks.length; i++) {
        if (tasks[i].name == n) {
            console.log(`status: ${tasks[i].status}`)
            return tasks[i].status
        }
    }
    console.log("status: not_started")
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

function setupNodes(tasks) {
    sort_order = []
    nodes = []

    x = il 
    y = it
    while (tasks.length > 0) {
        var to_remove = []
        var nodes_names = []
        var nodes_idx = []
        var to_sort = []
        for (let i = 0; i < tasks.length; i++) {
            should_add = true
            if (tasks[i].depends_on != null) {
                for (let j = 0; j < tasks[i].depends_on.length; j++) {
                    if (!nodes.includes(tasks[i].depends_on[j])) {
                        should_add = false
                        break
                    }
                } 
            }
            if (should_add) {
                to_sort.push(tasks[i].name)
                
                nodes_names.push(tasks[i].name)
                nodes_idx.push(i)
                to_remove.push(i)
                
                if (tasks[i].depends_on != null) {
                    for (let j = 0; j < tasks[i].depends_on.length; j++) {
                        link_data.push({
                            from: tasks[i].depends_on[j],
                            to: tasks[i].name,
                        })
                    }
                }
            }
        }

        sorted = sort_nodes(sort_order, nodes_idx, nodes_names, tasks)

        for (let i = 0; i < sorted.idx.length; i++) {
            node_data.push({
                name: sorted.name[i],
                x: x + hw,
                y: y + hh,
            })
            x += w + p
        }

        y += h + p
        x = il
        
        sort_order.push(...to_sort)
        nodes.push(...nodes_names)

        to_remove.sort(function(a, b){return b - a});

        for (let j = 0; j < to_remove.length; j++) {
            tasks.splice(to_remove[j], 1);
        }
    }
}

function renderNodes(cardID) {
    $(`#${cardID}`).find('div').remove()

    do {
        console.log(`cascade: ${cascade}`)
        console.log(`state: ${state}`)
    } while (cascade == null || state == null)

    for (let i = 0; i < node_data.length; i++) {
        current_status = getStatusFromName(node_data[i].name, state.tasks)
        color = state_colors[current_status]
        dependedOn = checkIfDependedOn(node_data[i].name, cascade.tasks)
        dependsOn = checkIfDependsOn(node_data[i].name, cascade.tasks)
        html = `<div 
            class="w3-round w3-border theme-border-light ${color} light"
            style="width:${w}px;height:${h}px;top:${node_data[i].y-hh}px;left:${node_data[i].x-hw}px;position:absolute;text-align:center;vertical-align:middle;line-height:${h}px;z-index:9999;"
            id="${node_data[i].name}">
            ${node_data[i].name}`
        if (dependsOn) {
            html += `<div class="arrow-down" style="top:-10px;left:${hw - 10}px;position:absolute;">`
        }
        if (dependedOn) {
            if (dependsOn) {
                html += `<div class="arrow-down" style="top:${h}px;left:-10px;position:absolute;">`
            } else {
                html += `<div class="arrow-down" style="top:${h}px;left:${hw - 10}px;position:absolute;">`
            }
        }
        html += '</div>'
        $(`#${cardID}`).append(html)
        $(`#${node_data[i].name}`).draggable()
    }
}

function renderLines(canvasID, cardID) {
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
        c_1 = {x: s.x, y: (s.y + e.y) / 2}
        c_2 = {x: e.x, y: (s.y + e.y) / 2}
        
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
            node_data[i].x = pos.left + hw
            node_data[i].y = pos.top + hh
        }
    }

    renderLines(canvasID, cardID)
}

function sleep(milliseconds) {
    const date = Date.now();
    let currentDate = null;
    do {
        currentDate = Date.now();
    } while (currentDate - date < milliseconds);
}

function initFlowchart(canvasID, cardID, tasks) {
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

    setupNodes(tasks)
    renderNodes(cardID)
    renderLines(canvasID, cardID)

    $("#spinner").css("display", "none")
    $("#page-darken").css("opacity", "0")

    setInterval(function() {
        updateFlowchart(canvasID, cardID)    
    }, flowchartUpdateIntervalMilliSeconds);
}
