package util

import (
	"context"
	"fmt"

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/api/s3"
	"github.com/google/uuid"
)

func DeleteBuildLogFilesForProject(ctx context.Context, loggingDB *db.LoggingDB, s3 s3.Client, projectID uuid.UUID) error {
	buildLogs, err := db.GetBuildLogsForProject(loggingDB, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch build log records (%s)", err.Error())
	}

	return DeleteBuildLogFiles(ctx, s3, buildLogs)
}

func DeleteBuildLogFilesForBuild(ctx context.Context, loggingDB *db.LoggingDB, s3 s3.Client, buildID uuid.UUID) error {
	buildLogs, err := db.GetBuildLogs(loggingDB, buildID)
	if err != nil {
		return fmt.Errorf("failed to fetch build log records (%s)", err.Error())
	}

	return DeleteBuildLogFiles(ctx, s3, buildLogs)
}

func DeleteBuildLogFiles(ctx context.Context, s3 s3.Client, buildLogs []*db.BuildLog) error {
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
