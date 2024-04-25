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


function doReset() {
    password = $("#password").val();
    confirm_password = $("#confirm_password").val();
    email = $("#email").val();

    $.ajax({
        type: "POST",
        url: '/auth/reset/do',
        contentType: "application/json",
        dataType: "json",
        data: JSON.stringify({
            "password": password, 
            "confirm_password": confirm_password,
            "email": email
        }),
        success: function(response) {
            alert("Password successfully reset")
            window.location.href = '/ui/login'
        },
        error: function(response) {
            obj = JSON.parse(response.responseText.slice(0, -2))
            alert("Password failed to be reset\n" + obj.error)
        }
    });
}
