package amqpclient

import (
	"github.com/go-nacelle/nacelle"
)

type ProducerInitializer struct {
	Logger    nacelle.Logger           `service:"logger"`
	Services nacelle.ServiceContainer `service:"services"`
}

const ServiceNameProducer = "amqp-producer"

func NewProducerInitializer() nacelle.Initializer {
	return &ProducerInitializer{
		Logger: nacelle.NewNilLogger(),
	}
}

func (pi *ProducerInitializer) Init(config nacelle.Config) error {
	producerConfig := &ProducerConfig{}
	if err := config.Load(producerConfig); err != nil {
		return err
	}

	var (
		exchange   = producerConfig.Exchange
		routingKey = producerConfig.RoutingKey
		uri        = producerConfig.URI
	)

	conn, err := makeConnection(uri, pi.Logger)
	if err != nil {
		return err
	}

	channel, err := makeChannel(conn, pi.Logger)
	if err != nil {
		conn.Close()
		return err
	}

	if err := makeExchange(channel, exchange, pi.Logger); err != nil {
		conn.Close()
		return err
	}

	confirms, returns, err := setupConfirms(channel, pi.Logger)
	if err != nil {
		conn.Close()
		return err
	}

	return pi.Services.Set(ServiceNameProducer, NewProducer(
		conn,
		channel,
		confirms,
		returns,
		exchange,
		routingKey,
	))
}
