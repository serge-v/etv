package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func mainPage(a *api, w http.ResponseWriter, r *http.Request) error {
	d := struct {
		Version string
	}{
		Version: version,
	}

	if err := uiT.ExecuteTemplate(w, "main", d); err != nil {
		return err
	}

	return nil
}

func activatePage(a *api, w http.ResponseWriter, r *http.Request) error {
	activation, err := a.getActivation()
	if err != nil {
		return err
	}

	if err := uiT.ExecuteTemplate(w, "activation", activation); err != nil {
		return err
	}

	return nil
}

func authorizeHandler(a *api, w http.ResponseWriter, r *http.Request) error {
	if err := a.authorize(); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

func itemsPage(a *api, w http.ResponseWriter, r *http.Request) error {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		return errors.New("invalid /items/ link")
	}
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return err
	}

	lock.Lock()
	mo, ok := mobjects[id]
	lock.Unlock()
	if ok {
		//		fmt.Fprintf(w, "%+v", mo)
		if err := uiT.ExecuteTemplate(w, "movie", mo); err != nil {
			return err
		}
		return nil
	}

	m, err := a.getChildren(id, 1)
	if err != nil {
		return err
	}

	d := struct {
		List       []Child
		Pagination Pagination
	}{
		List:       m.Data.Children,
		Pagination: m.Data.Pagination,
	}

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}

	return nil
}

func bookmarksPage(a *api, w http.ResponseWriter, r *http.Request) error {
	bm, err := a.getMyFavorites()
	if err != nil {
		return err
	}

	d := struct {
		List       []Child
		Pagination Pagination
	}{
		List:       bm.Data.Bookmarks,
		Pagination: bm.Data.Pagination,
	}

	if err := uiT.ExecuteTemplate(w, "bookmarks", d); err != nil {
		return err
	}

	return nil
}

func historyPage(a *api, w http.ResponseWriter, r *http.Request) error {
	bm, err := a.history(1)
	if err != nil {
		return err
	}

	d := struct {
		List       []Child
		Pagination Pagination
	}{
		List:       bm.Data.Media,
		Pagination: bm.Data.Pagination,
	}

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}

	return nil
}

func channelPage(a *api, w http.ResponseWriter, r *http.Request) error {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		return errors.New("invalid /items/ link")
	}
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return err
	}
	var d struct {
		List []Child
	}
	ch, err := a.getChannel(id)
	if err != nil {
		return err
	}

	d.List = ch.Data.Media

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}
	return nil
}

func getLocalFile(id int64) (string, error) {
	list, err := filepath.Glob(os.Getenv("HOME") + "/vid/*.*")
	if err != nil {
		return "", err
	}
	return list[id], nil
}

func cookiesPage(a *api, w http.ResponseWriter, r *http.Request) error {
	d := struct {
		Auth authorizationResp
	}{
		Auth: a.auth,
	}

	refresh := r.URL.Query().Get("refresh")
	if refresh == "1" {
		if err := a.refreshToken(); err != nil {
			return errors.Wrap(err, "cookiesPage")
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return nil
	}

	if err := uiT.ExecuteTemplate(w, "cookies", d); err != nil {
		return err
	}

	return nil
}

func localPage(a *api, w http.ResponseWriter, r *http.Request) error {
	list, err := filepath.Glob(os.Getenv("HOME") + "/vid/*.*")
	if err != nil {
		return err
	}

	type item struct {
		ID   int
		Name string
	}

	type data struct {
		List []item
	}

	d := data{}

	for i, fname := range list {
		base := filepath.Base(fname)
		d.List = append(d.List, item{ID: i, Name: base})
	}

	if err := uiT.ExecuteTemplate(w, "local", d); err != nil {
		return err
	}

	return nil
}

func searchPage(a *api, w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query().Get("q")
	if q == "" {
		if err := uiT.ExecuteTemplate(w, "search", nil); err != nil {
			return err
		}
		return nil
	}

	var d struct {
		List []Child
	}

	ch, err := a.search(q, 1)
	if err != nil {
		return err
	}

	d.List = ch.Data.Media

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}
	return nil
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
}

func archivePage(a *api, w http.ResponseWriter, r *http.Request) error {
	var d struct {
		List []Child
	}
	ch, err := a.getArchive(1)
	if err != nil {
		return err
	}

	d.List = ch.Data.Media

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}
	return nil
}

func channelsPage(a *api, w http.ResponseWriter, r *http.Request) error {
	var err error
	var d struct {
		List []Channel
	}

	ch, err := a.getChannels()
	if err != nil {
		return err
	}

	d.List = ch.Data

	if err := uiT.ExecuteTemplate(w, "channels", d); err != nil {
		return err
	}
	return nil
}

var uiT = template.New("")

func loadAuth() authorizationResp {
	var auth authorizationResp

	authURL := *confURL + "/auth.json"
	resp, err := http.Get(authURL)
	if err != nil {
		log.Println(err)
		return auth
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("status code", resp.StatusCode)
		return auth
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return auth
	}

	if err := json.Unmarshal(buf, &auth); err != nil {
		log.Println(err)
		return auth
	}
	auth.Expires = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	log.Printf("auth loaded: %+v", &auth)
	return auth
}

func etvHandler(h func(a *api, w http.ResponseWriter, r *http.Request) error) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		if version == "" {
			var err error
			uiT, err = template.New("").ParseGlob("templates/*.html")
			if err != nil {
				log.Println(err)
				return
			}
		}

		if a.auth.AccessToken == "" {
			a.auth = loadAuth()
		}

		a.deviceCode = r.URL.Query().Get("device_code")

		log.Printf("request url: %s, a.auth: %+v", r.URL.String(), a.auth)
		log.Printf("request handler: %T", h)
		if err := h(&a, w, r); err != nil {
			log.Printf("request error: %+v", err)
			if err == errInvalidGrant {
				err = activatePage(&a, w, r)
				if err == nil {
					return
				}
			}
			d := struct {
				Error string
			}{
				Error: err.Error(),
			}

			if err := uiT.ExecuteTemplate(w, "error", d); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}
			return
		}
		log.Printf("request done")

	}
	return http.HandlerFunc(f)
}

func errorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		if version == "" {
			var err error
			uiT, err = template.New("").ParseGlob("templates/*.html")
			if err != nil {
				log.Println(err)
				return
			}
		}

		if err := h(w, r); err != nil {
			log.Printf("request error: %+v", err)
			d := struct {
				Error string
			}{
				Error: err.Error(),
			}

			if err := uiT.ExecuteTemplate(w, "error", d); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}
			return
		}
		log.Printf("request done")

	}
	return http.HandlerFunc(f)
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/tmp/log")
}

var a api

func runServer() error {
	player = newPlayer("", 20, false)
	ipcamPlayer = newPlayer("omxplayer1", 10, true)
	a.auth = loadAuth()

	http.Handle("/", etvHandler(mainPage))
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.Handle("/bookmarks/", etvHandler(bookmarksPage))
	http.Handle("/history/", etvHandler(historyPage))
	http.Handle("/channels/", etvHandler(channelsPage))
	http.Handle("/channel/", etvHandler(channelPage))
	http.Handle("/search/", etvHandler(searchPage))
	http.Handle("/archive/", etvHandler(archivePage))
	http.Handle("/item/", etvHandler(itemsPage))
	http.Handle("/activate/", etvHandler(activatePage))
	http.Handle("/authorize/", etvHandler(authorizeHandler))
	http.Handle("/play/", etvHandler(playerHandler))
	http.Handle("/ipcam/", errorHandler(ipcamHandler))
	http.HandleFunc("/log", logHandler)
	http.Handle("/local/", etvHandler(localPage))
	http.Handle("/cookies", etvHandler(cookiesPage))
	if strings.HasPrefix(*server, ":") {
		log.Println("serving on http://localhost" + *server)
	} else {
		log.Println("serving on http://" + *server)
	}
	return http.ListenAndServe(*server, nil)
}
