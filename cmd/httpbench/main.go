package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/rnemeth90/httpbench"
	"github.com/spf13/pflag"
)

type config struct {
	url   string
	count int
}

var (
	url   string
	count int
)

func init() {
	pflag.StringVar(&url, "url", "", "url to test")
	pflag.IntVar(&count, "count", 4, "count of requests")
}

func usage() {
	fmt.Println(os.Args[0])
	fmt.Println()

	fmt.Println("Usage:")
	fmt.Printf("  httpbench --url https://mywebsite.com\n")
	fmt.Printf("  httpbench --url https://mywebsite.com --count 100\n\n")

	fmt.Println("Options:")
	pflag.PrintDefaults()
}

func main() {
	pflag.Parse()
	args := pflag.Args()

	if url == "" && len(args) == 0 {
		usage()
		os.Exit(1)
	}

	if len(args) > 1 {
		usage()
		os.Exit(1)
	}

	if len(args) == 1 {
		url = args[0]
	}

	c := config{
		url:   url,
		count: count,
	}

	if err := run(c, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func run(c config, w io.Writer) error {

	results := []httpbench.HTTPResponse{}
	rChan := make(chan httpbench.HTTPResponse)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	wg.Add(count)
	for i := 0; i < count; i++ {
		go httpbench.MakeRequest(c.url, rChan, &mu, &wg)
		results = append(results, <-rChan)
	}

	for _, r := range results {
		fmt.Println(url, r.Status, r.Latency)
	}

	wg.Wait()
	return nil
}
