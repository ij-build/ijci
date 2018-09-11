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
	"github.com/efritz/ijci/api/s3"
)

type (
	BuildResource struct {
		*chevron.EmptySpec

		Logger nacelle.Logger `service:"logger"`
		DB     *db.LoggingDB  `service:"db"`
		S3     s3.Client      `service:"s3"`
	}

	jsonBuildPatchPayload struct {
		BuildStatus string `json:"build_status"`
	}
)

func (r *BuildResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, err := db.GetBuild(r.DB, getBuildID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return response.Empty(http.StatusNotFound)
		}

		return internalError(
			r.Logger,
			fmt.Errorf("failed to fetch build record (%s)", err.Error()),
		)
	}

	buildLogs, err := db.GetBuildLogs(r.DB, build.BuildID)
	if err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to fetch build log records (%s)", err.Error()),
		)
	}

	buildLogContents := map[string]string{}
	for _, buildLog := range buildLogs {
		content, err := r.S3.Download(ctx, buildLog.Key)
		if err != nil {
			return internalError(
				r.Logger,
				fmt.Errorf("failed to fetch build log content (%s)", err.Error()),
			)
		}

		buildLogContents[buildLog.BuildLogID.String()] = content
	}

	return response.JSON(map[string]interface{}{
		"build":              build,
		"build_logs":         buildLogs,
		"build_log_contents": buildLogContents,
	})
}

func (r *BuildResource) Patch(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, err := db.GetBuild(r.DB, getBuildID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return response.Empty(http.StatusNotFound)
		}

		return internalError(
			r.Logger,
			fmt.Errorf("failed to fetch build record (%s)", err.Error()),
		)
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
