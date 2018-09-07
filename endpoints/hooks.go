package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	"github.com/efritz/ijci/amqp"
)

type (
	HookResource struct {
		*chevron.EmptySpec

		Logger   nacelle.Logger `service:"logger"`
		Producer *amqp.Producer `service:"amqp-producer"`
	}

	jsonRepo struct {
		Repo string `json:"repo"`
	}
)

func (hr *HookResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	jsonRepo := &jsonRepo{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), jsonRepo); err != nil {
		logger.Error("schema does not match shape of struct (%s)", err.Error())
		return response.Empty(http.StatusInternalServerError)
	}

	if err := hr.Producer.Publish([]byte(jsonRepo.Repo)); err != nil {
		return response.Empty(http.StatusInternalServerError)
	}

	return response.Empty(http.StatusOK)
}
