package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/puchidemy/puchi-backend/app/notification/internal/biz"
	"github.com/puchidemy/puchi-backend/app/notification/internal/conf"
)

// GotifyClient sends push notifications via the Gotify REST API.
type GotifyClient struct {
	config *conf.Gotify
	client *http.Client
}

// NewGotifyClient creates a new GotifyClient.
func NewGotifyClient(cfg *conf.Gotify) *GotifyClient {
	return &GotifyClient{
		config: cfg,
		client: &http.Client{},
	}
}

// gotifyRequestBody is the JSON payload sent to the Gotify API.
type gotifyRequestBody struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// Send implements biz.GotifySender.
func (g *GotifyClient) Send(msg biz.GotifyMessage) error {
	body, err := json.Marshal(gotifyRequestBody{
		Title:    msg.Title,
		Message:  msg.Message,
		Priority: msg.Priority,
	})
	if err != nil {
		return fmt.Errorf("gotify marshal: %w", err)
	}
	url := fmt.Sprintf("%s/message?token=%s", g.config.Url, g.config.Token)
	resp, err := g.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gotify send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("gotify error: %s", resp.Status)
	}
	return nil
}
