package db

import (
	"fmt"
	"time"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Build struct {
	BuildID       uuid.UUID `db:"build_id" json:"build_id"`
	RepositoryURL string    `db:"repository_url" json:"repository_url"`
	BuildStatus   string    `db:"build_status" json:"build_status"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func GetBuild(db sqlx.Queryer, buildID uuid.UUID) (*Build, error) {
	b := &Build{}
	if err := sqlx.Get(db, b, `select * from builds where build_id = $1`, buildID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return b, nil
}

func CreateBuild(db sqlx.Execer, logger nacelle.Logger, b *Build) error {
	now := time.Now()
	b.BuildStatus = "queued"
	b.CreatedAt = now
	b.UpdatedAt = now

	_, err := db.Exec(
		`insert into builds (
			build_id,
			repository_url,
			build_status,
			created_at,
			updated_at
		) values ($1, $2, $3, $4, $5)`,
		b.BuildID,
		b.RepositoryURL,
		b.BuildStatus,
		b.CreatedAt,
		b.UpdatedAt,
	)

	if err != nil {
		return handlePostgresError(err, "insert error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id": b.BuildID,
	}, "Build created")

	return nil
}

func UpdateBuild(db sqlx.Execer, logger nacelle.Logger, b *Build) error {
	b.UpdatedAt = time.Now()

	resp, err := db.Exec(
		`update builds
		set
			build_status = $1,
			updated_at = $2
		where
			build_id = $3`,
		b.BuildStatus,
		b.UpdatedAt,
		b.BuildID,
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
		"build_id": b.BuildID,
	}, "Build updated")

	return nil
}
