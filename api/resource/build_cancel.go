package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	"github.com/google/uuid"

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/util"
)

type BuildCancelResource struct {
	*chevron.EmptySpec
	DB *db.LoggingDB `service:"db"`
}

func (r *BuildCancelResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := getBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if build.Canceled || util.IsTerminal(build.BuildStatus) {
		return response.Empty(http.StatusConflict)
	}

	build.Canceled = true

	if err := db.UpdateBuild(r.DB, logger, build.Build); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to update build (%s)", err.Error()),
		)
	}

	if build.AgentAddr == nil {
		return response.Empty(http.StatusOK)
	}

	return r.cancelOnAgent(*build.AgentAddr, build.BuildID, logger)
}

func (r *BuildCancelResource) cancelOnAgent(agentAddr string, buildID uuid.UUID, logger nacelle.Logger) response.Response {
	url := fmt.Sprintf(
		"%s/builds/%s/cancel",
		agentAddr,
		buildID,
	)

	resp, err := http.DefaultClient.Post(url, "", nil)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to cancel build on agent (%s)", err.Error()),
		)
	}

	resp.Body.Close()

	if resp.StatusCode == 404 {
		return response.Empty(http.StatusNotFound)
	}

	if 200 > resp.StatusCode || resp.StatusCode >= 300 {
		return util.InternalError(
			logger,
			fmt.Errorf("unexpected %d status from agent", resp.StatusCode),
		)
	}

	return response.Empty(http.StatusOK)
}
