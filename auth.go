package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	clientID     = "a332b9d61df7254dffdc81a260373f25592c94c9"
	clientSecret = "744a52aff20ec13f53bcfd705fc4b79195265497"
	tokenURL     = "https://accounts.etvnet.com/auth/oauth/token"
	codeURL      = "https://accounts.etvnet.com/auth/oauth/device/code"
	apiRoot      = "https://secure.etvnet.com/api/v3.0/"
)

var scope = []string{
	"com.etvnet.media.browse",
	"com.etvnet.media.watch",
	"com.etvnet.media.bookmarks",
	"com.etvnet.media.history",
	"com.etvnet.media.live",
	"com.etvnet.media.fivestar",
	"com.etvnet.media.comments",
	"com.etvnet.persons",
	"com.etvnet.notifications",
}

func fetch(u string, d interface{}) {
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if *debug {
		fmt.Fprintln(os.Stderr, resp.Status)
		fmt.Println(string(buf))
	}
	if err := json.Unmarshal(buf, d); err != nil {
		log.Fatal(err)
	}
}

type activationResp struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
}

func getActivationCode() {
	u := codeURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20")

	var resp activationResp
	fetch(u, &resp)
	fmt.Printf("%+v\n", resp)
}

type authorizationResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
}

func authorize(deviceCode string) {
	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=http%3A%2F%2Foauth.net%2Fgrant_type%2Fdevice%2F1.0" +
		"&code=" + deviceCode
	var resp authorizationResp
	fetch(u, &resp)
}

func getFavorites() {
	u := apiRoot + "video/bookmarks/items.json?per_page=20&page=1&access_token=" + cfg.AccessToken
	var resp struct{}
	fetch(u, &resp)
}

type config struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

var debug = flag.Bool("d", false, "debug")
var getCode = flag.Bool("c", false, "get activation code")
var auth = flag.String("a", "", "authorize")
var cfg config

func main() {
	flag.Parse()
	f, err := os.Open("etvrc")
	if err != nil {
		log.Fatal(err)
	}
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	if *getCode {
		getActivationCode()
		return
	}
	if *auth != "" {
		authorize(*auth)
	}

	getFavorites()
}
