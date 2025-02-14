package adapter

import (
	"context"

	framework "github.com/sgnl-ai/adapter-framework"
)

type Client interface {
	GetPage(ctx context.Context, request *Request) (*Response, *framework.Error)
}

type Request struct {
	BaseURL          string
	Token            string
	PageSize         int64
	EntityExternalID string
	Cursor           string
}

type Response struct {
	StatusCode       int
	RetryAfterHeader string
	Objects          []map[string]any
	NextCursor       string
}
