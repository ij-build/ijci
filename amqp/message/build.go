package message

import (
	"encoding/json"

	"github.com/google/uuid"
)

type BuildMessage struct {
	BuildID       uuid.UUID `json:"id"`
	RepositoryURL string    `json:"repository_url"`
}

func (m *BuildMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *BuildMessage) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, m)
}
