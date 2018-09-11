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
	BuildResource struct {
		*chevron.EmptySpec
		DB *db.LoggingDB `service:"db"`
	}

	jsonBuildPatchPayload struct {
		BuildStatus string `json:"build_status"`
		AgentAddr   string `json:"agent_addr"`
	}
)

func (r *BuildResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := getBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	buildLogs, err := db.GetBuildLogs(r.DB, build.BuildID)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build log records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"build":      build,
		"build_logs": buildLogs,
	})
}

func (r *BuildResource) Patch(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := getBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	payload := &jsonBuildPatchPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	build.BuildStatus = payload.BuildStatus
	build.AgentAddr = &payload.AgentAddr

	if err := db.UpdateBuild(r.DB, logger, build); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to update build (%s)", err.Error()),
		)
	}

	return response.JSON(build)
}
