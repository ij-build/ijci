package message

import "encoding/json"

type BuildRequest struct {
	BuildID       string `json:"id"`
	RepositoryURL string `json:"repository_url"`
}

func (br *BuildRequest) Marshal() ([]byte, error) {
	return json.Marshal(br)
}

func (br *BuildRequest) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, br)
}
