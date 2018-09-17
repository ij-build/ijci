package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
		ProjectID     *string `json:"project_id"`
		RepositoryURL *string `json:"repository_url"`
		CommitBranch  *string `json:"commit_branch"`
		CommitHash    *string `json:"commit_hash"`
	}
)

func (r *BuildsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	builds, err := db.GetBuilds(r.DB)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"builds": builds,
	})
}

func (r *BuildsResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	payload := &jsonBuildPostPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	project, err := r.getProject(payload.ProjectID, payload.RepositoryURL, logger)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to create project record (%s)", err.Error()),
		)
	}

	build := &db.BuildWithProject{
		Project: project,
		Build: &db.Build{
			BuildID:      uuid.New(),
			CommitBranch: payload.CommitBranch,
			CommitHash:   payload.CommitHash,
			BuildStatus:  "queued",
			CreatedAt:    time.Now(),
		},
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

	resp := response.JSON(map[string]interface{}{
		"build": build,
	})

	resp.SetStatusCode(http.StatusCreated)
	resp.SetHeader("Location", fmt.Sprintf("/builds/%s", build.BuildID))
	return resp
}

func (r *BuildsResource) getProject(projectID, repositoryURL *string, logger nacelle.Logger) (*db.Project, error) {
	if projectID != nil {
		project, err := db.GetProject(r.DB, uuid.Must(uuid.Parse(*projectID)))
		if err != nil {
			return nil, err
		}

		return project.Project, nil
	}

	return db.GetOrCreateProject(r.DB, logger, *repositoryURL)
}

func (r *BuildsResource) queueBuild(build *db.BuildWithProject) error {
	message := &message.BuildMessage{
		BuildID:       build.BuildID,
		RepositoryURL: build.Project.RepositoryURL,
		CommitBranch:  orString(build.CommitBranch, ""),
		CommitHash:    orString(build.CommitHash, ""),
	}

	if err := r.Producer.Publish(message); err != nil {
		return fmt.Errorf("failed to publish message (%s)", err.Error())
	}

	return nil
}
