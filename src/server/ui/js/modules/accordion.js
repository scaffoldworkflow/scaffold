function toggleAccordion(id) {
  var x = document.getElementById(id);
  if (x.className.indexOf("w3-show") == -1) {
    x.className += " w3-show";
    $(`#${id}-icon`).removeClass("fa-caret-down");
    $(`#${id}-icon`).addClass("fa-caret-up");
  } else {
    x.className = x.className.replace(" w3-show", "");
    $(`#${id}-icon`).removeClass("fa-caret-up");
    $(`#${id}-icon`).addClass("fa-caret-down");
  }
}
