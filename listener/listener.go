package listener

import (
	"github.com/efritz/nacelle"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/handler"
	"github.com/efritz/ijci/message"
)

type Listener struct {
	Logger   nacelle.Logger  `service:"logger"`
	Consumer *amqp.Consumer  `service:"amqp-consumer"`
	Handler  handler.Handler `service:"handler"`
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
		if err := l.handle(delivery.Body); err != nil {
			delivery.Nack(false, false)
		} else {
			delivery.Ack(false)
		}
	}

	l.Logger.Info("No longer consuming")
	return nil
}

func (l *Listener) handle(payload []byte) error {
	message := &message.BuildRequest{}
	if err := message.Unmarshal(payload); err != nil {
		l.Logger.Error(
			"Failed to unmarshal message (%s)",
			err.Error(),
		)

		return nil
	}

	l.Logger.Info("Handling build %s", message.BuildID)

	if err := l.Handler.Handle(message); err != nil {
		l.Logger.Error(
			"Failed to handle message (%s)",
			err.Error(),
		)
	}

	return nil
}

func (l *Listener) Stop() error {
	return l.Consumer.Shutdown()
}
