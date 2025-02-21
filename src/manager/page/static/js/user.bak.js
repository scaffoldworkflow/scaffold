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

function addGroup() {
    let groupName = document.getElementById("group-to-add").value
    groupName = groupName.trim()
    if (groupName == "") {
        return
    }
    html = `<div ondblclick="removeGroup('${groupName}')" class="w3-tag w3-round ui-green user-group tag" style="padding:3px" id="group-${groupName}">${groupName}</div>`
    document.getElementById("group-to-add").value = ''
    document.getElementById("group-card").appendChild(html)
}

function removeGroup(name) {
    document.getElementById(`group-${name}`).remove()
}

function addRole() {
    let roleName = document.getElementById("role-to-add").value
    roleName = roleName.trim()
    if (roleName == "") {
        return
    }
    html = `<div ondblclick="removeRole('${roleName}')" class="w3-tag w3-round ui-green user-role tag" style="padding:3px" id="role-${roleName}">${roleName}</div>`
    document.getElementById("role-to-add").value = ''
    document.getElementById("role-card").appendChild(html)
}

function removeRole(name) {
    document.getElementById(`role-${name}`).remove()
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
        "username": document.getElementById("user-add-username").value,
        "password": document.getElementById("user-add-password").value,
        "given_name": document.getElementById("user-add-given-name").value,
        "family_name": document.getElementById("user-add-family-name").value,
        "email": document.getElementById("user-add-email").value,
        "reset_token": "",
        "reset_token_created": "",
        "created": "",
        "updated": "",
        "login_token": "",
        "api_tokens": [],
        "groups": groups,
        "roles": roles
    }

    let req = new Request("/api/v1/user/" + username, {
        method: "PUT",
        body: JSON.stringify(data),
        headers: {
            "Content-Type": "application/json",
        }
    });

    doreq(req,
        function(response) {
            window.location.reload()
        },
        function(response) {
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            // document.getElementById("error-container").textContent = response.responseJSON['error']
            // openModal('error-modal')
        }
    )

    // jQuery.ajax({
    //     url: "/api/v1/user/" + username,
    //     type: "PUT",
    //     contentType: 'application/json',
    //     data: JSON.stringify(data),
    //     success: function(response) {
    //         window.location.reload();
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

function generateAPIToken() {
    parts = window.location.href.split('/')
    username = parts[parts.length - 1]

    tokenName = document.getElementById("user-generate-api-token-name").value

    let req = new Request(`/auth/token/${username}/${tokenName}`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        }
    });

    doreq(req,
        function(response) {
            document.getElementById("spinner").style.display = "none"
            document.getElementById("page-darken").style.opacity = "0"
            token = response.token
            document.getElementById("api-token-field").textContent = token
            document.getElementById("api-token-card").style.display = "inline-block"
            document.getElementById("api-token-generate-button").style.display = "none"
            document.getElementById("api-token-cancel-button").style.display = "none"
            document.getElementById("api-token-done-button").style.display = "block"
            document.getElementById("api-token-copy-button").style.display = "inline-block"
        },
        function(response) {
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            // document.getElementById("error-container").textContent = response.responseJSON['error']
            // openModal('error-modal')
        }
    )

    // jQuery.ajax({
    //     url: `/auth/token/${username}/${tokenName}`,
    //     type: "POST",
    //     contentType: 'application/json',
    //     success: function(response) {
    //         jQuery("#spinner").css("display", "none")
    //         jQuery("#page-darken").css("opacity", "0")
    //         token = response.token
    //         jQuery("#api-token-field").text(token)
    //         jQuery("#api-token-card").css("display", "inline-block")
    //         jQuery("#api-token-generate-button").css("display", "none")
    //         jQuery("#api-token-cancel-button").css("display", "none")
    //         jQuery("#api-token-done-button").css("display", "block")
    //         jQuery("#api-token-copy-button").css("display", "inline-block")
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

function revokeAPIToken(name) {
    parts = window.location.href.split('/')
    username = parts[parts.length - 1]

    let req = new Request(`/auth/token/${username}/${tokenName}`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        }
    });

    doreq(req,
        function(response) {
            document.getElementById("spinner").style.display = "none"
            document.getElementById("page-darken").style.opacity = "0"
            window.location.reload();
        },
        function(response) {
            console.log(response)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            // document.getElementById("error-container").textContent = response.responseJSON['error']
            // openModal('error-modal')
        }
    )

    // jQuery.ajax({
    //     url: `/auth/token/${username}/${name}`,
    //     type: "DELETE",
    //     contentType: 'application/json',
    //     success: function(response) {
    //         jQuery("#spinner").css("display", "none")
    //         jQuery("#page-darken").css("opacity", "0")
    //         window.location.reload();
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
