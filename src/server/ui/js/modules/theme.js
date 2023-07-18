var theme;

$(document).ready(function() {
    theme = localStorage.getItem('scaffold-theme');
    if (theme) {
        if (theme == 'light') {
            $('.dark').addClass('light').removeClass('dark');
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
        if (typeof editor === 'undefined' || editor === null) {
            console.log("Editor not found!")
        } else {
            console.log("Editor found!")
            monaco.editor.setTheme('scaffoldDarkTheme');
        }
    } else {
        theme = 'light'
        $('.dark').addClass('light').removeClass('dark');
        if (typeof editor === 'undefined' || editor === null) {
            console.log("Editor not found!")
        } else {
            console.log("Editor found!")
            monaco.editor.setTheme('scaffoldLightTheme');
        }
    }
    localStorage.setItem('scaffold-theme', theme);
}
