package main

import (
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/httpbase"
	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	amqpclient "github.com/ij-build/ijci/amqp/client"
	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/resource"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	monitor := db.NewMonitor()
	if err := services.Set("monitor", monitor); err != nil {
		return err
	}

	migrationSourceDriver, err := source.Open("file:///migrations")
	if err != nil {
		return err
	}

	processes.RegisterInitializer(
		pgutil.NewInitializer(pgutil.WithMigrationSourceDriver(migrationSourceDriver)),
		nacelle.WithInitializerName("db"),
	)

	processes.RegisterInitializer(
		amqpclient.NewProducerInitializer(),
		nacelle.WithInitializerName("amqp-producer"),
	)

	processes.RegisterProcess(
		monitor,
		nacelle.WithProcessName("monitor"),
	)

	processes.RegisterProcess(
		httpbase.NewServer(chevron.NewInitializer(resource.SetupRoutesFunc)),
		nacelle.WithProcessName("server"),
	)

	return nil
}

func main() {
	nacelle.NewBootstrapper("icji-api", setup).BootAndExit()
}
