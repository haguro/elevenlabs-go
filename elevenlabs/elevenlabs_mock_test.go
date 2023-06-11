package elevenlabs

import (
	"context"
	"time"
)

func NewMockClient(ctx context.Context, baseURL, apiKey string, reqTimeout time.Duration) *Client {
	c := NewClient(ctx, apiKey, reqTimeout)
	c.baseURL = baseURL
	return c
}

func MockDefaultClient(baseURL string) *Client {
	getDefaultClient()
	defaultClient.baseURL = baseURL
	return defaultClient
}
