package db

import (
	"time"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type (
	Project struct {
		ProjectID            uuid.UUID  `db:"project_id" json:"project_id"`
		Name                 string     `db:"name" json:"name"`
		RepositoryURL        string     `db:"repository_url" json:"repository_url"`
		LastBuildID          *uuid.UUID `db:"last_build_id" json:"last_build_id"`
		LastBuildStatus      *string    `db:"last_build_status" json:"last_build_status"`
		LastBuildCompletedAt *time.Time `db:"last_build_completed_at" json:"last_build_completed_at"`
	}

	ProjectWithBuilds struct {
		*Project
		Builds []*Build `json:"builds"`
	}
)

func GetProjects(db *LoggingDB) ([]*Project, error) {
	query := `select * from projects order by last_build_completed_at desc, project_id`

	projects := []*Project{}
	if err := sqlx.Select(db, &projects, query); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	return projects, nil
}

func GetProject(db *LoggingDB, projectID uuid.UUID) (*ProjectWithBuilds, error) {
	query := `select * from projects where project_id = $1`

	p := &Project{}
	if err := sqlx.Get(db, p, query, projectID); err != nil {
		return nil, handlePostgresError(err, "select error")
	}

	builds, err := GetBuildsForProject(db, projectID)
	if err != nil {
		return nil, err
	}

	return &ProjectWithBuilds{Project: p, Builds: builds}, nil
}

func GetOrCreateProject(db *LoggingDB, logger nacelle.Logger, repositoryURL string) (*Project, error) {
	query := `
	insert into projects (
		project_id,
		name,
		repository_url
	) values ($1, $2, $3)
	on conflict ("repository_url") do update
		set project_id = projects.project_id
		returning projects.project_id
	`

	p := &Project{
		ProjectID:     uuid.New(),
		Name:          repositoryURL,
		RepositoryURL: repositoryURL,
	}

	var projectID uuid.UUID
	if err := sqlx.Get(db, &projectID, query, p.ProjectID, p.Name, p.RepositoryURL); err != nil {
		return nil, handlePostgresError(err, "insert error")
	}

	if projectID != p.ProjectID {
		p := &Project{}
		if err := sqlx.Get(db, p, `select * from projects where project_id = $1`, projectID); err != nil {
			return nil, handlePostgresError(err, "select error")
		}

		return p, nil
	}

	logger.InfoWithFields(nacelle.LogFields{
		"project_id": p.ProjectID,
	}, "Project created")

	return p, nil
}

func CreateProject(db *LoggingDB, logger nacelle.Logger, p *Project) error {
	query := `
	insert into projects (
		project_id,
		name,
		repository_url
	) values ($1, $2, $3)
	`

	_, err := db.Exec(
		query,
		p.ProjectID,
		p.Name,
		p.RepositoryURL,
	)

	if err != nil {
		return handlePostgresError(err, "insert error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"project_id": p.ProjectID,
	}, "Project created")

	return nil
}

func UpdateProject(db *LoggingDB, logger nacelle.Logger, p *Project) error {
	query := `
	update projects
	set
		name = $1,
		repository_url = $2
	where project_id = $3
	`

	_, err := db.Exec(
		query,
		p.Name,
		p.RepositoryURL,
		p.ProjectID,
	)

	if err != nil {
		return handlePostgresError(err, "update error")
	}

	logger.InfoWithFields(nacelle.LogFields{
		"project_id": p.ProjectID,
	}, "Project updated")

	return nil
}
