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

    AddNodeChildren(node_structure, name, depends) {
        for (let idx = 0; idx < node_structure.length; idx++) {
            if (depends.includes(node_structure[idx].name)) {
                node_structure[idx].children.push({name: name, children: []})
                return {node_structure: node_structure, was_found: true}
            }
        
            let data = this.AddNodeChildren(node_structure[idx].children, name, depends)
            node_structure[idx].children = data.node_structure
            let found = data.was_found
            if (found) {
                return {node_structure: node_structure, was_found: true}
            }
        }
        return {node_structure: node_structure, was_found: false}
    }

    // GetXY(node_structure, x, y) {
    GetXY(g) {

        let positions = {}
        
        let y = this.padding
        for (let r = 0; r < g.length; r++) {
            let x = this.padding
            for (let c = 0; c < g[r].length; c++) {
                let n = g[r][c];
                if (n == "") {
                    x += this.widths[c] + this.padding
                    continue
                }
                let s = this.GetNodeSize(n)
                let delta_x = (this.widths[c] - s.width) / 2
                positions[n] = {name: n, x: x + delta_x, y: y, width: this.widths[c], height: this.heights[r]}
                x += this.widths[c] + this.padding
            }
            y += this.heights[r] + this.padding
        }

        return {positions: positions}
        
        // let positions = {}
        // let max_width = 0

        // for (let node of node_structure) {
        //     let size_data = this.GetNodeSize(node.name)
        //     if (size_data.width > max_width) {
        //         max_width = size_data.width
        //     }
        // }

        // let height = 0
        
        // for (let node of node_structure) {
        //     let size_data = this.GetNodeSize(node.name)

        //     positions[node.name] = {name: node.name, x: x, y: y, width: size_data.width, height: size_data.height}

        //     let data = this.GetXY(node.children, x + max_width + this.padding, y)
        //     let to_add = data.positions
                
        //     y += size_data.height + this.padding

        //     for (let [key, val] of Object.entries(to_add)) {
        //         positions[key] = val
        //     }
        // }
        // return {positions: positions}
    }

    BuildNodeStructure() {
        // let node_structure = []
        // let to_add = {}

        // for (let [key, val] of Object.entries(this.structure)) {
        //     to_add[key] = []
        //     for (let idx = 0; idx < val.length; idx++) {
        //         to_add[key].push(val[idx])
        //     }
        // }

        // let found = true
        // while (found) {
        //     found = false
        //     let to_remove = []
        //     for (let [key, val] of Object.entries(to_add)) {
        //         if (val.length == 0) {
        //             found = true
        //             let data = this.AddNodeChildren(node_structure, key, this.structure[key])
        //             node_structure = data.node_structure
        //             let was_found = data.was_found
        //             if (!was_found) {
        //                 node_structure.push({name: key, children: []})
        //             }
        //             for (let [key2, val2] of Object.entries(to_add)) {
        //                 while(val2.indexOf(key) > -1) {
        //                     to_add[key2].splice(val2.indexOf(key), 1);
        //                 }
        //             }
        //             to_remove.push(key)
        //         }
        //     }
        //     for (let idx = 0; idx < to_remove.length; idx++) {
        //         delete to_add[to_remove[idx]]
        //     }
        // }
        // return node_structure
    }

    SetupNodes() {
        for (let [name, _] of Object.entries(this.tasks)) {
            this.structure[name] = []
        }
        for (let [name, task] of Object.entries(this.tasks)) {
            if (task.out != null && task.out != undefined) {
                for (let [pin, nodes] of Object.entries(task.out)) {
                    for (let node of nodes) {
                        this.links.push({
                            from: name,
                            to: node,
                            pin: pin,
                            color: this.pin_colors[pin],
                        })
                    }
                }
                for (let [_, children] of Object.entries(task.out)) {
                    for (let child of children) {
                        this.structure[child].push(name)
                    }
                }
            }
        }

        let grid = [[]]
        grid = this.GetGrid(0, 0, this.tasks, grid)
        grid = this.TransposeGrid(grid)
        grid = this.RemoveDuplicates(grid)

        this.PrintGrid(grid)
        
        // let node_structure = this.BuildNodeStructure(this.structure)

        // let data = this.GetXY(node_structure, this.canvas.offset().left + this.padding, this.padding)
        let data = this.GetXY(grid)
        let positions = data.positions

        for (let [key, val] of Object.entries(positions)) {
            this.nodes[key] = val
        }
    }

    RenderNodes() {
        $(this.parent).find('div').remove();

        for (let [key, val] of Object.entries(this.nodes)) {
            let bg_color = this.tasks[key].title.background_color
            let fg_color = this.tasks[key].title.foreground_color

            let rows = ''
            let outputs = []
            if (this.tasks[key].out != null && this.tasks[key].out != undefined) {
                outputs = Object.keys(this.tasks[key].out)
            }
            let input_cells = `<td><i class="fa-solid fa-circle-dot" style="color:${this.input_color}" data-pin="${key}.Input"></i></td>
                                <td style="color:${this.input_color}">Input</td>`
            if (this.structure[key].length == 0) {
                input_cells = `<td></td>
                                <td></td>`
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

    TransposeGrid(grid) {
        this.PrintGrid(grid)
        let out = []

        this.widths = []
        for (let i = 0; i < grid.length; i++) {
            this.widths.push(0)
        }

        this.heights = []
        for (let i = 0; i < grid[0].length; i++) {
            this.heights.push(0)
        }

        for (let c = 0; c < grid[0].length; c++) {
            out.push([])
            for (let r = 0; r < grid.length; r++) {
                let row = grid[r]
                out[c].push(row[c])
                if (row[c] == "") {
                    continue
                }
                let size = this.GetNodeSize(row[c])
                if (size.width > this.widths[r]) {
                    this.widths[r] = size.width
                }
                if (size.height > this.heights[c]) {
                    this.heights[c] = size.height
                }
            }
        }

        console.log(out)
        console.log(this.widths)
        console.log(this.heights)

        return out
    }

    GetChildTasks(ts, ns, tr) {
        let out = {}
        for (let [k, t] of Object.entries(ts)) {
            let tc = Object.assign({}, t)
            if (tr.includes(k)) {
                continue
            }
            let parents = []
            for (let p of tc.parents) {
                if (ns.includes(p)) {
                    continue
                }
                parents.push(p)
            }
            tc.parents = [...parents]
            out[k] = tc
        }
        return out
    }

    GetBlankRow(c) {
        let out = []
        for (let i = 0; i < c; i++) {
            out.push('')
        }

        return out
    }

    PrintGrid(g) {
        let ws = []
        for (let i = 0; i < g[0].length; i++) {
            ws.push(0)
        }

        for (let r of g) {
            for (let i = 0; i < r.length; i++) {
                let c = r[i]
                if (c.length > ws[i]) {
                    ws[i] = c.length
                }
            }
        }

        let out = "+"
        for (let w of ws) {
            for (let i = 0; i < w + 2; i++) {
                out += "-"
            }
            out += "+"
        } 
        console.log(out)

        for (let r of g) {
            out = "| "
            for (let i = 0; i < r.length; i++) {
                let c = r[i]
                while (c.length < ws[i]) {
                    c += " "
                }
                out += c + " | "
            }
            console.log(out)
        }

        out = "+"
        for (let w of ws) {
            for (let i = 0; i < w + 2; i++) {
                out += "-"
            }
            out += "+"
        } 
        console.log(out)
    }

    GetGrid(r, c, ts, g) {
        if (ts.length == 0) {
            return g
        }

        let to_process = []
        for (let [k, t] of Object.entries(ts)) {
            if (t.parents.length == 0) {
                to_process.push(k)
            }
        }

        let tsc = Object.assign({}, ts)

        let ns = []

        for (let i = 0; i < to_process.length; i++) {
            let n = to_process[i]
            ns.push(n)
            if (g[0].length == c) {
                for (let j = 0; j < g.length; j++) {
                    g[j].push('')
                }
            }
            if (i > 0) {
                g.push(this.GetBlankRow(g[0].length))
            }
            
            g[r][c] = n
            g = this.GetGrid(r, c + 1, this.GetChildTasks(tsc, ns, to_process), g)
            r = g.length
        }
        return g
    }

    RemoveDuplicates(g) {
        let out = []
        for (let r = 0; r < g.length; r++) {
            let row = g[r]
            let temp = []
            for (let c = 0; c < row.length; c++) {
                let col = row[c]
                if (temp.includes(col)) {
                    temp.push('')
                    continue
                }
                temp.push(col)
            }
            out.push(temp)
        }

        return out
    }
}
