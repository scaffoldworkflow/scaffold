
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import status.js
// import data.js

var users

function getUsers() {
    parts = window.location.href.split('/')
    $.ajax({
        url: "/api/v1/user",
        type: "GET",
        success: function (result) {
            users = result
        }
    });
}

function render() {
    var prefix = $("#search").val();
    var tempUsers = JSON.parse(JSON.stringify(users));
    prefix = prefix.toLowerCase();
    if (prefix != "") {
        for (var idx = tempUsers.length - 1; idx >= 0; idx--) {
            if (tempUsers[idx].name.toLowerCase().indexOf(prefix) == -1) {
                tempUsers.splice(idx, 1)
            }
        }
    }
    var table = document.getElementById('users-table');
    var tableHTMLString = '<tr><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Username</span></th><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Email</span></th><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Given Name</span></th><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Family Name</span></th><th><div class="w3-round w3-button scaffold-green"></div></th></tr>' +
        tempUsers.map(function (user) {
            return '<tr>' +
                '<td>' + user.username + '</td>' +
                '<td>' + user.email + '</td>' +
                '<td>' + user.given_name + '</td>' +
                '<td>' + user.family_name + '</td>' +
                '<td class="table-link-cell">' +
                '<a href="/ui/user/' + user.username + '" class="table-link-link w3-right-align light theme-text" style="float:right;margin-right:16px;"><i class="fa-solid fa-link"></i></a>' +
                '</td>' +
                '</tr>'
        }).join('');

    table.innerHTML = tableHTMLString;
}

function addUser() {
    parts = window.location.href.split('/')

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
            $("#log-container").text(response.responseJSON['error'])
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
            $("#log-container").text(response.responseJSON['error'])
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
