package apiclient

import "github.com/go-nacelle/nacelle"

type Initializer struct {
	Services nacelle.ServiceContainer `service:"services"`
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

	if err := i.Services.Inject(client); err != nil {
		return err
	}

	return i.Services.Set(ServiceName, client)
}
