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
	"github.com/google/uuid"

	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/util"
)

type (
	BuildLogsResource struct {
		*chevron.EmptySpec
		DB *db.LoggingDB `service:"db"`
	}

	jsonBuildLogPostPayload struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
)

func (r *BuildLogsResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	payload := &jsonBuildLogPostPayload{}
	if err := json.Unmarshal(middleware.GetJSONData(ctx), payload); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to unmarshal request body (%s)", err.Error()),
		)
	}

	build, err := db.GetBuild(r.DB, util.GetBuildID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return response.Empty(http.StatusNotFound)
		}

		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build record (%s)", err.Error()),
		)
	}

	buildLog := &db.BuildLog{
		BuildLogID: uuid.New(),
		BuildID:    build.BuildID,
		Name:       payload.Name,
		CreatedAt:  time.Now(),
	}

	if err := db.CreateBuildLog(r.DB, logger, buildLog); err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to create build log record (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"build_log": buildLog,
	}).SetStatusCode(http.StatusCreated)
}
