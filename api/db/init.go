package db

import (
	"github.com/efritz/nacelle"
	_ "github.com/lib/pq"
)

type Initializer struct {
	Logger    nacelle.Logger           `service:"logger"`
	Container nacelle.ServiceContainer `service:"container"`
}

const ServiceName = "db"

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	dbConfig := &Config{}
	if err := config.Load(dbConfig); err != nil {
		return err
	}

	logger := i.Logger
	if !dbConfig.LogSQLQueries {
		logger = nacelle.NewNilLogger()
	}

	db, err := Dial(dbConfig.PostgresURL, logger)
	if err != nil {
		return err
	}

	return i.Container.Set(ServiceName, db)
}
