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
	"github.com/efritz/ijci/api/s3"
	"github.com/efritz/ijci/util"
)

type BuildRequeueResource struct {
	*chevron.EmptySpec
	DB       *db.LoggingDB        `service:"db"`
	S3       s3.Client            `service:"s3"`
	Producer *amqpclient.Producer `service:"amqp-producer"`
}

func (r *BuildRequeueResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := getBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if err := deleteBuildLogFilesForBuild(ctx, r.DB, r.S3, build.BuildID); err != nil {
		return util.InternalError(logger, err)
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

	if err := queueBuild(r.Producer, build); err != nil {
		return util.InternalError(
			logger,
			err,
		)
	}

	return response.JSON(map[string]interface{}{
		"build": build,
	})
}
