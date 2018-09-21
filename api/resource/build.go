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

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/s3"
	"github.com/efritz/ijci/util"
)

type (
	BuildResource struct {
		*chevron.EmptySpec
		DB *db.LoggingDB `service:"db"`
		S3 s3.Client     `service:"s3"`
	}

	jsonBuildPatchPayload struct {
		BuildStatus          *string    `json:"build_status"`
		AgentAddr            *string    `json:"agent_addr"`
		CommitBranch         *string    `json:"commit_branch"`
		CommitHash           *string    `json:"commit_hash"`
		CommitMessage        *string    `json:"commit_message"`
		CommitAuthorName     *string    `json:"commit_author_name"`
		CommitAuthorEmail    *string    `json:"commit_author_email"`
		CommitAuthoredAt     *time.Time `json:"commit_authored_at"`
		CommitCommitterName  *string    `json:"commit_committer_name"`
		CommitCommitterEmail *string    `json:"commit_committer_email"`
		CommitCommittedAt    *time.Time `json:"commit_committed_at"`
		ErrorMessage         *string    `json:"error_message"`
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
		"build": &db.BuildWithLogs{build, buildLogs},
	})
}

func (r *BuildResource) Patch(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	payload := &jsonBuildPatchPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	build, resp := getBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if build.Canceled {
		return response.Empty(http.StatusConflict)
	}

	if payload.BuildStatus != nil {
		if util.JustStarted(build.BuildStatus, *payload.BuildStatus) {
			now := time.Now()
			build.StartedAt = &now
		}

		if util.JustCompleted(build.BuildStatus, *payload.BuildStatus) {
			now := time.Now()
			build.CompletedAt = &now
		}

		build.BuildStatus = *payload.BuildStatus
	}

	build.AgentAddr = orOptionalString(payload.AgentAddr, build.AgentAddr)
	build.CommitBranch = orOptionalString(payload.CommitBranch, build.CommitBranch)
	build.CommitHash = orOptionalString(payload.CommitHash, build.CommitHash)
	build.CommitMessage = orOptionalString(payload.CommitMessage, build.CommitMessage)
	build.CommitAuthorName = orOptionalString(payload.CommitAuthorName, build.CommitAuthorName)
	build.CommitAuthorEmail = orOptionalString(payload.CommitAuthorEmail, build.CommitAuthorEmail)
	build.CommitAuthoredAt = orOptionalTime(payload.CommitAuthoredAt, build.CommitAuthoredAt)
	build.CommitCommitterName = orOptionalString(payload.CommitCommitterName, build.CommitCommitterName)
	build.CommitCommitterEmail = orOptionalString(payload.CommitCommitterEmail, build.CommitCommitterEmail)
	build.CommitCommittedAt = orOptionalTime(payload.CommitCommittedAt, build.CommitCommittedAt)
	build.ErrorMessage = orOptionalString(payload.ErrorMessage, build.ErrorMessage)

	if err := db.UpdateBuild(r.DB, logger, build.Build); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to update build (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"build": build,
	})
}

func (r *BuildResource) Delete(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	build, resp := getBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if !build.Canceled && !util.IsTerminal(build.BuildStatus) {
		return response.Empty(http.StatusConflict)
	}

	buildLogs, err := db.GetBuildLogs(r.DB, build.BuildID)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build log records (%s)", err.Error()),
		)
	}

	keys := []string{}
	for _, buildLog := range buildLogs {
		if buildLog.Key != nil {
			keys = append(keys, *buildLog.Key)
		}
	}

	if err := r.S3.Delete(ctx, keys); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to delete build (%s)", err.Error()),
		)
	}

	if err := db.DeleteBuild(r.DB, logger, build.Build); err != nil {
		if err == db.ErrDoesNotExist {
			return response.Empty(http.StatusNotFound)
		}

		return util.InternalError(
			logger,
			fmt.Errorf("failed to delete build (%s)", err.Error()),
		)
	}

	return response.Empty(http.StatusNoContent)
}
