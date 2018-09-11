package main

import (
	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	basehttp "github.com/efritz/nacelle/base/http"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/db"
	"github.com/efritz/ijci/http"
	"github.com/efritz/ijci/s3"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	processes.RegisterInitializer(
		db.NewInitializer(),
		nacelle.WithInitializerName("db"),
	)

	processes.RegisterInitializer(
		s3.NewInitializer(),
		nacelle.WithInitializerName("s3"),
	)

	processes.RegisterInitializer(
		amqp.NewProducerInitializer(),
		nacelle.WithInitializerName("amqp-producer"),
	)

	setupRoutes := chevron.RouteInitializerFunc(http.SetupRoutes)

	processes.RegisterProcess(
		basehttp.NewServer(chevron.NewInitializer(setupRoutes)),
		nacelle.WithProcessName("server"),
	)

	return nil
}

func main() {
	nacelle.NewBootstrapper("icji-api", setup).BootAndExit()
}
