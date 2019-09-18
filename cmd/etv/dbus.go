package main

import (
	"fmt"
	"log"
	"os/exec"
)

type dbusControl struct {
	name string
}

func (b *dbusControl) send(args []string) (string, error) {
	cmd := exec.Command("dbus-send", args...)
	cmd.Env = []string{
		fmt.Sprintf(`DBUS_SESSION_BUS_ADDRESS="/tmp/omxplayerdbus.%s"`, b.name),
		fmt.Sprintf(`DBUS_SESSION_BUS_PID="/tmp/omxplayerdbus.%s.pid"`, b.name),
	}

	buf, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: dbus-send: %s", err, string(buf))
	}

	return string(buf), nil
}

func (b *dbusControl) status(arg string) (string, error) {
	args := []string{
		"--print-reply=literal",
		"--session",
		"--reply-timeout=500",
		"--dest=org.mpris.MediaPlayer2.omxplayer",
		"/org/mpris/MediaPlayer2",
		"org.freedesktop.DBus.Properties.Get",
		`string:"org.mpris.MediaPlayer2.Player"`,
		`string:"PlaybackStatus"`,
	}
	s, err := b.send(args)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (b *dbusControl) setVideoPos(x1, y1, x2, y2 int) {
	args := []string{
		"--print-reply=literal",
		"--session",
		"--dest=org.mpris.MediaPlayer2.omxplayer",
		"/org/mpris/MediaPlayer2",
		"org.mpris.MediaPlayer2.Player.VideoPos",
		"objpath:/not/used",
		fmt.Sprintf(`string:"%d %d %d %d`, x1, y1, x2, y2),
	}
	s, err := b.send(args)
	if err != nil {
		log.Println("setVideoPos", err, s)
	}
	log.Println("setVideoPos", s)
}
