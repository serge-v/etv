package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
		log.Fatal("Refresh token.")
		return
	}
	if base.Error != "" {
		log.Fatal(base.Error)
	}
}

func fetch(u, cachePath string, d interface{}) {
	// log.Print("fetch:", cachePath)
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

	if err := ioutil.WriteFile(cachePath, buf, 0660); err != nil {
		log.Fatal(err)
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
	fetch(u, "a.json", &resp)
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
	fetch(u, "auth.json", &resp)
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
	cfg.AccessToken = resp.AccessToken
	cfg.RefreshToken = resp.RefreshToken
	cfg.ExpiresIn = resp.ExpiresIn
	f, err := os.Create("etvrc.tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&newconf); err != nil {
		log.Fatal(err)
	}

	log.Println("new config saved to etvrc.tmp")

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

func getMovie(id int64, name, indent string, path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/%d/children.json?per_page=%d&page=%d&access_token=%s", apiRoot, id, N, page, cfg.AccessToken)
	var resp Children
	fetch(u, fmt.Sprintf("m%d-page%d.json", id, page), &resp)

	fmt.Println(">", indent, name, id)
	for i, c := range resp.Data.Children {
		if num == 0 {
			fmt.Printf("%2d %s %4d %d %-20s\n", i+1, c.OnAir, c.ChildrenCount, c.WatchStatus, c.ShortName)
			continue
		}

		if num != i+1 {
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

func getPage(path []string) (num, page int) {
	if len(path) == 0 {
		return 0, 1
	}
	cc := strings.Split(path[0], "p")
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
	movie, page := getPage(path)

	u := fmt.Sprintf("%svideo/bookmarks/folders/%d/items.json?per_page=20&page=1&access_token=%s", apiRoot, folder, cfg.AccessToken)
	var resp Bookmarks
	fetch(u, "b1.json", &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	p := resp.Data.Pagination
	fmt.Printf("page: %d/%d\n", p.Page, p.Pages)
	for i, b := range resp.Data.Bookmarks {
		if movie > 0 {
			if movie == i+1 {
				getMovie(b.ID, b.ShortName, "", path[1:], page)
			}
		} else if movie == 0 {
			fmt.Println(i+1, b.ID, b.OnAir, b.ShortName, b.ChildrenCount)
		}
	}
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

func getArchive(path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, page, cfg.AccessToken)
	var resp Media
	fetch(u, fmt.Sprintf("archive-%d.json", page), &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	fmt.Printf("%d/%d\n", pg.Page, pg.Pages)
	for i, c := range resp.Data.Media {
		if num == 0 {
			fmt.Printf("%2d %s %4d %d %-30s %-16s %s\n", i+1, c.OnAir, c.ChildrenCount, c.WatchStatus,
				fmt.Sprintf("%s:%d", c.Channel.Name, c.Channel.ID), c.Tag, c.ShortName)
			continue
		}
		if num != i+1 {
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
		getMovie(c.ID, c.ShortName, "", path[1:], childPage)
		break
	}
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
	fmt.Printf("%d/%d\n", pg.Page, pg.Pages)
	for i, c := range resp.Data.Media {
		if num == 0 {
			fmt.Printf("%2d %s %4d %d %-30s %-16s %s\n", i+1, c.OnAir, c.ChildrenCount, c.WatchStatus,
				fmt.Sprintf("%s:%d", c.Channel.Name, c.Channel.ID), c.Tag, c.ShortName)
			continue
		}
		if num != i+1 {
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
		getMovie(c.ID, c.ShortName, "", path[1:], childPage)
		break
	}
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

type config struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

var debug = flag.Bool("d", false, "debug")
var getCode = flag.Bool("code", false, "get activation code")
var refresh = flag.Bool("r", false, "refresh token")
var auth = flag.String("auth", "", "authorize")
var path = flag.String("p", "", "get movie by path like b/m3/p1")
var cfg config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
		} else if pp[0][0] == 'c' {
			_, page := getPage(pp)
			getChannels(pp[1:], page)
			return
		}
	}

	flag.Usage()
}
