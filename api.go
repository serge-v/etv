package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
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

const bitrate = 400

func getURL(u string) ([]byte, error) {
	log.Println("downloading url:", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if *debug {
		fmt.Fprintln(os.Stderr, u)
		fmt.Fprintln(os.Stderr, resp.Status)
		fmt.Println(string(buf))
	}
	return buf, nil
}

var ErrInvalidToken = errors.New("invalid token")
var ErrInvalidGrant = errors.New("invalid grant")

func checkToken(buf []byte) error {
	var base Base
	if err := json.Unmarshal(buf, &base); err != nil {
		log.Println(err, string(buf))
		return err
	}
	if base.Error == "invalid_token" {
		return ErrInvalidToken
	}
	if base.Error == "invalid_grant" {
		return ErrInvalidGrant
	}
	if base.Error != "" {
		return errors.New(base.Error)
	}
	return nil
}

func fetch(u, cachePath string, d interface{}) error {
	fname := cacheDir + cachePath
	fi, err := os.Stat(fname)
	maxTime := time.Now().Add(-30 * time.Second)
	if err == nil && fi.ModTime().Before(maxTime) {
		os.Remove(fname)
		cachePath = ""
	}
	if cachePath != "" {
		buf, err := ioutil.ReadFile(fname)
		if err == nil {
			err = checkToken(buf)
			if err == ErrInvalidToken {
				if err = refreshToken(); err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
			if err = json.Unmarshal(buf, d); err != nil {
				log.Println(cachePath, err)
				return err
			}
			return nil
		}
	}

	buf, err := getURL(u)
	if err != nil {
		log.Println(err)
		return err
	}
	err = checkToken(buf)
	if err == ErrInvalidToken {
		if err = refreshToken(); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	if cachePath != "" {
		if err := ioutil.WriteFile(fname, buf, 0600); err != nil {
			log.Println(err)
			return err
		}
		log.Print("cached:", cachePath)
	}
	if err := json.Unmarshal(buf, d); err != nil {
		return err
	}
	return nil
}

var mobjects = make(map[int64]Child)
var lock sync.Mutex

func getBookmarkFolder(folder int64) (*Bookmarks, error) {
	u := fmt.Sprintf("%svideo/bookmarks/folders/%d/items.json?per_page=20&page=1&access_token=%s", apiRoot, folder, cfg.AccessToken)
	var resp Bookmarks
	if err := fetch(u, "b1.json", &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}

	lock.Lock()
	for _, c := range resp.Data.Bookmarks {
		if c.Type == "MediaObject" {
			mobjects[c.ID] = c
		}
	}
	lock.Unlock()

	return &resp, nil
}

func getMyFavorites() (*Bookmarks, error) {
	u := fmt.Sprintf("%svideo/bookmarks/folders.json?per_page=20&access_token=%s", apiRoot, cfg.AccessToken)
	var resp Folders
	if err := fetch(u, "b.json", &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}
	for _, o := range resp.Data.Folders {
		if o.Title == "serge" {
			return getBookmarkFolder(o.ID)
		}
	}
	return nil, errors.New("no my favorites")
}

func getChannels() (*Channels, error) {
	u := fmt.Sprintf("%svideo/channels.json?per_page=20&page=%d&access_token=%s", apiRoot, 1, cfg.AccessToken)
	var resp Channels
	if err := fetch(u, fmt.Sprintf("channels-%d.json", page), &resp); err != nil {
		log.Println(u, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%+v", resp)
	}
	return &resp, nil
}

func getChannel(id int64) (*Media, error) {
	u := fmt.Sprintf("%svideo/media/channel/%d/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, id, 1, cfg.AccessToken)
	var resp Media
	if err := fetch(u, fmt.Sprintf("channel-%d-%d.json", id, page), &resp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}

	lock.Lock()
	for _, c := range resp.Data.Media {
		if c.Type == "MediaObject" {
			mobjects[c.ID] = c
		}
	}
	lock.Unlock()

	return &resp, nil
}

func getChildren(id int64, page int) (*Children, error) {
	u := fmt.Sprintf("%svideo/media/%d/children.json?per_page=%d&page=%d&access_token=%s", apiRoot, id, N, page, cfg.AccessToken)
	var resp Children
	if err := fetch(u, fmt.Sprintf("m%d-page%d.json", id, page), &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}
	lock.Lock()
	for _, c := range resp.Data.Children {
		if c.Type == "MediaObject" {
			mobjects[c.ID] = c
		}
	}
	lock.Unlock()
	return &resp, nil
}

func getArchive(page int) (*Media, error) {
	u := fmt.Sprintf("%svideo/media/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, page, cfg.AccessToken)
	var resp Media
	if err := fetch(u, fmt.Sprintf("archive-%d.json", page), &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	lock.Lock()
	for _, c := range resp.Data.Media {
		if c.Type == "MediaObject" {
			mobjects[c.ID] = c
		}
	}
	lock.Unlock()
	return &resp, nil
}

func getChild(id int64) Child {
	lock.Lock()
	defer lock.Unlock()
	return mobjects[id]
}

func getStreamURL(id int64) (string, error) {
	u := fmt.Sprintf("%svideo/media/%d/watch.json?format=%s&protocol=hls&bitrate=%d&access_token=%s", apiRoot, id, "mp4", bitrate, cfg.AccessToken)
	var resp StreamURL
	if err := fetch(u, fmt.Sprintf("url%d.json", id), &resp); err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.ErrorMessage)
	}
	if *debug {
		log.Println("stream:", resp.Data.URL)
	}
	return resp.Data.URL, nil
}

func search(query string, page int) (*Media, error) {
	u := fmt.Sprintf("%svideo/media/search.json?per_page=20&page=%d&access_token=%s&q=%s", apiRoot, page, cfg.AccessToken, url.QueryEscape(query))
	var resp Media
	if err := fetch(u, fmt.Sprintf("search-%s-%d.json", query, page), &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}
	lock.Lock()
	for _, c := range resp.Data.Media {
		if c.Type == "MediaObject" {
			mobjects[c.ID] = c
		}
	}
	lock.Unlock()
	return &resp, nil
}

func history(page int) (*Media, error) {
	u := fmt.Sprintf("%svideo/media/history.json?per_page=20&page=%d&access_token=%s", apiRoot, page, cfg.AccessToken)
	var resp Media
	if err := fetch(u, fmt.Sprintf("history-%d.json", page), &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}
	lock.Lock()
	for _, c := range resp.Data.Media {
		if c.Type == "MediaObject" {
			mobjects[c.ID] = c
		}
	}
	lock.Unlock()
	return &resp, nil
}
