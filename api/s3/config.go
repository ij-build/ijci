package s3

type Config struct {
	S3BucketName string `env:"s3_bucket" default:"build-logs"`
	S3Endpoint   string `env:"s3_endpoint"`
}
