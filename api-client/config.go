package api

type Config struct {
	APIAddr string `env:"api_addr" required:"true"`
}
