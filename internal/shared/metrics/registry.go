package metrics

import (
	"sort"
	"sync/atomic"
)

func (m *Metrics) ensure(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureLocked(name)
}

func (m *Metrics) ensureLocked(name string) {
	if _, ok := m.counters[name]; !ok {
		m.counters[name] = &atomic.Int64{}
	}
	if _, ok := m.histogramSum[name]; !ok {
		m.histogramSum[name] = &atomic.Int64{}
	}
	if _, ok := m.histogramCnt[name]; !ok {
		m.histogramCnt[name] = &atomic.Int64{}
	}
}

type metricSnapshot struct {
	keys         []string
	counters     map[string]int64
	histogramSum map[string]int64
	histogramCnt map[string]int64
}

func (m *Metrics) snapshot() metricSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snapshot := metricSnapshot{
		keys:         make([]string, 0, len(m.counters)+len(m.histogramCnt)),
		counters:     make(map[string]int64, len(m.counters)),
		histogramSum: make(map[string]int64, len(m.histogramSum)),
		histogramCnt: make(map[string]int64, len(m.histogramCnt)),
	}
	for key, value := range m.counters {
		snapshot.keys = append(snapshot.keys, "counter:"+key)
		snapshot.counters[key] = value.Load()
	}
	for key, value := range m.histogramCnt {
		snapshot.keys = append(snapshot.keys, "histogram:"+key)
		snapshot.histogramCnt[key] = value.Load()
		snapshot.histogramSum[key] = m.histogramSum[key].Load()
	}
	sort.Strings(snapshot.keys)
	return snapshot
}

func (m *Metrics) counterFor(name string) *atomic.Int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureLocked(name)
	return m.counters[name]
}

func (m *Metrics) histogramSumFor(name string) *atomic.Int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureLocked(name)
	return m.histogramSum[name]
}

func (m *Metrics) histogramCountFor(name string) *atomic.Int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureLocked(name)
	return m.histogramCnt[name]
}
