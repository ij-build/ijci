package db

type Config struct {
	PostgresURL   string `env:"postgres_url"`
	LogSQLQueries bool   `env:"log_sql_queries" default:"false"`
}
