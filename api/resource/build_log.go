package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	"github.com/google/uuid"

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/s3"
	"github.com/efritz/ijci/api/util"
)

type (
	BuildLogResource struct {
		*chevron.EmptySpec
		DB *db.LoggingDB `service:"db"`
		S3 s3.Client     `service:"s3"`
	}

	jsonBuildLogPatchPayload struct {
		Content string `json:"content"`
	}
)

func (r *BuildLogResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	for _, f := range []chevron.Handler{r.getContentFromS3, r.getContentFromAgent, r.getContentFromS3} {
		if resp := f(ctx, req, logger); resp != nil {
			return resp
		}
	}

	return response.Empty(http.StatusNotFound)
}

func (r *BuildLogResource) Patch(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	payload := &jsonBuildLogPatchPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	buildLog, resp := util.GetBuildLog(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	now := time.Now()
	key := buildLog.BuildLogID.String()

	buildLog.Key = &key
	buildLog.UploadedAt = &now

	if err := r.S3.Upload(ctx, key, payload.Content); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to upload log file (%s)", err.Error()),
		)
	}

	if err := db.UpdateBuildLog(r.DB, logger, buildLog); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to create build log record (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"build_log": buildLog,
	}).SetStatusCode(http.StatusCreated)
}

func (r *BuildLogResource) getContentFromS3(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	buildLog, resp := util.GetBuildLog(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if buildLog.Key == nil {
		return nil
	}

	logger.Info("Serving build log from S3")

	content, err := r.S3.Download(ctx, *buildLog.Key)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build log content from S3 (%s)", err.Error()),
		)
	}

	return response.Respond([]byte(content))
}

func (r *BuildLogResource) getContentFromAgent(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := util.GetBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if build.AgentAddr == nil {
		return nil
	}

	return r.streamFromAgent(
		logger,
		util.GetBuildID(req),
		util.GetBuildLogID(req),
		*build.AgentAddr,
	)
}

func (r *BuildLogResource) streamFromAgent(
	logger nacelle.Logger,
	buildID uuid.UUID,
	buildLogID uuid.UUID,
	agentAddr string,
) response.Response {
	url := fmt.Sprintf(
		"%s/builds/%s/logs/%s",
		agentAddr,
		buildID,
		buildLogID,
	)

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to request build log from agent (%s)", err.Error()),
		)
	}

	if resp.StatusCode == 404 {
		logger.Info("Build log not active on agent")
		resp.Body.Close()
		return nil
	}

	if 200 > resp.StatusCode || resp.StatusCode >= 300 {
		resp.Body.Close()

		return util.InternalError(
			logger,
			fmt.Errorf("unexpected %d status from agent", resp.StatusCode),
		)
	}

	logger.Info("Streaming active build log from S3")

	return response.Stream(resp.Body, response.WithFlush())
}
