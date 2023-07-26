var CurrentStateName

function updateStateStatus() {
    ids = ["state-header", "output-header", "status-header", "code-header"]
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
                $("#state-code").text(tasks[i].run)
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
