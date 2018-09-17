package listener

import (
	"github.com/efritz/ij/subcommand"
	"github.com/efritz/nacelle"

	"github.com/efritz/ijci/agent/api"
	"github.com/efritz/ijci/agent/handler"
	"github.com/efritz/ijci/amqp/client"
	"github.com/efritz/ijci/amqp/message"
)

type Listener struct {
	Logger    nacelle.Logger       `service:"logger"`
	Consumer  *amqpclient.Consumer `service:"amqp-consumer"`
	Handler   handler.Handler      `service:"handler"`
	APIClient apiclient.Client     `service:"api"`
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

	logger := l.Logger.WithFields(nacelle.LogFields{
		"build_id": message.BuildID,
	})

	initialStatus := "in-progress"

	if err := l.APIClient.UpdateBuild(message.BuildID, &apiclient.BuildPayload{
		BuildStatus: &initialStatus,
	}); err != nil {
		return err
	}

	logger.Info("Starting build")
	err := l.Handler.Handle(message, logger)
	status := getStatus(err)
	logger.Info("Build completed with status %s", status)

	return l.APIClient.UpdateBuild(message.BuildID, &apiclient.BuildPayload{
		BuildStatus:  &status,
		ErrorMessage: getErrorMessage(err),
	})
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

func getErrorMessage(err error) *string {
	if err == nil || err == subcommand.ErrBuildFailed {
		return nil
	}

	message := err.Error()
	return &message
}
