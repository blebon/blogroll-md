package types

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/blebon/blogroll-md/internal/utils"
	"golang.org/x/net/html"
)

type Post struct {
	Title   string    `yaml:"title"`
	Url     string    `yaml:"url"`
	Updated time.Time `yaml:"time"`
}

type Blog struct {
	Title string `yaml:"title"`
	Url   string `yaml:"url"`
	Feed  string `yaml:"feed,omitempty"`
	Post  Post   `yaml:"post,omitempty"`
}

type Blogs []*Blog

func (b *Blog) Process(ctx context.Context) error {
	if err := b.GetFeed(ctx); err != nil {
		return err
	}
	return b.GetLastPost(ctx)
}

func (b *Blog) GetFeed(ctx context.Context) error {
	if b.Feed == "" {
		resp, err := utils.Get(ctx, b.Url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, b.Url)
		}

		doc, err := html.Parse(resp.Body)
		if err != nil {
			return err
		}

		b.Feed = utils.FindFeed(doc)
		if b.Feed == "" {
			return fmt.Errorf("no feed found in %s", b.Url)
		}
		if feedURL, err := resolveReference(b.Url, b.Feed); err == nil {
			b.Feed = feedURL
		}
	}
	return nil
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type atomEntry struct {
	Title   string     `xml:"title"`
	Links   []atomLink `xml:"link"`
	Updated time.Time  `xml:"updated"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

type genericFeed struct {
	AtomEntries []atomEntry `xml:"entry"`
	RssItems    []rssItem   `xml:"channel>item"`
}

func (b *Blog) GetLastPost(ctx context.Context) error {
	resp, err := utils.Get(ctx, b.Feed)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d fetching feed %s", resp.StatusCode, b.Feed)
	}

	var feed genericFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return err
	}

	if len(feed.AtomEntries) > 0 {
		entry := feed.AtomEntries[0]
		b.Post = Post{
			Title:   entry.Title,
			Updated: entry.Updated,
			Url:     b.Url,
		}
		for _, link := range entry.Links {
			if link.Rel == "alternate" || link.Rel == "" {
				if postURL, err := resolveReference(b.Url, link.Href); err == nil {
					b.Post.Url = postURL
				} else {
					b.Post.Url = link.Href
				}
				return nil
			}
		}
		for _, link := range entry.Links {
			if link.Rel == "self" {
				b.Post.Url = link.Href
				return nil
			}
		}
		return nil
	}

	if len(feed.RssItems) > 0 {
		item := feed.RssItems[0]
		t, _ := time.Parse(time.RFC1123Z, item.PubDate)
		if t.IsZero() {
			t, _ = time.Parse(time.RFC1123, item.PubDate)
		}
		b.Post = Post{
			Title:   item.Title,
			Url:     item.Link,
			Updated: t,
		}
		return nil
	}

	return fmt.Errorf("no entries found in feed %s", b.Feed)
}

func resolveReference(base, ref string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref, err
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref, err
	}
	return baseURL.ResolveReference(refURL).String(), nil
}
