package service

import (
	"bytes"
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
)

type S3Service interface {
	UploadFile(ctx context.Context, folder, filename string, data []byte, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
	ExtractKeyFromURL(url string) string
}

type s3Service struct {
	client     *s3.Client
	bucketName string
	region     string
}

func NewS3Service(client *s3.Client, cfg *config.Config) S3Service {
	return &s3Service{
		client:     client,
		bucketName: cfg.S3BucketName,
		region:     cfg.AWSRegion,
	}
}

func (s *s3Service) UploadFile(ctx context.Context, folder, filename string, data []byte, contentType string) (string, error) {
	key := folder + "/" + filename
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}
	url := "https://" + s.bucketName + ".s3." + s.region + ".amazonaws.com/" + key
	return url, nil
}

func (s *s3Service) DeleteFile(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}

func (s *s3Service) ExtractKeyFromURL(url string) string {
	if url == "" {
		return ""
	}
	prefix := "https://" + s.bucketName + ".s3." + s.region + ".amazonaws.com/"
	if strings.HasPrefix(url, prefix) {
		return strings.TrimPrefix(url, prefix)
	}
	return ""
}
