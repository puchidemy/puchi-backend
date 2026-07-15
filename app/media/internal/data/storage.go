package data

import "time"

// StorageProvider defines the interface for S3-compatible storage operations.
type StorageProvider interface {
	GenerateUploadURL(objectKey, contentType string, contentLength int64) (string, error)
	GenerateDownloadURL(objectKey string, ttl time.Duration) (string, error)
	ObjectExists(objectKey string) (bool, error)
}

// MockStorage is a simple mock implementation of StorageProvider for development.
type MockStorage struct{}

// GenerateUploadURL returns a fake upload URL.
func (m *MockStorage) GenerateUploadURL(objectKey, contentType string, contentLength int64) (string, error) {
	return "http://localhost:3900/puchi-media/" + objectKey, nil
}

// GenerateDownloadURL returns a fake download URL.
func (m *MockStorage) GenerateDownloadURL(objectKey string, ttl time.Duration) (string, error) {
	return "http://localhost:3900/puchi-media/" + objectKey, nil
}

// ObjectExists always returns true for the mock.
func (m *MockStorage) ObjectExists(objectKey string) (bool, error) {
	return true, nil
}
