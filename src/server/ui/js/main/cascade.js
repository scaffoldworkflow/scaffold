// import flowchart.js
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import status.js
// import data.js

stateIntervalMilliSeconds = "500"

var cascade
var state
var link_data = []
var node_data = []

it = 20
il = 20
w = 100
h = 50
p = 50
hw = w / 2
hh = h / 2

function getCascade() {
    parts = window.location.href.split('/')
    cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/cascade/" + cascadeName,
        type: "GET",
        success: function (result) {
            cascade = result
            getState(true)
            
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function getState(shouldInit) {
    parts = window.location.href.split('/')
    stateName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/state/" + stateName,
        type: "GET",
        success: function (result) {
            state = result
            if (shouldInit) {
                initFlowchart("cascade-canvas", "cascade-card", [...cascade.tasks])
            }
        },
        error: function (result) {
            console.log(result)
        }
    });
}

$(document).ready(
    function () {
        $("#spinner").css("display", "block")
        $("#page-darken").css("opacity", "1")
        
        getCascade()
    }
)
