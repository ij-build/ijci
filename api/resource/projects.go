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

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/util"
)

type (
	ProjectsResource struct {
		*chevron.EmptySpec
		DB *db.LoggingDB `service:"db"`
	}

	jsonProjectPostPayload struct {
		Name          string `json:"name"`
		RepositoryURL string `json:"repository_url"`
	}
)

func (r *ProjectsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	pageMeta, resp := util.GetPageMeta(req)
	if resp != nil {
		return resp
	}

	projects, pagedResultsMeta, err := db.GetProjects(r.DB, pageMeta)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch project records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"projects": projects,
		"meta":     pagedResultsMeta,
	})
}

func (r *ProjectsResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	payload := &jsonProjectPostPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	project := &db.Project{
		ProjectID:     uuid.New(),
		Name:          payload.Name,
		RepositoryURL: payload.RepositoryURL,
	}

	if err := db.CreateProject(r.DB, logger, project); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to create project record (%s)", err.Error()),
		)
	}

	resp := response.JSON(map[string]interface{}{
		"project": project,
	})

	resp.SetStatusCode(http.StatusCreated)
	resp.SetHeader("Location", fmt.Sprintf("/projects/%s", project.ProjectID))
	return resp
}
