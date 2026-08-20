package metrics

import "sync/atomic"

func (m *Metrics) ensure(name string) {
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

func (m *Metrics) counterFor(name string) *atomic.Int64 {
	m.ensure(name)
	return m.counters[name]
}

func (m *Metrics) histogramSumFor(name string) *atomic.Int64 {
	m.ensure(name)
	return m.histogramSum[name]
}

func (m *Metrics) histogramCountFor(name string) *atomic.Int64 {
	m.ensure(name)
	return m.histogramCnt[name]
}
