var datastore
var inputs  

function openInput() {
    getDataStore()
    toggleCurrentInput()
}

function getDataStore() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/datastore/" + cascadeName,
        type: "GET",
        success: function (result) {
            datastore = result
            getInputs(true)
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function getInputs(trigger) {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/input/" + cascadeName,
        type: "GET",
        success: function (result) {
            inputs = result
            if (trigger) {
                loadInputData()
            }
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function loadInputData() {
    $("#current-input-div").empty()
    theme = localStorage.getItem('scaffold-theme');

    let html = `<div class="w3-bar-item ${theme} scaffold-green w3-border-bottom theme-border-light w3-button" onclick="saveInputs()" >
        <i class="fa-solid fa-floppy-disk" id="save-icon"></i>&nbsp;Save input changes
    </div>`
    $("#current-input-div").append(html)

    for (let idx = 0; idx < inputs.length; idx++) {
        let i = inputs[idx]
        let value = datastore.env[i.name]
        let html = `<div class="w3-bar-item ${theme} theme-base w3-border-bottom theme-border-light">
            <b>${i.description}</b>
        </div>
        <div class="w3-bar-item ${theme} theme-light w3-border-bottom theme-border-light">
            <input
                class="w3-input ${theme} theme-light"
                type="${i.type}"
                id="${i.name}"
            >
        </div>`
        $("#current-input-div").append(html)
        $(`#${i.name}`).val(value)
    }
}

function changeIcon() {
    if ($("#save-icon").hasClass("fa-floppy-disk")) {
        $("#save-icon").removeClass("fa-floppy-disk")
        $("#save-icon").addClass("fa-check")
    } else {
        $("#save-icon").removeClass("fa-check")
        $("#save-icon").addClass("fa-floppy-disk")
    }
}

function saveInputs() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")
    
    for (let i = 0; i < inputs.length; i++) {
        value = $(`#${inputs[i].name}`).val()
        datastore.env[inputs[i].name] = value
    }

    $.ajax({
        type: "PUT",
        url: `/api/v1/datastore/${cascadeName}`,
        contentType: "application/json",
        dataType: "json",
        data: JSON.stringify(datastore),
        success: function(response) {
            console.log(response)
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            changeIcon()
            setInterval(changeIcon, 1500)
        },
        error: function(response) {
            console.log(response)
            $("#error-container").text(response.responseJSON['error'])
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            openModal('error-modal')
        }
    });
}

$(document).ready(
    function() {
        getInputs(false)
    }
)
