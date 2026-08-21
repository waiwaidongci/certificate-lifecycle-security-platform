package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Metrics struct {
	mu           sync.RWMutex
	counters     map[string]*atomic.Int64
	histogramSum map[string]*atomic.Int64
	histogramCnt map[string]*atomic.Int64
}

func New() *Metrics {
	return &Metrics{
		counters:     make(map[string]*atomic.Int64),
		histogramSum: make(map[string]*atomic.Int64),
		histogramCnt: make(map[string]*atomic.Int64),
	}
}

func (m *Metrics) Inc(name string) {
	m.counterFor(name).Add(1)
}

func (m *Metrics) Observe(name string, value float64) {
	m.histogramSumFor(name).Add(int64(value))
	m.histogramCountFor(name).Add(1)
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		m.mu.RLock()
		defer m.mu.RUnlock()
		keys := make([]string, 0, len(m.counters)+len(m.histogramCnt))
		for key := range m.counters {
			keys = append(keys, "counter:"+key)
		}
		for key := range m.histogramCnt {
			keys = append(keys, "histogram:"+key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if strings.HasPrefix(key, "counter:") {
				name := strings.TrimPrefix(key, "counter:")
				fmt.Fprintf(w, "# TYPE %s counter\n%s %d\n", name, name, m.counters[name].Load())
			} else {
				name := strings.TrimPrefix(key, "histogram:")
				fmt.Fprintf(w, "# TYPE %s histogram\n%s_sum %d\n%s_count %d\n", name, name, m.histogramSum[name].Load(), name, m.histogramCnt[name].Load())
			}
		}
	})
}
