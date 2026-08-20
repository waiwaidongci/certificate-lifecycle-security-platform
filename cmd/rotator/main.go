package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	certapp "github.com/acme/certpilot/internal/certificate/application"
	certsqlite "github.com/acme/certpilot/internal/certificate/infrastructure/sqlite"
	issuermock "github.com/acme/certpilot/internal/issuer/adapter/mock"
	issuerapp "github.com/acme/certpilot/internal/issuer/application"
	issuersqlite "github.com/acme/certpilot/internal/issuer/infrastructure/sqlite"
	policyapp "github.com/acme/certpilot/internal/policy/application"
	policysqlite "github.com/acme/certpilot/internal/policy/infrastructure/sqlite"
	rotapp "github.com/acme/certpilot/internal/rotation/application"
	rotsqlite "github.com/acme/certpilot/internal/rotation/infrastructure/sqlite"
	serviceapp "github.com/acme/certpilot/internal/servicecatalog/application"
	servicesqlite "github.com/acme/certpilot/internal/servicecatalog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/config"
	"github.com/acme/certpilot/internal/shared/database"
	eventapp "github.com/acme/certpilot/internal/shared/eventlog/application"
	eventsqlite "github.com/acme/certpilot/internal/shared/eventlog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "", "path to YAML config")
	advanceDays := flag.Int("advance-days", 0, "override rotation advance days")
	execute := flag.Bool("execute", true, "execute due tasks after generating plans")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if *advanceDays > 0 {
		cfg.Rotation.DefaultAdvanceDays = *advanceDays
	}
	log := logger.New(os.Getenv("CERTPILOT_LOG_LEVEL"))
	clk := clock.RealClock{}
	db, err := database.Open(cfg.Database, log)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if cfg.Database.Driver == "sqlite" {
		if err := db.MigrateSQLite(ctx); err != nil {
			return err
		}
	}

	eventService := eventapp.NewService(db, eventsqlite.NewRepository(), clk)
	serviceService := serviceapp.NewService(db, servicesqlite.NewRepository(), clk)
	policyService := policyapp.NewService(db, policysqlite.NewRepository(), clk)
	issuerService := issuerapp.NewService(db, issuersqlite.NewRepository(), clk, issuermock.NewIssuer(clk.Now, cfg.Issuer.DefaultLifetimeDays, cfg.Issuer.CommonName))
	certService := certapp.NewService(db, certsqlite.NewRepository(), serviceService, policyService, issuerService, eventService, clk, cfg.Issuer.DefaultLifetimeDays)
	rotationService := rotapp.NewService(db, rotsqlite.NewRepository(), certService, eventService, clk)

	plans, err := rotationService.Generate(ctx, cfg.Rotation.DefaultAdvanceDays)
	if err != nil {
		return err
	}
	fmt.Printf("generated_plans=%d\n", len(plans))
	if *execute {
		tasks, execErr := rotationService.ExecuteDueTasks(ctx, 20)
		if execErr != nil {
			return execErr
		}
		fmt.Printf("executed_tasks=%d\n", len(tasks))
	}
	return nil
}
