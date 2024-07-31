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

        console.log(this.tasks)

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
            console.log("Node " + v + ": " + JSON.stringify(g.node(v)));
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
