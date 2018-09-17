package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
)

type (
	Client interface {
		UpdateBuild(buildID uuid.UUID, payload *BuildPayload) error
		OpenBuildLog(buildID uuid.UUID, prefix string) (uuid.UUID, error)
		UploadBuildLog(buildID, buildLogID uuid.UUID, content string) error
	}

	BuildPayload struct {
		BuildStatus          *string    `json:"build_status,omitempty"`
		AgentAddr            *string    `json:"agent_addr,omitempty"`
		CommitBranch         *string    `json:"commit_branch,omitempty"`
		CommitHash           *string    `json:"commit_hash,omitempty"`
		CommitMessage        *string    `json:"commit_message,omitempty"`
		CommitAuthorName     *string    `json:"commit_author_name,omitempty"`
		CommitAuthorEmail    *string    `json:"commit_author_email,omitempty"`
		CommitAuthoredAt     *time.Time `json:"commit_authored_at,omitempty"`
		CommitCommitterName  *string    `json:"commit_committer_name,omitempty"`
		CommitCommitterEmail *string    `json:"commit_committer_email,omitempty"`
		CommitCommitedAt     *time.Time `json:"commit_committed_at,omitempty"`
		ErrorMessage         *string    `json:"error_message,omitempty"`
	}

	client struct {
		Logger     nacelle.Logger `service:"logger"`
		apiAddr    string
		publicAddr string
	}

	jsonBuildLog struct {
		BuildLogID uuid.UUID `json:"build_log_id"`
	}

	jsonBuildLogEnvelope struct {
		BuildLog *jsonBuildLog `json:"build_log"`
	}
)

func NewClient(apiAddr, publicAddr string) *client {
	return &client{
		Logger:     nacelle.NewNilLogger(),
		apiAddr:    apiAddr,
		publicAddr: publicAddr,
	}
}

func (c *client) UpdateBuild(buildID uuid.UUID, payload *BuildPayload) error {
	logger := c.Logger.WithFields(nacelle.LogFields{
		"build_id": buildID,
	})

	// Always update this
	payload.AgentAddr = &c.publicAddr

	logger.Info("Updating build")
	_, err := c.do("PATCH", fmt.Sprintf("/builds/%s", buildID), payload)
	logger.Info("Updated build")
	return err
}

func (c *client) OpenBuildLog(buildID uuid.UUID, prefix string) (uuid.UUID, error) {
	logger := c.Logger.WithFields(nacelle.LogFields{
		"build_id": buildID,
	})

	logger.Info("Opening build log %s", prefix)

	resp, err := c.do("POST", fmt.Sprintf("/builds/%s/logs", buildID), map[string]interface{}{
		"name": prefix,
	})

	if err != nil {
		return uuid.Nil, err
	}

	payload := &jsonBuildLog{}
	if err := json.Unmarshal(resp, &jsonBuildLogEnvelope{payload}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal response (%s)", err.Error())
	}

	logger.InfoWithFields(
		nacelle.LogFields{
			"build_log_id": payload.BuildLogID,
		},
		"Opened build log %s",
		prefix,
	)

	return payload.BuildLogID, nil
}

func (c *client) UploadBuildLog(buildID, buildLogID uuid.UUID, content string) error {
	logger := c.Logger.WithFields(nacelle.LogFields{
		"build_id":     buildID,
		"build_log_id": buildLogID,
	})

	logger.Info("Uploading build log")

	_, err := c.do("PATCH", fmt.Sprintf("/builds/%s/logs/%s", buildID, buildLogID), map[string]interface{}{
		"content": content,
	})

	logger.Info("Uploaded build log")

	return err
}

func (c *client) do(method, uri string, body interface{}) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload (%s)", err.Error())
	}

	req, err := http.NewRequest(method, c.apiAddr+uri, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to construct API request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to patch build (%s)", err.Error())
	}

	defer resp.Body.Close()

	if 200 > resp.StatusCode || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected %d status from API", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body (%s)", err.Error())
	}

	return content, nil
}
