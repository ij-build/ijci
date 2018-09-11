package db

type Config struct {
	PostgresURL string `env:"postgres_url"`
}
