function addGroup() {
    let groupName = $("#group-to-add").val()
    groupName = groupName.trim()
    if (groupName == "") {
        return
    }
    html = `<div ondblclick="removeGroup('${groupName}')" class="w3-tag w3-round scaffold-green user-group tag" style="padding:3px" id="group-${groupName}">${groupName}</div>`
    $("#group-to-add").val("")
    $("#group-card").append(html)
}

function removeGroup(name) {
    $(`#group-${name}`).remove()
}

function addRole() {
    let roleName = $("#role-to-add").val()
    roleName = roleName.trim()
    if (roleName == "") {
        return
    }
    html = `<div ondblclick="removeRole('${roleName}')" class="w3-tag w3-round scaffold-green user-role tag" style="padding:3px" id="role-${roleName}">${roleName}</div>`
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
        groupData = document.getElementById('group-tags').value.split(',')
    }
    groups = []
    for (var i = 0; i < groupData.length; i++) {
        groups.push(groupData[i].trim())
    }

    roleData = []
    if (document.getElementById('role-tags').value != "") {
        roleData = document.getElementById('role-tags').value.split(',')
    }
    roles = []
    for (var i = 0; i < roleData.length; i++) {
        roles.push(roleData[i].trim())
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

    $.ajax({
        url: "/api/v1/user/" + username,
        type: "PUT",
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            window.location.reload();
        },
        error: function(response) {
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            openModal('error-modal')

        }
    });
}

function generateAPIToken() {
    parts = window.location.href.split('/')
    username = parts[parts.length - 1]

    tokenName = $("#user-generate-api-token-name").val()

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
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            openModal('error-modal')
        }
    });
}

function revokeAPIToken(name) {
    parts = window.location.href.split('/')
    username = parts[parts.length - 1]

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
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            openModal('error-modal')
        }
    });
}
