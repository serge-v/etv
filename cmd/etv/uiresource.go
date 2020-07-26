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
td { padding: 2px; }
.button {
    background-color: #4CAF50; /* Green */
    border: none;
    color: white;
    padding: 20px 10px 10px 10px;
    text-align: center;
    text-decoration: none;
    font-size: 36px;
    font-family: helvetica;
    vertical-align: middle;
    width: 80%;
    height: 70px;
    display: block;
}
.error {
	background-color: pink;
	color: red;
	font-size: 24px;
}
.title {
	font-size: 16px;
	padding: 4px 4px 4px 4px;
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
.qfield { width: 330px; border: solid 2px #4CAF50; font-size: 32px; }
</style>
</head>
<body>
{{end}}
`

const ipcamText  = `{{define "ipcam"}}
{{template "header"}}
<table>
        <tr>
	        <td><a href="/" class="button blue">Main</a></td>
	        <td><a href="/ipcam?cmd=move&arg=full" class="button blue">Full</a></td>
        </tr>
        <tr>
	        <td><a href="/ipcam?cmd=move&arg=nw" class="button blue">NW</a></td>
	        <td><a href="/ipcam?cmd=move&arg=ne" class="button blue">NE</a></td>
        </tr>
        <tr>
	        <td><a href="/ipcam?cmd=move&arg=sw" class="button blue">SW</a></td>
	        <td><a href="/ipcam?cmd=move&arg=se" class="button blue">SE</a></td>
        </tr>
        <tr>
	        <td><a href="/ipcam?cmd=start" class="button blue">Start</a></td>
	        <td><a href="/ipcam?cmd=stop" class="button blue">Stop</a></td>
        </tr>
        <tr>
	        <td><a href="/ipcam?cmd=volume&arg=-10" class="button blue">Vol -</a></td>
	        <td><a href="/ipcam?cmd=volume&arg=10" class="button blue">Vol +</a></td>
        </tr>
</table>
{{template "footer"}}
{{end}}
`

const playText  = `{{define "play"}}
{{template "header"}}
<table>
        <tr>
	        <td><a href="/" class="button blue">Main</a></td>
	        <td><a href="/play?cmd=pause" class="button blue">Pause</a></td>
        <tr>
        <tr>
	        <td><a href="/play?cmd=window" class="button blue">Window</a></td>
	        <td><a href="/play?cmd=stop" class="button blue">Stop</a></td>
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
	<tr><td><a class="button blue" href="/">Main</a></td></tr>
</table>
</form>
{{template "footer"}}
{{end}}

`

const uiText  = `{{define "error"}}
{{template "header"}}
<table>
	<tr><td><div class="button error">{{.Error}}</div></td></tr>
	<tr><td><a class="button blue" href="/">Main</a></td></tr>
</table>
{{template "footer"}}
{{end}}

{{define "main"}}
{{template "header"}}
<table>
	<tr><td><a class="button" href="/bookmarks">Bookmarks</a></td></tr>
	<tr><td><a class="button" href="/history">History</a></td></tr>
	<tr><td><a class="button" href="/channels">Channels</a></td></tr>
	<tr><td><a class="button" href="/archive">Archive</a></td></tr>
	<tr><td><a class="button" href="/search">Search</a></td></tr>
	<tr><td><a class="button" href="/local">Local</a></td></tr>
	<tr><td><a class="button blue" href="/play/">Player</a></td></tr>
	<tr><td><a class="button blue" href="/ipcam">IP Camera</a></td></tr>
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
	<tr><td><a class="button title" href="/item/{{.ID}}">{{.ShortName}}<br>{{.OnAir}}</a></td></tr>
	{{end}}
	<tr><td><a class="button blue" href="/">Main</a></td></tr>
</table>
{{template "footer"}}
{{end}}

{{define "items"}}
{{template "header"}}
<table>
	{{range .List}}
	<tr><td><a class="button title watch{{.WatchStatus}}" href="/item/{{.ID}}">{{.ShortName}}<br>{{.ChildrenCount}}<br>{{.OnAir}}</a></td></tr>
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
	<tr><td><a class="button watch0 title" href="/play/?lid={{.ID}}">{{.Name}}</a></td></tr>
	{{end}}
	<tr><td><a class="button blue" href="/">Main</a></td></tr>
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
	uiT  = template.Must(uiT.New("ipcam").Parse(ipcamText))
	uiT  = template.Must(uiT.New("play").Parse(playText))
	uiT  = template.Must(uiT.New("search").Parse(searchText))
	uiT  = template.Must(uiT.New("ui").Parse(uiText))
}
