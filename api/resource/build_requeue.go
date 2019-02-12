package resource

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	"github.com/efritz/ijci/amqp/client"
	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/util"
)

type BuildRequeueResource struct {
	*chevron.EmptySpec
	DB       *db.LoggingDB        `service:"db"`
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
