package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/alp/suumo-hunter-go/internal/models"
)

// mockS3Client is a mock implementation of S3API for testing.
type mockS3Client struct {
	getObjectFunc func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	putObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, params, optFns...)
	}
	return nil, errors.New("getObjectFunc not implemented")
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, params, optFns...)
	}
	return nil, errors.New("putObjectFunc not implemented")
}

func TestDownload(t *testing.T) {
	csvData := `id,name,address,age,floor,rent,management_fee,deposit,key_money,layout,area,walk_minutes,url
jnc_001,テストマンション,東京都渋谷区,5,3,79000,5000,1ヶ月,1ヶ月,1K,25.5,8,https://suumo.jp/chintai/jnc_001/
jnc_002,テストアパート,東京都新宿区,10,2,65000,3000,-,1ヶ月,1R,20,5,https://suumo.jp/chintai/jnc_002/
`

	mock := &mockS3Client{
		getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte(csvData))),
			}, nil
		},
	}

	storage := NewStorage(mock, "test-bucket", "properties.csv")
	ctx := context.Background()

	properties, err := storage.Download(ctx)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	if len(properties) != 2 {
		t.Fatalf("Download() returned %d properties, want 2", len(properties))
	}

	// Verify first property
	p1 := properties[0]
	if p1.ID != "jnc_001" {
		t.Errorf("Property 1 ID = %q, want %q", p1.ID, "jnc_001")
	}
	if p1.Name != "テストマンション" {
		t.Errorf("Property 1 Name = %q, want %q", p1.Name, "テストマンション")
	}
	if p1.Rent != 79000 {
		t.Errorf("Property 1 Rent = %f, want %f", p1.Rent, 79000.0)
	}
}

func TestDownloadNoSuchKey(t *testing.T) {
	mock := &mockS3Client{
		getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return nil, &types.NoSuchKey{}
		},
	}

	storage := NewStorage(mock, "test-bucket", "properties.csv")
	ctx := context.Background()

	properties, err := storage.Download(ctx)
	if err != nil {
		t.Fatalf("Download() error = %v, expected nil for NoSuchKey", err)
	}

	if len(properties) != 0 {
		t.Errorf("Download() returned %d properties, want 0 for NoSuchKey", len(properties))
	}
}

func TestDownloadError(t *testing.T) {
	mock := &mockS3Client{
		getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return nil, errors.New("access denied")
		},
	}

	storage := NewStorage(mock, "test-bucket", "properties.csv")
	ctx := context.Background()

	_, err := storage.Download(ctx)
	if err == nil {
		t.Error("Download() expected error, got nil")
	}
}

func TestUpload(t *testing.T) {
	var uploadedData []byte
	var uploadedBucket, uploadedKey string

	mock := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			uploadedBucket = *params.Bucket
			uploadedKey = *params.Key
			data, _ := io.ReadAll(params.Body)
			uploadedData = data
			return &s3.PutObjectOutput{}, nil
		},
	}

	storage := NewStorage(mock, "test-bucket", "properties.csv")
	ctx := context.Background()

	properties := []models.Property{
		{
			ID:            "jnc_001",
			Name:          "テストマンション",
			Address:       "東京都渋谷区",
			Age:           5,
			Floor:         3,
			Rent:          79000,
			ManagementFee: 5000,
			Deposit:       "1ヶ月",
			KeyMoney:      "1ヶ月",
			Layout:        "1K",
			Area:          25.5,
			WalkMinutes:   8,
			URL:           "https://suumo.jp/chintai/jnc_001/",
		},
	}

	err := storage.Upload(ctx, properties)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	// Verify upload parameters
	if uploadedBucket != "test-bucket" {
		t.Errorf("Upload bucket = %q, want %q", uploadedBucket, "test-bucket")
	}
	if uploadedKey != "properties.csv" {
		t.Errorf("Upload key = %q, want %q", uploadedKey, "properties.csv")
	}

	// Verify CSV content
	if len(uploadedData) == 0 {
		t.Error("Upload data is empty")
	}

	// Parse the uploaded CSV to verify it's valid
	parsed, err := models.LoadFromCSV(bytes.NewReader(uploadedData))
	if err != nil {
		t.Fatalf("Failed to parse uploaded CSV: %v", err)
	}
	if len(parsed) != 1 {
		t.Errorf("Parsed %d properties from uploaded CSV, want 1", len(parsed))
	}
}

func TestUploadError(t *testing.T) {
	mock := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return nil, errors.New("upload failed")
		},
	}

	storage := NewStorage(mock, "test-bucket", "properties.csv")
	ctx := context.Background()

	err := storage.Upload(ctx, []models.Property{})
	if err == nil {
		t.Error("Upload() expected error, got nil")
	}
}

func TestStorageAccessors(t *testing.T) {
	storage := NewStorage(nil, "my-bucket", "my-key.csv")

	if storage.BucketName() != "my-bucket" {
		t.Errorf("BucketName() = %q, want %q", storage.BucketName(), "my-bucket")
	}

	if storage.BucketKey() != "my-key.csv" {
		t.Errorf("BucketKey() = %q, want %q", storage.BucketKey(), "my-key.csv")
	}
}

func TestDownloadAndUploadRoundTrip(t *testing.T) {
	// Test that we can upload and download properties without data loss
	var storedData []byte

	mock := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			data, _ := io.ReadAll(params.Body)
			storedData = data
			return &s3.PutObjectOutput{}, nil
		},
		getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader(storedData)),
			}, nil
		},
	}

	storage := NewStorage(mock, "test-bucket", "properties.csv")
	ctx := context.Background()

	original := []models.Property{
		{
			ID:            "jnc_001",
			Name:          "テストマンション",
			Address:       "東京都渋谷区渋谷1-1-1",
			Age:           5,
			Floor:         3,
			Rent:          79000,
			ManagementFee: 5000,
			Deposit:       "1ヶ月",
			KeyMoney:      "1ヶ月",
			Layout:        "1K",
			Area:          25.5,
			WalkMinutes:   8,
			URL:           "https://suumo.jp/chintai/jnc_001/",
		},
		{
			ID:            "jnc_002",
			Name:          "テストアパート",
			Address:       "東京都新宿区新宿2-2-2",
			Age:           10,
			Floor:         2,
			Rent:          65000,
			ManagementFee: 3000,
			Deposit:       "-",
			KeyMoney:      "1ヶ月",
			Layout:        "1R",
			Area:          20.0,
			WalkMinutes:   5,
			URL:           "https://suumo.jp/chintai/jnc_002/",
		},
	}

	// Upload
	if err := storage.Upload(ctx, original); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	// Download
	loaded, err := storage.Download(ctx)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	// Compare
	if len(loaded) != len(original) {
		t.Fatalf("Downloaded %d properties, want %d", len(loaded), len(original))
	}

	for i := range original {
		if loaded[i].ID != original[i].ID {
			t.Errorf("Property[%d].ID = %q, want %q", i, loaded[i].ID, original[i].ID)
		}
		if loaded[i].Name != original[i].Name {
			t.Errorf("Property[%d].Name = %q, want %q", i, loaded[i].Name, original[i].Name)
		}
		if loaded[i].Rent != original[i].Rent {
			t.Errorf("Property[%d].Rent = %f, want %f", i, loaded[i].Rent, original[i].Rent)
		}
	}
}
