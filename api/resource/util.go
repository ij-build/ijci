package resource

import (
	"fmt"
	"time"

	"github.com/efritz/ijci/amqp/client"
	"github.com/efritz/ijci/amqp/message"
	"github.com/efritz/ijci/api/db"
)

func queueBuild(producer *amqpclient.Producer, build *db.BuildWithProject) error {
	message := &message.BuildMessage{
		BuildID:       build.BuildID,
		RepositoryURL: build.Project.RepositoryURL,
		CommitBranch:  orString(build.CommitBranch, ""),
		CommitHash:    orString(build.CommitHash, ""),
	}

	if err := producer.Publish(message); err != nil {
		return fmt.Errorf("failed to publish message (%s)", err.Error())
	}

	return nil
}

//
// Optional Value Helpers

func orString(newVal *string, oldVal string) string {
	if newVal != nil {
		return *newVal
	}

	return oldVal
}

func orOptionalString(newVal, oldVal *string) *string {
	if newVal != nil {
		return newVal
	}

	return oldVal
}

func orOptionalTime(newVal, oldVal *time.Time) *time.Time {
	if newVal != nil {
		return newVal
	}

	return oldVal
}
