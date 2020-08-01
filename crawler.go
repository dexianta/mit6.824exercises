package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"golang.org/x/net/html"
)

type Fetcher interface {
	// Fetch returns a slice of URLs found on the page
	Fetch(url string) (urls []string, err error)
}


type linkFetcher map[string][]string

func patchSlash(url string) string {
	if url[len(url)-1] != '/' {
		url = url + "/"
	}

	return url
}

func getDomain(url string) string {
	if len(url) < 1 {
		return ""
	}

	url = patchSlash(url)

	cnt, idx := 0, 0
	for i, e := range url {
		if e == '/' {
			cnt++
		}
		if cnt == 3 {
			idx = i
			break
		}
	}

	return url[0:idx]
}

func (f linkFetcher) Fetch(url string) ([]string, error) {

	domain := getDomain(url)

	resp, err := http.Get(url)
	if err != nil {
		return []string{}, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, nil
	}
	doc, err := html.Parse(strings.NewReader(string(body)))

	res := make([]string, 0)

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
					res = append(res, patchSlash(u))
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

////////////////////////////////////////////////////////////////////
// serial crawler
func Serial(url string, fetcher Fetcher, fetched map[string]int, depth int) {
	fmt.Println(len(fetched))
	if len(fetched) > 1000 {
		println("passed target: ", len(fetched))
		os.Exit(0)
	}
	_, exist := fetched[url]
	if (depth < 0) || exist {
		return
	}

	// this is depth first search, every request return a bunch results
	fetched[url] = depth
	urls, err := fetcher.Fetch(url)
	if err != nil {
		return
	}
	for _, u := range urls {
		Serial(u, fetcher, fetched, depth-1)
	}
	return
}

//
// Concurrent crawler with shared state and Mutex
//
type fetchState struct {
	mu      sync.Mutex
	fetched map[string]int
}

func makeState() *fetchState {
	f := &fetchState{}
	f.fetched = make(map[string]int)
	return f
}

func ConcurrentMutex(url string, fetcher Fetcher, f *fetchState, depth int) {
	if depth < 0 {
		return
	}

	f.mu.Lock()
	fmt.Println(len(f.fetched))
	_, ok := f.fetched[url]
	if !ok {
		f.fetched[url] = depth
	}
	f.mu.Unlock()

	if ok {
		return
	}

	urls, err := fetcher.Fetch(url)
	if err != nil {
		return
	}
	var done sync.WaitGroup
	for _, u := range urls {
		done.Add(1)
		go func(u string) {
			// making sure the procedure started to run
			// otherwise it gets returned
			defer done.Done()
			ConcurrentMutex(u, fetcher, f, depth-1)
		}(u)
	}
	done.Wait()
	return
}

//
// Concurrent Crawler with channels
//

func worker(url string, ch chan []string, fetcher Fetcher) {
	urls, err := fetcher.Fetch(url)
	if err != nil {
		ch <- []string{}
	} else {
		ch <- urls
	}
}

func master(ch chan []string, fetcher Fetcher) {
	n := 1
	fetched := make(map[string]bool)
	for urls := range ch {
		for _, u := range urls {
			if fetched[u] == false {
				fetched[u] = true
				n += 1 // for every new link we add one
				go worker(u, ch, fetcher)
			}
		}
		println(len(fetched))
		n -= 1 // for every empty link we subtract one
		if n == 0 {
			break // when n = 0, means we've exhausted the internet, return without hanging
		}
	}
}

func ConcurrentChannel(url string, fetcher Fetcher) {
	ch := make(chan []string)
	go func() {
		ch <- []string{url}
	}()
	master(ch, fetcher)
}

func main() {
	const startUrl = "https://en.wikipedia.org/wiki/Germany"
	// fmt.Println("=== Serial ===")
	// res := make(map[string]int)
	// Serial(startUrl, linkFetcher{}, res, 2)
	// fmt.Println(len(res))

	// fmt.Println("=== ConcurrentMutex ===")
	// ConcurrentMutex(startUrl,
	// 	linkFetcher{},
	// 	makeState(),
	// 	2)

	fmt.Println("=== ConcurrentChannel ===")
	ConcurrentChannel(startUrl, linkFetcher{})

}
