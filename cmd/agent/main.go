package main

import (
	"github.com/efritz/nacelle"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/listener"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	processes.RegisterInitializer(
		amqp.NewConsumerInitializer(),
		nacelle.WithInitializerName("amqp-consumer"),
	)

	processes.RegisterProcess(
		listener.NewListener(),
		nacelle.WithProcessName("listener"),
	)

	return nil
}

func main() {
	nacelle.NewBootstrapper("ijci-agent", setup).BootAndExit()
}
