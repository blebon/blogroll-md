package utils_test

import (
	"testing"

	"github.com/blebon/blogroll-md/internal/utils"
	"golang.org/x/net/html"
)

func atomLinkNode(href string) *html.Node {
	return &html.Node{
		Type: html.ElementNode,
		Data: "link",
		Attr: []html.Attribute{
			{Key: "type", Val: "application/atom+xml"},
			{Key: "href", Val: href},
		},
	}
}

func TestFindFeed(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n    *html.Node
		want string
	}{
		{
			name: "returns feed href from atom link element",
			n:    atomLinkNode("https://example.com/atom.xml"),
			want: "https://example.com/atom.xml",
		},
		{
			name: "returns empty string for non-feed link element",
			n: &html.Node{
				Type: html.ElementNode,
				Data: "link",
				Attr: []html.Attribute{
					{Key: "rel", Val: "stylesheet"},
					{Key: "href", Val: "style.css"},
				},
			},
			want: "",
		},
		{
			name: "finds feed in child node",
			n: func() *html.Node {
				parent := &html.Node{Type: html.ElementNode, Data: "head"}
				parent.AppendChild(atomLinkNode("https://example.com/atom.xml"))
				return parent
			}(),
			want: "https://example.com/atom.xml",
		},
		{
			name: "returns empty string for empty document",
			n:    &html.Node{Type: html.DocumentNode},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.FindFeed(tt.n)
			if got != tt.want {
				t.Errorf("FindFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}
