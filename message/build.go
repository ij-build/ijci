package message

import "encoding/json"

type BuildMessage struct {
	BuildID       string `json:"id"`
	RepositoryURL string `json:"repository_url"`
}

func (m *BuildMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *BuildMessage) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, m)
}
