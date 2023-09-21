package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

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
	pflag.StringVarP(&url, "url", "u", "", "url to send requests to")
	pflag.IntVarP(&count, "requests", "r", 4, "count of requests per second")
	pflag.IntVarP(&duration, "duration", "d", 10, "duration (seconds)")
	pflag.BoolVarP(&insecure, "insecure", "i", false, "use HTTP instead of HTTPS")
	pflag.StringVarP(&headers, "headers", "h", "", "key/value request headers <string:string>")
	pflag.StringVarP(&bodyFileName, "bodyFile", "b", "", "body file in json")
	pflag.Int64VarP(&timeout, "timeout", "t", 10, "timeout")
	pflag.BoolVarP(&keepalives, "keepalives", "k", true, "use keepalives")
	pflag.BoolVarP(&compression, "compression", "c", true, "use compression")
	pflag.Usage = usage
}

func usage() {
	fmt.Println(`
| |   | | | |       | |                   | |    
| |__ | |_| |_ _ __ | |__   ___ _ __   ___| |__  
| '_ \| __| __| '_ \| '_ \ / _ \ '_ \ / __| '_ \ 
| | | | |_| |_| |_) | |_) |  __/ | | | (__| | | |
|_| |_|\__|\__| .__/|_.__/ \___|_| |_|\___|_| |_|
              | |                                
              |_|                                
	`)
	fmt.Println(os.Args[0])
	fmt.Println()

	fmt.Println("Usage:")
	fmt.Printf("  httpbench --url https://mywebsite.com\n")
	fmt.Printf("  httpbench --url https://mywebsite.com --requests 100\n\n")

	fmt.Println("Options:")
	pflag.PrintDefaults()
}

func main() {
	pflag.Parse()
	args := pflag.Args()

	runtime.GOMAXPROCS(runtime.NumCPU())

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
			return fmt.Errorf("Failed to read the body file: %v\n", err)
		}
	}

	if !httpbench.IsValidURL(c.url) {
		return errors.New("The provided URL is not valid. Please check and try again.")
	}

	color.Green("Making %d calls per second for %d seconds...", c.count, c.duration)

	numjobs := c.count * c.duration
	respChan := make(chan httpbench.HTTPResponse, numjobs)
	reqChan := make(chan *http.Request, numjobs)

	httpbench.Dispatcher(reqChan, c.count, c.duration, c.useHTTP, c.url, "GET", body, c.headers)

	httpbench.WorkerPool(reqChan, respChan, c.duration, c.count, c.timeout, c.keepalives, c.compression)

	var resultslice []httpbench.HTTPResponse

	for i := 1; i <= numjobs; i++ {
		r := <-respChan
		resultslice = append(resultslice, r)
	}

	stats := httpbench.CalculateStatistics(resultslice)

	fmt.Println("\nResults:")
	fmt.Printf("  Average Response Time: %v\n", stats.AvgTimePerRequest)
	fmt.Printf("  Fastest Response Time: %v\n", stats.FastestRequest)
	fmt.Printf("  Slowest Response Time: %v\n", stats.SlowestRequest)
	fmt.Printf("  Total Calls Made: %d\n", stats.TotalCalls)
	fmt.Println()
	fmt.Printf("  200s Responses: %d\n", stats.TwoHundredResponses)
	fmt.Printf("  300s Responses: %d\n", stats.ThreeHundredResponses)
	fmt.Printf("  400s Responses: %d\n", stats.FourHundredResponses)
	fmt.Printf("  500s Responses: %d\n", stats.FiveHundredResponses)

	return nil
}
