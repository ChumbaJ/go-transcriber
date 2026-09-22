// Package s3 uses seaweedfs to store files
package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Storage struct {
	Bucket   string
	S3Client *s3.Client
}

func New(ctx context.Context, bucket string) *Storage {
	sdkConfig, err := config.LoadDefaultConfig(ctx,
		config.WithBaseEndpoint("http://localhost:3900"),
	)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return nil
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &Storage{
		S3Client: s3Client,
		Bucket:   bucket,
	}
}

var ErrEOF = errors.New("storage: no more data")

func (s *Storage) Upload(ctx context.Context, b []byte) (addr string, err error) {
	// The file is over, EOF
	if len(b) == 0 {
		return "", ErrEOF
	}

	key := uuid.NewString()

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(b),
	}

	_, err = s.S3Client.PutObject(ctx, input)

	if err != nil {
		fmt.Println("error putting into bucket: ", err.Error())
		return "", fmt.Errorf("error putting into bucket: %w", err)
	}

	return key, nil
}
