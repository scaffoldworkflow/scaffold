var globalDataStore
var globalInputs

function openInput() {
    getDataStore()
    openModal("input-modal")
}

function getDataStore() {
    cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/datastore/" + cascadeName,
        type: "GET",
        success: function (result) {
            globalDataStore = result
            getInputs()
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function getInputs() {
    cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/input/" + cascadeName,
        type: "GET",
        success: function (result) {
            globalInputs = result
            loadInputData()
        },
        error: function (result) {
            console.log(result)
        }
    });
}

function loadInputData() {
    inputs = [...globalInputs]
    datastore = globalDataStore.env

    $("#input-card").empty()
    $("#input-card").append("<br>")

    for (let i = 0; i < inputs.length; i++) {
        // value = encodeURIComponent(datastore[inputs[i].name])
        value = datastore[inputs[i].name]
        html = `<label>
                ${inputs[i].description}
            </label>
            <input
                class="w3-input"
                type="${inputs[i].type}"
                id="${inputs[i].name}"
            >`
        $("#input-card").append(html)
        $("#input-card").append("<br>")
        $(`#${inputs[i].name}`).val(value)
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
    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    inputs = [...globalInputs]
    datastore = globalDataStore
    
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
