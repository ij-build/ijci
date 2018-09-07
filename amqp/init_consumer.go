package amqp

import (
	"fmt"

	"github.com/efritz/nacelle"
)

type ConsumerInitializer struct {
	Logger    nacelle.Logger           `service:"logger"`
	Container nacelle.ServiceContainer `service:"container"`
}

const ServiceNameConsumer = "amqp-consumer"

func NewConsumerInitializer() nacelle.Initializer {
	return &ConsumerInitializer{
		Logger: nacelle.NewNilLogger(),
	}
}

func (ci *ConsumerInitializer) Init(config nacelle.Config) error {
	consumerConfig := &ConsumerConfig{}
	if err := config.Load(consumerConfig); err != nil {
		return err
	}

	var (
		consumerTag  = consumerConfig.ConsumerTag
		exchange     = consumerConfig.Exchange
		exchangeType = consumerConfig.ExchangeType
		queueName    = consumerConfig.QueueName
		routingKey   = consumerConfig.RoutingKey
		uri          = consumerConfig.URI
	)

	conn, channel, err := makeChannelAndEnsureExchange(
		uri,
		exchange,
		exchangeType,
		ci.Logger,
	)

	if err != nil {
		return err
	}

	if err := makeAndBindQueue(
		channel,
		exchange,
		queueName,
		routingKey,
		ci.Logger,
	); err != nil {
		return err
	}

	ci.Logger.DebugWithFields(nacelle.LogFields{
		"consumer_tag": consumerTag,
	}, "Consuming from queue")

	deliveries, err := channel.Consume(
		queueName,
		consumerTag,
		false, // noAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to get deliveries from queue (%s)", err.Error())
	}

	return ci.Container.Set(ServiceNameConsumer, NewConsumer(
		conn,
		channel,
		consumerTag,
		deliveries,
	))
}
