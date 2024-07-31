
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import deploy_status.js
// import data.js
// import table.js

var webhooks

function getWebhooks() {
    parts = window.location.href.split('/')
    $.ajax({
        url: "/api/v1/webhook",
        type: "GET",
        success: function (result) {
            webhooks = result
        },
        error: function(result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    });
}

function render() {
    var prefix = $("#search").val();
    prefix = prefix.toLowerCase();

    if (prefix == "") {
        for (let idx = 0; idx < webhooks.length; idx++) {
            let id = webhooks[idx].id
            $(`#webhooks-row-${id}`).removeClass("table-hide")
            $(`#webhooks-row-${id}`).addClass("table-show")
        }
        return
    }

    for (let idx = 0; idx < webhooks.length; idx++) {
        let id = webhooks[idx].id
        if (id.toLowerCase().indexOf(prefix) == -1) {
            $(`#webhooks-row-${id}`).removeClass("table-show")
            $(`#webhooks-row-${id}`).addClass("table-hide")
            continue
        }
        $(`#webhooks-row-${id}`).removeClass("table-hide")
        $(`#webhooks-row-${id}`).addClass("table-show")
    } 
}

function addUser() {
    data = {
        "entrypoint": $("#webhooks-add-entrypoint").val(),
        "cascade": $("#webhooks-add-cascade").val(),
    }

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: "/api/v1/webhook",
        type: "POST",
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            closeModal('webhooks-add-modal');
            window.location.reload()
        },
        error: function(response) {
            closeModal('webhooks-delete-modal');
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            openModal('error-modal')
        }
    });
}

function openDeleteModal(id) {
    $("#webhook-delete-id").text(id)
    openModal("webhooks-delete-modal");
}

function deleteWebhook() {
    parts = window.location.href.split('/')

    id = $("#webhook-delete-id").text()

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: "/api/v1/webhook/" + id,
        type: "DELETE",
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            closeModal('webhooks-delete-modal');
            window.location.reload()
        },
        error: function(response) {
            closeModal('webhooks-delete-modal');
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            openModal('error-modal')
        }
    });
}

$(document).ready(
    function() {
        getWebhooks()
    }
)
