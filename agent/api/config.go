package apiclient

type Config struct {
	APIAddr string `env:"api_addr" required:"true"`
}
