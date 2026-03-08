package utils

import "golang.org/x/net/html"

func FindFeed(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "link" {
		attrs := map[string]string{}
		for _, a := range n.Attr {
			attrs[a.Key] = a.Val
		}
		if attrs["type"] == "application/atom+xml" || attrs["type"] == "application/rss+xml" {
			return attrs["href"]
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if feed := FindFeed(c); feed != "" {
			return feed
		}
	}
	return ""
}
