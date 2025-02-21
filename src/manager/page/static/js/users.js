function addUser() {
    groupData = []
    if (document.getElementById('users-add-groups').value != "") {
        groupData = document.getElementById('users-add-groups').value.split(',')
    }
    groups = []
    for (var i = 0; i < groupData.length; i++) {
        groups.push(groupData[i].trim())
    }

    roleData = []
    if (document.getElementById('users-add-roles').value != "") {
        roleData = document.getElementById('users-add-roles').value.split(',')
    }
    roles = []
    for (var i = 0; i < roleData.length; i++) {
        roles.push(roleData[i].trim())
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
            username = $("#users-add-username").val()
            window.location.assign('/ui/users/' + username);
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

function deleteUser(username) {
    $("#spinner").css("display", "block")
    $("#page-darken").css("opacity", "1")

    $.ajax({
        url: "/api/v1/user/" + username,
        type: "DELETE",
        success: function(response) {
            closeModal('users-delete-modal');
            window.location.assign("/ui/users/" + username);
        },
        error: function(response) {
            closeModal('users-delete-modal');
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            $("#error-container").text(response.responseJSON['error'])
            openModal('error-modal')
        }
    });
}
