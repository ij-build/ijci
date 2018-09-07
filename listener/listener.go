package listener

import (
	"github.com/efritz/nacelle"

	"github.com/efritz/ijci/amqp"
)

type Listener struct {
	Logger   nacelle.Logger `service:"logger"`
	Consumer *amqp.Consumer `service:"amqp-consumer"`
}

func NewListener() *Listener {
	return &Listener{
		Logger: nacelle.NewNilLogger(),
	}
}

func (l *Listener) Init(config nacelle.Config) error {
	return nil
}

func (l *Listener) Start() error {
	for delivery := range l.Consumer.Deliveries() {
		l.Logger.Info(
			"Processing message %d: '%s'\n",
			delivery.DeliveryTag,
			string(delivery.Body),
		)

		delivery.Ack(false)
	}

	l.Logger.Info("No longer consuming")
	return nil
}

func (l *Listener) Stop() error {
	return l.Consumer.Shutdown()
}
