package amqpclient

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Consumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	consumerTag string
	deliveries  <-chan amqp.Delivery
}

func NewConsumer(
	conn *amqp.Connection,
	channel *amqp.Channel,
	consumerTag string,
	deliveries <-chan amqp.Delivery,
) *Consumer {
	return &Consumer{
		conn:        conn,
		channel:     channel,
		consumerTag: consumerTag,
		deliveries:  deliveries,
	}
}

func (c *Consumer) Shutdown() error {
	if err := c.channel.Cancel(c.consumerTag, true); err != nil {
		return fmt.Errorf("failed to cancel consumer (%s)", err.Error())
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close AMQP connection (%s)", err.Error())
	}

	return nil
}

func (c *Consumer) Deliveries() <-chan amqp.Delivery {
	return c.deliveries
}
