package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
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

func getURL(u string) []byte {
	log.Println("downloading url:", u)
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if *debug {
		fmt.Fprintln(os.Stderr, u)
		fmt.Fprintln(os.Stderr, resp.Status)
		fmt.Println(string(buf))
	}
	return buf
}

func checkToken(buf []byte) {
	var base Base
	if err := json.Unmarshal(buf, &base); err != nil {
		log.Fatal(err, string(buf))
	}
	if base.Error == "invalid_token" {
		panic("Refresh token.")
		return
	}
	if base.Error != "" {
		log.Fatal(base.Error)
	}
}

func fetch(u, cachePath string, d interface{}) {
	log.Print("fetch:", cachePath)
	if cachePath != "" {
		buf, err := ioutil.ReadFile(cachePath)
		if err == nil {
			checkToken(buf)
			if err := json.Unmarshal(buf, d); err != nil {
				log.Fatalln(cachePath, err)
			}
			return
		}
	}

	buf := getURL(u)
	checkToken(buf)

	if cachePath != "" {
		if err := ioutil.WriteFile(cachePath, buf, 0660); err != nil {
			log.Fatal(err)
		}
		log.Print("cached:", cachePath)
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
	os.Remove("activation.json")
	fetch(u, "activation.json", &resp)
	fmt.Printf("Activation code: %s\n", resp.UserCode)
	fmt.Println("Open http://etvnet.com/device/ and enter activation code.")
	fmt.Println("Then run: etv -auth")
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

func authorize() {
	var aresp activationResp
	os.Remove("auth.json")
	fetch("", "activation.json", &aresp)

	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=http%3A%2F%2Foauth.net%2Fgrant_type%2Fdevice%2F1.0" +
		"&code=" + aresp.DeviceCode
	var resp authorizationResp
	fetch(u, "", &resp)
	fmt.Printf("+%v", resp)
	var newconf = cfg
	newconf.AccessToken = resp.AccessToken
	newconf.RefreshToken = resp.RefreshToken
	newconf.ExpiresIn = resp.ExpiresIn
	f, err := os.Create("etvrc.tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&newconf); err != nil {
		log.Fatal(err)
	}
	if err = os.Rename("etvrc", "etvrc.old"); err != nil {
		log.Fatal(err)
	}
	if err = os.Rename("etvrc.tmp", "etvrc"); err != nil {
		log.Fatal(err)
	}
	log.Println("etvrc updated")
}

func refreshToken() {
	log.Println("refresh token")
	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=refresh_token" +
		"&refresh_token=" + cfg.RefreshToken
	var resp authorizationResp
	fetch(u, "", &resp)
	fmt.Printf("+%v", resp)
	var newconf = cfg
	newconf.AccessToken = resp.AccessToken
	newconf.RefreshToken = resp.RefreshToken
	newconf.ExpiresIn = resp.ExpiresIn
	f, err := os.Create("etvrc.tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&newconf); err != nil {
		log.Fatal(err)
	}
	if err = os.Rename("etvrc", "etvrc.old"); err != nil {
		log.Fatal(err)
	}
	if err = os.Rename("etvrc.tmp", "etvrc"); err != nil {
		log.Fatal(err)
	}
	log.Println("etvrc updated")
}

const N = 20

func getStreamURL(id int64) string {
	u := fmt.Sprintf("%svideo/media/%d/watch.json?format=%s&protocol=hls&bitrate=%d&access_token=%s", apiRoot, id, "mp4", 400, cfg.AccessToken)
	var resp StreamURL
	fetch(u, fmt.Sprintf("url%d.json", id), &resp)
	fmt.Println("stream:", resp.Data.URL)
	return resp.Data.URL
}

func startPlayer(u string) {
	cmd := exec.Command("mpv", u)
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
}

func walkChildren(selected, childPage int, children []Child, path []string, indent string) {
	for i, c := range children {
		if selected == 0 {
			printChild(i+1, c)
			continue
		}

		if selected != i+1 {
			continue
		}

		if c.ChildrenCount == 0 {
			fmt.Printf("=================\n%s\n%s, watch: %d\n%s\n", c.Name, c.OnAir, c.WatchStatus, c.Description)
			if len(path) > 1 && path[1] == "s" {
				u := getStreamURL(c.ID)
				startPlayer(u)
			}
			return
		}
		getMovie(c.ID, c.ShortName, indent+"  ", path[1:], childPage)
		break
	}
}

func getMovie(id int64, name, indent string, path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/%d/children.json?per_page=%d&page=%d&access_token=%s", apiRoot, id, N, page, cfg.AccessToken)
	var resp Children
	fetch(u, fmt.Sprintf("m%d-page%d.json", id, page), &resp)

	fmt.Println(">", indent, name, id)
	walkChildren(num, childPage, resp.Data.Children, path, indent)
}

func getPage(path []string) (num, page int) {
	if len(path) == 0 {
		return 0, 1
	}
	cc := strings.Split(path[0], ",")
	if len(cc) > 0 {
		num, _ = strconv.Atoi(cc[0])
	}
	if len(cc) > 1 {
		page, _ = strconv.Atoi(cc[1])
	}
	if page == 0 {
		page = 1
	}
	return
}

func getBookmarks(folder int64, path []string) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/bookmarks/folders/%d/items.json?per_page=20&page=1&access_token=%s", apiRoot, folder, cfg.AccessToken)
	var resp Bookmarks
	fetch(u, "b1.json", &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	p := resp.Data.Pagination
	if num == 0 {
		fmt.Printf("page %d of %d\n", p.Page, p.Pages)
	}
	walkChildren(num, childPage, resp.Data.Bookmarks, path, "")
}

func getFavorites(path []string) {
	u := fmt.Sprintf("%svideo/bookmarks/folders.json?per_page=20&access_token=%s", apiRoot, cfg.AccessToken)
	var resp Folders
	fetch(u, "b.json", &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	for _, o := range resp.Data.Folders {
		if o.Title == "serge" {
			getBookmarks(o.ID, path)
		}
	}
}

func printChild(n int, c Child) {
	if c.Tag == "худ. фильм" {
		c.Tag = "х.ф."
	}
	fmt.Printf("%2d %s %4d %4d %d %-24s %s. %s, %s, %v\n", n, c.OnAir, c.Rating, c.ChildrenCount, c.WatchStatus,
		c.Channel.Name, c.ShortName, c.Tag, c.Country, c.Year)
}

func getArchive(path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, page, cfg.AccessToken)
	var resp Media
	fetch(u, fmt.Sprintf("archive-%d.json", page), &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	fmt.Printf("archive, page %d of %d\n", pg.Page, pg.Pages)
	walkChildren(num, childPage, resp.Data.Media, path, "")
}

func getChannel(id int64, name string, path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/channel/%d/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, id, page, cfg.AccessToken)
	var resp Media
	fetch(u, fmt.Sprintf("channel-%d-%d.json", id, page), &resp)

	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	fmt.Printf("channel %s, page %d of %d\n", name, pg.Page, pg.Pages)
	walkChildren(num, childPage, resp.Data.Media, path, "")
}

func getChannels(path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/channels.json?per_page=20&page=%d&access_token=%s", apiRoot, page, cfg.AccessToken)
	var resp Channels
	fetch(u, fmt.Sprintf("channels-%d.json", page), &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	for i, c := range resp.Data {
		if num == 0 {
			fmt.Printf("%2d %s\n", i+1, c.Name)
			continue
		}
		if num != i+1 {
			continue
		}

		getChannel(c.ID, c.Name, path[1:], childPage)
		break
	}
}

func getHistory(path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/history.json?per_page=20&page=%d&access_token=%s", apiRoot, page, cfg.AccessToken)
	var resp Media
	fetch(u, fmt.Sprintf("history-%d.json", page), &resp)

	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	if num == 0 {
		fmt.Printf("history, page %d of %d\n", pg.Page, pg.Pages)
	}
	walkChildren(num, childPage, resp.Data.Media, path, "")
}

func search(query string, page int, path []string) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/search.json?per_page=20&page=%d&access_token=%s&q=%s", apiRoot, page, cfg.AccessToken, url.QueryEscape(query))
	var resp Media
	fetch(u, fmt.Sprintf("search-%s.json", query), &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	fmt.Printf("search, page %d of %d\n", pg.Page, pg.Pages)
	walkChildren(num, childPage, resp.Data.Media, path, "")
}

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

func main() {
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
		pp := strings.Split(*path, "/")
		if pp[0] == "b" {
			getFavorites(pp[1:])
			return
		} else if pp[0][0] == 'a' {
			_, page := getPage(pp)
			getArchive(pp[1:], page)
			return
		} else if pp[0][0] == 'q' {
			if *query == "" {
				panic("-q parameter required")
			}
			_, page := getPage(pp)
			search(*query, page, pp[1:])
			return
		} else if pp[0][0] == 'c' {
			_, page := getPage(pp)
			getChannels(pp[1:], page)
			return
		} else if pp[0][0] == 'h' {
			_, page := getPage(pp)
			getHistory(pp[1:], page)
			return
		}
		return
	}

	flag.Usage()
}
