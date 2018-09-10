package amqp

import (
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type (
	Producer struct {
		conn       *amqp.Connection
		channel    *amqp.Channel
		confirms   <-chan amqp.Confirmation
		returns    <-chan amqp.Return
		exchange   string
		routingKey string
		mutex      sync.Mutex
	}

	Marshaler interface {
		Marshal() ([]byte, error)
	}
)

func NewProducer(
	conn *amqp.Connection,
	channel *amqp.Channel,
	confirms <-chan amqp.Confirmation,
	returns <-chan amqp.Return,
	exchange string,
	routingKey string,
) *Producer {
	return &Producer{
		conn:       conn,
		channel:    channel,
		confirms:   confirms,
		returns:    returns,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (p *Producer) Shutdown() error {
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("failed to close AMQP connection (%s)", err.Error())
	}

	return nil
}

func (p *Producer) Publish(message Marshaler) error {
	body, err := message.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message (%s)", err.Error())
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.channel.Publish(
		p.exchange,
		p.routingKey,
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "text/plain",
			Body:         body,
		},
	); err != nil {
		return err
	}

	if confirm := <-p.confirms; !confirm.Ack {
		return fmt.Errorf("publish was nacked")
	}

	select {
	case <-p.returns:
		return fmt.Errorf("message was not routed")
	default:
	}

	return nil
}
