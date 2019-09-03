// etv is a player for Raspberry PI with web UI.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var (
	server      = flag.String("http", ":8099", "start server on `[ip]:port`")
	showVersion = flag.Bool("version", false, "print version")
	confURL     = flag.String("conf", "", "config server URL")
)

var version string

func weatherLoop() {
	printWeather()
	tick := time.NewTicker(time.Minute * 10)
	for range tick.C {
		printWeather()
	}
}

func printWeather() {
	cmd := exec.Command("curl", "wttr.in?m")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func processArgs() (err error) {
	switch {
	case *showVersion:
		fmt.Println(version)
	default:
		go weatherLoop()
		return runServer()
	}
	return nil
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := processArgs(); err != nil {
		log.Fatalf("%+v", err)
	}
}
