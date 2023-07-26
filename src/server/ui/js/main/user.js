
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import status.js
// import data.js

var groupTagify
var roleTagify

$(document).ready(
    function() {
        var groupInput = document.getElementById('group-tags'),
        groupTagify = new Tagify(groupInput, {
                id: 'group_tags',
            }
        )
        var roleInput = document.getElementById('role-tags'),
        roleTagify = new Tagify(roleInput, {
                id: 'role_tags',
            }
        )
    }
)

function addGroup() {
    groupName = $("#group-to-add").val()
    html = `<div ondblclick="removeGroup('{{ . }}')" class="w3-tag w3-round scaffold-green user-group tag" style="padding:3px" id="group-${groupName}">
                ${groupName}
            </div>`
    $("#group-to-add").val("")
    $("#group-card").append(html)
}

function removeGroup(name) {
    $(`#group-${name}`).remove()
}

function addRole() {
    roleName = $("#role-to-add").val()
    html = `<div ondblclick="removeRole('{{ . }}')" class="w3-tag w3-round scaffold-green user-role tag" style="padding:3px" id="role-${roleName}">
                ${roleName}
            </div>`
    $("#role-to-add").val("")
    $("#role-card").append(html)
}

function removeRole(name) {
    $(`#role-${name}`).remove()
}

function saveUser() {
    parts = window.location.href.split('/')
    username = parts[parts.length - 1]

    groupData = []
    if (document.getElementById('group-tags').value != "") {
        groupData = JSON.parse(document.getElementById('group-tags').value)
    }
    groups = []
    for (var i = 0; i < groupData.length; i++) {
        groups.push(groupData[i]["value"])
    }

    roleData = []
    if (document.getElementById('role-tags').value != "") {
        roleData = JSON.parse(document.getElementById('role-tags').value)
    }
    roles = []
    for (var i = 0; i < roleData.length; i++) {
        roles.push(roleData[i]["value"])
    }

    data = {
        "username": $("#user-add-username").val(),
        "password": $("#user-add-password").val(),
        "given_name": $("#user-add-given-name").val(),
        "family_name": $("#user-add-family-name").val(),
        "email": $("#user-add-email").val(),
        "reset_token": "",
        "reset_token_created": "",
        "created": "",
        "updated": "",
        "login_token": "",
        "api_tokens": [],
        "groups": groups,
        "roles": roles
    }

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: "/api/v1/user/" + username,
        type: "PUT",
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            window.location.reload();
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

function generateAPIToken() {
    parts = window.location.href.split('/')
    username = parts[parts.length - 1]

    tokenName = $("#user-generate-api-token-name").val()

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: `/auth/token/${username}/${tokenName}`,
        type: "POST",
        contentType: 'application/json',
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            token = response.token
            $("#api-token-field").text(token)
            $("#api-token-card").css("display", "inline-block")
            $("#api-token-generate-button").css("display", "none")
            $("#api-token-cancel-button").css("display", "none")
            $("#api-token-done-button").css("display", "block")
            $("#api-token-copy-button").css("display", "inline-block")
            
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

function revokeAPIToken(name) {
    parts = window.location.href.split('/')
    username = parts[parts.length - 1]

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: `/auth/token/${username}/${name}`,
        type: "DELETE",
        contentType: 'application/json',
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            window.location.reload();
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

function openAPITokenModal() {
    $("#api-token-card").css("display", "none")
    $("#api-token-generate-button").css("display", "block")
    $("#api-token-cancel-button").css("display", "block")
    $("#api-token-done-button").css("display", "none")
    $("#api-token-copy-button").css("display", "none")
    openModal("user-generate-api-token-modal")
}

function closeAPITokenModal() {
    // navigator.clipboard.writeText($("#api-token-field").text());
    closeModal("user-generate-api-token-modal")
    window.location.reload();
}
