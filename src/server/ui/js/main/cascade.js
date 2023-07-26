// import flowchart.js
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import status.js
// import data.js
// import state.js
// import input.js

stateIntervalMilliSeconds = "500"

tasks = []
states = []
link_data = []
node_data = []
structure = {}
elements = []
positions = {}

it = 20
il = 20
w = 200
h = 50
p = 50
hw = w / 2
hh = h / 2
m = 10

function getTasks() {
    parts = window.location.href.split('/')
    cascadeName = parts[parts.length - 1]

    var foo

    $.ajax({
        url: "/api/v1/task/" + cascadeName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            tasks = result
            getStates(true)
            
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function getStates(shouldInit) {
    parts = window.location.href.split('/')
    cascadeName = parts[parts.length - 1]

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
    cascadeName = parts[parts.length - 1]
    taskName = CurrentStateName
    
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
        $("#spinner").css("display", "block")
        $("#page-darken").css("opacity", "1")
        
        getTasks()
        setInterval(function() {
            getStates(false)
        }, stateIntervalMilliSeconds)

    }
)
