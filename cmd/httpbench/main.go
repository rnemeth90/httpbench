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
	headers      string
	bodyFileName string
	timeout      int64
	keepalives   bool
	compression  bool
	duration     int
	method       string
	goroutines   int
	proxyAddress string
	proxyUser    string
	proxyPass    string
	username     string
	password     string
	insecure     bool
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
	method       string
	goroutines   int
	proxyAddress string
	proxyUser    string
	proxyPass    string
	username     string
	password     string

	defaultGoRoutines = runtime.NumCPU()
)

func init() {
	pflag.StringVarP(&url, "url", "u", "", "Target URL to which the HTTP requests will be sent.")
	pflag.IntVarP(&count, "requests", "r", 4, "Number of requests to be sent per second.")
	pflag.IntVarP(&duration, "duration", "d", 10, "Duration of the test in seconds.")
	pflag.IntVarP(&goroutines, "goroutines", "g", defaultGoRoutines, "Number of concurrent goroutines to spawn for handling requests.")
	pflag.BoolVarP(&insecure, "insecure", "i", false, "Skip SSL certificate validation")
	pflag.StringVarP(&headers, "headers", "h", "", "Set request headers in a key:value format. Multiple headers can be separated by commas.")
	pflag.StringVarP(&proxyAddress, "proxyAddress", "p", "", "The URL of a proxy to use for all requests.")
	pflag.StringVarP(&proxyUser, "proxy-user", "", "", "Username for proxy authentication")
	pflag.StringVarP(&proxyUser, "proxy-pass", "", "", "Password for proxy authentication")
	pflag.StringVarP(&method, "method", "m", "GET", "HTTP method to use for the requests (e.g., GET, POST, PUT).")
	pflag.StringVarP(&bodyFileName, "bodyFile", "b", "", "Path to a JSON file containing the request body. Used for methods like POST or PUT.")
	pflag.Int64VarP(&timeout, "timeout", "t", 10, "Timeout in seconds for each request.")
	pflag.BoolVarP(&keepalives, "keepalives", "k", true, "Enable HTTP keep-alive, allowing re-use of TCP connections.")
	pflag.BoolVarP(&compression, "compression", "c", true, "Enable request and response compression (usually gzip or deflate).")
	pflag.StringVarP(&username, "username", "", "", "Username for URL authentication")
	pflag.StringVarP(&password, "password", "", "", "Password for URL authentication")
	pflag.Usage = usage
}

const header = `
| |   | | | |       | |                   | |    
| |__ | |_| |_ _ __ | |__   ___ _ __   ___| |__  
| '_ \| __| __| '_ \| '_ \ / _ \ '_ \ / __| '_ \ 
| | | | |_| |_| |_) | |_) |  __/ | | | (__| | | |
|_| |_|\__|\__| .__/|_.__/ \___|_| |_|\___|_| |_|
              | |                                
              |_|                                
`

func usage() {
	fmt.Printf("%s", header)
	fmt.Printf("%s\n", os.Args[0])

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
		headers:      headers,
		bodyFileName: bodyFileName,
		timeout:      timeout,
		keepalives:   keepalives,
		compression:  compression,
		duration:     duration,
		method:       method,
		goroutines:   goroutines,
		proxyAddress: proxyAddress,
		proxyUser:    proxyUser,
		proxyPass:    proxyPass,
		username:     username,
		password:     password,
		insecure:     insecure,
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

	if c.goroutines > numjobs {
		fmt.Fprintln(os.Stderr, "Number of goroutines exceeds the total number of requests. Adjusting goroutines to match request count.")
		goroutines = numjobs
	}

	httpbench.Dispatcher(reqChan, c.goroutines, c.count, c.url, c.method, body, c.headers, c.username, c.password)

	httpbench.WorkerPool(reqChan, respChan, c.goroutines, c.duration, c.timeout, c.keepalives, c.compression, c.proxyAddress, c.proxyUser, c.proxyPass, c.insecure)

	var resultslice []httpbench.HTTPResponse

	// this is slow...
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
