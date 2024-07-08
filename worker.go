package httpbench

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"
)

type worker struct {
	client   *http.Client
	reqChan  chan *http.Request
	respChan chan HTTPResponse
	wg       *sync.WaitGroup
}

func newWorker(client *http.Client, reqChan chan *http.Request, respChan chan HTTPResponse, wg *sync.WaitGroup) *worker {
	return &worker{
		client:   client,
		reqChan:  reqChan,
		respChan: respChan,
		wg:       wg,
	}
}

func WorkerPool(reqChan chan *http.Request, respChan chan HTTPResponse, goroutines int, requestsPerSecond int, duration int, timeout int64, keepalives, compression bool, proxyAddress, proxyUser, proxyPass string, skipSSLValidation bool) {
	var wg sync.WaitGroup
	client := createHTTPClient(timeout, keepalives, compression, proxyAddress, proxyUser, proxyPass, skipSSLValidation)
	for durationCounter := 1; durationCounter <= duration; durationCounter++ {
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			worker := newWorker(client, reqChan, respChan, &wg)
			worker.work()
		}

		var finished = durationCounter * requestsPerSecond
		color.Cyan("Finished sending %d requests...", finished)
		time.Sleep(1 * time.Second)
	}
	wg.Wait()
	close(respChan)
}

func (w *worker) work() {
	defer w.wg.Done()
	for req := range w.reqChan {
		start := time.Now()
		httpResponse := HTTPResponse{}

		resp, err := w.client.Transport.RoundTrip(req)
		if err != nil {
			httpResponse.Err = err
			w.respChan <- httpResponse
			continue
		}

		if resp == nil {
			httpResponse.Err = errors.New("received nil response")
			w.respChan <- httpResponse
			continue
		}

		end := time.Since(start)
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}

		httpResponse.Status = resp.StatusCode
		httpResponse.Latency = end

		w.respChan <- httpResponse
	}
}
