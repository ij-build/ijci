package apiclient

import "fmt"

type Config struct {
	APIAddr      string `env:"api_addr" required:"true"`
	PublicHost   string `env:"public_host" required:"true"`
	PublicPort   int    `env:"public_port" default:"5000"`
	PublicScheme string `env:"public_scheme" default:"http"`
	PublicAddr   string
}

func (c *Config) PostLoad() error {
	c.PublicAddr = fmt.Sprintf(
		"%s://%s:%d",
		c.PublicScheme,
		c.PublicHost,
		c.PublicPort,
	)

	return nil
}
