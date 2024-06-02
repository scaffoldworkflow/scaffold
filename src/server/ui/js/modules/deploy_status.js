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
