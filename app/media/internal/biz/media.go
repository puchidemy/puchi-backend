package biz

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/puchidemy/puchi-backend/app/media/internal/data/sqlc/gen"
)

// Domain errors
var (
	ErrMediaNotFound      = errors.New("media not found")
	ErrInvalidCategory    = errors.New("invalid category")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrMediaTooLarge      = errors.New("media file too large")
)

// Allowed categories
var validCategories = map[string]bool{
	"avatar":       true,
	"lesson_image": true,
	"lesson_audio": true,
	"recording":    true,
}

// MediaRepo defines the repository contract for media objects.
type MediaRepo interface {
	CreateMediaObject(ctx context.Context, arg gen.CreateMediaObjectParams) (*gen.MediaObject, error)
	GetMediaObject(ctx context.Context, id int64) (*gen.MediaObject, error)
	GetMediaObjectByKey(ctx context.Context, objectKey string) (*gen.MediaObject, error)
	ListUserMedia(ctx context.Context, userID string) ([]*gen.MediaObject, error)
	UpdateMediaStatus(ctx context.Context, id int64, status string) (*gen.MediaObject, error)
	DeleteMediaObject(ctx context.Context, id int64) error
}

// StorageProvider defines the contract for external storage (MinIO/Garage).
type StorageProvider interface {
	GenerateUploadURL(objectKey, contentType string, contentLength int64) (string, error)
	GenerateDownloadURL(objectKey string, ttl time.Duration) (string, error)
	ObjectExists(objectKey string) (bool, error)
}

// MediaUsecase handles media operations.
type MediaUsecase struct {
	repo    MediaRepo
	storage StorageProvider
}

// NewMediaUsecase creates a new MediaUsecase.
func NewMediaUsecase(repo MediaRepo, storage StorageProvider) *MediaUsecase {
	return &MediaUsecase{repo: repo, storage: storage}
}

// RequestUpload generates a presigned upload URL and creates a pending media object.
func (uc *MediaUsecase) RequestUpload(ctx context.Context, userID, category, contentType string, contentLength int64) (*gen.MediaObject, string, string, int64, error) {
	if !validCategories[category] {
		return nil, "", "", 0, ErrInvalidCategory
	}

	if !isValidContentType(contentType) {
		return nil, "", "", 0, ErrInvalidContentType
	}

	objectKey := generateObjectKey(category, userID, contentType)

	obj, err := uc.repo.CreateMediaObject(ctx, gen.CreateMediaObjectParams{
		UserID:      userID,
		ObjectKey:   objectKey,
		Bucket:      "puchi-media",
		ContentType: contentType,
		Category:    category,
		SizeBytes:   contentLength,
		Status:      "uploading",
	})
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("create media object: %w", err)
	}

	uploadURL, err := uc.storage.GenerateUploadURL(objectKey, contentType, contentLength)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("generate upload url: %w", err)
	}

	ttl := int64(900) // 15 minutes default
	return obj, uploadURL, objectKey, ttl, nil
}

// FinalizeUpload marks a media object as uploaded after verifying the upload.
func (uc *MediaUsecase) FinalizeUpload(ctx context.Context, mediaID int64) (*gen.MediaObject, error) {
	obj, err := uc.repo.GetMediaObject(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMediaNotFound, err)
	}

	exists, err := uc.storage.ObjectExists(obj.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("check object exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("upload not completed: object %s not found in storage", obj.ObjectKey)
	}

	obj, err = uc.repo.UpdateMediaStatus(ctx, mediaID, "ready")
	if err != nil {
		return nil, fmt.Errorf("finalize upload: %w", err)
	}
	return obj, nil
}

// GetMedia retrieves a media object with a download URL.
func (uc *MediaUsecase) GetMedia(ctx context.Context, id int64) (*gen.MediaObject, string, error) {
	obj, err := uc.repo.GetMediaObject(ctx, id)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrMediaNotFound, err)
	}

	downloadURL, err := uc.storage.GenerateDownloadURL(obj.ObjectKey, 1*time.Hour)
	if err != nil {
		return nil, "", fmt.Errorf("generate download url: %w", err)
	}

	return obj, downloadURL, nil
}

// DeleteMedia deletes a media object.
func (uc *MediaUsecase) DeleteMedia(ctx context.Context, id int64) error {
	_, err := uc.repo.GetMediaObject(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMediaNotFound, err)
	}

	if err := uc.repo.DeleteMediaObject(ctx, id); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}
	return nil
}

// isValidContentType checks if the content type is supported.
func isValidContentType(contentType string) bool {
	for prefix := range validContentTypes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	return false
}

// Allowed content type prefixes
var validContentTypes = map[string]bool{
	"image/": true,
	"audio/": true,
	"video/": true,
}

// generateObjectKey creates an object key with pattern: {category}/{user_id}/{uuid}.{ext}
func generateObjectKey(category, userID, contentType string) string {
	ext := ".bin"
	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
		ext = exts[0]
	}
	ext = normalizeExt(ext)
	uid := uuid.New().String()
	return path.Join(category, userID, uid+ext)
}

// normalizeExt normalizes common file extensions.
func normalizeExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpeg":
		return ".jpg"
	case ".tiff":
		return ".tif"
	default:
		return ext
	}
}
