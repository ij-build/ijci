package main

import (
	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	basehttp "github.com/efritz/nacelle/base/http"

	"github.com/efritz/ijci/agent/api"
	"github.com/efritz/ijci/agent/handler"
	"github.com/efritz/ijci/agent/listener"
	"github.com/efritz/ijci/agent/logs"
	"github.com/efritz/ijci/agent/resource"
	"github.com/efritz/ijci/amqp/client"
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
		logs.NewInitializer(),
		nacelle.WithInitializerName("logs"),
	)

	processes.RegisterInitializer(
		handler.NewInitializer(),
		nacelle.WithInitializerName("handler"),
	)

	processes.RegisterProcess(
		basehttp.NewServer(chevron.NewInitializer(resource.SetupRoutesFunc)),
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
