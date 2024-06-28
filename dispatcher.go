package httpbench

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Dispatcher struct {
	ReqChan  chan *http.Request
	Duration int
	RPS      int
	URL      string
	Method   string
	Body     []byte
	Headers  string
	Username string
	Password string
}

func (d *Dispatcher) Dispatch() error {
	if !isValidMethod(d.Method) {
		log.Printf("Invalid HTTP Method: %s", d.Method)
		os.Exit(1)
	}

	parsedURL, err := url.Parse(d.URL)
	if err != nil {
		return fmt.Errorf("Invalid URL: %s", err)
	}

	parsedURL.Host = strings.ToLower(parsedURL.Host)
	d.URL = parsedURL.String()

	totalRequests := d.RPS * d.Duration
	for i := 0; i < totalRequests; i++ {
		req, err := http.NewRequest(d.Method, d.URL, bytes.NewBuffer(d.Body))
		if err != nil {
			log.Println(err)
			continue // if error, skip the current iteration and proceed with the next
		}

		if d.Username != "" && d.Password != "" {
			req.SetBasicAuth(d.Username, d.Password)
		}

		if d.Headers != "" {
			headerLines := strings.Split(d.Headers, ",")

			if err := setHeaders(req, headerLines); err != nil {
				log.Printf("%s\n", err)
				continue // continue with the next request
			}
		}
		d.ReqChan <- req
	}

	close(d.ReqChan)
	return nil
}
