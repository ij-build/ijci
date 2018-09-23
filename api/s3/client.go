package s3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type (
	Client interface {
		Upload(ctx context.Context, key, content string) error
		Download(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, keys []string) error
	}

	client struct {
		bucketName string
		api        s3iface.S3API
		uploader   *s3manager.Uploader
		downloader *s3manager.Downloader
	}
)

const MaxBatchSize = 1000

func NewClient(bucketName string, api s3iface.S3API) *client {
	return &client{
		bucketName: bucketName,
		api:        api,
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

func (c *client) Delete(ctx context.Context, keys []string) error {
	for len(keys) > 0 {
		batch := keys
		if len(keys) > MaxBatchSize {
			batch = keys[:MaxBatchSize]
		}

		objects := []*s3.ObjectIdentifier{}
		for _, key := range batch {
			objects = append(objects, &s3.ObjectIdentifier{Key: aws.String(key)})
		}

		resp, err := c.api.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(c.bucketName),
			Delete: &s3.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		})

		if err != nil {
			return err
		}

		if len(resp.Errors) > 0 {
			return fmt.Errorf("failed to delete build logs (%s)", resp.Errors[0])
		}

		// Prepare next batch
		keys = keys[len(batch):]
	}

	return nil
}
