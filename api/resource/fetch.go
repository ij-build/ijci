package resource

import (
	"fmt"
	"net/http"

	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/util"
)

func getBuild(loggingDB *db.LoggingDB, logger nacelle.Logger, req *http.Request) (*db.Build, response.Response) {
	build, err := db.GetBuild(loggingDB, util.GetBuildID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return nil, response.Empty(http.StatusNotFound)
		}

		return nil, util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build record (%s)", err.Error()),
		)
	}

	return build, nil
}

func getBuildLog(loggingDB *db.LoggingDB, logger nacelle.Logger, req *http.Request) (*db.BuildLog, response.Response) {
	buildLog, err := db.GetBuildLog(loggingDB, util.GetBuildID(req), util.GetBuildLogID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return nil, response.Empty(http.StatusNotFound)
		}

		return nil, util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build log record (%s)", err.Error()),
		)
	}

	return buildLog, nil
}
