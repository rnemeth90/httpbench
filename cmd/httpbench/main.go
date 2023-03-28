package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rnemeth90/httpbench"
	"github.com/spf13/pflag"
)

type config struct {
	url          string
	count        int
	useHTTP      bool
	headers      string
	bodyFileName string
	timeout      int64
	keepalives   bool
	compression  bool
	duration     int
}

var (
	url          string
	count        int
	insecure     bool
	headers      string
	bodyFileName string
	timeout      int64
	keepalives   bool
	compression  bool
	duration     int
)

func init() {
	pflag.StringVarP(&url, "url", "u", "", "url to test")
	pflag.IntVarP(&count, "requests", "r", 4, "count of requests per second")
	pflag.IntVarP(&duration, "duration", "d", 10, "duration")
	pflag.BoolVarP(&insecure, "insecure", "i", false, "insecure")
	pflag.StringVarP(&headers, "headers", "h", "", "request headers <string:string>")
	pflag.StringVarP(&bodyFileName, "bodyFile", "b", "", "body file in json")
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
		url:          url,
		count:        count,
		useHTTP:      insecure,
		headers:      headers,
		bodyFileName: bodyFileName,
		timeout:      timeout,
		keepalives:   keepalives,
		compression:  compression,
		duration:     duration,
	}

	if err := run(c, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func run(c config, w io.Writer) error {

	body := make([]byte, 0)
	err := errors.New("")
	if c.bodyFileName != "" {
		body, err = ioutil.ReadFile(bodyFileName)
		if err != nil {
			return err
		}
	}

	color.Green("Making %d calls per second for %d seconds...", c.count, c.duration)

	// create channels: request chan, response chan,
	respChan := make(chan httpbench.HTTPResponse, c.count)
	reqChan := make(chan *http.Request, c.count)

	// create dispatcher
	httpbench.Dispatcher(reqChan, c.count, c.useHTTP, c.url, "GET", body, c.headers)

	// create worker pool
	httpbench.WorkerPool(reqChan, respChan, c.count, c.timeout, c.keepalives, c.compression)

	// build results
	results := httpbench.BuildResults(c.count, respChan)

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Color("yellow")
	s.Prefix = "Processing..."
	s.Start() // Start the spinner
	s.Stop()

	stats := httpbench.CalculateStatistics(results)
	fmt.Println()

	fmt.Println("----------------------------------")
	fmt.Println("Total Requests:", stats.TotalCalls)
	fmt.Println("Total Time Taken:", stats.TotalTime)
	fmt.Println("----------------------------------")
	fmt.Println("fastest:", stats.FastestRequest)
	fmt.Println("slowest:", stats.SlowestRequest)
	fmt.Println("average:", stats.AvgTimePerRequest)
	fmt.Println("----------------------------------")
	fmt.Println("20x count:", stats.TwoHundredResponses)
	fmt.Println("30x count:", stats.ThreeHundredResponses)
	fmt.Println("40x count:", stats.FourHundredResponses)
	fmt.Println("50x count:", stats.FiveHundredResponses)

	return nil
}
