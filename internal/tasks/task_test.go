package tasks_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/blebon/blogroll-md/internal/client"
	"github.com/blebon/blogroll-md/internal/persistence"
	"github.com/blebon/blogroll-md/internal/tasks"
	"github.com/blebon/blogroll-md/internal/types"
	mock "github.com/stretchr/testify/mock"
)

func TestTask_GetBlogUpdates(t *testing.T) {
	const atomFeed = `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Test Post</title><link rel="alternate" href="https://example.com/post"/><updated>2026-01-01T00:00:00Z</updated></entry></feed>`
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		filename string
		wantErr  bool
	}{
		{
			name:     "processes blogs successfully",
			filename: "testdata/task.yml",
			wantErr:  false,
		},
		{
			name:     "returns error when HTTP client fails",
			filename: "testdata/task.yml",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta, err := tasks.NewTask(tt.filename)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
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
			gotErr := ta.GetBlogUpdates(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetBlogUpdates() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetBlogUpdates() succeeded unexpectedly")
			}
		})
	}
}

func TestTask_FilterInputCutoff(t *testing.T) {
	cutoff := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		blogs         types.Blogs
		wantRemaining int
	}{
		{
			name: "keeps blog updated after cutoff",
			blogs: types.Blogs{
				{Title: "Recent", Post: types.Post{Updated: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}},
			},
			wantRemaining: 1,
		},
		{
			name: "removes blog updated before cutoff",
			blogs: types.Blogs{
				{Title: "Old", Post: types.Post{Updated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}},
			},
			wantRemaining: 0,
		},
		{
			name: "keeps blog with zero updated time",
			blogs: types.Blogs{
				{Title: "Zero", Post: types.Post{}},
			},
			wantRemaining: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := persistence.NewMockPersistence(t)
			mockDB.EXPECT().CutOff().Return(cutoff)
			ta := &tasks.Task{
				Database: mockDB,
				Blogs:    tt.blogs,
			}
			ta.FilterInputCutoff()
			if len(ta.Blogs) != tt.wantRemaining {
				t.Errorf("FilterInputCutoff() left %d blogs, want %d", len(ta.Blogs), tt.wantRemaining)
			}
		})
	}
}

func TestTask_LoadBlogs(t *testing.T) {
	blogs := types.Blogs{
		{Title: "Test Blog", Url: "https://example.com"},
	}
	tests := []struct {
		name    string
		blogs   types.Blogs
		readErr error
		wantErr bool
	}{
		{
			name:    "populates blogs on successful read",
			blogs:   blogs,
			wantErr: false,
		},
		{
			name:    "returns error when read fails",
			readErr: errors.New("disk I/O error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := persistence.NewMockPersistence(t)
			mockDB.EXPECT().Read().Return(tt.blogs, tt.readErr)
			ta := &tasks.Task{Database: mockDB}
			gotErr := ta.LoadBlogs()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("LoadBlogs() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("LoadBlogs() succeeded unexpectedly")
			}
			if len(ta.Blogs) != len(tt.blogs) {
				t.Errorf("LoadBlogs() got %d blogs, want %d", len(ta.Blogs), len(tt.blogs))
			}
		})
	}
}

func TestTask_WriteDatabase(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		filename string
		wantErr  bool
	}{
		{
			name:     "skips write when update is false",
			filename: "testdata/task.yml",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta, err := tasks.NewTask(tt.filename)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := ta.WriteDatabase(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("WriteDatabase() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("WriteDatabase() succeeded unexpectedly")
			}
		})
	}
}

func TestTask_OutputMarkdown(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		filename string
		wantErr  bool
	}{
		{
			name:     "writes markdown output file",
			filename: "testdata/task.yml",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta, err := tasks.NewTask(tt.filename)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := ta.OutputMarkdown(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("OutputMarkdown() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("OutputMarkdown() succeeded unexpectedly")
			}
		})
	}
}
