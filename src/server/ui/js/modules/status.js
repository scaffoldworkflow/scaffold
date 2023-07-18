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
        url: "/status",
        type: "GET",
        success: function (result) {
            tableInnerHTML = '<tr>' +
                '<th class="table-title w3-medium scaffold-text-green">' +
                    '<span class="table-title-text">Service</span>' +
                '</th>' +
                '<th class="table-title w3-medium scaffold-text-green">' +
                    '<span class="table-title-text">Status</span>' +
                '</th>' +
            '</tr>'

            totalCount = countJSON(result);
            healthyCount = 0;
            for (var key in result) {
                if (result[key]) {
                    healthyCount++;
                }
                serviceStatusColor = result[key] ? "scaffold-text-green" : "scaffold-text-green"
                serviceStatusText = result[key] ? "Up" : "Down"
                tableInnerHTML += '<tr>' +
                    '<td>' + key + '</td>' +
                    '<td class="' + serviceStatusColor + '">' + serviceStatusText + '</td>' +
                '</tr>'
            }
            $("#status-table").html(tableInnerHTML)
            $("#status-icon").removeClass(allHealthClasses)
            if (totalCount == healthyCount) {
                $("#status-icon").addClass(healthIcons["healthy"] + " " + healthColors["healthy"])
            } else if (healthyCount == 0) {
                $("#status-icon").addClass(healthIcons["unhealthy"] + " " + healthColors["unhealthy"])
            } else {
                $("#status-icon").addClass(healthIcons["degraded"] + " " + healthColors["degraded"])
            }
        },
        error: function (result) {
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
        // updateStatuses()
        // setInterval(updateStatuses, healthIntervalMilliSeconds);
    }
)
