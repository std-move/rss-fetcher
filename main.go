package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var now = time.Now()

var client = http.Client{
	Timeout: 15 * time.Second,
}

const USER_AGENT = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0"

func main() {
	parser := gofeed.NewParser()
	parser.Client = &client
	parser.UserAgent = USER_AGENT

	origFeed, err := parser.ParseURL("https://feeds.a.dj.com/rss/RSSWorldNews.xml")
	if err != nil {
		log.Fatal("failed to fetch feed: ", err, "\n")
	}

	myFeed := &feeds.Feed{
		Title:       origFeed.Title + " - bigTop stories",
		Link:        &feeds.Link{Href: "https://std-move.github.io/rss-fetcher/wsj-world-bigtop.xml"},
		Description: origFeed.Description,
		Created:     now,
	}

	for _, itm := range origFeed.Items {
		if err := processArticle(itm, &myFeed.Items); err != nil {
			log.Println(err)
		}
	}

	if err = serializeFeed(myFeed, "public/wsj-world-bigtop.xml"); err != nil {
		log.Fatal("failed to serialize my feed: ", err, "\n")
	}
	log.Print("Successfully updated feed, article count [", len(myFeed.Items), "] out of [", len(origFeed.Items), "]", "\n")
}

func processArticle(item *gofeed.Item, myItems *[]*feeds.Item) error {
	ampLink := func() string {
		if !strings.Contains(item.Link, "wsj.com/amp/") {
			return strings.Replace(item.Link, "wsj.com/", "wsj.com/amp/", -1)
		} else {
			return item.Link
		}
	}()

	req, err := http.NewRequest("GET", ampLink, nil)
	if err != nil {
		return fmt.Errorf("%v [%v]: %w", "error creating req", ampLink, err)
	}
	req.Header.Set("User-Agent", USER_AGENT)
	rsp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%v [%v]: %w", "error fetching", ampLink, err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("%v [%v]: %w", "error reading", ampLink, err)
	}
	if !bytes.Contains(body, []byte("class=\"bigTop-hero")) {
		log.Print("not a bigTop article [", ampLink, "]", "\n")
		return nil
	}
	log.Print("bigTop article [", ampLink, "]", "\n")

	created := func() time.Time {
		if item.PublishedParsed != nil {
			return *item.PublishedParsed
		} else {
			return now
		}
	}()
	updated := func() time.Time {
		if item.UpdatedParsed != nil {
			return *item.UpdatedParsed
		} else {
			return created
		}
	}()

	*myItems = append(*myItems, &feeds.Item{
		Title:   item.Title,
		Link:    &feeds.Link{Href: item.Link},
		Content: item.Description,
		Created: created,
		Updated: updated,
	})

	return nil
}

func serializeFeed(myFeed *feeds.Feed, pathToFile string) error {
	ser, err := myFeed.ToAtom()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(pathToFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("%v [%v]: %w", "failed to mkdir", filepath.Dir(pathToFile), err)
	}
	f, err := os.Create(pathToFile)
	if err != nil {
		return fmt.Errorf("%v [%v]: %w", "failed to create file", pathToFile, err)
	}
	defer f.Close()
	_, err = f.WriteString(ser)
	if err != nil {
		return fmt.Errorf("%v [%v]: %w", "failed to write my feed", pathToFile, err)
	}

	return nil
}
