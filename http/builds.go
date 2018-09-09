package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	"github.com/google/uuid"

	"github.com/efritz/ijci/amqp"
	"github.com/efritz/ijci/db"
	"github.com/efritz/ijci/message"
)

type (
	BuildsResource struct {
		*chevron.EmptySpec

		Logger   nacelle.Logger `service:"logger"`
		DB       *db.LoggingDB  `service:"db"`
		Producer *amqp.Producer `service:"amqp-producer"`
	}

	jsonBuildRequest struct {
		RepositoryURL string `json:"repository_url"`
	}
)

func (br *BuildsResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	requestPayload := &jsonBuildRequest{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), requestPayload); err != nil {
		return internalError(
			br.Logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	message := message.BuildRequest{
		BuildID:       uuid.New().String(),
		RepositoryURL: requestPayload.RepositoryURL,
	}

	messagePayload, err := message.Marshal()
	if err != nil {
		return internalError(
			br.Logger,
			fmt.Errorf("failed to marshal message (%s)", err.Error()),
		)
	}

	// TODO - put record in DB

	if err := br.Producer.Publish(messagePayload); err != nil {
		return internalError(
			br.Logger,
			fmt.Errorf("failed to publish message (%s)", err.Error()),
		)
	}

	return response.Empty(http.StatusOK)
}
