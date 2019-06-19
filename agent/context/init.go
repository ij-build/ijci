package context

import "github.com/go-nacelle/nacelle"

type Initializer struct {
	Services nacelle.ServiceContainer `service:"services"`
}

const ServiceName = "context-processor"

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	processor := NewContextProcessor()
	if err := i.Services.Inject(processor); err != nil {
		return err
	}

	return i.Services.Set(ServiceName, processor)
}
