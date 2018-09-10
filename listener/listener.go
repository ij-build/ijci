package listener

import (
	"github.com/efritz/ij/subcommand"
	"github.com/efritz/nacelle"
	"github.com/google/uuid"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/api-client"
	"github.com/efritz/ijci/handler"
	"github.com/efritz/ijci/message"
)

type Listener struct {
	Logger    nacelle.Logger  `service:"logger"`
	Consumer  *amqp.Consumer  `service:"amqp-consumer"`
	Handler   handler.Handler `service:"handler"`
	APIClient api.Client      `service:"api"`
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
			l.Logger.Error(
				"Failed to handle message (%s)",
				err.Error(),
			)

			delivery.Nack(false, false)
			continue
		}

		delivery.Ack(false)
	}

	l.Logger.Info("No longer consuming")
	return nil
}

func (l *Listener) Stop() error {
	return l.Consumer.Shutdown()
}

func (l *Listener) handle(payload []byte) error {
	message := &message.BuildMessage{}
	if err := message.Unmarshal(payload); err != nil {
		l.Logger.Error(
			"Failed to unmarshal message (%s)",
			err.Error(),
		)

		return nil
	}

	buildID := uuid.Must(uuid.Parse(message.BuildID))

	logger := l.Logger.WithFields(nacelle.LogFields{
		"build_id": buildID,
	})

	if err := l.APIClient.UpdateBuildStatus(buildID, "in-progress"); err != nil {
		return err
	}

	logger.Info("Starting build")
	err := l.Handler.Handle(message, logger)
	status := getStatus(err)
	logger.Info("Build completed with status %s", status)

	return l.APIClient.UpdateBuildStatus(buildID, status)
}

//
// Helpers

func getStatus(err error) string {
	if err != nil {
		if err == subcommand.ErrBuildFailed {
			return "failed"
		}

		return "errored"
	}

	return "succeeded"
}
