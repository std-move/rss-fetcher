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
	origFeed, err := gofeed.NewParser().ParseURL("https://feeds.a.dj.com/rss/RSSWorldNews.xml")
	if err != nil {
		log.Fatal("failed to fetch feed: ", err)
	}

	now := time.Now()
	myFeed := &feeds.Feed{
		Title:       origFeed.Title + " - bigTop stories",
		Link:        &feeds.Link{Href: "https://std-move.github.io/rss-fetcher/wsj-world-bigtop.xml"},
		Description: origFeed.Description,
		Created:     now,
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	for _, itm := range origFeed.Items {
		ampLink := func() string {
			if !strings.Contains(itm.Link, "wsj.com/amp/") {
				return strings.Replace(itm.Link, "wsj.com/", "wsj.com/amp/", -1)
			} else {
				return itm.Link
			}
		}()

		req, err := http.NewRequest("GET", ampLink, nil)
		if err != nil {
			log.Println("error creating req: ", ampLink, err)
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0")
		rsp, err := client.Do(req)
		if err != nil {
			log.Println("error fetching: ", ampLink, err)
			continue
		}
		body, err := func /* ReadAndCloseBody */ () ([]byte, error) {
			defer rsp.Body.Close()
			return io.ReadAll(rsp.Body)
		}()
		if err != nil {
			log.Println("error reading: ", ampLink, err)
			continue
		}
		if !bytes.Contains(body, []byte("class=\"bigTop-hero")) {
			log.Println("not a bigTop article: ", ampLink)
			continue
		}
		log.Println("bigTop article: ", itm.Link)

		created := func() time.Time {
			if itm.PublishedParsed != nil {
				return *itm.PublishedParsed
			} else {
				return now
			}
		}()
		updated := func() time.Time {
			if itm.UpdatedParsed != nil {
				return *itm.UpdatedParsed
			} else {
				return created
			}
		}()

		myFeed.Items = append(myFeed.Items, &feeds.Item{
			Title:   itm.Title,
			Link:    &feeds.Link{Href: itm.Link},
			Content: itm.Description,
			Created: created,
			Updated: updated,
		})
	}

	ser, err := myFeed.ToAtom()
	if err != nil {
		log.Fatal("failed to serialize my feed: ", err)
	}

	err = os.MkdirAll("public", os.ModePerm)
	if err != nil {
		log.Fatal("failed to mkdir: ", err)
	}
	f, err := os.Create("public/wsj-world-bigtop.xml")
	if err != nil {
		log.Fatal("failed to create file: ", err)
	}
	defer f.Close()
	_, err = f.WriteString(ser)
	if err != nil {
		log.Fatal("failed to write my feed: ", err)
	}

	log.Println("Successfully updated feed, article count [", len(myFeed.Items), "] out of [", len(origFeed.Items), "]")
}
