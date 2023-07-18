
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import status.js
// import data.js

var cascades
var healthIntervalMilliSeconds = "5000"

function getCascades() {
    $.ajax({
        url: "/api/v1/cascade",
        type: "GET",
        success: function (result) {
            cascades = result
        },
        error: function(result) {
            console.log(result)
        }
    });
}

function render() {
    var prefix = $("#search").val();
    var tempCascades = JSON.parse(JSON.stringify(cascades));
    prefix = prefix.toLowerCase();
    if (prefix != "") {
        for (var idx = tempCascades.length - 1; idx >= 0; idx--) {
            if (tempCascades[idx].name.toLowerCase().indexOf(prefix) == -1) {
                tempCascades.splice(idx, 1)
            }
        }
    }
    var table = document.getElementById('cascades-table');
    var tableHTMLString = '<tr><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Name</span></th><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Created</span></th><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Updated</span></th><th class="table-title w3-medium scaffold-text-green"><span class="table-title-text">Version</span></th><th><span class="table-title-text"></span></th></tr>' +
        tempCascades.cascades.map(function (cascade) {
            return '<tr>' +
                '<td>' + cascade.name + '</td>' +
                '<td>' + cascade.created + '</td>' +
                '<td>' + cascade.updated + '</td>' +
                '<td>' + cascade.version + '</td>' +
                '<td class="table-link-cell">' +
                '<a href="/ui/cascade/' + cascade.name + '" class="table-link-link w3-right-align light theme-text" style="float:right;margin-right:16px;"><i class="fa-solid fa-link"></i></a>' +
                '</td>' +
                '</tr>'
        }).join('');

    table.innerHTML = tableHTMLString;
}

$(document).ready(
    function() {
        getCascades()
    }
)
