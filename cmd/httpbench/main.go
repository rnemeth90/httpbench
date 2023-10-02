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
	url               string
	requestsPerSecond int
	headers           string
	bodyFileName      string
	timeout           int64
	keepalives        bool
	compression       bool
	duration          int
	method            string
	goroutines        int
	proxyAddress      string
	proxyUser         string
	proxyPass         string
	username          string
	password          string
	insecure          bool
}

var (
	url               string
	requestsPerSecond int
	insecure          bool
	headers           string
	bodyFileName      string
	timeout           int64
	keepalives        bool
	compression       bool
	duration          int
	method            string
	goroutines        int
	proxyAddress      string
	proxyUser         string
	proxyPass         string
	username          string
	password          string

	defaultGoRoutines = runtime.NumCPU()
)

func init() {
	pflag.SetInterspersed(false) // don't allow flags to be interspersed with positional args for POSIX compliance

	pflag.StringVarP(&url, "url", "u", "", "Target URL to which the HTTP requests will be sent.")
	pflag.IntVarP(&requestsPerSecond, "rps", "r", 4, "Number of requests to be sent per second.")
	pflag.IntVarP(&duration, "duration", "d", 10, "Duration of the test in seconds.")
	pflag.IntVarP(&goroutines, "goroutines", "g", defaultGoRoutines, "Number of concurrent goroutines to spawn for handling requests.")
	pflag.BoolVarP(&insecure, "insecure", "i", false, "Skip SSL certificate validation")
	pflag.StringVarP(&headers, "headers", "h", "", "Set request headers in a key:value format. Multiple headers can be separated by commas.")
	pflag.StringVarP(&proxyAddress, "proxyAddress", "p", "", "The URL of a proxy to use for all requests.")
	pflag.StringVarP(&proxyUser, "proxy-user", "", "", "Username for proxy authentication")
	pflag.StringVarP(&proxyPass, "proxy-pass", "", "", "Password for proxy authentication")
	pflag.StringVarP(&method, "method", "m", "GET", "HTTP method to use for the requests (e.g., GET, POST, PUT).")
	pflag.StringVarP(&bodyFileName, "bodyFile", "b", "", "Path to a JSON file containing the request body. Used for methods like POST or PUT.")
	pflag.Int64VarP(&timeout, "timeout", "t", 10, "Timeout in seconds for each request.")
	pflag.BoolVarP(&keepalives, "keepalives", "k", true, "Enable HTTP keep-alive, allowing re-use of TCP connections.")
	pflag.BoolVarP(&compression, "compression", "c", true, "Enable request and response compression.")
	pflag.StringVarP(&username, "username", "", "", "Username for basic auth to the endpoint")
	pflag.StringVarP(&password, "password", "", "", "Password for basic auth to the endpoint")

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

	fmt.Println("Usage:")
	fmt.Printf("  Basic use-case:\n")
	fmt.Printf("    httpbench -u, --url https://mywebsite.com\n\n")

	fmt.Printf("  Specify number of requests per second and duration:\n")
	fmt.Printf("    httpbench -u https://mywebsite.com -r, --rps 10 -d, --duration 30\n\n")

	fmt.Printf("  Making POST requests with a body file:\n")
	fmt.Printf("    httpbench -u https://mywebsite.com/api/posts -m, --method POST -b, --bodyFile /path/to/file.json\n\n")

	fmt.Printf("  Adding request headers:\n")
	fmt.Printf("    httpbench -u https://mywebsite.com -h, --headers \"Authorization:Bearer XYZ,Content-Type:application/json\"\n\n")

	fmt.Printf("  Using proxy with authentication:\n")
	fmt.Printf("    httpbench -u https://mywebsite.com -p, --proxyAddress http://proxy.com:8080 --proxy-user myuser --proxy-pass mypass\n\n")

	fmt.Printf("  Disabling keep-alive:\n")
	fmt.Printf("    httpbench -u https://mywebsite.com -k, --keepalives=false\n\n")

	fmt.Printf("  Skipping SSL validation:\n")
	fmt.Printf("    httpbench -u https://mywebsite.com -i, --insecure\n\n")

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

	if proxyPass != "" && (proxyUser == "" || proxyAddress == "") {
		fmt.Println("Specifying a proxy password must be accompanied by a proxy username and address.")
		os.Exit(1)
	}

	if proxyUser != "" && (proxyPass == "" || proxyAddress == "") {
		fmt.Println("Specifying a proxy username must be accompanied by a proxy password and address.")
		os.Exit(1)
	}

	c := config{
		url:               url,
		requestsPerSecond: requestsPerSecond,
		headers:           headers,
		bodyFileName:      bodyFileName,
		timeout:           timeout,
		keepalives:        keepalives,
		compression:       compression,
		duration:          duration,
		method:            method,
		goroutines:        goroutines,
		proxyAddress:      proxyAddress,
		proxyUser:         proxyUser,
		proxyPass:         proxyPass,
		username:          username,
		password:          password,
		insecure:          insecure,
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

	color.Green("Making %d calls per second for %d seconds...", c.requestsPerSecond, c.duration)

	numjobs := c.requestsPerSecond * c.duration
	respChan := make(chan httpbench.HTTPResponse, numjobs)
	reqChan := make(chan *http.Request, numjobs)

	if c.goroutines > numjobs {
		fmt.Fprintln(os.Stderr, "Number of goroutines exceeds the total number of requests. Adjusting goroutines to match request count.")
		goroutines = numjobs
	}

	httpbench.Dispatcher(reqChan, c.duration, c.requestsPerSecond, c.url, c.method, body, c.headers, c.username, c.password)

	httpbench.WorkerPool(reqChan, respChan, c.goroutines, c.requestsPerSecond, c.duration, c.timeout, c.keepalives, c.compression, c.proxyAddress, c.proxyUser, c.proxyPass, c.insecure)

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
