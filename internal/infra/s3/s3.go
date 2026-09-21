// Package s3 uses seaweedfs to store files
package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Storage struct {
	Bucket    string
	S3Manager *transfermanager.Client
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

	s3Client := s3.NewFromConfig(sdkConfig)
	s3Manager := transfermanager.New(s3Client)

	// Just a formal check log
	fmt.Printf("Let's list up buckets for your account.\n")
	result, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Printf("S3 error: %+v\n", err)
		return nil
	}
	for _, bucket := range result.Buckets {
		fmt.Printf("\t%v\n", *bucket.Name)
	}

	return &Storage{
		S3Manager: s3Manager,
		Bucket:    bucket,
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

	_, err = s.S3Manager.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   cr,
	})
	if err != nil {
		return "", 0, fmt.Errorf("error putting into bucket: %w", err)
	}

	return key, cr.n, nil
}
