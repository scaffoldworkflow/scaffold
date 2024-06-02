
function toggleSidebar() {
    var sidebar = document.getElementById("sidebar")
    var page_darken = document.getElementById("page-darken")
    if (sidebar.className.indexOf("show") == -1) {
        sidebar.classList.add("show");
        sidebar.classList.remove("left-slide-out-300");
        void sidebar.offsetWidth;
        sidebar.classList.add("left-slide-in-300")
        $("#sidebar").css("left", "0px")

        page_darken.classList.remove("fade-out");
        void page_darken.offsetWidth;
        page_darken.classList.add("fade-in");
        $("#page-darken").css("opacity", "1")
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("left-slide-in-300");
        void sidebar.offsetWidth;
        sidebar.classList.add("left-slide-out-300")
        $("#sidebar").css("left", "-300px")

        page_darken.classList.remove("fade-in");
        void page_darken.offsetWidth;
        page_darken.classList.add("fade-out");
        $("#page-darken").css("opacity", "0")
    }
}

var theme;

$(document).ready(function() {
    theme = localStorage.getItem('scaffold-theme');
    if (theme) {
        if (theme == 'light') {
            $('.dark').addClass('light').removeClass('dark');
        } else {
            $('.light').addClass('dark').removeClass('light');
        }
    } else {
        theme = 'light'
        localStorage.setItem('scaffold-theme', theme);
    }
})

function toggleTheme() {
    if (theme == 'light') {
        theme = 'dark'
        $('.light').addClass('dark').removeClass('light');
    } else {
        theme = 'light'
        $('.dark').addClass('light').removeClass('dark');
    }
    localStorage.setItem('scaffold-theme', theme);
}

function closeModal(modalID) {
    document.getElementById(modalID).style.display='none'
}

function openModal(modalID) {
    document.getElementById(modalID).style.display='block'
}
function toggleMenu() {
    var x = document.getElementById("user-menu");
    if (x.className.indexOf("w3-show") == -1) { 
        x.className += " w3-show";
    } else {
        x.className = x.className.replace(" w3-show", "");
    }
}
allHealthClasses = ""
healthIntervalMilliSeconds = "5000"

function populateVariables() {
    for (var key in healthIcons) {
        allHealthClasses += healthIcons[key] + " "
    }
    for (var key in healthColors) {
        allHealthClasses += healthColors[key] + " "
    }
}

function countJSON(json) {
    var count = 0;
    for (var key in json) {
        if (json.hasOwnProperty(key)) {
            count++;
        }
    }
    return count
}

function updateStatuses() {
    $.ajax({
        url: "/health/status",
        type: "GET",
        success: function (result) {
            tableInnerHTML = '<tr>' +
                '<th class="table-title w3-medium scaffold-text-green">' +
                '</th>' +
                '<th class="table-title w3-medium scaffold-text-green">' +
                    '<span class="table-title-text">Service</span>' +
                '</th>' +
                '<th class="table-title w3-medium scaffold-text-green">' +
                    '<span class="table-title-text">IP</span>' +
                '</th>' +
                '<th class="table-title w3-medium scaffold-text-green">' +
                    '<span class="table-title-text">Status</span>' +
                '</th>' +
                '<th class="table-title w3-medium scaffold-text-green">' +
                    '<span class="table-title-text">Version</span>' +
                '</th>' +
            '</tr>'

            let is_healthy = true
            let down_count = 0

            for (let i = 0; i < result.nodes.length; i++) {
                serviceStatusColor = healthColors[result.nodes[i].status]
                serviceStatusText = healthText[result.nodes[i].status]
                serviceStatusIcon = healthIcons[result.nodes[i].status]
                serviceStatusVersion = result.nodes[i].version
                if (result.nodes[i].status != 'healthy') {
                    is_healthy = false
                    down_count += 1
                }
                tableInnerHTML += '<tr>' +
                    '<td class="status-table-icon fa-solid ' + serviceStatusColor + ' ' + serviceStatusIcon + '">' + '</td>' +
                    '<td>' + result.nodes[i].name + '</td>' +
                    '<td>' + result.nodes[i].ip + '</td>' +
                    '<td class="status-table-status">' + serviceStatusText + '</td>' +
                    '<td>' + serviceStatusVersion + '</td>' +
                '</tr>'
            }
            $("#status-table").html(tableInnerHTML)
            $("#status-icon").removeClass(allHealthClasses)
            if (down_count == result.nodes.length) {
                $("#status-icon").addClass(healthIcons["unhealthy"] + " " + healthColors["unhealthy"])
            } else if (is_healthy) {
                $("#status-icon").addClass(healthIcons["healthy"] + " " + healthColors["healthy"])
            } else {
                $("#status-icon").addClass(healthIcons["degraded"] + " " + healthColors["healthy"])
            }
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
            tableElements = $('.status-table-icon')
            for (let i = 0; i < tableElements.length; i++) {
                $(tableElements[i]).removeClass(allHealthClasses)
                $(tableElements[i]).addClass(healthColors['unknown'])
                $(tableElements[i]).addClass(healthIcons['unknown'])
            }

            tableElements = $('.status-table-status')
            for (let i = 0; i < tableElements.length; i++) {
                $(tableElements[i]).text(healthText['unknown'])
            }

            $("#status-icon").removeClass(allHealthClasses)
            $("#status-icon").addClass(healthIcons["unknown"] + " " + healthColors["unknown"])
        }
    });
}

function showStatus() {
    toggleSidebar()
    openModal('status-modal')
}

$(document).ready(
    function() {
        populateVariables()
        updateStatuses()
        setInterval(updateStatuses, healthIntervalMilliSeconds);
    }
)

var healthIcons = {
	'healthy': 'fa-circle-check',
	'degraded': 'fa-circle-exclamation',
	'unhealthy': 'fa-circle-xmark',
	'unknown': 'fa-circle-question'
}

var healthColors = {
	'healthy': 'scaffold-text-green',
	'degraded': 'scaffold-text-yellow',
	'unhealthy': 'scaffold-text-red',
	'unknown': 'scaffold-text-charcoal'
}

var healthText = {
	'healthy': 'Up',
	'degraded': 'Degraded',
	'unhealthy': 'Down',
	'unknown': 'Unknown'
}

var icons = {
	'healthy': 'fa-circle-check',
	'warn': 'fa-circle-exclamation',
	'error': 'fa-circle-xmark',
	'deploying': 'fa-spinner fa-pulse',
	'unknown': 'fa-circle-question',
	'not-deployed': 'fa-circle'
}

var colors = {
	'healthy': 'green',
	'warn': 'yellow',
	'error': 'red',
	'deploying': 'grey',
	'unknown': 'orange',
	'not-deployed': 'grey'
}


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
