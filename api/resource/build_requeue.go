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
	"github.com/ij-build/ijci/amqp/client"
	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/util"
)

type BuildRequeueResource struct {
	*chevron.EmptySpec
	DB       *pgutil.LoggingDB    `service:"db"`
	Producer *amqpclient.Producer `service:"amqp-producer"`
}

func (r *BuildRequeueResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := util.GetBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if err := db.DeleteBuildLogsForBuild(r.DB, logger, build.BuildID); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to clear build log records for build (%s)", err.Error()),
		)
	}

	build.Build = &db.Build{
		BuildID:      build.BuildID,
		CommitBranch: build.CommitBranch,
		CommitHash:   build.CommitHash,
		BuildStatus:  "queued",
		StartedAt:    build.StartedAt,
		QueuedAt:     time.Now(),
	}

	if err := db.UpdateBuild(r.DB, logger, build.Build); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to update build (%s)", err.Error()),
		)
	}

	if err := util.QueueBuild(r.Producer, build); err != nil {
		return util.InternalError(
			logger,
			err,
		)
	}

	return response.JSON(map[string]interface{}{
		"build": build,
	})
}
