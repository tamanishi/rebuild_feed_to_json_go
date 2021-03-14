package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	net_html "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	Enclosure   string `xml:"enclosure"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Generator   string `xml:"generator"`
	Link        string `xml:"link"`
	Language    string `xml:"language"`
	Items       []Item `xml:"item"`
}

type Rss struct {
	Channel Channel `xml:"channel"`
}

type Shownote struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type Episode struct {
	Title           string      `json:"title"`
	MediaUrl        string      `json:"mediaUrl"`
	PublicationDate string      `json:"publicationDate"`
	Shownotes       []*Shownote `json:"shownotes"`
}

type Episodes struct {
	Episodes []Episode `json:"episodes"`
}

func getNode(node *net_html.Node, shownotes *[]*Shownote) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == net_html.ElementNode {
			if c.DataAtom == atom.A {
				*shownotes = append(*shownotes, getAnchor(c))
			}
			getNode(c, shownotes)
		}
	}
}

func getAnchor(node *net_html.Node) *Shownote {
	var title = ""
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == net_html.TextNode {
			title = c.Data
			break
		}
	}

	var url = ""
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			url = attr.Val
			break
		}
	}

	return &Shownote{Title: title, Url: url}
}

func main() {
	resp, err := http.Get("http://feeds.rebuild.fm/rebuildfm")
	if err != nil {
		fmt.Println("http.Get failed.")
		return
	}
	defer resp.Body.Close()

	rss := Rss{}

	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(&rss)
	if err != nil {
		fmt.Println("decoder.Decode failed.")
		return
	}

	episodes := Episodes{}

	for _, item := range rss.Channel.Items {
		node, err := net_html.Parse(strings.NewReader(html.UnescapeString(item.Description)))
		if err != nil {
			fmt.Println("net_html.Parse failed.")
			return
		}

		var shownotes []*Shownote
		getNode(node, &shownotes)

		loc, _ := time.LoadLocation("America/Los_Angeles")
		pubDate, err := time.ParseInLocation(time.RFC1123Z, item.PubDate, loc)
		if err != nil {
			fmt.Println("time.ParseInLocation failed.")
			return
		}

		episode := Episode{
			Title:           item.Title,
			MediaUrl:        item.Link,
			PublicationDate: pubDate.UTC().Format(time.RFC3339),
			Shownotes:       shownotes,
		}

		episodes.Episodes = append(episodes.Episodes, episode)
	}

	jsonBytes, err := json.Marshal(episodes)
	if err != nil {
		fmt.Println("json.Marshal failed.")
		return
	}

	fmt.Println(string(jsonBytes))
}
