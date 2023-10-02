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
	TotalCalls            int           `json:"total_calls"`
	TotalTime             time.Duration `json:"total_time"`
	AvgTimePerRequest     time.Duration `json:"avg_time_per_request"`
	FastestRequest        time.Duration `json:"fastest_request"`
	SlowestRequest        time.Duration `json:"slowest_request"`
	TwoHundredResponses   int           `json:"two_hundreds"`
	ThreeHundredResponses int           `json:"three_hundreds"`
	FourHundredResponses  int           `json:"four_hundreds"`
	FiveHundredResponses  int           `json:"five_hundreds"`
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
func Dispatcher(reqChan chan *http.Request, duration int, requestsPerSecond int, u string, method string, body []byte, headers string, username string, password string) {
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

	totalRequests := requestsPerSecond * duration
	for i := 0; i < totalRequests; i++ {
		req, err := http.NewRequest(method, u, bytes.NewBuffer(body))
		if err != nil {
			log.Println(err)
			continue // if error, skip the current iteration and proceed with the next
		}

		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}

		if headers != "" {
			headerLines := strings.Split(headers, ",")

			if err := setHeaders(req, headerLines); err != nil {
				log.Printf("%s\n", err)
				continue // continue with the next request
			}
		}
		reqChan <- req
	}

	close(reqChan)
}

func setHeaders(r *http.Request, headers []string) error {
	requestHeaders, err := parseHeaders(headers)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to parse headers: %s", err))
	}

	for k, v := range requestHeaders {
		r.Header.Set(k, v)
	}
	return nil
}

// worker pool
func WorkerPool(reqChan chan *http.Request, respChan chan HTTPResponse, goroutines int, requestsPerSecond int, duration int, timeout int64, keepalives, compression bool, proxyAddress, proxyUser, proxyPass string, skipSSLValidation bool) {
	var wg sync.WaitGroup
	client := createHTTPClient(timeout, keepalives, compression, proxyAddress, proxyUser, proxyPass, skipSSLValidation)
	for durationCounter := 1; durationCounter <= duration; durationCounter++ {
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go worker(client, reqChan, respChan, &wg)
		}

		var finished = durationCounter * requestsPerSecond
		color.Cyan("Finished sending %d requests...", finished)
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
