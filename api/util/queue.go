package util

import (
	"fmt"

	"github.com/ij-build/ijci/amqp/client"
	"github.com/ij-build/ijci/amqp/message"
	"github.com/ij-build/ijci/api/db"
)

func QueueBuild(producer *amqpclient.Producer, build *db.BuildWithProject) error {
	message := &message.BuildMessage{
		BuildID:       build.BuildID,
		RepositoryURL: build.Project.RepositoryURL,
		CommitBranch:  OrString(build.CommitBranch, ""),
		CommitHash:    OrString(build.CommitHash, ""),
	}

	if err := producer.Publish(message); err != nil {
		return fmt.Errorf("failed to publish message (%s)", err.Error())
	}

	return nil
}
