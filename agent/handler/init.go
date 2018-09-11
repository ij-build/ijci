package handler

import "github.com/efritz/nacelle"

type Initializer struct {
	Container nacelle.ServiceContainer `service:"container"`
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
	if err := i.Container.Inject(handler); err != nil {
		return err
	}

	return i.Container.Set(ServiceName, handler)
}
