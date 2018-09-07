package amqp

import (
	"fmt"

	"github.com/efritz/nacelle"
)

type ProducerInitializer struct {
	Logger    nacelle.Logger           `service:"logger"`
	Container nacelle.ServiceContainer `service:"container"`
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
		exchange     = producerConfig.Exchange
		exchangeType = producerConfig.ExchangeType
		routingKey   = producerConfig.RoutingKey
		uri          = producerConfig.URI
	)

	conn, channel, err := makeChannelAndEnsureExchange(
		uri,
		exchange,
		exchangeType,
		pi.Logger,
	)

	if err != nil {
		return err
	}

	pi.Logger.Debug("Putting AMQP channel into confirm mode")

	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("failed to put channel into confirm mode (%s)", err.Error())
	}

	return pi.Container.Set(ServiceNameProducer, NewProducer(
		conn,
		channel,
		exchange,
		routingKey,
	))
}
