package httpbench

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
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

func CreateHTTPClient(timeout int64, keepalives bool, compression bool) *http.Client {

	t := &http.Transport{}

	if !keepalives {
		t.MaxConnsPerHost = -1
		t.DisableKeepAlives = true
	}

	if !compression {
		t.DisableCompression = compression
	}

	return &http.Client{
		Timeout:   time.Second * time.Duration(timeout),
		Transport: t,
	}
}

// Dispatcher
func Dispatcher(reqChan chan *http.Request, requestCount int, useHTTP bool, url string, method string, body []byte, headers string) {
	defer close(reqChan)

	if !strings.Contains(url, "http") {
		url = ParseURL(url, useHTTP)
	}

	url = strings.ToLower(url)

	// create the requests
	for i := 0; i < requestCount; i++ {
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
		if err != nil {
			log.Println(err)
		}

		var requestHeaders map[string]string
		if headers != "" {
			requestHeaders = ParseHeaders(headers)

			for k, v := range requestHeaders {
				req.Header.Set(k, v)
			}
		}
		reqChan <- req
	}
}

// Worker Pool
func WorkerPool(reqChan chan *http.Request, respChan chan HTTPResponse, duration int, maxConnections int, timeout int64, keepalives, compression bool) {
	client := CreateHTTPClient(timeout, keepalives, compression)
	for durationCounter := 1; durationCounter <= duration; durationCounter++ {
		for i := 0; i < maxConnections; i++ {
			go worker(client, reqChan, respChan)
		}

		var finished = durationCounter * maxConnections
		color.Cyan("Finished sending %d requests...", finished)
		time.Sleep(1 * time.Second)
	}
}

// Worker
func worker(client *http.Client, reqChan chan *http.Request, respChan chan HTTPResponse) {
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

	// build results here?
}

func BuildResults(requestCount int, respChan chan HTTPResponse) []HTTPResponse {
	results := []HTTPResponse{}
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

func ParseURL(url string, useHTTP bool) string {
	if !useHTTP {
		return fmt.Sprintf("https://%s", strings.ToLower(url))
	}
	return fmt.Sprintf("http://%s", strings.ToLower(url))
}

func ParseHeaders(headers string) map[string]string {
	m := make(map[string]string)

	csvs := strings.Split(headers, ",")
	for _, v := range csvs {
		headers := strings.Split(v, ":")
		for i := 0; i < len(headers)-1; i++ {
			m[headers[i]] = headers[i+1]
		}
	}
	return m
}
