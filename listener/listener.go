package listener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/efritz/ij/subcommand"
	"github.com/efritz/nacelle"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/handler"
	"github.com/efritz/ijci/message"
)

type Listener struct {
	Logger   nacelle.Logger  `service:"logger"`
	Consumer *amqp.Consumer  `service:"amqp-consumer"`
	Handler  handler.Handler `service:"handler"`
	apiAddr  string
}

func NewListener() *Listener {
	return &Listener{
		Logger: nacelle.NewNilLogger(),
	}
}

func (l *Listener) Init(config nacelle.Config) error {
	listenerConfig := &Config{}
	if err := config.Load(listenerConfig); err != nil {
		return err
	}

	l.apiAddr = listenerConfig.APIAddr
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

	logger.Info("Starting build")
	err := l.Handler.Handle(message)
	status := getStatus(err)
	logger.Info("Build completed with status %s", status)

	return l.updateBuild(message.BuildID, status)
}

func (l *Listener) updateBuild(buildID, status string) error {
	payload, err := json.Marshal(map[string]string{"build_status": status})
	if err != nil {
		return fmt.Errorf("failed to marshal API payload (%s)", err.Error())
	}

	url := fmt.Sprintf("%s/builds/%s", l.apiAddr, buildID)

	req, err := http.NewRequest("PATCH", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to construct API request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to patch build (%s)", err.Error())
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected %d status from API", resp.StatusCode)
	}

	return nil
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
