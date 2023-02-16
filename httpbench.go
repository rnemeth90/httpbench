package httpbench

import (
	"net/http"
	"sync"
	"time"
)

type HTTPResponse struct {
	Latency time.Duration
	Status  int
	Err     error
}

// MakeRequest makes a HTTP request
func MakeRequest(url string, rChan chan HTTPResponse, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer mu.Unlock()
	defer wg.Done()

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	httpResponse := HTTPResponse{}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		httpResponse.Err = err
	}

	mu.Lock()
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

func parseURL(url string) string {
	return ""
}
