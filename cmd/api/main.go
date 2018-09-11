package main

import (
	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	basehttp "github.com/efritz/nacelle/base/http"

	"github.com/efritz/ijci/amqp/client"
	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/resource"
	"github.com/efritz/ijci/api/s3"
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
		amqpclient.NewProducerInitializer(),
		nacelle.WithInitializerName("amqp-producer"),
	)

	setupRoutes := chevron.RouteInitializerFunc(resource.SetupRoutes)

	processes.RegisterProcess(
		basehttp.NewServer(chevron.NewInitializer(setupRoutes)),
		nacelle.WithProcessName("server"),
	)

	return nil
}

func main() {
	nacelle.NewBootstrapper("icji-api", setup).BootAndExit()
}
