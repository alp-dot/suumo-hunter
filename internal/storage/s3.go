// Package storage provides S3 operations for storing and retrieving property data.
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/alp/suumo-hunter-go/internal/models"
)

// S3API defines the interface for S3 operations.
// This allows for easy mocking in tests.
type S3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// Storage handles S3 operations for property data.
type Storage struct {
	client     S3API
	bucketName string
	bucketKey  string
}

// NewStorage creates a new Storage instance.
func NewStorage(client S3API, bucketName, bucketKey string) *Storage {
	return &Storage{
		client:     client,
		bucketName: bucketName,
		bucketKey:  bucketKey,
	}
}

// Download fetches the property CSV from S3 and returns the parsed properties.
// If the file doesn't exist, returns an empty slice (not an error).
func (s *Storage) Download(ctx context.Context) ([]models.Property, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s.bucketKey),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		// Check if the error is "NoSuchKey" (file doesn't exist)
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			// File doesn't exist yet, return empty slice
			return []models.Property{}, nil
		}

		// Also check for NotFound error message (some S3-compatible services)
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return []models.Property{}, nil
		}

		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the entire body
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	// Parse CSV
	properties, err := models.LoadFromCSV(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV from S3: %w", err)
	}

	return properties, nil
}

// Upload saves the properties as CSV to S3.
func (s *Storage) Upload(ctx context.Context, properties []models.Property) error {
	// Convert properties to CSV
	var buf bytes.Buffer
	if err := models.SaveToCSV(&buf, properties); err != nil {
		return fmt.Errorf("failed to convert properties to CSV: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(s.bucketKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("text/csv; charset=utf-8"),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// BucketName returns the configured bucket name.
func (s *Storage) BucketName() string {
	return s.bucketName
}

// BucketKey returns the configured bucket key.
func (s *Storage) BucketKey() string {
	return s.bucketKey
}
