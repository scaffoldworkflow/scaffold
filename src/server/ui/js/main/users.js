
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import deploy_status.js
// import data.js

var users
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
    groupName = groupName.trim()
    if (groupName == "") {
        return
    }
    html = `<div ondblclick="removeGroup('{{ . }}')" class="w3-tag w3-round scaffold-green user-group tag" style="padding:3px" id="group-${groupName}">${groupName}</div>`
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
    html = `<div ondblclick="removeRole('{{ . }}')" class="w3-tag w3-round scaffold-green user-role tag" style="padding:3px" id="role-${roleName}">${roleName}</div>`
    $("#role-to-add").val("")
    $("#role-card").append(html)
}

function removeRole(name) {
    $(`#role-${name}`).remove()
}

function getUsers() {
    parts = window.location.href.split('/')
    $.ajax({
        url: "/api/v1/user",
        type: "GET",
        success: function (result) {
            users = result
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
        for (let idx = 0; idx < users.length; idx++) {
            let name = users[idx].username
            $(`#users-row-${name}`).removeClass("table-hide")
            $(`#users-row-${name}`).addClass("table-show")
        }
        return
    }

    for (let idx = 0; idx < users.length; idx++) {
        let name = users[idx].username
        if (name.toLowerCase().indexOf(prefix) == -1) {
            $(`#users-row-${name}`).removeClass("table-show")
            $(`#users-row-${name}`).addClass("table-hide")
            continue
        }
        $(`#users-row-${name}`).removeClass("table-hide")
        $(`#users-row-${name}`).addClass("table-show")
    } 
}

function addUser() {
    // roles = []
    // roleElements = $('.user-role')
    // for (let i = 0; i < roleElements.length; i++) {
    //     roles.push($(roleElements[i]).text().trim())
    // }

    // groups= []
    // groupElements = $('.user-group')
    // for (let i = 0; i < groupElements.length; i++) {
    //     groups.push($(groupElements[i]).text().trim())
    // }

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
        "username": $("#users-add-username").val(),
        "password": $("#users-add-password").val(),
        "given_name": $("#users-add-given-name").val(),
        "family_name": $("#users-add-family-name").val(),
        "email": $("#users-add-email").val(),
        "reset_token": "",
        "reset_token_created": "",
        "created": "",
        "updated": "",
        "login_token": "",
        "api_tokens": [],
        "groups": groups,
        "roles": roles,
    }

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: "/api/v1/user",
        type: "POST",
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            username = $("#users-add-username").val()
            closeModal('users-add-modal');
            window.location.assign('/ui/user/' + username);
        },
        error: function(response) {
            closeModal('users-delete-modal');
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

function openDeleteModal(username, id) {
    $("#user-delete-username").text(username)
    openModal("users-delete-modal");
}

function deleteUser() {
    parts = window.location.href.split('/')

    username = $("#user-delete-username").text()

    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: "/api/v1/user/" + username,
        type: "DELETE",
        success: function(response) {
            $("#spinner").css("display", "none")
            $("#page-darken").css("opacity", "0")
            closeModal('users-delete-modal');
            window.location.reload()
        },
        error: function(response) {
            closeModal('users-delete-modal');
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
        getUsers()
    }
)
