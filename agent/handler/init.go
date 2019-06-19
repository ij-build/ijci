package handler

import "github.com/go-nacelle/nacelle"

type Initializer struct {
	Services nacelle.ServiceContainer `service:"services"`
}

const ServiceName = "handler"

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	handlerConfig := &Config{}
	if err := config.Load(handlerConfig); err != nil {
		return err
	}

	handler := NewHandler(handlerConfig.ScratchRoot)
	if err := i.Services.Inject(handler); err != nil {
		return err
	}

	return i.Services.Set(ServiceName, handler)
}
