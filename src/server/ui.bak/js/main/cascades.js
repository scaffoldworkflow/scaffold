
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import deploy_status.js
// import data.js

var cascades = {}
var healthIntervalMilliSeconds = "5000"

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
                window.location.href = "/ui/login";
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
