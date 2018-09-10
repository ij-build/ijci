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

	jsonBuildPostPayload struct {
		RepositoryURL string `json:"repository_url"`
	}
)

func (r *BuildsResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	requestPayload := &jsonBuildPostPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), requestPayload); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	buildID := uuid.New()
	repositoryURL := requestPayload.RepositoryURL

	build := &db.Build{
		BuildID:       buildID,
		RepositoryURL: repositoryURL,
	}

	if err := db.CreateBuild(r.DB, r.Logger, build); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to create build (%s)", err.Error()),
		)
	}

	message := &message.BuildMessage{
		BuildID:       buildID.String(),
		RepositoryURL: repositoryURL,
	}

	if err := r.Producer.Publish(message); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to publish message (%s)", err.Error()),
		)
	}

	resp := response.JSON(build)
	resp.SetStatusCode(http.StatusCreated)
	resp.SetHeader("Location", fmt.Sprintf("/builds/%s", buildID))
	return resp
}
