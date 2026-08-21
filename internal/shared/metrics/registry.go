package metrics

import "sync/atomic"

// ensure lazily registers all three series (counter, histogram sum, histogram
// count) for the given name. It uses a double-checked locking pattern so that
// the hot path (an already-registered name) only takes a read lock, while the
// cold path (first observation of a name) takes a write lock and re-checks to
// avoid racing with another goroutine performing the same first registration.
func (m *Metrics) ensure(name string) {
	m.mu.RLock()
	_, counterOk := m.counters[name]
	_, sumOk := m.histogramSum[name]
	_, cntOk := m.histogramCnt[name]
	m.mu.RUnlock()
	if counterOk && sumOk && cntOk {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if !counterOk {
		if _, ok := m.counters[name]; !ok {
			m.counters[name] = &atomic.Int64{}
		}
	}
	if !sumOk {
		if _, ok := m.histogramSum[name]; !ok {
			m.histogramSum[name] = &atomic.Int64{}
		}
	}
	if !cntOk {
		if _, ok := m.histogramCnt[name]; !ok {
			m.histogramCnt[name] = &atomic.Int64{}
		}
	}
}

func (m *Metrics) counterFor(name string) *atomic.Int64 {
	m.ensure(name)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

func (m *Metrics) histogramSumFor(name string) *atomic.Int64 {
	m.ensure(name)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.histogramSum[name]
}

func (m *Metrics) histogramCountFor(name string) *atomic.Int64 {
	m.ensure(name)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.histogramCnt[name]
}
