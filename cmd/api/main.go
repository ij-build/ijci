package main

import (
	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	basehttp "github.com/efritz/nacelle/base/http"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/http"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
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
