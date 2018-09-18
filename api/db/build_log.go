package db

import (
	"time"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BuildLog struct {
	BuildLogID uuid.UUID  `db:"build_log_id" json:"build_log_id"`
	BuildID    uuid.UUID  `db:"build_id" json:"build_id"`
	Name       string     `db:"name" json:"name"`
	Key        *string    `db:"key" json:"key"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UploadedAt *time.Time `db:"uploaded_at" json:"uploaded_at"`
}

func GetBuildLogs(db *LoggingDB, buildID uuid.UUID) ([]*BuildLog, error) {
	query := `select * from build_logs where build_id = $1 order by created_at`

	buildLogs := []*BuildLog{}
	if err := sqlx.Select(db, &buildLogs, query, buildID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return buildLogs, nil
}

func GetBuildLog(db *LoggingDB, buildID, buildLogID uuid.UUID) (*BuildLog, error) {
	query := `select * from build_logs where build_id = $1 AND build_log_id = $2`

	buildLog := &BuildLog{}
	if err := sqlx.Get(db, buildLog, query, buildID, buildLogID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return buildLog, nil
}

func CreateBuildLog(db *LoggingDB, logger nacelle.Logger, l *BuildLog) error {
	query := `
	insert into build_logs (
		build_log_id,
		build_id,
		name,
		created_at
	) values ($1, $2, $3, $4)
	`

	_, err := db.Exec(
		query,
		l.BuildLogID,
		l.BuildID,
		l.Name,
		l.CreatedAt,
	)

	if err != nil {
		return handlePostgresError(err, "insert error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id":     l.BuildID,
		"build_log_id": l.BuildLogID,
	}, "Build log created")

	return nil
}

func UpdateBuildLog(db *LoggingDB, logger nacelle.Logger, l *BuildLog) error {
	query := `
	update build_logs
	set
		key = $1,
		uploaded_at = $2
	where
		build_log_id = $3
	`

	if _, err := db.Exec(
		query,
		l.Key,
		l.UploadedAt,
		l.BuildLogID,
	); err != nil {
		return handlePostgresError(err, "update error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id":     l.BuildID,
		"build_log_id": l.BuildLogID,
	}, "Build log updated")

	return nil
}
