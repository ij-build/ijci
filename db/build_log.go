package db

import (
	"github.com/efritz/nacelle"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BuildLog struct {
	BuildLogID uuid.UUID `db:"build_log_id" json:"build_log_id"`
	BuildID    uuid.UUID `db:"build_id" json:"build_id"`
	Name       string    `db:"name" json:"name"`
	Content    string    `db:"content" json:"content"`
}

func GetBuildLogs(db sqlx.Queryer, buildID uuid.UUID) ([]*BuildLog, error) {
	buildLogs := []*BuildLog{}
	if err := sqlx.Select(db, &buildLogs, `select * from build_logs where build_id = $1`, buildID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return buildLogs, nil
}

func CreateBuildLog(db sqlx.Execer, logger nacelle.Logger, l *BuildLog) error {
	_, err := db.Exec(
		`insert into build_logs (
			build_log_id,
			build_id,
			name,
			content
		) values ($1, $2, $3, $4)`,
		l.BuildLogID,
		l.BuildID,
		l.Name,
		l.Content,
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
