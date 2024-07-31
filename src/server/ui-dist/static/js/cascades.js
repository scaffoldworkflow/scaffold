
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

function sortTable(n, tableID) {
    var table, rows, switching, i, x, y, shouldSwitch, dir, switchcount = 0;
    table = document.getElementById(tableID);
    switching = true;
    // Set the sorting direction to ascending:
    dir = "asc";
    /* Make a loop that will continue until
    no switching has been done: */
    while (switching) {
        // Start by saying: no switching is done:
        switching = false;
        rows = table.rows;
        /* Loop through all table rows (except the
        first, which contains table headers): */
        for (i = 1; i < (rows.length - 1); i++) {
            // Start by saying there should be no switching:
            shouldSwitch = false;
            /* Get the two elements you want to compare,
            one from current row and one from the next: */
            x = rows[i].getElementsByTagName("TD")[n];
            y = rows[i + 1].getElementsByTagName("TD")[n];
            /* Check if the two rows should switch place,
            based on the direction, asc or desc: */
            if (dir == "asc") {
                if (x.innerHTML.toLowerCase() > y.innerHTML.toLowerCase()) {
                    // If so, mark as a switch and break the loop:
                    shouldSwitch = true;
                    break;
                }
            } else if (dir == "desc") {
                if (x.innerHTML.toLowerCase() < y.innerHTML.toLowerCase()) {
                    // If so, mark as a switch and break the loop:
                    shouldSwitch = true;
                    break;
                }
            }
        }
        if (shouldSwitch) {
            /* If a switch has been marked, make the switch
            and mark that a switch has been done: */
            rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
            switching = true;
            // Each time a switch is done, increase this count by 1:
            switchcount++;
        } else {
            /* If no switching has been done AND the direction is "asc",
            set the direction to "desc" and run the while loop again. */
            if (switchcount == 0 && dir == "asc") {
                dir = "desc";
                switching = true;
            }
        }
    }
}


var cascades = {}
var healthIntervalMilliSeconds = "5000"

function deleteCascade(cascadeName) {
    $.ajax({
        url: "/api/v1/cascade/" + cascadeName,
        type: "DELETE",
        success: function (result) {
            console.log(`Cascade ${cascadeName} deleted`)
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.assign("/ui/login");
            }
        }
    })
}

function getCascades() {
    $.ajax({
        url: "/api/v1/cascade",
        type: "GET",
        success: function (result) {
            for (let idx = 0; idx < result.length; idx++) {
                cascades[result[idx].name] = result[idx]
            }
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
    let prefix = $("#search").val();
    prefix = prefix.toLowerCase();
    
    if (prefix == "") {
        for (let idx = 0; idx < cascades.length; idx++) {
            let name = cascades[idx].name
            $(`#cascades-row-${name}`).removeClass("table-hide")
            $(`#cascades-row-${name}`).addClass("table-show")
        }
        return
    }
    for (let idx = 0; idx < cascades.length; idx++) {
        let name = cascades[idx].name
        if (name.toLowerCase().indexOf(prefix) == -1) {
            $(`#cascades-row-${name}`).removeClass("table-show")
            $(`#cascades-row-${name}`).addClass("table-hide")
            continue
        }
        $(`#cascades-row-${name}`).removeClass("table-hide")
        $(`#cascades-row-${name}`).addClass("table-show")
    }
}

$(document).ready(
    function() {
        getCascades()
    }
)
