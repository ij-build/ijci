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
	"github.com/efritz/ijci/api/util"
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
	payload := &jsonProjectPatchPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	project, resp := util.GetProject(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	project.Name = util.OrString(payload.Name, project.Name)
	project.RepositoryURL = util.OrString(payload.RepositoryURL, project.RepositoryURL)

	if err := db.UpdateProject(r.DB, logger, project); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to update project (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"project": project,
	})
}

func (r *ProjectResource) Delete(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	if err := db.DeleteProject(r.DB, logger, util.GetProjectID(req)); err != nil {
		if err == db.ErrDoesNotExist {
			return response.Empty(http.StatusNotFound)
		}

		return util.InternalError(
			logger,
			fmt.Errorf("failed to delete project (%s)", err.Error()),
		)
	}

	return response.Empty(http.StatusNoContent)
}
