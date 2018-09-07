package endpoints

import (
	"context"
	"net/http"
	"time"

	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	"github.com/efritz/ijci/amqp"
)

type HookResource struct {
	*chevron.EmptySpec

	Logger   nacelle.Logger `service:"logger"`
	Producer *amqp.Producer `service:"amqp-producer"`
}

func (hr *HookResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	message := []byte(time.Now().String())

	if err := hr.Producer.Publish(message); err != nil {
		return response.Empty(http.StatusInternalServerError)
	}

	return response.Empty(http.StatusOK)
}
