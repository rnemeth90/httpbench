package httpbench

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
