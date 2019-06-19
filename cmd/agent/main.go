package main

import (
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/httpbase"
	"github.com/go-nacelle/nacelle"

	"github.com/ij-build/ijci/agent/api"
	"github.com/ij-build/ijci/agent/context"
	"github.com/ij-build/ijci/agent/handler"
	"github.com/ij-build/ijci/agent/listener"
	"github.com/ij-build/ijci/agent/log"
	"github.com/ij-build/ijci/agent/resource"
	"github.com/ij-build/ijci/amqp/client"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	processes.RegisterInitializer(
		amqpclient.NewConsumerInitializer(),
		nacelle.WithInitializerName("amqp-consumer"),
	)

	processes.RegisterInitializer(
		apiclient.NewInitializer(),
		nacelle.WithInitializerName("api-client"),
	)

	processes.RegisterInitializer(
		context.NewInitializer(),
		nacelle.WithInitializerName("context"),
	)

	processes.RegisterInitializer(
		log.NewInitializer(),
		nacelle.WithInitializerName("log"),
	)

	processes.RegisterInitializer(
		handler.NewInitializer(),
		nacelle.WithInitializerName("handler"),
	)

	processes.RegisterProcess(
		httpbase.NewServer(chevron.NewInitializer(resource.SetupRoutesFunc)),
		nacelle.WithProcessName("server"),
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
