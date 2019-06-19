package log

import "github.com/go-nacelle/nacelle"

type Initializer struct {
	Services nacelle.ServiceContainer `service:"services"`
}

const ServiceName = "log-processor"

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	processor := NewLogProcessor()
	if err := i.Services.Inject(processor); err != nil {
		return err
	}

	return i.Services.Set(ServiceName, processor)
}
