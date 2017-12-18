package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// Decrease crawlCount when the function returns
	defer func() {
		count <- -1
	}()

	if depth <= 0 {
		return
	}
	
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		count <- 1 // Increase crawlCount for each new call to Crawl
		go Crawl(u, depth-1, fetcher)
	}
}

func main() {
	go Crawl("http://golang.org/", 4, fetcher)
	for crawlCount := 1; crawlCount > 0; {
		crawlCount += <-count
	}
}

type Cache struct {
	visited map[string]bool
	mux     sync.Mutex
}

type fakeResult struct {
	body string
	urls []string
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher struct {
	urls  map[string]*fakeResult
	cache Cache
}

func (f *fakeFetcher) Fetch(url string) (string, []string, error) {
	f.cache.mux.Lock()
	defer f.cache.mux.Unlock()

	if f.cache.visited[url] == true {
		return "", nil, fmt.Errorf("already visited %s", url)
	}

	if res, ok := f.urls[url]; ok {
		f.cache.visited[url] = true
		return res.body, res.urls, nil
	}

	return "", nil, fmt.Errorf("not found: %s", url)
}

// a channel to alter the count of how many Crawl routines are still running
var count = make(chan int)

// fetcher is a populated fakeFetcher.
var fetcher = &fakeFetcher{
	urls: map[string]*fakeResult{
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
				"ç/",
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
	},
	cache: Cache{visited: map[string]bool{}},
}
