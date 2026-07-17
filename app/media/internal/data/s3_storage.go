package data

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/puchidemy/puchi-backend/app/media/internal/conf"
)

var publicCategories = map[string]bool{
	"lesson_image": true,
	"lesson_audio": true,
}

// S3Storage implements biz.StorageProvider against an S3-compatible API (Cloudflare R2).
type S3Storage struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucket     string
	CDNBase    string
	presignTTL time.Duration
}

// NewS3StorageFromConfig builds an R2/S3 client from media storage config.
func NewS3StorageFromConfig(storage *conf.Media_Storage, upload *conf.Media_Upload) (*S3Storage, error) {
	if storage == nil {
		return nil, fmt.Errorf("storage config is required")
	}
	if storage.GetEndpoint() == "" {
		return nil, fmt.Errorf("storage endpoint is required")
	}
	if storage.GetBucket() == "" {
		return nil, fmt.Errorf("storage bucket is required")
	}

	accessKey, secretKey := storageCredentials(storage)
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("storage credentials are required")
	}

	region := storage.GetRegion()
	if region == "" {
		region = "auto"
	}

	endpoint := storage.GetEndpoint()
	if !storage.GetUseSsl() && !strings.HasPrefix(endpoint, "http://") {
		endpoint = "http://" + strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	ttl := 15 * time.Minute
	if upload != nil && upload.GetPresignedUrlTtl() > 0 {
		ttl = time.Duration(upload.GetPresignedUrlTtl()) * time.Second
	}

	return &S3Storage{
		client:     client,
		presigner:  s3.NewPresignClient(client),
		bucket:     storage.GetBucket(),
		CDNBase:    strings.TrimRight(storage.GetCdnBaseUrl(), "/"),
		presignTTL: ttl,
	}, nil
}

func storageCredentials(storage *conf.Media_Storage) (accessKey, secretKey string) {
	accessKey = storage.GetAccessKeyId()
	secretKey = storage.GetSecretAccessKey()
	if v := os.Getenv("R2_ACCESS_KEY_ID"); v != "" {
		accessKey = v
	}
	if v := os.Getenv("R2_SECRET_ACCESS_KEY"); v != "" {
		secretKey = v
	}
	return accessKey, secretKey
}

// PublicURL returns the CDN URL for a public object key.
func (s *S3Storage) PublicURL(objectKey string) string {
	key := strings.TrimPrefix(objectKey, "/")
	return s.CDNBase + "/" + key
}

// GenerateUploadURL returns a presigned PUT URL for direct client upload.
func (s *S3Storage) GenerateUploadURL(objectKey, contentType string, contentLength int64) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}
	req, err := s.presigner.PresignPutObject(context.Background(), input, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return "", fmt.Errorf("presign put object: %w", err)
	}
	return req.URL, nil
}

// GenerateDownloadURL returns a CDN URL for public categories or a presigned GET for private ones.
func (s *S3Storage) GenerateDownloadURL(objectKey string, ttl time.Duration) (string, error) {
	if publicCategories[categoryFromObjectKey(objectKey)] && s.CDNBase != "" {
		return s.PublicURL(objectKey), nil
	}

	if ttl <= 0 {
		ttl = time.Hour
	}
	req, err := s.presigner.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}
	return req.URL, nil
}

// ObjectExists checks whether an object is present in the bucket.
func (s *S3Storage) ObjectExists(objectKey string) (bool, error) {
	_, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err == nil {
		return true, nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey") {
		return false, nil
	}
	return false, fmt.Errorf("head object: %w", err)
}

func categoryFromObjectKey(objectKey string) string {
	if i := strings.Index(objectKey, "/"); i > 0 {
		return objectKey[:i]
	}
	return objectKey
}
