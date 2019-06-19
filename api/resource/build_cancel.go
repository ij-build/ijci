package resource

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/efritz/response"
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/google/uuid"
	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/util"
)

type BuildCancelResource struct {
	*chevron.EmptySpec
	DB *pgutil.LoggingDB `service:"db"`
}

func (r *BuildCancelResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := util.GetBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if util.IsTerminal(build.BuildStatus) {
		return response.Empty(http.StatusConflict)
	}

	now := time.Now()
	build.BuildStatus = "canceled"
	build.CompletedAt = &now

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
