package httpbench

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HTTPResponse struct {
	Latency time.Duration
	Status  int
	Err     error
}

// MakeRequest makes a HTTP request
func MakeRequest(url string, useHTTP bool, connCount int, headers string, rChan chan HTTPResponse, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer mu.Unlock()
	defer wg.Done()
	var requestHeaders map[string]string

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = connCount
	t.MaxIdleConnsPerHost = connCount
	t.MaxConnsPerHost = connCount

	mu.Lock()
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}

	httpResponse := HTTPResponse{}

	if !strings.Contains(url, "http") {
		url = parseURL(url, useHTTP)
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		httpResponse.Err = err
	}

	if headers != "" {
		requestHeaders = parseHeaders(headers)

		for k, v := range requestHeaders {
			request.Header.Set(k, v)
		}
	}

	start := time.Now()
	response, err := client.Do(request)
	if err != nil {
		httpResponse.Err = err
	}
	defer response.Body.Close()

	end := time.Since(start)
	status := response.StatusCode

	httpResponse = HTTPResponse{
		Latency: end,
		Status:  status,
	}

	rChan <- httpResponse
}

func parseURL(url string, useHTTP bool) string {
	if !useHTTP {
		return fmt.Sprintf("https://%s", strings.ToLower(url))
	}

	return fmt.Sprintf("http://%s", strings.ToLower(url))
}

func parseHeaders(headers string) map[string]string {
	m := make(map[string]string)
	return m
}
