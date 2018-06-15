package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type console struct{}

func (con *console) startPlayer(u string) {
	cmd := exec.Command("omxplayer", u)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func (con *console) walkChildren(selected, childPage int, children []Child, path []string, indent string) {
	for i, c := range children {
		if selected == 0 {
			con.printChild(i+1, c)
			continue
		}

		if selected != i+1 {
			continue
		}

		if c.ChildrenCount == 0 {
			fmt.Printf("=================\n%s\n%s, watch: %d\n%s\n", c.Name, c.OnAir, c.WatchStatus, c.Description)
			if len(path) > 1 && path[1] == "s" {
				u, err := getStreamURL(c.ID)
				if err != nil {
					return
				}
				con.startPlayer(u)
			}
			return
		}
		con.getMovie(c.ID, c.ShortName, indent+"  ", path[1:], childPage)
		break
	}
}

func (con *console) getMovie(id int64, name, indent string, path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/%d/children.json?per_page=%d&page=%d&access_token=%s", apiRoot, id, N, page, cfg.AccessToken)
	var resp Children
	fetch(u, fmt.Sprintf("m%d-page%d.json", id, page), &resp)

	fmt.Println(">", indent, name, id)
	con.walkChildren(num, childPage, resp.Data.Children, path, indent)
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

func (con *console) getBookmarks(folder int64, path []string) {
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
	con.walkChildren(num, childPage, resp.Data.Bookmarks, path, "")
}

func (con *console) getFavorites(path []string) {
	u := fmt.Sprintf("%svideo/bookmarks/folders.json?per_page=20&access_token=%s", apiRoot, cfg.AccessToken)
	var resp Folders
	fetch(u, "b.json", &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	for _, o := range resp.Data.Folders {
		if o.Title == "serge" {
			con.getBookmarks(o.ID, path)
		}
	}
}

func (con *console) printChild(n int, c Child) {
	if c.Tag == "худ. фильм" {
		c.Tag = "х.ф."
	}
	fmt.Printf("%2d %s %4d %4d %d %-24s %s. %s, %s, %v\n", n, c.OnAir, c.Rating, c.ChildrenCount, c.WatchStatus,
		c.Channel.Name, c.ShortName, c.Tag, c.Country, c.Year)
}

func (con *console) getArchive(path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, page, cfg.AccessToken)
	var resp Media
	fetch(u, fmt.Sprintf("archive-%d.json", page), &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	fmt.Printf("archive, page %d of %d\n", pg.Page, pg.Pages)
	con.walkChildren(num, childPage, resp.Data.Media, path, "")
}

func (con *console) getChannel(id int64, name string, path []string, page int) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/channel/%d/archive.json?per_page=20&page=%d&access_token=%s", apiRoot, id, page, cfg.AccessToken)
	var resp Media
	fetch(u, fmt.Sprintf("channel-%d-%d.json", id, page), &resp)

	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	fmt.Printf("channel %s, page %d of %d\n", name, pg.Page, pg.Pages)
	con.walkChildren(num, childPage, resp.Data.Media, path, "")
}

func (con *console) getChannels(path []string, page int) {
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

		con.getChannel(c.ID, c.Name, path[1:], childPage)
		break
	}
}

func (con *console) getHistory(path []string, page int) {
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
	con.walkChildren(num, childPage, resp.Data.Media, path, "")
}

func (con *console) search(query string, page int, path []string) {
	num, childPage := getPage(path)
	u := fmt.Sprintf("%svideo/media/search.json?per_page=20&page=%d&access_token=%s&q=%s", apiRoot, page, cfg.AccessToken, url.QueryEscape(query))
	var resp Media
	fetch(u, fmt.Sprintf("search-%s-%d.json", query, page), &resp)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.ErrorMessage)
	}
	pg := resp.Data.Pagination
	fmt.Printf("search, page %d of %d\n", pg.Page, pg.Pages)
	con.walkChildren(num, childPage, resp.Data.Media, path, "")
}
