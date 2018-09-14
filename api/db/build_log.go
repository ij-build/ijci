package db

import (
	"fmt"
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

func GetBuildLogs(db sqlx.Queryer, buildID uuid.UUID) ([]*BuildLog, error) {
	buildLogs := []*BuildLog{}
	if err := sqlx.Select(db, &buildLogs, `select * from build_logs where build_id = $1`, buildID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return buildLogs, nil
}

func GetBuildLog(db sqlx.Queryer, buildID, buildLogID uuid.UUID) (*BuildLog, error) {
	buildLog := &BuildLog{}

	if err := sqlx.Get(
		db,
		buildLog,
		`select * from build_logs where build_id = $1 AND build_log_id = $2`,
		buildID,
		buildLogID,
	); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return buildLog, nil
}

func CreateBuildLog(db sqlx.Execer, logger nacelle.Logger, l *BuildLog) error {
	l.CreatedAt = time.Now()

	_, err := db.Exec(
		`insert into build_logs (
			build_log_id,
			build_id,
			name,
			created_at
		) values ($1, $2, $3, $4)`,
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

func UpdateBuildLog(db sqlx.Execer, logger nacelle.Logger, l *BuildLog) error {
	now := time.Now()
	l.UploadedAt = &now

	resp, err := db.Exec(
		`update build_logs
		set
			key = $1,
			uploaded_at = $2
		where
			build_log_id = $3`,
		l.Key,
		l.UploadedAt,
		l.BuildLogID,
	)

	if err != nil {
		return handlePostgresError(err, "update error")
	}

	ra, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows (%s)", err.Error())
	}

	if ra == 0 {
		return ErrDoesNotExist
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id":     l.BuildID,
		"build_log_id": l.BuildLogID,
	}, "Build log updated")

	return nil
}