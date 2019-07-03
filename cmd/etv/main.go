// etv is a player for Raspberry PI with web UI.
package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	server      = flag.String("http", ":8099", "start server on `[ip]:port`")
	showVersion = flag.Bool("version", false, "print version")
	confURL     = flag.String("conf", "", "config server URL")
)

var version string

func processArgs() (err error) {
	switch {
	case *showVersion:
		fmt.Println(version)
	default:
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
