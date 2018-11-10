package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/pkg/errors"
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
	log.Println("getURL:", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrapf(err, "get url: %s", u)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "get url body: %s", u)
	}
	return buf, nil
}

var errInvalidToken = errors.New("invalid token")
var errInvalidGrant = errors.New("invalid grant")

func checkToken(buf []byte) error {
	var base Base
	if err := json.Unmarshal(buf, &base); err != nil {
		return errors.Wrap(err, "check token")
	}
	if base.Error == "invalid_token" {
		return errInvalidToken
	}
	if base.Error == "invalid_grant" {
		return errInvalidGrant
	}
	if base.Error != "" {
		return errors.New(base.Error)
	}
	return nil
}

type api struct {
	deviceCode string
	auth       authorizationResp
}

func (a *api) fetch(u, cachePath string, d interface{}) error {
	buf, err := getURL(u)
	if err != nil {
		log.Println(err)
		return errors.Wrapf(err, "fetch: %s", u)
	}
	err = checkToken(buf)
	if err != nil {
		return errors.Wrap(err, "fetch")
	}
	if err := json.Unmarshal(buf, d); err != nil {
		return errors.Wrap(err, "fetch unmarshal")
	}
	return nil
}

var mobjects = make(map[int64]Child)
var lock sync.Mutex

func (a *api) getBookmarkFolder(folder int64) (*Bookmarks, error) {
	u := fmt.Sprintf("%svideo/bookmarks/folders/%d/items.json?per_page=20&page=1&access_token=%s", apiRoot, folder, a.auth.AccessToken)
	var resp Bookmarks
	if err := a.fetch(u, "b1.json", &resp); err != nil {
		return nil, errors.Wrapf(err, "bookmarks: %d", folder)
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

func (a *api) getMyFavorites() (*Bookmarks, error) {
	u := fmt.Sprintf("%svideo/bookmarks/folders.json?per_page=20&access_token=%s", apiRoot, a.auth.AccessToken)
	var resp Folders
	if err := a.fetch(u, "b.json", &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.ErrorMessage)
	}
	for _, o := range resp.Data.Folders {
		if o.Title == "serge" {
			return a.getBookmarkFolder(o.ID)
		}
	}
	return nil, errors.New("no my favorites")
}

func (a *api) getChannels() (*Channels, error) {
	const page = 1
	u := fmt.Sprintf("%svideo/channels.json?per_page=20&page=%d&access_token=%s", apiRoot, page, a.auth.AccessToken)
	var resp Channels
	if err := a.fetch(u, fmt.Sprintf("channels-%d.json", page), &resp); err != nil {
		log.Println(u, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%+v", resp)
	}
	return &resp, nil
}

func (a *api) getChannel(id int64) (*Media, error) {
	const page = 1
	u := fmt.Sprintf("%svideo/media/channel/%d/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, id, page, a.auth.AccessToken)
	var resp Media
	if err := a.fetch(u, fmt.Sprintf("channel-%d-%d.json", id, page), &resp); err != nil {
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

func (a *api) getChildren(id int64, page int) (*Children, error) {
	u := fmt.Sprintf("%svideo/media/%d/children.json?per_page=%d&page=%d&access_token=%s", apiRoot, id, N, page, a.auth.AccessToken)
	var resp Children
	if err := a.fetch(u, fmt.Sprintf("m%d-page%d.json", id, page), &resp); err != nil {
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

func (a *api) getArchive(page int) (*Media, error) {
	u := fmt.Sprintf("%svideo/media/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, page, a.auth.AccessToken)
	var resp Media
	if err := a.fetch(u, fmt.Sprintf("archive-%d.json", page), &resp); err != nil {
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

func (a *api) getChild(id int64) Child {
	lock.Lock()
	defer lock.Unlock()
	return mobjects[id]
}

func (a *api) getStreamURL(id int64) (string, error) {
	u := fmt.Sprintf("%svideo/media/%d/watch.json?format=%s&protocol=hls&bitrate=%d&access_token=%s", apiRoot, id, "mp4", bitrate, a.auth.AccessToken)
	var resp StreamURL
	if err := a.fetch(u, fmt.Sprintf("url%d.json", id), &resp); err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.ErrorMessage)
	}
	return resp.Data.URL, nil
}

func (a *api) search(query string, page int) (*Media, error) {
	u := fmt.Sprintf("%svideo/media/search.json?per_page=20&page=%d&access_token=%s&q=%s", apiRoot, page, a.auth.AccessToken, url.QueryEscape(query))
	var resp Media
	if err := a.fetch(u, fmt.Sprintf("search-%s-%d.json", query, page), &resp); err != nil {
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

func (a *api) history(page int) (*Media, error) {
	u := fmt.Sprintf("%svideo/media/history.json?per_page=20&page=%d&access_token=%s", apiRoot, page, a.auth.AccessToken)
	var resp Media
	if err := a.fetch(u, fmt.Sprintf("history-%d.json", page), &resp); err != nil {
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
