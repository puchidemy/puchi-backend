package data

import "testing"

func TestPublicURL_UsesCDNBase(t *testing.T) {
	s := &S3Storage{CDNBase: "https://cdn.puchi.io.vn"}
	u := s.PublicURL("lesson_audio/x.mp3")
	if u != "https://cdn.puchi.io.vn/lesson_audio/x.mp3" {
		t.Fatal(u)
	}
}

func TestGenerateDownloadURL_PublicCategoryUsesCDN(t *testing.T) {
	s := &S3Storage{CDNBase: "https://cdn.puchi.io.vn"}
	u, err := s.GenerateDownloadURL("lesson_image/user/abc.jpg", 0)
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://cdn.puchi.io.vn/lesson_image/user/abc.jpg" {
		t.Fatalf("got %q", u)
	}
}
