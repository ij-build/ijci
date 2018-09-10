package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
)

type (
	Client interface {
		UpdateBuildStatus(buildID uuid.UUID, status string) error
		UploadBuildLog(buildID uuid.UUID, name, content string) error
	}

	client struct {
		Logger  nacelle.Logger `service:"logger"`
		apiAddr string
	}
)

func NewClient(apiAddr string) *client {
	return &client{
		Logger:  nacelle.NewNilLogger(),
		apiAddr: apiAddr,
	}
}

func (c *client) UpdateBuildStatus(buildID uuid.UUID, status string) error {
	return c.do("PATCH", fmt.Sprintf("/builds/%s", buildID), map[string]interface{}{
		"build_status": status,
	})
}

func (c *client) UploadBuildLog(buildID uuid.UUID, name, content string) error {
	return c.do("POST", fmt.Sprintf("/builds/%s/logs", buildID), map[string]interface{}{
		"name":    name,
		"content": content,
	})
}

func (c *client) do(method, uri string, body interface{}) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal payload (%s)", err.Error())
	}

	req, err := http.NewRequest(method, c.apiAddr+uri, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to construct API request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to patch build (%s)", err.Error())
	}

	defer resp.Body.Close()

	if 200 > resp.StatusCode || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected %d status from API", resp.StatusCode)
	}

	return nil
}
