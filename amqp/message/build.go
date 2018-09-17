package message

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type BuildMessage struct {
	BuildID       uuid.UUID `json:"id"`
	RepositoryURL string    `json:"repository_url"`
	CommitBranch  string    `json:"commit_branch"`
	CommitHash    string    `json:"commit_hash"`
}

const DefaultCommitBranch = "master"

func (m *BuildMessage) Normalize() error {
	if m.CommitBranch == "" {
		m.CommitBranch = DefaultCommitBranch
	}

	if m.CommitBranch != DefaultCommitBranch && m.CommitHash != "" {
		return fmt.Errorf("commit_branch and commit_hash were both supplied")
	}

	return nil
}

func (m *BuildMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *BuildMessage) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, m)
}
