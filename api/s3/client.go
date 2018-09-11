package s3

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type (
	Client interface {
		Upload(ctx context.Context, key, content string) error
		Download(ctx context.Context, key string) (string, error)
	}

	client struct {
		bucketName string
		uploader   *s3manager.Uploader
		downloader *s3manager.Downloader
	}
)

func NewClient(bucketName string, api s3iface.S3API) *client {
	return &client{
		bucketName: bucketName,
		uploader:   s3manager.NewUploaderWithClient(api),
		downloader: s3manager.NewDownloaderWithClient(api),
	}
}

func (c *client) Upload(ctx context.Context, key, content string) error {
	_, err := c.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(content)),
	})

	return err
}

func (c *client) Download(ctx context.Context, key string) (string, error) {
	buffer := &aws.WriteAtBuffer{}

	_, err := c.downloader.DownloadWithContext(ctx, buffer, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", err
	}

	return string(buffer.Bytes()), nil
}
