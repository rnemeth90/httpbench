package httpbench

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateHTTPClient(t *testing.T) {
	testCases := []struct {
		name              string
		timeout           int64
		keepalives        bool
		compression       bool
		proxyAddress      string
		proxyUser         string
		proxyPass         string
		skipSSLValidation bool
	}{
		{name: "KeepAlivesAndCompression", timeout: 10, keepalives: true, compression: true, proxyAddress: "", proxyUser: "", proxyPass: "", skipSSLValidation: false},
		{name: "KeepAlivesAndNoCompression", timeout: 10, keepalives: true, compression: false, proxyAddress: "", proxyUser: "", proxyPass: "", skipSSLValidation: false},
		{name: "NoKeepAlivesAndCompression", timeout: 10, keepalives: false, compression: true, proxyAddress: "", proxyUser: "", proxyPass: "", skipSSLValidation: false},
		{name: "NoKeepAlivesAndNoCompression", timeout: 10, keepalives: false, compression: false, proxyAddress: "", proxyUser: "", proxyPass: "", skipSSLValidation: false},
		{name: "NoKeepAlivesAndNoCompressionSkipSSLValidation", timeout: 10, keepalives: false, compression: false, proxyAddress: "", proxyUser: "", proxyPass: "", skipSSLValidation: true},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := createHTTPClient(int64(test.timeout), test.keepalives, test.compression, "", "", "", true)

			expectedDuration := time.Duration(test.timeout) * time.Second
			if got.Timeout != expectedDuration {
				t.Errorf("expected: %v\ngot: %v", test.timeout, got.Timeout)
			}
		})
	}
}

func TestDispatcher(t *testing.T) {
	duration := 2
	rps := 5 // Requests per second
	expectedTotalRequests := duration * rps
	reqChan := make(chan *http.Request, expectedTotalRequests)

	server := createTestServer()

	// in the future, we might want to test different configurations using table driven tests.
	// Here's an example for a single configuration.
	dispatcher := Dispatcher{
		ReqChan:  reqChan,
		Duration: duration,
		RPS:      rps,
		URL:      server.URL,
		Method:   "GET",
		Body:     nil,
		Headers:  "",
		Username: "",
		Password: "",
	}
	if err := dispatcher.Dispatch(); err != nil {
		t.Fatal(err)
	}

	count := 0
	for range reqChan {
		count++
	}

	if count != expectedTotalRequests {
		t.Errorf("Expected %d requests, but got %d", expectedTotalRequests, count)
	}
}

func createTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, Gopher!"))
	}))
}
