package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func getActivationCode() (string, error) {
	u := codeURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20")

	var err error
	var resp activationResp
	os.Remove("activation.json")
	if err = fetch(u, "activation.json", &resp); err != nil {
		return "", err
	}
	fmt.Printf("Activation code: %s\n", resp.UserCode)
	fmt.Println("Open http://etvnet.com/device/ and enter activation code.")
	fmt.Println("Then run: etv -auth")
	return resp.UserCode, nil
}

func authorize() error {
	var aresp activationResp
	if err := os.Remove("auth.json"); err != nil {
		log.Println(err)
	}
	if err := fetch("", "activation.json", &aresp); err != nil {
		return err
	}

	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=http%3A%2F%2Foauth.net%2Fgrant_type%2Fdevice%2F1.0" +
		"&code=" + aresp.DeviceCode
	var resp authorizationResp
	if err := fetch(u, "", &resp); err != nil {
		return err
	}
	fmt.Printf("+%v", resp)
	var newconf = cfg
	newconf.AccessToken = resp.AccessToken
	newconf.RefreshToken = resp.RefreshToken
	newconf.ExpiresIn = resp.ExpiresIn
	f, err := os.Create("etvrc.tmp")
	if err != nil {
		return err
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&newconf); err != nil {
		return err
	}
	if err = os.Rename("etvrc", "etvrc.old"); err != nil {
		return err
	}
	if err = os.Rename("etvrc.tmp", "etvrc"); err != nil {
		return err
	}
	log.Println("etvrc updated")
	return nil
}

func refreshToken() error {
	log.Println("refresh token")
	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=refresh_token" +
		"&refresh_token=" + cfg.RefreshToken
	var resp authorizationResp
	if err := fetch(u, "", &resp); err != nil {
		return err
	}
	log.Printf("+%v", resp)
	var newconf = cfg
	newconf.AccessToken = resp.AccessToken
	newconf.RefreshToken = resp.RefreshToken
	newconf.ExpiresIn = resp.ExpiresIn
	f, err := os.Create("etvrc.tmp")
	if err != nil {
		return err
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&newconf); err != nil {
		return err
	}
	if err = os.Rename("etvrc", "etvrc.old"); err != nil {
		return err
	}
	if err = os.Rename("etvrc.tmp", "etvrc"); err != nil {
		return err
	}
	log.Println("etvrc updated")
	return nil
}

const N = 20
const bitrate = 400

type config struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

var debug = flag.Bool("d", false, "debug")
var getCode = flag.Bool("code", false, "get activation code from etvnet.com")
var refresh = flag.Bool("r", false, "refresh token")
var auth = flag.Bool("auth", false, "authorize after entering activation code")
var path = flag.String("path", "", "get movie by `path` [abch][p]/num/... : [archive,bookmarks,channels,history][pPAGE]/NUM")
var page = flag.Int("p", 1, "page number")
var num = flag.Int("n", 0, "movie number")
var play = flag.Bool("play", false, "start player")
var query = flag.String("q", "", "specify query for -path q request")
var cfg config

func main2() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if *debug {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	if *getCode {
		getActivationCode()
		return
	}

	f, err := os.Open("etvrc")
	if err != nil {
		log.Fatal(err)
	}
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	if *auth {
		authorize()
		return
	}

	if *refresh {
		refreshToken()
		return
	}

	if *path != "" {
		var c console
		pp := strings.Split(*path, "/")
		if pp[0] == "b" {
			c.getFavorites(pp[1:])
			return
		} else if pp[0][0] == 'a' {
			_, page := getPage(pp)
			c.getArchive(pp[1:], page)
			return
		} else if pp[0][0] == 'q' {
			if *query == "" {
				panic("-q parameter required")
			}
			_, page := getPage(pp)
			println("=== page", page)
			c.search(*query, page, pp[1:])
			return
		} else if pp[0][0] == 'c' {
			_, page := getPage(pp)
			c.getChannels(pp[1:], page)
			return
		} else if pp[0][0] == 'h' {
			_, page := getPage(pp)
			c.getHistory(pp[1:], page)
			return
		} else {
			fmt.Println("unknown path")
		}
		return
	}

	flag.Usage()
}
