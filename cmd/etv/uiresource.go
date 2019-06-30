package main

import (
	"html/template"
)

const channelsText  = `{{define "channels"}}
{{template "header"}}
<table>
	{{range .List}}
	<tr><td><a class="button" href="/channel/{{.ID}}">{{.Name}}</a></td></tr>
	{{end}}
</table>
{{template "footer"}}
{{end}}

`

const footerText  = `{{define "footer"}}
</body>
</html>
{{end}}


`

const headerText  = `{{define "header"}}
<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, maximum-scale=1, minimum-scale=1" />
<style>
table{ width:100%; }
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
    padding: 120px 0px 0px 0px;
    text-align: center;
    text-decoration: none;
    font-size: 72px;
    font-family: helvetica;
    vertical-align: middle;
    width: 100%;
    height: 280px;
    display: block;
}
.watch0{
    background-color: #4CAF50; /* Green */
}
.watch1{
    background-color: #4C2250; /* Green */
}
.watch2{
    background-color: #123456; /* Green */
}
a:active {
    background-color: #8888FF;
}
.blue {
	background-color: #6C70BF;
}
.pre { width: 330px; white-space: pre-wrap; }
.qfield { width: 330px; border: solid 2px #4CAF50; font-size: 32px; }
</style>
</head>
<body>
{{end}}

`

const playText  = `{{define "play"}}
{{template "header"}}
<table>
        {{if .Error}}
        <tr><td style="color: red">ERROR: {{.Error}}</td><tr>
        {{end}}
        <tr>
        <td><a href="/" class="button blue">Main</a></td>
        <td><a href="/play?cmd=pause" class="button blue">Pause</a></td>
        <tr>
        <tr>
        <td><a href="/play?cmd=volume&arg=-10" class="button blue">Vol -</a></td>
        <td><a href="/play?cmd=volume&arg=10" class="button blue">Vol +</a></td>
        </tr>
        <tr>
        <td><a href="/play?cmd=seek&arg=-30" class="button blue">-30</a></td>
        <td><a href="/play?cmd=seek&arg=30" class="button blue">+30</a></td>
        </tr>
        <tr>
        <td><a href="/play?cmd=seek&arg=-600" class="button blue">-600</a></td>
        <td><a href="/play?cmd=seek&arg=600" class="button blue">+600</a></td>
        </tr>
</table>
{{template "footer"}}
{{end}}

`

const searchText  = `{{define "search"}}
{{template "header"}}
<form method="GET" action="/search/">
<table>
	<tr><td><input class="qfield" type="text" name="q"></td></tr>
	<tr><td><input class="button" type="submit" value="Search"></td></tr>
</table>
</form>
{{template "footer"}}
{{end}}

`

const uiText  = `{{define "main"}}
{{template "header"}}
<table>
	<tr><td><a class="button" href="/bookmarks">Bookmarks</a></td></tr>
	<tr><td><a class="button" href="/history">History</a></td></tr>
	<tr><td><a class="button" href="/channels">Channels</a></td></tr>
	<tr><td><a class="button" href="/archive">Archive</a></td></tr>
	<tr><td><a class="button" href="/search">Search</a></td></tr>
	<tr><td><a class="button" href="/local">Local</a></td></tr>
	<tr><td><a class="button blue" href="/play/">Player</a></td></tr>
	<tr><td><a class="button blue" href="/log">Log</a></td></tr>
	<tr><td><a class="button blue" href="/cookies">Cookies</a></td></tr>
</table>
version: {{.Version}}
{{template "footer"}}
{{end}}

{{define "activation"}}
{{template "header"}}
<h2>Activation</h2>
To activate the box enter go to <a href="https://www.etvnet.com/device/">etvnet activation page</a><br>
and enter the code:<br>
<h3>{{.UserCode}}</h3>
<table>
<tr><td><a class="button" href="/authorize?device_code={{.DeviceCode}}">Code entered</a></td></tr>
<tr><td><a class="button blue" href="/">Main</a></td></tr>
</table>
{{template "footer"}}
{{end}}

{{define "bookmarks"}}
{{template "header"}}
<h1>Bookmarks</h1>
<table>
	{{range .List}}
	<tr><td><a class="button" href="/item/{{.ID}}">{{.ShortName}}<br>{{.OnAir}}</a></td></tr>
	{{end}}
	<tr><td><a class="button blue" href="/">Main</a></td></tr>
</table>
{{template "footer"}}
{{end}}

{{define "items"}}
{{template "header"}}
<table>
	{{range .List}}
	<tr><td><a class="button watch{{.WatchStatus}}" href="/item/{{.ID}}">{{.ShortName}}<br>{{.ChildrenCount}}<br>{{.OnAir}}</a></td></tr>
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

{{define "local"}}
{{template "header"}}
<table>
	{{range .List}}
	<tr><td><a class="button watch0" href="/play/?lid={{.ID}}">{{.Name}}</a></td></tr>
	{{end}}
	<tr><td><a class="button blue" href="/">main</a></td></tr>
</table>
{{template "footer"}}
{{end}}

{{define "cookies"}}
{{template "header"}}

<table>
	<tr><td>AccessToken:</td><td>{{.Auth.AccessToken}}</td></tr>
	<tr><td>Expires:</td><td>{{.Auth.Expires}}</td></tr>
	<tr><td>RefreshToken:</td><td>{{.Auth.RefreshToken}}</td></tr>
</table>
<table>
	<tr><td><a class="button blue" href="/cookies?refresh=1">Refresh</a></td></tr>
	<tr><td><a class="button blue" href="/activate">Activate</a></td></tr>
	<tr><td><a class="button blue" href="/">Main</a></td></tr>
</table>
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
