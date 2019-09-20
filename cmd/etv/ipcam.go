package main

import (
	"log"
	"net/http"
	"strconv"
)

var ipcamPlayer *videoPlayer

func ipcamHandler(w http.ResponseWriter, r *http.Request) error {
	return handleIPCamRequests(ipcamPlayer, w, r)
}

const camURL = "rtsp://192.168.1.203/11"

func setPlayerPos(player *videoPlayer, arg string) {
	switch arg {
	case "nw":
		player.dbus.setVideoPos(0, 0, 960, 540)
	case "sw":
		player.dbus.setVideoPos(0, 540, 960, 1080)
	case "ne":
		player.dbus.setVideoPos(0, 0, 960, 540)
	case "se":
		player.dbus.setVideoPos(960, 540, 1920, 1080)
	case "full":
		player.dbus.setVideoPos(0, 0, 1920, 1080)
	}
	log.Println("setPlayerPos", arg)
}

func handleIPCamRequests(player *videoPlayer, w http.ResponseWriter, r *http.Request) error {
	var d struct {
		Status playerStatus
		Error  error
	}

	command := r.URL.Query().Get("cmd")
	arg := r.URL.Query().Get("arg")

	switch command {
	case "volume":
		n, _ := strconv.Atoi(arg)
		player.volume(n)
	case "move":
		setPlayerPos(player, arg)
	}

	if player.pid > 0 && command == "stop" {
		player.stop()
	}
	if player.pid == 0 && command == "start" {
		d.Error = player.start(camURL)
	}

	st, err := player.dbus.status()
	log.Println(st, err)

	if err := uiT.ExecuteTemplate(w, "ipcam", d); err != nil {
		return err
	}
	return nil
}
