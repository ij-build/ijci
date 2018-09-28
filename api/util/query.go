package util

import (
	"fmt"
	"net/http"

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
)

func GetProject(loggingDB *db.LoggingDB, logger nacelle.Logger, req *http.Request) (*db.Project, response.Response) {
	build, err := db.GetProject(loggingDB, GetProjectID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return nil, response.Empty(http.StatusNotFound)
		}

		return nil, InternalError(
			logger,
			fmt.Errorf("failed to fetch project record (%s)", err.Error()),
		)
	}

	return build, nil
}

func GetBuild(loggingDB *db.LoggingDB, logger nacelle.Logger, req *http.Request) (*db.BuildWithProject, response.Response) {
	build, err := db.GetBuild(loggingDB, GetBuildID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return nil, response.Empty(http.StatusNotFound)
		}

		return nil, InternalError(
			logger,
			fmt.Errorf("failed to fetch build record (%s)", err.Error()),
		)
	}

	return build, nil
}

func GetBuildLog(loggingDB *db.LoggingDB, logger nacelle.Logger, req *http.Request) (*db.BuildLog, response.Response) {
	buildLog, err := db.GetBuildLog(loggingDB, GetBuildID(req), GetBuildLogID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return nil, response.Empty(http.StatusNotFound)
		}

		return nil, InternalError(
			logger,
			fmt.Errorf("failed to fetch build log record (%s)", err.Error()),
		)
	}

	return buildLog, nil
}
