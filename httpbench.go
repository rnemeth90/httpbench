package httpbench

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

type Statistics struct {
	UsedConnections       int
	UsedThreads           int
	TotalCalls            int
	TotalTime             time.Time
	AvgTimePerRequest     time.Duration
	RequestsPerSecond     float64
	FastestRequest        time.Duration
	SlowestRequest        time.Duration
	TwoHundredResponses   int
	ThreeHundredResponses int
	FourHundredResponses  int
	FiveHundredResponses  int
}

func CreateHTTPClient(timeout int64, keepalives bool, compression bool) http.Client {

	t := &http.Transport{}

	if !keepalives {
		t.MaxConnsPerHost = -1
		t.DisableKeepAlives = true
	}

	if !compression {
		t.DisableCompression = compression
	}

	return http.Client{
		Timeout:   time.Second * time.Duration(timeout),
		Transport: t,
	}
}

// MakeRequest makes a HTTP request
func MakeRequestAsync(url string, useHTTP bool, headers string, mu *sync.Mutex, wg *sync.WaitGroup, client *http.Client, results *[]HTTPResponse) {
	var requestHeaders map[string]string
	defer wg.Done()
	// client trace to log whether the request's underlying tcp connection was re-used
	//clientTrace := &httptrace.ClientTrace{
	//	GotConn: func(info httptrace.GotConnInfo) { log.Printf("conn was reused: %t", info.Reused) },
	//}
	//traceCtx := httptrace.WithClientTrace(context.Background(), clientTrace)

	mu.Lock()
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
	if _, err := io.Copy(ioutil.Discard, response.Body); err != nil {
		log.Fatal(err)
	}
	response.Body.Close()

	end := time.Since(start)
	status := response.StatusCode

	httpResponse = HTTPResponse{
		Latency: end,
		Status:  status,
	}
	mu.Unlock()

	*results = append(*results, httpResponse)
}

func parseURL(url string, useHTTP bool) string {
	if !useHTTP {
		return fmt.Sprintf("https://%s", strings.ToLower(url))
	}

	return fmt.Sprintf("http://%s", strings.ToLower(url))
}

func parseHeaders(headers string) map[string]string {
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
