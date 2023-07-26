
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import status.js
// import data.js

function togglePane(id) {
    pane = $(`#${id}`)
    icon = $(`#${id}-icon`)
    
    if (pane.hasClass("w3-show")) {
        icon.removeClass("fa-caret-up")
        icon.addClass("fa-caret-down")

        pane.removeClass('w3-show')
    } else {
        icon.removeClass("fa-caret-down")
        icon.addClass("fa-caret-up")

        pane.addClass('w3-show')
    }
}
