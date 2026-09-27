// Package s3 uses seaweedfs to store files
package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Storage struct {
	Bucket   string
	S3Client *s3.Client
}

func New(ctx context.Context, bucket string) (*Storage, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx,
		config.WithBaseEndpoint("http://localhost:3900"),
	)
	if err != nil {
		return nil, fmt.Errorf("load s3 config: %w", err)
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &Storage{
		S3Client: s3Client,
		Bucket:   bucket,
	}, nil
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
		return "", fmt.Errorf("error putting into bucket: %w", err)
	}

	return key, nil
}

func (s *Storage) Get(ctx context.Context, addr string) ([]byte, error) {
	output, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(addr),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 getobject: %w", err)
	}

	defer output.Body.Close()

	b, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	return b, nil
}
