package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Fetcher interface {
	// Fetch returns a slice of URLs found on the page
	Fetch(url string) (urls []string, err error)
}

type fakeFetcher map[string]*fakeResult

type linkFetcher map[string][]string

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) ([]string, error) {
	// url already found before
	if res, ok := f[url]; ok {
		fmt.Printf("found: %s\n", url)
		return res.urls, nil
	}
	// we didn't find anything
	fmt.Printf("missing: %s\n", url)
	return nil, fmt.Errorf("not found: %s")
}

func (f linkFetcher) Fetch(url string) ([]string, error) {
	if url[len(url) - 1] != '/' {
		url = url + "/"
	}

	cnt := 0
	idx := 0
	for i, e := range url {
		if e == '/' {
			cnt++
		}
		if cnt == 3 {
			idx = i
			break
		}
	}

	domain := url[0:idx]
	fmt.Println("url: ", url)
	fmt.Println("domain: ", domain)

	resp, err := http.Get(url)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, nil
	}
	doc, err := html.Parse(strings.NewReader(string(body)))

	res := make([]string, 1)

	var ff func(*html.Node)
	ff = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					var u string
					if !strings.Contains(a.Val, "http") {
						u = domain + a.Val
					} else {
						u = a.Val
					}
					res = append(res, u)
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ff(c)
		}
	}
	ff(doc)

	return res, nil
}

// this is a pre populated dictionary
// for actual ones, it should provide some result upon every request
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

////////////////////////////////////////////////////////////////////
// serial crawler
func Serial(url string, fetcher Fetcher, fetched map[string]bool) {
	if fetched[url] {
		return
	}

	// this is depth first search, every request return a bunch results
	fetched[url] = true
	urls, err := fetcher.Fetch(url)
	if err != nil {
		return
	}
	for _, u := range urls {
		Serial(u, fetcher, fetched)
	}
	return
}

func main() {
	// fmt.Printf("=== Serial ===\n")
	// Serial("http://golang.org/", fetcher, make(map[string]bool))
	urls, err := (linkFetcher{}).Fetch("http://golang.org")
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, url := range urls {
		fmt.Println(url)
	}
}
