package httpbench

import (
	"net/http"
	"time"
)

type HTTPResponse struct {
	Latency time.Duration
	Status  int
}

// MakeRequest makes a HTTP request
func MakeRequest(url string) (*HTTPResponse, error) {

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	end := time.Since(start)
	status := response.StatusCode

	httpResponse := HTTPResponse{
		Latency: end,
		Status:  status,
	}

	return &httpResponse, nil
}

func parseURL(url string) string {
	return ""
}
