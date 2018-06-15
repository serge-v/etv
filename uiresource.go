package main

import (
	"html/template"
)

const channelsText  = `{{template "header"}}
<table>
	{{range .List}}
	<tr><td><a class="button" href="/channel/{{.ID}}">{{.Name}}</a></td></tr>
	{{end}}
</table>
{{template "footer"}}
`

const footerText  = `</body>
</html>

`

const headerText  = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width; maximum-scale=1; minimum-scale=1;" />
<style>
table.zebra tr:nth-child(even) { background-color: #E6E6E6; }
.err tr:nth-child(4n) { background-color: #E6E6E6; }
.err tr:nth-child(4n+1) { background-color: #E6E6E6; }
th { background-color: lightsteelblue; }
td { padding: 2px; }
td.completed { background-color:transparent; }
td.completed_with_errors { background-color: #DDDD88; }
td.failed { background-color: #FF8888; }
td.known_error { background-color:plum; }
td.submitted { background-color:#8888FF; }
td.processing { background-color:#BBBBFF; }
td.aborted { background-color:darkgray; }
.button {
    background-color: #4CAF50; /* Green */
    border: none;
    color: white;
    padding: 15px 32px;
    text-align: center;
    vertical-align: middle;
    text-decoration: none;
    display: inline-block;
    font-size: 32px;
    width: 240px;
    min-height: 100px;
}
.pre { width: 640px; white-space: pre-wrap; }
</style>
</head>
<body>

`

const playText  = `<!DOCTYPE html>
<html>
<head>
<title>Player</title>
<meta name="viewport" content="width=device-width; maximum-scale=1; minimum-scale=1;" />
<style type="text/css">
.button {
	background-color: lightgray;
	border: none;
	color: white;
	padding: 15px 32px;
	text-align: center;
	vertical-align: middle;
	text-decoration: none;
	display: inline-block;
	font-size: 32px;
	width: 100px;
	height: 80px;
}

.left {
	background-color: #6C70BF;
}

.right {
	background-color: #4C50AF;
}

</style>
</head>
<body>
	<table border="0" style="width: 200px">
		{{if .Error}}
		<tr><td style="color: red">ERROR: {{.Error}}</td><tr>
		{{end}}
		<tr>
			<td><a href="/" class="button left">Stop</a></td>
			<td><a href="/play?cmd=pause" class="button right">Pause</a></td>
		<tr>
		<tr>
			<td><a href="/play?cmd=volume&arg=-10" class="button left">Vol -</a></td>
			<td><a href="/play?cmd=volume&arg=10" class="button right">Vol +</a></td>
		</tr>
		<tr>
			<td><a href="/play?cmd=seek&arg=-30" class="button left">-30</a></td>
			<td><a href="/play?cmd=seek&arg=30" class="button right">+30</a></td>
		</tr>
		<tr>
			<td><a href="/play?cmd=seek&arg=-600" class="button left">-600</a></td>
			<td><a href="/play?cmd=seek&arg=600" class="button right">+600</a></td>
		</tr>
	</table>
</body>
</html>
`

const searchText  = `{{template "header"}}
<form method="POST" action="/search/">
<table>
	<tr><td>
	</td></tr>
</table>
</form>
{{template "footer"}}
`

const uiText  = `{{define "main"}}
{{template "header"}}
<table>
	<tr><td><a class="button" href="/bookmarks">bookmarks</a></td></tr>
	<tr><td><a class="button" href="/channels">channels</a></td></tr>
	<tr><td><a class="button" href="/archive">archive</a></td></tr>
	<tr><td><a class="button" href="/search">search</a></td></tr>
</table>
{{template "footer"}}
{{end}}

{{define "activation"}}
{{template "header"}}
<h2>Activation</h2>
To activate the box enter go to <a href="http://etvnet.com/activate">etvnet activation page</a><br>
and enter the code:<br>
<h3>{{.Code}}</h3>
<a class="button" href="/authorize">Code entered</a><br>
<a class="button" href="/">Cancel</a><br>
{{template "footer"}}
{{end}}

{{define "bookmarks"}}
{{template "header"}}
<h1>Bookmarks</h1>
<table>
	{{range .List}}
	<tr><td><a class="button" href="/item/{{.ID}}">{{.ShortName}}<br>{{.OnAir}}</a></td></tr>
	{{end}}
</table>
{{template "footer"}}
{{end}}

{{define "items"}}
{{template "header"}}
<table>
	{{range .List}}
	<tr><td><a class="button" href="/item/{{.ID}}">{{.ShortName}}<br>{{.ChildrenCount}}<br>{{.OnAir}}</a></td></tr>
	{{end}}
</table>
{{template "footer"}}
{{end}}

{{define "movie"}}
{{template "header"}}
	<h1>{{.Name}}</h1>
	<h3>{{.Country}} {{.Year}}</h3>
	<h3>{{.OnAir}}</h3>
	<h3>{{.Tag}}</h3>
	<a class="button" href="/play/?id={{.ID}}">Play</a>
	<br>
	<br>
	<img src="{{.Thumb}}">
	<div class="pre">{{.Description}}</div>
{{template "footer"}}
{{end}}

`
func init() {
	uiT  = template.Must(uiT.New("channels").Parse(channelsText))
	uiT  = template.Must(uiT.New("footer").Parse(footerText))
	uiT  = template.Must(uiT.New("header").Parse(headerText))
	uiT  = template.Must(uiT.New("play").Parse(playText))
	uiT  = template.Must(uiT.New("search").Parse(searchText))
	uiT  = template.Must(uiT.New("ui").Parse(uiText))
}