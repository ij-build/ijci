package main

import (
	"github.com/efritz/nacelle"
	"github.com/efritz/scarf"
	"github.com/efritz/scarf/logging"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/grpc"
	"github.com/efritz/ijci/handler"
	"github.com/efritz/ijci/listener"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	processes.RegisterInitializer(
		amqp.NewConsumerInitializer(),
		nacelle.WithInitializerName("amqp-consumer"),
	)

	processes.RegisterInitializer(
		handler.NewInitializer(),
		nacelle.WithInitializerName("handler"),
	)

	processes.RegisterProcess(
		listener.NewListener(),
		nacelle.WithProcessName("listener"),
	)

	// TODO - make a scarf function for this

	processes.RegisterInitializer(
		logging.NewInitializer(scarf.DefaultExtractors),
		nacelle.WithInitializerName("log decorator"),
	)

	processes.RegisterProcess(
		scarf.NewServer(grpc.NewEndpointSet()),
		nacelle.WithProcessName("server"),
	)

	return nil
}

func main() {
	nacelle.NewBootstrapper("ijci-agent", setup).BootAndExit()
}
