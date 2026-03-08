package utils_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/blebon/blogroll-md/internal/client"
	"github.com/blebon/blogroll-md/internal/utils"
	mock "github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		url     string
		want    *http.Response
		wantErr bool
	}{
		{
			name:    "returns response on successful request",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "returns error when client fails",
			url:     "https://example.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTP := client.NewMockHTTPClient(t)
			orig := client.Client
			t.Cleanup(func() { client.Client = orig })
			client.Client = mockHTTP
			if tt.wantErr {
				mockHTTP.EXPECT().Do(mock.AnythingOfType("*http.Request")).Return(nil, errors.New("connection refused"))
			} else {
				mockHTTP.EXPECT().Do(mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil)
			}
			got, gotErr := utils.Get(context.Background(), tt.url)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Get() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Get() succeeded unexpectedly")
			}
			if got.StatusCode != http.StatusOK {
				t.Errorf("Get() status = %v, want %v", got.StatusCode, http.StatusOK)
			}
		})
	}
}
