package main

import (
	"bytes"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://feeds.a.dj.com/rss/RSSWorldNews.xml")
	if err != nil {
		log.Fatal(err)
	}
	// log.Println(feed.Title)

	now := time.Now()
	myFeed := &feeds.Feed{
		Title:       feed.Title + " - bigTop stories",
		Link:        &feeds.Link{Href: "https://std-move.github.io/rss-fetcher/wsj-world-bigtop.xml"},
		Description: feed.Description,
		Created:     now,
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	for _, itm := range feed.Items {
		origLink := itm.Link
		if !strings.Contains(itm.Link, "wsj.com/amp/") {
			itm.Link = strings.Replace(itm.Link, "wsj.com/", "wsj.com/amp/", -1)
		}
		req, err := http.NewRequest("GET", itm.Link, nil)
		if err != nil {
			log.Println("error creating: ", itm.Link, err)
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0")
		rsp, err := client.Do(req)
		if err != nil {
			log.Println("error fetching: ", itm.Link, err)
			continue
		}
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Println("error reading: ", itm.Link, err)
			continue
		}
		if !bytes.Contains(body, []byte("class=\"bigTop-hero")) {
			continue
		}
		log.Println("bigTop link: ", itm.Link)

		var created time.Time
		if itm.PublishedParsed != nil {
			created = *itm.PublishedParsed
		} else {
			created = now
		}

		var updated time.Time
		if itm.UpdatedParsed != nil {
			updated = *itm.UpdatedParsed
		} else {
			updated = created
		}
		myFeed.Items = append(myFeed.Items, &feeds.Item{
			Title:       itm.Title,
			Link:        &feeds.Link{Href: origLink},
			Content:     itm.Description,
			Created:     created,
			Updated:	 updated,
		})
	}

	ser, err := myFeed.ToAtom()
	if err != nil {
		log.Fatal("failed to serialize my feed: ", err)
	}

	_ = os.MkdirAll("public", os.ModePerm)
	f, err := os.Create("public/wsj-world-bigtop.xml")
	if err != nil {
		log.Fatal("failed to serialize my feed: ", err)
	}
	defer f.Close()
	_, err = f.WriteString(ser)
	if err != nil {
		log.Fatal("failed to write my feed: ", err)
	}

	log.Println(ser)
}
