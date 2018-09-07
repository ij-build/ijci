package amqp

import (
	"fmt"

	"github.com/efritz/nacelle"
	"github.com/streadway/amqp"
)

func makeChannelAndEnsureExchange(
	uri string,
	exchange string,
	exchangeType string,
	logger nacelle.Logger,
) (*amqp.Connection, *amqp.Channel, error) {
	logger.Debug("Dialing AMQP broker")

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to amqp broker (%s)", err.Error())
	}


	logger.Debug("Getting AMQP channel")

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create channel (%s)", err.Error())
	}

	logger.Debug("Declaring AMQP exchange '%s'", exchange)

	if err := channel.ExchangeDeclare(
		exchange,
		exchangeType,
		true,  // durable
		false, // delete when complete
		false, // internal
		false, // noWait
		nil,   // arguments
	); err != nil {
		return nil, nil, fmt.Errorf("failed to declare exchange (%s)", err.Error())
	}

	return conn, channel, nil
}

func makeAndBindQueue(
	channel *amqp.Channel,
	exchange string,
	queueName string,
	routingKey string,
	logger nacelle.Logger,
) error {
	logger.Debug("Declaring AMQP queue '%s'", queueName)

	queue, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue (%s)", err.Error())
	}

	logger.DebugWithFields(nacelle.LogFields{
		"messages":  queue.Messages,
		"consumers": queue.Consumers,
	}, "Declared queue")

	logger.Debug("Binding AMQP exchange '%s'", routingKey)

	if err := channel.QueueBind(
		queue.Name,
		routingKey,
		exchange,
		false, // noWait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("failed to bind queue (%s)", err.Error())
	}

	return nil
}
