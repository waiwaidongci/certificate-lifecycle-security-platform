package httpx

import (
	"net/http"
	"time"

	"github.com/acme/certpilot/internal/shared/config"
)

func NewServer(addr string, handler http.Handler, cfg config.ServerConfig) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

func DefaultServerTimeouts() (read, write, idle time.Duration) {
	return 10 * time.Second, 15 * time.Second, 60 * time.Second
}
