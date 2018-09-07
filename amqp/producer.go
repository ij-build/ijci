package amqp

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Producer struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	confirmations chan amqp.Confirmation
	exchange      string
	routingKey    string
}

const MaxInFlight = 5

func NewProducer(
	conn *amqp.Connection,
	channel *amqp.Channel,
	exchange string,
	routingKey string,
) *Producer {
	confirmations := channel.NotifyPublish(make(
		chan amqp.Confirmation,
		MaxInFlight,
	))

	return &Producer{
		conn:          conn,
		channel:       channel,
		confirmations: confirmations,
		exchange:      exchange,
		routingKey:    routingKey,
	}
}

func (p *Producer) Shutdown() error {
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("failed to close AMQP connection (%s)", err.Error())
	}

	return nil
}

func (p *Producer) Publish(body []byte) error {
	if err := p.channel.Publish(
		p.exchange,
		p.routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "text/json",
			Body:         body,
			DeliveryMode: amqp.Transient,
		},
	); err != nil {
		return fmt.Errorf("failed to publish (%s)", err.Error())
	}

	confirmation := <-p.confirmations

	if !confirmation.Ack {
		return fmt.Errorf(
			"failed to deliver message (delivery tag %d)",
			confirmation.DeliveryTag,
		)
	}

	return nil
}
