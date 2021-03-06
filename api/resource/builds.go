package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/efritz/response"
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/chevron/middleware"
	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/google/uuid"
	"github.com/ij-build/ijci/amqp/client"
	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/util"
)

type (
	BuildsResource struct {
		*chevron.EmptySpec
		DB       *pgutil.LoggingDB    `service:"db"`
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
	pageMeta, resp := util.GetPageMeta(req)
	if resp != nil {
		return resp
	}

	builds, pagedResultsMeta, err := db.GetBuilds(
		r.DB,
		pageMeta,
		req.URL.Query().Get("filter"),
	)

	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"builds": builds,
		"meta":   pagedResultsMeta,
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

	now := time.Now()

	build := &db.BuildWithProject{
		Project: project,
		Build: &db.Build{
			BuildID:      uuid.New(),
			CommitBranch: payload.CommitBranch,
			CommitHash:   payload.CommitHash,
			BuildStatus:  "queued",
			CreatedAt:    now,
			QueuedAt:     now,
		},
	}

	if err := db.CreateBuild(r.DB, logger, build); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to create build record (%s)", err.Error()),
		)
	}

	if err := util.QueueBuild(r.Producer, build); err != nil {
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

		return project, nil
	}

	return db.GetOrCreateProject(r.DB, logger, *repositoryURL)
}
