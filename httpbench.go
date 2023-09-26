package httpbench

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type HTTPResponse struct {
	Latency time.Duration
	Status  int
	Err     error
}

type Statistics struct {
	TotalCalls            int
	TotalTime             time.Duration
	AvgTimePerRequest     time.Duration
	FastestRequest        time.Duration
	SlowestRequest        time.Duration
	TwoHundredResponses   int
	ThreeHundredResponses int
	FourHundredResponses  int
	FiveHundredResponses  int
}

var validHTTPMethods = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
	"HEAD":   true,
}

func isValidMethod(method string) bool {
	return validHTTPMethods[method]
}

func createHTTPClient(timeout int64, keepalives bool, compression bool, proxyAddress, proxyUser, proxyPass string, skipSSLValidation bool) *http.Client {
	t := &http.Transport{}

	if !keepalives {
		t.MaxConnsPerHost = -1
		t.DisableKeepAlives = true
	}

	if !compression {
		t.DisableCompression = compression
	}

	if skipSSLValidation {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if proxyAddress != "" {
		proxyURL, err := url.Parse(proxyAddress)
		if err != nil {
			log.Fatal("Invalid proxy URL: ", err)
			os.Exit(1)
		}

		t.Proxy = http.ProxyURL(proxyURL)

		if proxyUser != "" && proxyPass != "" {
			proxyURL.User = url.UserPassword(proxyUser, proxyPass)
		}
	}

	return &http.Client{
		Timeout:   time.Second * time.Duration(timeout),
		Transport: t,
	}
}

// Dispatcher
func Dispatcher(reqChan chan *http.Request, goroutines int, requestCount int, duration int, u string, method string, body []byte, headers string, username string, password string) {
	if !isValidMethod(method) {
		log.Printf("Invalid HTTP Method: %s", method)
		os.Exit(1)
	}

	parsedURL, err := url.Parse(u)
	if err != nil {
		log.Fatal("Invalid URL:", err)
		os.Exit(1)
	}
	parsedURL.Host = strings.ToLower(parsedURL.Host)
	u = parsedURL.String()

	headerLines := strings.Split(headers, ",")

	// create the requests
	for i := 0; i < goroutines; i++ {
		req, err := http.NewRequest(method, u, bytes.NewBuffer(body))
		if err != nil {
			log.Println(err)
			continue // if error, skip the current iteration and proceed with the next
		}

		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}

		// should we move this outside the loop that creates requests?
		if headers != "" {
			requestHeaders, err := parseHeaders(headerLines)
			if err != nil {
				log.Printf("failed to parse headers: %s", err)
				continue
			}

			for k, v := range requestHeaders {
				req.Header.Set(k, v)
			}
		}
		reqChan <- req
	}

	close(reqChan)
}

// worker pool
func WorkerPool(reqChan chan *http.Request, respChan chan HTTPResponse, goroutines int, duration int, timeout int64, keepalives, compression bool, proxyAddress, proxyUser, proxyPass string, skipSSLValidation bool) {
	var wg sync.WaitGroup
	client := createHTTPClient(timeout, keepalives, compression, proxyAddress, proxyUser, proxyPass, skipSSLValidation)
	for durationCounter := 1; durationCounter <= duration; durationCounter++ {
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go worker(client, reqChan, respChan, &wg)
			fmt.Println("GoRoutines that currently exist: ", runtime.NumGoroutine())
		}

		var finished = durationCounter * goroutines
		color.Cyan("Finished sending %d requests per second...", finished)
		time.Sleep(1 * time.Second)
	}
	wg.Wait()
	close(respChan)
}

// Worker
func worker(client *http.Client, reqChan chan *http.Request, respChan chan HTTPResponse, wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range reqChan {
		start := time.Now()
		httpResponse := HTTPResponse{}

		resp, err := client.Transport.RoundTrip(req)
		if err != nil {
			httpResponse.Err = err
		}
		end := time.Since(start)
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}

		httpResponse.Status = resp.StatusCode
		httpResponse.Latency = end

		respChan <- httpResponse
	}
}

func BuildResults(requestCount int, respChan chan HTTPResponse) []HTTPResponse {
	results := make([]HTTPResponse, 0, requestCount)
	var conns int64

	for conns < int64(requestCount) {
		select {
		case r, ok := <-respChan:
			if ok {
				if r.Err != nil {
					log.Println(r.Err.Error())
				} else {
					results = append(results, r)
				}
				conns++
			}
		}
	}
	return results
}

func parseHeader(headerString string) (string, string, error) {
	colonIndex := strings.Index(headerString, ":")
	if colonIndex == -1 {
		return "", "", errors.New(fmt.Sprintf("invalid header format: %s", headerString))
	}

	headerName := strings.TrimSpace(headerString[:colonIndex])
	headerValue := strings.TrimSpace(headerString[colonIndex+1:])

	return headerName, headerValue, nil
}

func parseHeaders(headers []string) (map[string]string, error) {
	headerMap := make(map[string]string)

	for _, headerLine := range headers {
		headerName, headerValue, err := parseHeader(headerLine)
		if err != nil {
			return nil, err
		}
		headerMap[headerName] = headerValue
	}

	return headerMap, nil
}
