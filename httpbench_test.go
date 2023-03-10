package httpbench_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/rnemeth90/httpbench"
)

func TestCreateHTTPClient(t *testing.T) {
	testCases := []struct {
		name        string
		timeout     int64
		keepalives  bool
		compression bool
	}{
		{name: "KeepAlivesAndCompression", timeout: 10, keepalives: true, compression: true},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := httpbench.CreateHTTPClient(int64(test.timeout), test.keepalives, test.compression)

			if int64(got.Timeout) != test.timeout {
				t.Errorf("expected: %v\ngot: %v", test.timeout, got.Timeout)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		useHTTP bool
		want    string
	}{
		{name: "secureURL", url: "google.com", useHTTP: false, want: "https://google.com"},
		{name: "insecureURL", url: "google.com", useHTTP: true, want: "http://google.com"},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := httpbench.ParseURL(test.url, test.useHTTP)

			if got != test.want {
				t.Errorf("expected: %s\ngot: %s", test.want, got)
			}
		})
	}
}

func TestParseHeaders(t *testing.T) {
	testCases := []struct {
		name         string
		headerString string
	}{
		{name: "test1", headerString: "server:aws,x-sid:apollo"},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			want := make(map[string]string)
			want["server"] = "aws"
			want["x-sid"] = "apollo"

			got := httpbench.ParseHeaders(test.headerString)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("expected: %v\ngot: %v", want, got)
			}
		})
	}
}

func TestMakeRequestAsync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("host", "tester")
		w.WriteHeader(http.StatusTeapot)
	}))

	results := []httpbench.HTTPResponse{}
	client := http.Client{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	httpbench.MakeRequestAsync(server.URL, false, "", nil, &mu, &wg, &client, &results)

	expect := httpbench.HTTPResponse{
		Latency: results[0].Latency, //this obviously isn't testing anything...
		Status:  http.StatusTeapot,
		Err:     nil,
	}

	if !reflect.DeepEqual(results[0], expect) {
		t.Errorf("expected: %v\ngot: %v\n", expect, results[0])
	}
}
