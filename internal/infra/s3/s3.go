// Package s3 uses seaweedfs to store files
package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Storage struct {
	Bucket string
	client *s3.Client
}

func New(ctx context.Context, awsAK string, awsSK string, bucket string) *Storage {
	// creds := credentials.NewStaticCredentialsProvider()

	sdkConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("garage"),
	)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return nil
	}

	s3Client := s3.NewFromConfig(sdkConfig)
	result, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Printf("Couldn't list buckets, here is why: %v \n", err.Error())
		return nil
	}

	fmt.Println("Buckets: ", result.Buckets)

	return &Storage{
		client: s3Client,
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
