// import flowchart.js
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import deploy_status.js
// import data.js
// import run_state.js
// import input.js
// import accordion.js

stateIntervalMilliSeconds = "500"

tasks = {}
states = []
datastore = {}
link_data = []
node_data = []
structure = {}
elements = []
positions = {}

right_panel_width = 500

var workflow

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

function toggleCurrentInput() {
    let sidebar = document.getElementById("current-input")
    if (sidebar.className.indexOf("show") == -1) {
        // Close state
        let state = document.getElementById("current-state")
        state.classList.remove("show");
        $("#current-state").css("left", `calc(100%)`)
        // Close legend
        let legend = document.getElementById("current-legend")
        legend.classList.remove("show");
        $("#current-legend").css("left", `calc(100%)`)
        // Show input
        sidebar.classList.add("show");
        sidebar.classList.remove("right-slide-out-500");
        sidebar.classList.add("right-slide-in-500")
        $("#current-input").css("left", `calc(100% - ${right_panel_width}px)`)
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("right-slide-in-500");
        sidebar.classList.add("right-slide-out-500")
        $("#current-input").css("left", `calc(100%)`)
    }
}

function toggleCurrentLegend() {
    let sidebar = document.getElementById("current-legend")
    if (sidebar.className.indexOf("show") == -1) {
        // Close state
        let state = document.getElementById("current-state")
        state.classList.remove("show");
        $("#current-state").css("left", `calc(100%)`)
        // Close input
        let input = document.getElementById("current-input")
        input.classList.remove("show");
        $("#current-input").css("left", `calc(100%)`)
        // Show legend
        sidebar.classList.add("show");
        sidebar.classList.remove("right-slide-out-500");
        sidebar.classList.add("right-slide-in-500")
        $("#current-legend").css("left", `calc(100% - ${right_panel_width}px)`)
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("right-slide-in-500");
        sidebar.classList.add("right-slide-out-500")
        $("#current-legend").css("left", `calc(100%)`)
    }
}

function updateDatastore() {
    $("#current-state-div").empty()
    theme = localStorage.getItem('scaffold-theme');

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
        html = `<div class="w3-bar-item ${theme} theme-base w3-border-bottom theme-border-light">
            <b>${key}</b>
        </div>
        <div class="w3-bar-item ${theme} theme-light w3-border-bottom theme-border-base">
            ${value}
        </div>`
        $("#current-state-div").append(html)
    }
}

function getDatastore() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/datastore/" + cascadeName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            datastore = result
            updateDatastore()
            
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function getTasks() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/task/" + cascadeName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            rawTasks = result
            for (let task of rawTasks) {
                out = {}
                if (task.depends_on.success != null && task.depends_on.success != undefined && task.depends_on.success.length > 0) {
                    out['Success'] = task.depends_on.success
                }
                if (task.depends_on.error != null && task.depends_on.error != undefined && task.depends_on.error.length > 0) {
                    out['Error'] = task.depends_on.error
                }
                if (task.depends_on.always != null && task.depends_on.always != undefined && task.depends_on.always.length > 0) {
                    out['Always'] = task.depends_on.always
                }
                tasks[task.name] = {
                    "title": {
                        "background_color": "",
                        "foreground_color": "",
                        "text": task.name 
                    },
                    "out": out,
                    "func": "changeStateName"
                }
            }
            getStates(true)
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function getStates(shouldInit) {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/state/" + cascadeName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            states = result 
            if (shouldInit) {
                initFlowchart("cascade-canvas", "cascade-card")
                setInterval(updateStateStatus, stateIntervalMilliSeconds)
            }
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function triggerRun() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]
    let taskName = CurrentStateName
    
    if (taskName != "") {
        $.ajax({
        url: "/api/v1/run/" + cascadeName + "/" + taskName,
        type: "POST",
        success: function (result) {
            console.log("Run triggered")
        },
        error: function (result) {
            console.log(result)
        }
    });
    }
}

$(document).ready(
    function () {
        // let width = $( document ).width();
        $("#current-state").css("left", `calc(100%)`)
        $("#current-input").css("left", `calc(100%)`)
        $("#sidebar").css("left", "-300px")

        // $("#spinner").css("display", "block")
        // $("#page-darken").css("opacity", "1")

        getTasks()
        setInterval(function() {
            getStates(false)
        }, stateIntervalMilliSeconds)
        setInterval(getDatastore, stateIntervalMilliSeconds)
    }
)
