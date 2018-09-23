package resource

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/efritz/ijci/amqp/client"
	"github.com/efritz/ijci/amqp/message"
	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/s3"
	"github.com/efritz/ijci/util"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	"github.com/google/uuid"
)

//
// Query Helpers

func getProject(loggingDB *db.LoggingDB, logger nacelle.Logger, req *http.Request) (*db.ProjectWithBuilds, response.Response) {
	build, err := db.GetProject(loggingDB, util.GetProjectID(req))
	if err != nil {
		if err == db.ErrDoesNotExist {
			return nil, response.Empty(http.StatusNotFound)
		}

		return nil, util.InternalError(
			logger,
			fmt.Errorf("failed to fetch project record (%s)", err.Error()),
		)
	}

	return build, nil
}

func getBuild(loggingDB *db.LoggingDB, logger nacelle.Logger, req *http.Request) (*db.BuildWithProject, response.Response) {
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

//
// Update Helpers

func deleteBuildLogFilesForProject(ctx context.Context, loggingDB *db.LoggingDB, s3 s3.Client, projectID uuid.UUID) error {
	buildLogs, err := db.GetBuildLogsForProject(loggingDB, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch build log records (%s)", err.Error())
	}

	return deleteBuildLogFiles(ctx, s3, buildLogs)
}

func deleteBuildLogFilesForBuild(ctx context.Context, loggingDB *db.LoggingDB, s3 s3.Client, buildID uuid.UUID) error {
	buildLogs, err := db.GetBuildLogs(loggingDB, buildID)
	if err != nil {
		return fmt.Errorf("failed to fetch build log records (%s)", err.Error())
	}

	return deleteBuildLogFiles(ctx, s3, buildLogs)
}

func deleteBuildLogFiles(ctx context.Context, s3 s3.Client, buildLogs []*db.BuildLog) error {
	keys := []string{}
	for _, buildLog := range buildLogs {
		if buildLog.Key != nil {
			keys = append(keys, *buildLog.Key)
		}
	}

	if err := s3.Delete(ctx, keys); err != nil {
		return fmt.Errorf("failed to delete build logs (%s)", err.Error())
	}

	return nil
}

//
// Queue Helpers

func queueBuild(producer *amqpclient.Producer, build *db.BuildWithProject) error {
	message := &message.BuildMessage{
		BuildID:       build.BuildID,
		RepositoryURL: build.Project.RepositoryURL,
		CommitBranch:  orString(build.CommitBranch, ""),
		CommitHash:    orString(build.CommitHash, ""),
	}

	if err := producer.Publish(message); err != nil {
		return fmt.Errorf("failed to publish message (%s)", err.Error())
	}

	return nil
}

//
// Optional Value Helpers

func orString(newVal *string, oldVal string) string {
	if newVal != nil {
		return *newVal
	}

	return oldVal
}

func orOptionalString(newVal, oldVal *string) *string {
	if newVal != nil {
		return newVal
	}

	return oldVal
}

func orOptionalTime(newVal, oldVal *time.Time) *time.Time {
	if newVal != nil {
		return newVal
	}

	return oldVal
}
