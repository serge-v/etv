package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const stampZ = "2006-01-02 15:04:05Z"

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
	log.Println("empty favicon served")
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

var funcs = template.FuncMap{
	"nozero": func(n int) string {
		if n == 0 {
			return ""
		}
		if n < 1000 {
			return fmt.Sprintf("%d", n)
		} else if n < 1000000 {
			return fmt.Sprintf("%d,%03d", n/1000, n%1000)
		} else if n < 1000000000 {
			return fmt.Sprintf("%d,%03d,%03d", n/1000000, n%1000000/1000, n%1000)
		} else {
			return fmt.Sprintf("%d,%03d,%03d,%03d", n/1000000000, n%1000000000/1000000, n%1000000/1000, n%1000)
		}
	},
}

var uiT = template.New("")

func errorHandler(h func(a *api, w http.ResponseWriter, r *http.Request) error) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		if version == "" {
			var err error
			uiT, err = template.New("").ParseGlob("templates/*.tmpl")
			if err != nil {
				log.Println(err)
				return
			}
		}

		var auth authorizationResp
		var a api

		cookie, err := r.Cookie("atoken")
		if err == nil {
			auth.AccessToken = cookie.Value
		}
		cookie, err = r.Cookie("rtoken")
		if err == nil {
			auth.RefreshToken = cookie.Value
		}
		cookie, err = r.Cookie("expires")
		if err == nil {
			n, _ := strconv.ParseInt(cookie.Value, 10, 32)
			auth.ExpiresIn = int(n)
		}

		a.deviceCode = r.URL.Query().Get("device_code")

		if auth.AccessToken == "" && a.deviceCode == "" {
			activatePage(&a, w, r)
			return
		}

		a.auth = auth
		log.Printf("url: %s, a.auth: %+v", r.URL.String(), a.auth)

		if err := h(&a, w, r); err != nil {
			log.Println("error:", err)
			if err == errInvalidGrant {
				activatePage(&a, w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if auth.AccessToken != a.auth.AccessToken {
			http.SetCookie(w, &http.Cookie{Name: "atoken", Value: a.auth.AccessToken, Path: "/"})
			http.SetCookie(w, &http.Cookie{Name: "rtoken", Value: a.auth.RefreshToken, Path: "/"})
			http.SetCookie(w, &http.Cookie{Name: "expires", Value: strconv.Itoa(a.auth.ExpiresIn), Path: "/"})
			http.Redirect(w, r, "/bookmarks/", http.StatusMovedPermanently)
		}
	}
	return http.HandlerFunc(f)
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/tmp/log")
}

func runServer() error {
	http.Handle("/", errorHandler(mainPage))
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.Handle("/bookmarks/", errorHandler(bookmarksPage))
	http.Handle("/history/", errorHandler(historyPage))
	http.Handle("/channels/", errorHandler(channelsPage))
	http.Handle("/channel/", errorHandler(channelPage))
	http.Handle("/search/", errorHandler(searchPage))
	http.Handle("/archive/", errorHandler(archivePage))
	http.Handle("/item/", errorHandler(itemsPage))
	http.Handle("/activate/", errorHandler(activatePage))
	http.Handle("/authorize/", errorHandler(authorizeHandler))
	http.Handle("/play/", errorHandler(playerHandler))
	http.HandleFunc("/log", logHandler)
	http.Handle("/local/", errorHandler(localPage))
	log.Println("serving on http://localhost" + *server)
	return http.ListenAndServe(*server, nil)
}
