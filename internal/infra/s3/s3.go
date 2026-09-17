// Package s3 uses seaweedfs to store files
package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Storage struct {
	Bucket string
	client *s3.Client
}

func New(ctx context.Context, awsAK string, awsSK string, bucket string) *Storage {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(awsAK, awsSK, ""),
		),
		config.WithBaseEndpoint("http://localhost:8333"),
	)
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &Storage{
		client: client,
		Bucket: bucket,
	}
}

// countingReader counts bytes from reader
type countingReader struct {
	n int64
	r io.Reader
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.n += int64(n)

	return n, err
}

func (s *Storage) Upload(ctx context.Context, r io.Reader) (addr string, n int64, err error) {
	key := uuid.NewString()

	cr := &countingReader{
		r: r,
		n: 0,
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   cr,
	}, func(o *s3.Options) {
		o.APIOptions = append(o.APIOptions,
			v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
		)
	})
	if err != nil {
		return "", 0, fmt.Errorf("error putting into bucket: %w", err)
	}

	return key, cr.n, nil
}
