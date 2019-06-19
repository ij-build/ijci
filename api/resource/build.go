package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/chevron/middleware"
	"github.com/go-nacelle/nacelle"
	"github.com/efritz/response"

	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/util"
)

type (
	BuildResource struct {
		*chevron.EmptySpec
		DB *db.LoggingDB `service:"db"`
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
	build, resp := util.GetBuild(r.DB, logger, req)
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

	build, resp := util.GetBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if util.IsTerminal(build.BuildStatus) {
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

	build.AgentAddr = util.OrOptionalString(payload.AgentAddr, build.AgentAddr)
	build.CommitBranch = util.OrOptionalString(payload.CommitBranch, build.CommitBranch)
	build.CommitHash = util.OrOptionalString(payload.CommitHash, build.CommitHash)
	build.CommitMessage = util.OrOptionalString(payload.CommitMessage, build.CommitMessage)
	build.CommitAuthorName = util.OrOptionalString(payload.CommitAuthorName, build.CommitAuthorName)
	build.CommitAuthorEmail = util.OrOptionalString(payload.CommitAuthorEmail, build.CommitAuthorEmail)
	build.CommitAuthoredAt = util.OrOptionalTime(payload.CommitAuthoredAt, build.CommitAuthoredAt)
	build.CommitCommitterName = util.OrOptionalString(payload.CommitCommitterName, build.CommitCommitterName)
	build.CommitCommitterEmail = util.OrOptionalString(payload.CommitCommitterEmail, build.CommitCommitterEmail)
	build.CommitCommittedAt = util.OrOptionalTime(payload.CommitCommittedAt, build.CommitCommittedAt)
	build.ErrorMessage = util.OrOptionalString(payload.ErrorMessage, build.ErrorMessage)

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
	build, resp := util.GetBuild(r.DB, logger, req)
	if resp != nil {
		return resp
	}

	if !util.IsTerminal(build.BuildStatus) {
		return response.Empty(http.StatusConflict)
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
