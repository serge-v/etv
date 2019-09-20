package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type dbusControl struct {
}

func (b *dbusControl) send(args []string) (string, error) {
	cmd := exec.Command("dbus-send", args...)
	u := os.Getenv("USER")

	fname := fmt.Sprintf("/tmp/omxplayerdbus.%s", u)
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}
	cmd.Env = append(cmd.Env, "DBUS_SESSION_BUS_ADDRESS="+strings.TrimSpace(string(buf)))

	fname = fmt.Sprintf("/tmp/omxplayerdbus.%s.pid", u)
	buf, err = ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}
	cmd.Env = append(cmd.Env, "DBUS_SESSION_BUS_PID="+strings.TrimSpace(string(buf)))
	log.Printf("start omxplayer: %+v", cmd)

	buf, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: dbus-send: %s", err, string(buf))
	}

	return string(buf), nil
}

func (b *dbusControl) status() (string, error) {
	args := []string{
		"--print-reply=literal",
		"--session",
		"--reply-timeout=500",
		"--dest=org.mpris.MediaPlayer2.omxplayer1",
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
		"--dest=org.mpris.MediaPlayer2.omxplayer1",
		"/org/mpris/MediaPlayer2",
		"org.mpris.MediaPlayer2.Player.VideoPos",
		"objpath:/not/used",
		fmt.Sprintf(`string:"%d %d %d %d"`, x1, y1, x2, y2),
	}
	s, err := b.send(args)
	if err != nil {
		log.Println("setVideoPos error:", err, s)
		return
	}
	log.Println("setVideoPos", s)
}
