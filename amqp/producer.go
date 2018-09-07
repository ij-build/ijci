package amqp

import (
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type Producer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	confirms   <-chan amqp.Confirmation
	returns    <-chan amqp.Return
	exchange   string
	routingKey string
	mutex      sync.Mutex
}

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

func (p *Producer) Publish(body []byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.channel.Publish(
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
	)

	if err != nil {
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
