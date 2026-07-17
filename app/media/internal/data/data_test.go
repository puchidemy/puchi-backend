package data

import (
	"testing"

	"github.com/puchidemy/puchi-backend/app/media/internal/conf"
)

func TestNewStorageProvider_EmptyEndpoint_ReturnsMock(t *testing.T) {
	provider, err := NewStorageProvider(&conf.Media{
		Storage: &conf.Media_Storage{Endpoint: ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := provider.(*MockStorage); !ok {
		t.Fatalf("expected MockStorage, got %T", provider)
	}
}

func TestNewStorageProvider_EndpointWithoutCredentials_ReturnsError(t *testing.T) {
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_SECRET_ACCESS_KEY", "")

	provider, err := NewStorageProvider(&conf.Media{
		Storage: &conf.Media_Storage{
			Endpoint: "https://example.r2.cloudflarestorage.com",
			Bucket:   "puchi-media",
		},
	})
	if err == nil {
		t.Fatal("expected error when endpoint is set without credentials")
	}
	if provider != nil {
		t.Fatalf("expected nil provider, got %T", provider)
	}
}
