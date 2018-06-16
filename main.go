// etv is a player for Raspberry PI with web UI.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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

var (
	debug   = flag.Bool("d", false, "debug")
	getCode = flag.Bool("code", false, "get activation code from etvnet.com")
	refresh = flag.Bool("r", false, "refresh token")
	auth    = flag.Bool("auth", false, "authorize after entering activation code")
	path    = flag.String("path", "", "get movie by `path` [abch][p]/num/... : [archive,bookmarks,channels,history][pPAGE]/NUM")
	page    = flag.Int("p", 1, "page number")
	num     = flag.Int("n", 0, "movie number")
	play    = flag.Bool("play", false, "start player")
	query   = flag.String("q", "", "specify query for -path q request")
)

var cfg config
var cacheDir = "/tmp/etvnet-cache." + os.Getenv("USER") + "/"

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if *debug {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	if err := os.Mkdir(cacheDir, 0700); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	if err := processArgs(); err != nil {
		log.Fatalf("%+v", err)
	}
}
