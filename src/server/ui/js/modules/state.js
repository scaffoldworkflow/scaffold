var CurrentStateName

function updateStateStatus() {
    ids = ["state-header", "output-header", "status-header", "code-header"]
    for (let k = 0; k < color_keys.length; k++) {
        $(`#check-header`).removeClass(state_colors[color_keys[k]])
    }
    $(`#check-header`).addClass(state_colors['not_started'])
    $("#state-check").text('')
    for (let k = 0; k < color_keys.length; k++) {
        $(`#previous-header`).removeClass(state_colors[color_keys[k]])
    }
    $(`#previous-header`).addClass(state_colors['not_started'])
    $("#state-previous").text('')
    if (CurrentStateName != "" && states != undefined) {
        for (var i = 0; i < states.length; i++) {
            if (states[i].task == CurrentStateName) {
                tempIds = [...ids]
                tempIds.push(states[i].task)
                for (let j = 0; j < tempIds.length; j++) {
                    for (let k = 0; k < color_keys.length; k++) {
                        $(`#${tempIds[j]}`).removeClass(state_colors[color_keys[k]])
                    }
                }
                for (let j = 0; j < tempIds.length; j++) {
                    $(`#${tempIds[j]}`).addClass(state_colors[states[i].status])
                }
                $("#state-name").text(states[i].task)
                $("#state-status").text(`Status: ${states[i].status}`)
                $("#state-started").text(`Started: ${states[i].started}`)
                $("#state-finished").text(`Finished: ${states[i].finished}`)
                $("#state-output").text(states[i].output)
                continue
            }
            if (states[i].task == `SCAFFOLD_CHECK-${CurrentStateName}`) {
                for (let k = 0; k < color_keys.length; k++) {
                    $(`#check-header`).removeClass(state_colors[color_keys[k]])
                }
                $(`#check-header`).addClass(state_colors[states[i].status])
                $("#state-check").text(states[i].output)
                continue
            }
            if (states[i].task == `SCAFFOLD_PREVIOUS-${CurrentStateName}`) {
                for (let k = 0; k < color_keys.length; k++) {
                    $(`#previous-header`).removeClass(state_colors[color_keys[k]])
                }
                $(`#previous-header`).addClass(state_colors[states[i].status])
                $("#state-previous").text(states[i].output)
                continue
            }
        }
    }
}

function changeStateName(name) {
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
