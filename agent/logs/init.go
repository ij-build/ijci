package logs

import "github.com/efritz/nacelle"

type Initializer struct {
	Container nacelle.ServiceContainer `service:"container"`
}

const ServiceName = "log-processor"

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	processor := NewLogProcessor()
	if err := i.Container.Inject(processor); err != nil {
		return err
	}

	return i.Container.Set(ServiceName, processor)
}
