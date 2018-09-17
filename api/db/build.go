package db

import (
	"time"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type (
	Build struct {
		BuildID              uuid.UUID  `db:"build_id" json:"build_id"`
		ProjectID            uuid.UUID  `db:"project_id" json:"project_id"`
		BuildStatus          string     `db:"build_status" json:"build_status"`
		AgentAddr            *string    `db:"agent_addr" json:"agent_addr"`
		CommitBranch         *string    `db:"commit_branch" json:"commit_branch"`
		CommitHash           *string    `db:"commit_hash" json:"commit_hash"`
		CommitMessage        *string    `db:"commit_message" json:"commit_message"`
		CommitAuthorName     *string    `db:"commit_author_name" json:"commit_author_name"`
		CommitAuthorEmail    *string    `db:"commit_author_email" json:"commit_author_email"`
		CommitAuthoredAt     *time.Time `db:"commit_authored_at" json:"commit_authored_at"`
		CommitCommitterName  *string    `db:"commit_committer_name" json:"commit_committer_name"`
		CommitCommitterEmail *string    `db:"commit_committer_email" json:"commit_committer_email"`
		CommitCommittedAt    *time.Time `db:"commit_committed_at" json:"commit_committed_at"`
		CreatedAt            time.Time  `db:"created_at" json:"created_at"`
		StartedAt            *time.Time `db:"started_at" json:"started_at"`
		CompletedAt          *time.Time `db:"completed_at" json:"completed_at"`
	}

	BuildWithProject struct {
		*Build
		Project *Project `db:"project" json:"project"`
	}

	BuildWithLogs struct {
		*BuildWithProject
		BuildLogs []*BuildLog `json:"build_logs"`
	}
)

func GetBuilds(db sqlx.Queryer) ([]*BuildWithProject, error) {
	query := `
	select
		builds.*,
		projects.project_id "project.project_id",
		projects.name "project.name",
		projects.repository_url "project.repository_url",
		projects.last_build_id "project.last_build_id",
		projects.last_build_status "project.last_build_status",
		projects.last_build_completed_at "project.last_build_completed_at"
	from builds
	join projects on builds.project_id = projects.project_id
	order by created_at desc
	`

	builds := []*BuildWithProject{}
	if err := sqlx.Select(db, &builds, query); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return builds, nil
}

func GetBuildsForProject(db sqlx.Queryer, projectID uuid.UUID) ([]*Build, error) {
	query := `select * from builds where project_id = $1 order by created_at desc`

	builds := []*Build{}
	if err := sqlx.Select(db, &builds, query, projectID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return builds, nil
}

func GetBuild(db sqlx.Queryer, buildID uuid.UUID) (*BuildWithProject, error) {
	query := `
	select
		builds.*,
		projects.project_id "project.project_id",
		projects.name "project.name",
		projects.repository_url "project.repository_url",
		projects.last_build_id "project.last_build_id",
		projects.last_build_status "project.last_build_status",
		projects.last_build_completed_at "project.last_build_completed_at"
	from builds
	join projects on builds.project_id = projects.project_id
	where build_id = $1
	`

	b := &BuildWithProject{}
	if err := sqlx.Get(db, b, query, buildID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return b, nil
}

func CreateBuild(db sqlx.Execer, logger nacelle.Logger, b *BuildWithProject) error {
	query := `
	insert into builds (
		build_id,
		project_id,
		build_status,
		created_at
	) values ($1, $2, $3, $4)
	`

	_, err := db.Exec(
		query,
		b.Build.BuildID,
		b.Project.ProjectID,
		b.BuildStatus,
		b.CreatedAt,
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
	buildsQuery := `
	update builds
	set
		build_status = $1,
		agent_addr = $2,
		commit_branch = $3,
		commit_hash = $4,
		commit_message = $5,
		commit_author_name = $6,
		commit_author_email = $7,
		commit_authored_at = $8,
		commit_committer_name = $9,
		commit_committer_email = $10,
		commit_committed_at = $11,
		started_at = $13,
		completed_at = $14
	where
		build_id = $10
	`

	projectsQuery := `
	update projects
	set
		last_build_id = $1,
		last_build_status = $2,
		last_build_completed_at = $3
	where project_id = $4
	`

	_, err := db.Exec(
		buildsQuery,
		b.BuildStatus,
		b.AgentAddr,
		b.CommitBranch,
		b.CommitHash,
		b.CommitMessage,
		b.CommitAuthorName,
		b.CommitAuthorEmail,
		b.CommitAuthoredAt,
		b.CommitCommitterName,
		b.CommitCommitterEmail,
		b.CommitCommittedAt,
		b.ErrorMessage,
		b.StartedAt,
		b.CompletedAt,
		b.BuildID,
	)

	if err != nil {
		return handlePostgresError(err, "update error")
	}

	// TODO - do in a transaction

	_, err = db.Exec(
		projectsQuery,
		b.BuildID,
		b.BuildStatus,
		b.CompletedAt,
		b.ProjectID,
	)

	if err != nil {
		return handlePostgresError(err, "update error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id": b.BuildID,
	}, "Build updated")

	return nil
}
