package db

import (
	"time"

	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type (
	Build struct {
		BuildID              uuid.UUID   `db:"build_id" json:"build_id"`
		ProjectID            uuid.UUID   `db:"project_id" json:"project_id"`
		BuildStatus          string      `db:"build_status" json:"build_status"`
		AgentAddr            *string     `db:"agent_addr" json:"agent_addr"`
		CommitBranch         *string     `db:"commit_branch" json:"commit_branch"`
		CommitHash           *string     `db:"commit_hash" json:"commit_hash"`
		CommitMessage        *string     `db:"commit_message" json:"commit_message"`
		CommitAuthorName     *string     `db:"commit_author_name" json:"commit_author_name"`
		CommitAuthorEmail    *string     `db:"commit_author_email" json:"commit_author_email"`
		CommitAuthoredAt     *time.Time  `db:"commit_authored_at" json:"commit_authored_at"`
		CommitCommitterName  *string     `db:"commit_committer_name" json:"commit_committer_name"`
		CommitCommitterEmail *string     `db:"commit_committer_email" json:"commit_committer_email"`
		CommitCommittedAt    *time.Time  `db:"commit_committed_at" json:"commit_committed_at"`
		ErrorMessage         *string     `db:"error_message" json:"error_message"`
		CreatedAt            time.Time   `db:"created_at" json:"created_at"`
		QueuedAt             time.Time   `db:"queued_at" json:"queued_at"`
		StartedAt            *time.Time  `db:"started_at" json:"started_at"`
		CompletedAt          *time.Time  `db:"completed_at" json:"completed_at"`
		TextSearchVector     interface{} `db:"tsv" json:"-"`
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

func GetBuilds(db *pgutil.LoggingDB, meta *pgutil.PageMeta, filter string) ([]*BuildWithProject, *pgutil.PagedResultMeta, error) {
	query := `
	select
		b.*,
		p.project_id "project.project_id",
		p.name "project.name",
		p.repository_url "project.repository_url",
		p.last_build_id "project.last_build_id",
		p.last_build_status "project.last_build_status",
		p.last_build_completed_at "project.last_build_completed_at"
	from builds b
	join projects p on b.project_id = p.project_id
	where $1 = '' or (b.tsv @@ plainto_tsquery($1))
	order by created_at desc
	`

	builds := []*BuildWithProject{}
	pageResults, err := pgutil.PagedSelect(db, meta, query, &builds, filter)
	if err != nil {
		return nil, nil, err
	}

	return builds, pageResults, nil
}

func GetQueuedBuilds(db *pgutil.LoggingDB, meta *pgutil.PageMeta, filter string) ([]*BuildWithProject, *pgutil.PagedResultMeta, error) {
	query := `
	select
		b.*,
		p.project_id "project.project_id",
		p.name "project.name",
		p.repository_url "project.repository_url",
		p.last_build_id "project.last_build_id",
		p.last_build_status "project.last_build_status",
		p.last_build_completed_at "project.last_build_completed_at"
	from builds b
	join projects p on b.project_id = p.project_id
	where build_status = 'queued' and ($1 = '' or (b.tsv @@ plainto_tsquery($1)))
	order by created_at desc
	`

	builds := []*BuildWithProject{}
	pagedResults, err := pgutil.PagedSelect(db, meta, query, &builds, filter)
	if err != nil {
		return nil, nil, err
	}

	return builds, pagedResults, nil
}

func GetActiveBuilds(db *pgutil.LoggingDB, meta *pgutil.PageMeta, filter string) ([]*BuildWithProject, *pgutil.PagedResultMeta, error) {
	query := `
	select
		b.*,
		p.project_id "project.project_id",
		p.name "project.name",
		p.repository_url "project.repository_url",
		p.last_build_id "project.last_build_id",
		p.last_build_status "project.last_build_status",
		p.last_build_completed_at "project.last_build_completed_at"
	from builds b
	join projects p on b.project_id = p.project_id
	where build_status = 'in-progress' and ($1 = '' or (b.tsv @@ plainto_tsquery($1)))
	order by created_at desc
	`

	builds := []*BuildWithProject{}
	pagedResults, err := pgutil.PagedSelect(db, meta, query, &builds, filter)
	if err != nil {
		return nil, nil, err
	}

	return builds, pagedResults, nil
}

func GetBuildsForProject(db *pgutil.LoggingDB, projectID uuid.UUID, meta *pgutil.PageMeta, filter string) ([]*Build, *pgutil.PagedResultMeta, error) {
	query := `
	select * from builds
	where project_id = $1 and ($2 = '' or (tsv @@ plainto_tsquery($2)))
	order by created_at desc
	`

	builds := []*Build{}
	pagedResults, err := pgutil.PagedSelect(db, meta, query, &builds, projectID, filter)
	if err != nil {
		return nil, nil, err
	}

	return builds, pagedResults, nil
}

func GetBuild(db *pgutil.LoggingDB, buildID uuid.UUID) (*BuildWithProject, error) {
	query := `
	select
		b.*,
		p.project_id "project.project_id",
		p.name "project.name",
		p.repository_url "project.repository_url",
		p.last_build_id "project.last_build_id",
		p.last_build_status "project.last_build_status",
		p.last_build_completed_at "project.last_build_completed_at"
	from builds b
	join projects p on b.project_id = p.project_id
	where build_id = $1
	`

	b := &BuildWithProject{}
	if err := sqlx.Get(db, b, query, buildID); err != nil {
		return nil, pgutil.HandleError(err, "select error")
	}

	return b, nil
}

func CreateBuild(db *pgutil.LoggingDB, logger nacelle.Logger, b *BuildWithProject) error {
	query := `
	insert into builds (
		build_id,
		project_id,
		build_status,
		created_at,
		queued_at
	) values ($1, $2, $3, $4, $5)
	`

	_, err := db.Exec(
		query,
		b.Build.BuildID,
		b.Project.ProjectID,
		b.BuildStatus,
		b.CreatedAt,
		b.QueuedAt,
	)

	if err != nil {
		return pgutil.HandleError(err, "insert error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id": b.BuildID,
	}, "Build created")

	return nil
}

func UpdateBuild(db *pgutil.LoggingDB, logger nacelle.Logger, b *Build) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	query := `
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
		error_message = $12,
		queued_at = $13,
		started_at = $14,
		completed_at = $15
	where
		build_id = $16
	`

	if _, err := tx.Exec(
		query,
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
		b.QueuedAt,
		b.StartedAt,
		b.CompletedAt,
		b.BuildID,
	); err != nil {
		return pgutil.HandleError(err, "update error")
	}

	if _, err := tx.Exec(
		`select update_last_build($1, null)`,
		b.ProjectID,
	); err != nil {
		return pgutil.HandleError(err, "update error")
	}

	if err := tx.Commit(); err != nil {
		return pgutil.HandleError(err, "commit error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id": b.BuildID,
	}, "Build updated")

	return nil
}

func DeleteBuild(db *pgutil.LoggingDB, logger nacelle.Logger, b *Build) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		`select update_last_build($1, $2)`,
		b.ProjectID,
		b.BuildID,
	); err != nil {
		return pgutil.HandleError(err, "delete error")
	}

	if _, err := tx.Exec(
		`delete from builds where build_id = $1`,
		b.BuildID,
	); err != nil {
		return pgutil.HandleError(err, "delete error")
	}

	if err := tx.Commit(); err != nil {
		return pgutil.HandleError(err, "commit error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"build_id": b.BuildID,
	}, "Build deleted")

	return nil
}
