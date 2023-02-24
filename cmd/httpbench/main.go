package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rnemeth90/httpbench"
	"github.com/spf13/pflag"
)

type config struct {
	url         string
	count       int
	threads     int
	useHTTP     bool
	headers     string
	timeout     int64
	keepalives  bool
	compression bool
	duration    int
}

var (
	url         string
	count       int
	threads     int
	insecure    bool
	headers     string
	timeout     int64
	keepalives  bool
	compression bool
	duration    int
)

func init() {
	pflag.StringVarP(&url, "url", "u", "", "url to test")
	pflag.IntVarP(&count, "requests", "r", 4, "count of requests per second")
	pflag.IntVarP(&threads, "threads", "l", 1, "threads")
	pflag.IntVarP(&duration, "duration", "d", 10, "duration")
	pflag.BoolVarP(&insecure, "insecure", "i", false, "insecure")
	pflag.StringVarP(&headers, "headers", "h", "", "request headers <string:string>")
	pflag.Int64VarP(&timeout, "timeout", "t", 10, "timeout")
	pflag.BoolVarP(&keepalives, "keepalives", "k", true, "keepalives")
	pflag.BoolVarP(&compression, "compression", "c", true, "compression")
	pflag.Usage = usage
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
		url:         url,
		count:       count,
		threads:     threads,
		useHTTP:     insecure,
		headers:     headers,
		timeout:     timeout,
		keepalives:  keepalives,
		compression: compression,
		duration:    duration,
	}

	fmt.Printf("making %d connections to %s...\n", count, url)

	if err := run(c, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func run(c config, w io.Writer) error {

	results := make([]httpbench.HTTPResponse, 0)
	var client http.Client
	var mu sync.Mutex
	var wg sync.WaitGroup
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)

	client = httpbench.CreateHTTPClient(c.timeout, c.keepalives, c.compression)

	color.Green("Making %d calls per second for %d seconds...", c.count, c.duration)

	for durationCounter := 0; durationCounter <= duration; durationCounter++ {
		for i := 0; i < count; i++ {
			wg.Add(1)
			go httpbench.MakeRequestAsync(c.url, c.useHTTP, c.headers, &mu, &wg, &client, &results)
		}

		var finished = durationCounter * c.count
		color.Cyan("Finished sending %d requests...", finished)
		time.Sleep(1 * time.Second)
	}

	s.Color("yellow")
	s.Prefix = "Processing..."
	s.Start() // Start the spinner
	wg.Wait()
	s.Stop()

	for _, r := range results {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("%s, %d, %v, %v", url, r.Status, r.Latency, r.Err))
	}

	return nil
}
