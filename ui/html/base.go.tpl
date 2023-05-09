{{define "base"}}
<!doctype html>
<html lang='en'>

<head>
    <meta charset='utf-8'>
    <title> Snippetbox</title>
    <!-- Link to the CSS stylesheet and favicon -->
    <link rel='stylesheet' href='/static/css/main.css'>
    <link rel='shortcut icon' href='/static/img/favicon.ico' type='image/x-icon'>
    <!-- Also link to some fonts hosted by Google -->
    <link rel='stylesheet' href='https://fonts.googleapis.com/css?family=Ubuntu+Mono:400,700'>
</head>

<body>
    <header>
        <h1><a href='/'>Snippetbox</a></h1>
    </header>
    {{template "nav" .}}
    <main>
        {{with.Flash}}
        <div class='flash'>{{.}}</div>
        {{end}}
        {{template "main" .}}
    </main>
    <footer>
        <!-- Update the footer to include the current year -->
        Powered by <a href='https://golang.org/'>Go</a> in {{.CurrentYear}}
    </footer>
    <!-- And include the JavaScript file -->
    <script src="/static/js/main.js" type="text/javascript"></script>
</body>

</html>
{{end}}