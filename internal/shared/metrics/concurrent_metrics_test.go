package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentMetricCollectionIsRaceFree(t *testing.T) {
	metrics := New()
	start := make(chan struct{})
	var writers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		worker := worker
		writers.Add(1)
		go func() {
			defer writers.Done()
			<-start
			for sample := 0; sample < 200; sample++ {
				name := "request_" + string(rune('a'+worker))
				metrics.Inc(name)
				metrics.Observe("latency_"+name, float64(sample))
			}
		}()
	}
	var readers sync.WaitGroup
	readers.Add(1)
	go func() {
		defer readers.Done()
		<-start
		for sample := 0; sample < 100; sample++ {
			recorder := httptest.NewRecorder()
			metrics.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		}
	}()
	close(start)
	writers.Wait()
	readers.Wait()
	recorder := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(recorder.Body.String(), "request_a") || !strings.Contains(recorder.Body.String(), "latency_request_a") {
		t.Fatalf("metrics output omitted collected series: %s", recorder.Body.String())
	}
}
