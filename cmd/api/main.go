package main

import (
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/httpbase"
	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/ij-build/ijci/amqp/client"
	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/resource"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	processes.RegisterInitializer(
		pgutil.NewInitializer(),
		nacelle.WithInitializerName("db"),
	)

	processes.RegisterInitializer(
		amqpclient.NewProducerInitializer(),
		nacelle.WithInitializerName("amqp-producer"),
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
