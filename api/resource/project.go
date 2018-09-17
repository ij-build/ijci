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

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/util"
)

type (
	ProjectResource struct {
		*chevron.EmptySpec
		DB *db.LoggingDB `service:"db"`
	}

	jsonProjectPatchPayload struct {
		Name          *string `json:"name"`
		RepositoryURL *string `json:"repository_url"`
	}
)

func (r *ProjectResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	project, err := db.GetProject(r.DB, util.GetProjectID(req))
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch project record (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"project": project,
	})
}

func (r *ProjectResource) Patch(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	project, resp := getProject(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	payload := &jsonProjectPatchPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	project.Name = orString(payload.Name, project.Name)
	project.RepositoryURL = orString(payload.RepositoryURL, project.RepositoryURL)

	if err := db.UpdateProject(r.DB, logger, project.Project); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to update project (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"project": project,
	})
}
