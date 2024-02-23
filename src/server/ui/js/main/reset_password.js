// import theme.js

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
