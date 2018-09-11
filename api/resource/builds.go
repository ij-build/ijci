package resource

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

	"github.com/efritz/ijci/amqp/client"
	"github.com/efritz/ijci/amqp/message"
	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/util"
)

type (
	BuildsResource struct {
		*chevron.EmptySpec
		DB       *db.LoggingDB        `service:"db"`
		Producer *amqpclient.Producer `service:"amqp-producer"`
	}

	jsonBuildPostPayload struct {
		RepositoryURL string `json:"repository_url"`
	}
)

func (r *BuildsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	builds, err := db.GetBuilds(r.DB)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to build records (%s)", err.Error()),
		)
	}

	return response.JSON(builds)
}

func (r *BuildsResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	requestPayload := &jsonBuildPostPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), requestPayload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	buildID := uuid.New()
	repositoryURL := requestPayload.RepositoryURL

	build := &db.Build{
		BuildID:       buildID,
		RepositoryURL: repositoryURL,
	}

	if err := db.CreateBuild(r.DB, logger, build); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to create build record (%s)", err.Error()),
		)
	}

	if err := r.queueBuild(build); err != nil {
		return util.InternalError(
			logger,
			err,
		)
	}

	resp := response.JSON(build)
	resp.SetStatusCode(http.StatusCreated)
	resp.SetHeader("Location", fmt.Sprintf("/builds/%s", buildID))
	return resp
}

func (r *BuildsResource) queueBuild(build *db.Build) error {
	message := &message.BuildMessage{
		BuildID:       build.BuildID,
		RepositoryURL: build.RepositoryURL,
	}

	if err := r.Producer.Publish(message); err != nil {
		return fmt.Errorf("failed to publish message (%s)", err.Error())
	}

	return nil
}
