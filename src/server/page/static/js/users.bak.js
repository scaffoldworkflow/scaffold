async function doreq(request, successFunc, errorFunc) {
    try {
        const response = await fetch(request);
        const result = await response.json();
        successFunc(result)
    } catch (error) {
        errorFunc(error)
    }
}

function closeModal(modalID) {
    document.getElementById(modalID).style.display='none'
}
function openModal(modalID) {
    document.getElementById(modalID).style.display='block'
}

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
        "username": document.getElementById("users-add-username").value,
        "password": document.getElementById("users-add-password").value,
        "given_name": document.getElementById("users-add-given-name").value,
        "family_name": document.getElementById("users-add-family-name").value,
        "email": document.getElementById("users-add-email").value,
        "reset_token": "",
        "reset_token_created": "",
        "created": "",
        "updated": "",
        "login_token": "",
        "api_tokens": [],
        "groups": groups,
        "roles": roles
    }

    document.getElementById("spinner").style.display = "block"
    document.getElementById("page-darken").style.opacity = "1"

    let req = new Request("/api/v1/user", {
        method: "POST",
        body: JSON.stringify(data),
        headers: {
            "Content-Type": "application/json",
        }
    });

    doreq(req,
        function(response) {
            username = document.getElementById("users-add-username").value
            window.location.assign('/ui/users/' + username);
        },
        function(response) {
            console.log(response)
            if (response.status == 401) {
                window.location.assign("/ui/login");
            }
            // document.getElementById("error-container").textContent = response.responseJSON['error']
            // openModal('error-modal')
        }
    )

    // data = {
    //     "username": jQuery("users-add-username").val(),
    //     "password": jQuery("users-add-password").val(),
    //     "given_name": jQuery("users-add-given-name").val(),
    //     "family_name": jQuery("users-add-family-name").val(),
    //     "email": jQuery("users-add-email").val(),
    //     "reset_token": "",
    //     "reset_token_created": "",
    //     "created": "",
    //     "updated": "",
    //     "login_token": "",
    //     "api_tokens": [],
    //     "groups": groups,
    //     "roles": roles,
    // }

    // document.getElementById("spinner").style.display = "block"
    // document.getElementById("page-darken").style.opacity = "1"

    // jQuery.ajax({
    //     url: "/api/v1/user",
    //     type: "POST",
    //     contentType: 'application/json',
    //     data: JSON.stringify(data),
    //     success: function(response) {
    //         username = jQuery("#users-add-username").val()
    //         window.location.assign('/ui/users/' + username);
    //     },
    //     error: function(response) {
    //         console.log(response)
    //         if (result.status == 401) {
    //             window.location.assign("/ui/login");
    //         }
    //         jQuery("#error-container").text(response.responseJSON['error'])
    //         openModal('error-modal')
    //     }
    // });
}

function deleteUser(username) {
    // jQuery("#spinner").css("display", "block")
    // jQuery("#page-darken").css("opacity", "1")

    document.getElementById("spinner").style.display = "block"
    document.getElementById("page-darken").style.opacity = "1"

    let req = new Request("/api/v1/user/" + username, {
        method: "DELETE",
        body: JSON.stringify(data),
        headers: {
            "Content-Type": "application/json",
        }
    });

    doreq(req,
        function(response) {
            closeModal('users-delete-modal');
            window.location.reload()
        },
        function(response) {
            closeModal('users-delete-modal');
            console.log(response)
            if (response.status == 401) {
                window.location.assign("/ui/login");
            }
            // document.getElementById("error-container").textContent = response.responseJSON['error']
            // openModal('error-modal')
        }
    )

    // jQuery.ajax({
    //     url: "/api/v1/user/" + username,
    //     type: "DELETE",
    //     success: function(response) {
    //         closeModal('users-delete-modal');
    //         window.location.assign("/ui/users/" + username);
    //     },
    //     error: function(response) {
    //         closeModal('users-delete-modal');
    //         console.log(response)
    //         if (result.status == 401) {
    //             window.location.assign("/ui/login");
    //         }
    //         jQuery("#error-container").text(response.responseJSON['error'])
    //         openModal('error-modal')
    //     }
    // });
}
