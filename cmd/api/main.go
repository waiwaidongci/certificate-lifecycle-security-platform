package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	certhttp "github.com/acme/certpilot/internal/certificate/adapter/http"
	certapp "github.com/acme/certpilot/internal/certificate/application"
	certsqlite "github.com/acme/certpilot/internal/certificate/infrastructure/sqlite"
	disthttp "github.com/acme/certpilot/internal/distribution/adapter/http"
	distlog "github.com/acme/certpilot/internal/distribution/adapter/log"
	distwebhook "github.com/acme/certpilot/internal/distribution/adapter/webhook"
	distapp "github.com/acme/certpilot/internal/distribution/application"
	distsqlite "github.com/acme/certpilot/internal/distribution/infrastructure/sqlite"
	issuerhttp "github.com/acme/certpilot/internal/issuer/adapter/http"
	issuermock "github.com/acme/certpilot/internal/issuer/adapter/mock"
	issuerapp "github.com/acme/certpilot/internal/issuer/application"
	issuersqlite "github.com/acme/certpilot/internal/issuer/infrastructure/sqlite"
	notifhttp "github.com/acme/certpilot/internal/notification/adapter/http"
	notiflog "github.com/acme/certpilot/internal/notification/adapter/log"
	notifwebhook "github.com/acme/certpilot/internal/notification/adapter/webhook"
	notifapp "github.com/acme/certpilot/internal/notification/application"
	notifsqlite "github.com/acme/certpilot/internal/notification/infrastructure/sqlite"
	policyhttp "github.com/acme/certpilot/internal/policy/adapter/http"
	policyapp "github.com/acme/certpilot/internal/policy/application"
	policysqlite "github.com/acme/certpilot/internal/policy/infrastructure/sqlite"
	rothttp "github.com/acme/certpilot/internal/rotation/adapter/http"
	rotapp "github.com/acme/certpilot/internal/rotation/application"
	rotsqlite "github.com/acme/certpilot/internal/rotation/infrastructure/sqlite"
	servicehttp "github.com/acme/certpilot/internal/servicecatalog/adapter/http"
	serviceapp "github.com/acme/certpilot/internal/servicecatalog/application"
	servicesqlite "github.com/acme/certpilot/internal/servicecatalog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/config"
	"github.com/acme/certpilot/internal/shared/database"
	eventhttp "github.com/acme/certpilot/internal/shared/eventlog/adapter/http"
	eventapp "github.com/acme/certpilot/internal/shared/eventlog/application"
	eventsqlite "github.com/acme/certpilot/internal/shared/eventlog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/httpx"
	"github.com/acme/certpilot/internal/shared/logger"
	"github.com/acme/certpilot/internal/shared/metrics"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "", "path to YAML config")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	log := logger.New(os.Getenv("CERTPILOT_LOG_LEVEL"))
	clk := clock.RealClock{}
	db, err := database.Open(cfg.Database, log)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if cfg.Database.Driver == "sqlite" {
		if err := db.MigrateSQLite(ctx); err != nil {
			cancel()
			return err
		}
	}
	if err := db.Ping(ctx); err != nil {
		cancel()
		return err
	}
	cancel()

	eventRepo := eventsqlite.NewRepository()
	eventService := eventapp.NewService(db, eventRepo, clk)

	serviceRepo := servicesqlite.NewRepository()
	serviceService := serviceapp.NewService(db, serviceRepo, clk)

	policyRepo := policysqlite.NewRepository()
	policyService := policyapp.NewService(db, policyRepo, clk)

	issuerRepo := issuersqlite.NewRepository()
	issuerPort := issuermock.NewIssuer(clk.Now, cfg.Issuer.DefaultLifetimeDays, cfg.Issuer.CommonName)
	issuerService := issuerapp.NewService(db, issuerRepo, clk, issuerPort)

	certRepo := certsqlite.NewRepository()
	certService := certapp.NewService(db, certRepo, serviceService, policyService, issuerService, eventService, clk, cfg.Issuer.DefaultLifetimeDays)

	distAdapters := map[string]distapp.DistributionAdapter{
		"log":     distlog.NewAdapter(log),
		"webhook": distwebhook.NewAdapter(cfg.Webhook.Timeout),
	}
	distRepo := distsqlite.NewRepository()
	distService := distapp.NewService(db, distRepo, serviceService, certService, eventService, clk, distAdapters)

	rotRepo := rotsqlite.NewRepository()
	rotService := rotapp.NewService(db, rotRepo, certService, eventService, clk)

	notifAdapters := map[string]notifapp.NotificationAdapter{
		"log":     notiflog.NewAdapter(log),
		"webhook": notifwebhook.NewAdapter(cfg.Webhook.Timeout),
	}
	notifRepo := notifsqlite.NewRepository()
	notifService := notifapp.NewService(db, notifRepo, certService, serviceService, eventService, clk, notifAdapters)

	mux := http.NewServeMux()
	m := metrics.New()
	mux.Handle("GET /healthz", healthHandler(db))
	mux.Handle("GET /readyz", readyHandler(db))
	mux.Handle("GET /metrics", m.Handler())

	servicehttp.NewHandler(serviceService).Register(mux)
	policyhttp.NewHandler(policyService).Register(mux)
	issuerhttp.NewHandler(issuerService).Register(mux)
	certhttp.NewHandler(certService).Register(mux)
	disthttp.NewHandler(distService).Register(mux)
	rothttp.NewHandler(rotService).Register(mux)
	notifhttp.NewHandler(notifService).Register(mux)
	eventhttp.NewHandler(eventService).Register(mux)

	handler := buildMiddleware(mux, log, m, cfg.Server)

	server := httpx.NewServer(cfg.Server.Address, handler, cfg.Server)

	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()
	go reminderWorker(workerCtx, notifService, log, 10*time.Minute)

	errCh := make(chan error, 1)
	go func() {
		log.Info(context.Background(), "api server starting", "address", cfg.Server.Address, "database", cfg.Database.Driver)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		log.Info(context.Background(), "shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info(context.Background(), "api server stopped")
	return nil
}

func buildMiddleware(handler http.Handler, log *logger.Logger, m *metrics.Metrics, cfg config.ServerConfig) http.Handler {
	handler = httpx.AccessLog(log, m)(handler)
	handler = httpx.PanicRecovery(log)(handler)
	handler = httpx.Timeout(cfg.WriteTimeout)(handler)
	handler = httpx.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst).Middleware(handler)
	handler = httpx.AuthPlaceholder(handler)
	handler = httpx.RequestID(handler)
	return handler
}

func healthHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func readyHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func reminderWorker(ctx context.Context, service *notifapp.Service, log *logger.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := service.Scan(ctx, 30, "log"); err != nil {
				log.Error(ctx, "reminder scan failed", "error", err)
			}
			if _, err := service.SendPending(ctx, 20); err != nil {
				log.Error(ctx, "reminder send failed", "error", err)
			}
		}
	}
}
