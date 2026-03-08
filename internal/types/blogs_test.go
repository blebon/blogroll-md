package types_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/blebon/blogroll-md/internal/client"
	"github.com/blebon/blogroll-md/internal/types"
	mock "github.com/stretchr/testify/mock"
)

func TestBlog_Process(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		wantErr bool
	}{
		{
			name:    "processes blog with valid atom feed",
			wantErr: false,
		},
		{
			name:    "returns error when HTTP client fails",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const atomFeed = `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Test Post</title><link rel="alternate" href="https://example.com/post"/><updated>2026-01-01T00:00:00Z</updated></entry></feed>`
			mockHTTP := client.NewMockHTTPClient(t)
			orig := client.Client
			t.Cleanup(func() { client.Client = orig })
			client.Client = mockHTTP
			if tt.wantErr {
				mockHTTP.EXPECT().Do(mock.AnythingOfType("*http.Request")).Return(nil, errors.New("connection refused"))
			} else {
				mockHTTP.EXPECT().Do(mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(atomFeed)),
				}, nil)
			}
			b := types.Blog{
				Title: "Test Blog",
				Url:   "https://example.com",
				Feed:  "https://example.com/feed",
			}
			gotErr := b.Process(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Process() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Process() succeeded unexpectedly")
			}
		})
	}
}
