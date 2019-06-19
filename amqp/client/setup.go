package amqpclient

import (
	"fmt"
	"time"

	"github.com/go-nacelle/nacelle"
	"github.com/streadway/amqp"
)

const MaxDialAttempts = 15

func makeConnection(uri string, logger nacelle.Logger) (*amqp.Connection, error) {
	logger.Debug("Dialing AMQP broker")

	for attempts := 0; ; attempts++ {
		conn, err := amqp.Dial(uri)
		if err == nil {
			return conn, nil
		}

		if attempts >= MaxDialAttempts {
			return nil, fmt.Errorf("failed to dial broker in within timeout")
		}

		logger.Error("Failed to connect to amqp broker, will retry in 2s (%s)", err.Error())
		<-time.After(time.Second * 2)
	}
}

func makeChannel(
	conn *amqp.Connection,
	logger nacelle.Logger,
) (*amqp.Channel, error) {
	logger.Debug("Getting AMQP channel")

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()

		return nil, fmt.Errorf(
			"failed to create channel (%s)",
			err.Error(),
		)
	}

	return channel, nil
}

func makeExchange(
	channel *amqp.Channel,
	exchange string,
	logger nacelle.Logger,
) error {
	logger.Debug("Declaring AMQP exchange '%s'", exchange)

	if err := channel.ExchangeDeclare(
		exchange,
		"direct",
		true,  // durable
		false, // delete when complete
		false, // internal
		false, // noWait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf(
			"failed to declare exchange (%s)",
			err.Error(),
		)
	}

	return nil
}

func makeBoundQueue(
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

func setupConfirms(
	channel *amqp.Channel,
	logger nacelle.Logger,
) (<-chan amqp.Confirmation, <-chan amqp.Return, error) {
	logger.Debug("Putting AMQP channel into confirm mode")

	if err := channel.Confirm(false); err != nil {
		return nil, nil, fmt.Errorf(
			"failed to put channel into confirm mode (%s)",
			err.Error(),
		)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	returns := channel.NotifyReturn(make(chan amqp.Return, 1))

	return confirms, returns, nil
}
