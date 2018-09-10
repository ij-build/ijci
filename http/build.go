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
	"github.com/gorilla/mux"

	"github.com/efritz/ijci/db"
)

type (
	BuildResource struct {
		*chevron.EmptySpec

		Logger nacelle.Logger `service:"logger"`
		DB     *db.LoggingDB  `service:"db"`
	}

	jsonBuildPatchPayload struct {
		BuildStatus string `json:"build_status"`
	}
)

func (r *BuildResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := r.getBuild(req)
	if resp != nil {
		return resp
	}

	return response.JSON(build)
}

func (r *BuildResource) Patch(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := r.getBuild(req)
	if resp != nil {
		return resp
	}

	payload := &jsonBuildPatchPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	build.BuildStatus = payload.BuildStatus

	if err := db.UpdateBuild(r.DB, r.Logger, build); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to update build (%s)", err.Error()),
		)
	}

	return response.JSON(build)
}

func (r *BuildResource) getBuild(req *http.Request) (*db.Build, response.Response) {
	build, err := db.GetBuild(r.DB, uuid.Must(uuid.Parse(mux.Vars(req)["build_id"])))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return nil, response.Empty(http.StatusNotFound)
		}

		return nil, internalError(
			r.Logger,
			fmt.Errorf("failed to fetch build record (%s)", err.Error()),
		)
	}

	return build, nil
}
