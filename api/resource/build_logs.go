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
	"github.com/google/uuid"

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/s3"
)

type (
	BuildLogsResource struct {
		*chevron.EmptySpec

		Logger nacelle.Logger `service:"logger"`
		DB     *db.LoggingDB  `service:"db"`
		S3     s3.Client      `service:"s3"`
	}

	jsonBuildLogPostPayload struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
)

func (r *BuildLogsResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	payload := &jsonBuildLogPostPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

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

	buildLogID := uuid.New()
	key := buildLogID.String()

	if err := r.S3.Upload(ctx, key, payload.Content); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to upload log file (%s)", err.Error()),
		)
	}

	buildLog := &db.BuildLog{
		BuildLogID: buildLogID,
		BuildID:    build.BuildID,
		Name:       payload.Name,
		Key:        key,
	}

	if err := db.CreateBuildLog(r.DB, r.Logger, buildLog); err != nil {
		return internalError(
			r.Logger,
			fmt.Errorf("failed to create build log record (%s)", err.Error()),
		)
	}

	resp := response.JSON(buildLog)
	resp.SetStatusCode(http.StatusCreated)
	return resp
}
