var CurrentStateName

function updateStateStatus() {
    let ids = ["cascade-state-header", "cascade-output-header", "cascade-status-header", "cascade-code-header"]
    // for (let k = 0; k < color_keys.length; k++) {
    //     $(`#cascade-check-header`).removeClass(state_colors[color_keys[k]])
    // }
    // $(`#cascade-check-header`).addClass(state_colors['not_started'])
    // $("#state-check").text('')
    // for (let k = 0; k < color_keys.length; k++) {
    //     $(`#cascade-previous-header`).removeClass(state_colors[color_keys[k]])
    // }
    // $(`#cascade-previous-header`).addClass(state_colors['not_started'])
    // $("#state-previous").text('')
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
                $("#state-output").text(state.output)
                $("#toggle-icon").removeClass("fa-toggle-off");
                $("#toggle-icon").removeClass("fa-toggle-on");
                if (state.disabled) {
                    $("#toggle-icon").addClass("fa-toggle-off");
                } else {
                    $("#toggle-icon").addClass("fa-toggle-on");
                }

                // console.log("building display!")
                // console.log(state)
                // console.log(state.display)
                buildDisplay(state.display, "current", color, text_color)
                continue
            }
            // if (state.task == `SCAFFOLD_PREVIOUS-${CurrentStateName}`) {
            //     let color = state_colors[state.status]
            //     let text_color = state_text_colors[state.status]
            //     for (let color of color_keys) {
            //         $(`#previous-header`).removeClass(state_colors[color])
            //     }
            //     $("#previous-run").text(`Run: ${state.number}`)
            //     $(`#previous-header`).addClass(color)
            //     $("#state-previous").text(state.output)
            //     buildDisplay(state.display, "previous", color, text_color)
            //     continue
            // }
            // if (state.task == `SCAFFOLD_CHECK-${CurrentStateName}`) {
            //     let color = state_colors[state.status]
            //     let text_color = state_text_colors[state.status]
            //     for (let color of color_keys) {
            //         $(`#check-header`).removeClass(state_colors[color])
            //     }
            //     $("#check-run").text(`Run: ${state.number}`)
            //     $(`#check-header`).addClass(color)
            //     $("#state-check").text(state.output)
            //     buildDisplay(state.display, "check", color, text_color)
            //     continue
            // }
        }
    }
}

function buildDisplayTable(cd, color, text_color) {
    // create the card
    let output = `<div class="w3-border w3-card light theme-light theme-border-light w3-round">`
    if (cd.name != null && cd.name != undefined && cd.name != "") {
        output += `
            <header class="w3-container ${color}">
                <h4>${cd.name}</h4>
            </header>
        `
    }

    // create the table
    output += `<table class="w3-table light w3-border theme-border-light theme-table-striped">`
    
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
    let output = `<div class="w3-border w3-card light theme-light theme-border-light w3-round">`
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
    let output = `<div class="w3-border w3-card light theme-light theme-border-light w3-round">`
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
    if (display == null || display == undefined || display.length == 0) {
        $(`#state-${item}-display-div`).css("display", "none")
        return
    }

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
    $(`#state-${item}-display-div`).css("display", "block")
}

function killRun() {
    let parts = window.location.href.split('/')
    let c = parts[parts.length - 1]

    let t = CurrentStateName

    if (c == "" || t == "") {
        $("#error-container").text(`Invalid task information, cascade: ${c}, task: ${t}`)
        openModal('error-modal')
        return
    }

    console.log(`Killing run ${c}.${t}`)

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
                window.location.href = "/ui/login";
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
        $("#error-container").text(`Invalid task information, cascade: ${c}, task: ${t}`)
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
                window.location.href = "/ui/login";
            }
            $("#error-container").text(response.responseJSON['error'])
            openModal('error-modal')
        }
    });
}

function changeStateName(name) {
    // Close input
    let input = document.getElementById("current-input")
    input.classList.remove("show");
    $("#current-input").css("left", `calc(100%)`)
    // Close legend
    let legend = document.getElementById("current-legend")
    legend.classList.remove("show");
    $("#current-legend").css("left", `calc(100%)`)
    // Close state
    let state = document.getElementById("current-state")
    state.classList.remove("show");
    $("#current-state").css("left", `calc(100%)`)
    
    if (CurrentStateName != "") {
        closeModal("state-modal")
    }
    CurrentStateName = name
    updateStateStatus()
    openModal("state-modal")
}

function closeStateModal(){
    closeModal("state-modal")
    CurrentStateName = ""
}
