package apiclient

import "github.com/efritz/nacelle"

type Initializer struct {
	Container nacelle.ServiceContainer `service:"container"`
}

const ServiceName = "api"

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	apiConfig := &Config{}
	if err := config.Load(apiConfig); err != nil {
		return err
	}

	client := NewClient(
		apiConfig.APIAddr,
		apiConfig.PublicAddr,
	)

	if err := i.Container.Inject(client); err != nil {
		return err
	}

	return i.Container.Set(ServiceName, client)
}
