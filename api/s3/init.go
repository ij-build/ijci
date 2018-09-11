package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/efritz/nacelle"
)

type Initializer struct {
	Logger    nacelle.Logger           `service:"logger"`
	Container nacelle.ServiceContainer `service:"container"`
}

const ServiceName = "s3"

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	s3Config := &Config{}
	if err := config.Load(s3Config); err != nil {
		return err
	}

	awsConfig := &aws.Config{}

	if s3Config.S3Endpoint != "" {
		awsConfig.Endpoint = aws.String(s3Config.S3Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	return i.Container.Set(ServiceName, NewClient(
		s3Config.S3BucketName,
		s3.New(session.New(awsConfig)),
	))
}
