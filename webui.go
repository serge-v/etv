package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const stampZ = "2006-01-02 15:04:05Z"

func mainPage(w http.ResponseWriter, r *http.Request) error {
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

func activatePage(w http.ResponseWriter, r *http.Request) error {
	var err error
	var d struct {
		Code string
	}

	d.Code, err = getActivationCode()
	if err != nil {
		return err
	}

	if err := uiT.ExecuteTemplate(w, "activation", d); err != nil {
		return err
	}

	return nil
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) error {
	if err := authorize(); err != nil {
		return err
	}
	http.Redirect(w, r, "/bookmarks", http.StatusFound)
	return nil
}

func itemsPage(w http.ResponseWriter, r *http.Request) error {
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

	m, err := getChildren(id, 1)
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

func bookmarksPage(w http.ResponseWriter, r *http.Request) error {
	bm, err := getMyFavorites()
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

func historyPage(w http.ResponseWriter, r *http.Request) error {
	bm, err := history(1)
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

func channelPage(w http.ResponseWriter, r *http.Request) error {
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
	ch, err := getChannel(id)
	if err != nil {
		return err
	}

	d.List = ch.Data.Media

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}
	return nil
}

func searchPage(w http.ResponseWriter, r *http.Request) error {
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

	ch, err := search(q, 1)
	if err != nil {
		return err
	}

	d.List = ch.Data.Media

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}
	return nil
}

func faviconHandler(w http.ResponseWriter, r *http.Request) error {
	log.Println("empty favicon served")
	return nil
}

func archivePage(w http.ResponseWriter, r *http.Request) error {
	var d struct {
		List []Child
	}
	ch, err := getArchive(1)
	if err != nil {
		return err
	}

	d.List = ch.Data.Media

	if err := uiT.ExecuteTemplate(w, "items", d); err != nil {
		return err
	}
	return nil
}

func channelsPage(w http.ResponseWriter, r *http.Request) error {
	var err error
	var d struct {
		List []Channel
	}

	ch, err := getChannels()
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

func errorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		if version == "" {
			var err error
			uiT, err = template.New("").ParseGlob("templates/*.tmpl")
			if err != nil {
				log.Println(err)
				return
			}
		}
		f, err := os.Open("etvrc")
		if err == nil {
			if err = json.NewDecoder(f).Decode(&cfg); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Println("config error", err)
		}
		if err := h(w, r); err != nil {
			log.Println("error:", err)
			if err == ErrInvalidGrant {
				http.Redirect(w, r, "/activate", http.StatusFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return http.HandlerFunc(f)
}

func runServer() error {
	http.Handle("/", errorHandler(mainPage))
	http.Handle("/favicon.ico", errorHandler(faviconHandler))
	http.Handle("/bookmarks/", errorHandler(bookmarksPage))
	http.Handle("/history/", errorHandler(historyPage))
	http.Handle("/channels/", errorHandler(channelsPage))
	http.Handle("/channel/", errorHandler(channelPage))
	http.Handle("/search/", errorHandler(searchPage))
	http.Handle("/archive/", errorHandler(archivePage))
	http.Handle("/item/", errorHandler(itemsPage))
	http.Handle("/activate/", errorHandler(activatePage))
	http.Handle("/authorize/", errorHandler(authorizeHandler))
	http.HandleFunc("/play/", playerHandler)
	log.Println("serving on http://localhost" + *server)
	return http.ListenAndServe(*server, nil)
}
