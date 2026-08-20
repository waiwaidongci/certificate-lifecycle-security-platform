package application

import (
	"sync"

	"github.com/acme/certpilot/internal/distribution/domain"
)

type dispatchGate struct {
	mu       sync.Mutex
	activeBy map[domain.DispatchKey]struct{}
}

func newDispatchGate() *dispatchGate {
	return &dispatchGate{activeBy: make(map[domain.DispatchKey]struct{})}
}

func (g *dispatchGate) Claim(key domain.DispatchKey) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, active := g.activeBy[key]; active {
		return false
	}
	g.activeBy[key] = struct{}{}
	return true
}

func (g *dispatchGate) Release(key domain.DispatchKey) {
	g.mu.Lock()
	delete(g.activeBy, key)
	g.mu.Unlock()
}
