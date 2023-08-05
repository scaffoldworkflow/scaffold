// import flowchart.js
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import status.js
// import data.js
// import state.js
// import input.js
// import accordion.js

stateIntervalMilliSeconds = "500"

tasks = []
states = []
datastore = {}
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

function toggleCurrentState() {
    let sidebar = document.getElementById("current-state")
    if (sidebar.className.indexOf("show") == -1) {
        sidebar.classList.add("show");
        sidebar.classList.remove("right-slide-out");
        void sidebar.offsetWidth;
        sidebar.classList.add("right-slide-in")
        $("#current-state").css("left", `calc(100% - 300px)`)
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("right-slide-in");
        void sidebar.offsetWidth;
        sidebar.classList.add("right-slide-out")
        $("#current-state").css("left", `calc(100%)`)
    }
}

function updateDatastore() {
    $("#current-state-div").empty()

    for (let [key, value] of Object.entries(datastore.env)) {
        html = `<div class="w3-bar-item light scaffold-yellow w3-border-bottom theme-border-light">
            <b>${key}</b>
        </div>
        <div class="w3-bar-item light theme-light w3-border-bottom theme-border-light">
            ${value}
        </div>`
        $("#current-state-div").append(html)
    }
}

function getDatastore() {
    parts = window.location.href.split('/')
    cascadeName = parts[parts.length - 1]

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
    parts = window.location.href.split('/')
    cascadeName = parts[parts.length - 1]

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
        // let width = $( document ).width();
        $("#current-state").css("left", `calc(100%)`)
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
