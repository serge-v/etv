// etv is a player for Raspberry PI with web UI.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	server      = flag.String("http", ":8099", "start server on `[ip]:port`")
	showVersion = flag.Bool("version", false, "print version")
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

var cacheDir = "/tmp/etvnet-cache." + os.Getenv("USER") + "/"

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := os.Mkdir(cacheDir, 0700); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	if err := processArgs(); err != nil {
		log.Fatalf("%+v", err)
	}
}
