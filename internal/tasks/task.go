package tasks

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/blebon/blogroll-md/internal/persistence"
	"github.com/blebon/blogroll-md/internal/persistence/sqlite"
	"github.com/blebon/blogroll-md/internal/persistence/yaml"
	"github.com/blebon/blogroll-md/internal/types"

	yamlv3 "gopkg.in/yaml.v3"
)

type Worker struct {
	PoolSizeMultiplier int `yaml:"poolSizeMultiplier,omitempty"`
	PoolSize           int
}

type Task struct {
	Header   *types.Header    `yaml:"header"`
	Sqlite   *sqlite.Database `yaml:"sqlite"`
	Yaml     *yaml.Database   `yaml:"yaml"`
	Output   *types.Output    `yaml:"output"`
	Worker   *Worker          `yaml:"worker"`
	Blogs    types.Blogs
	Database persistence.Persistence
}

func NewTask(filename string) (*Task, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	t := &Task{
		Header: &types.Header{},
		Sqlite: &sqlite.Database{},
		Yaml:   &yaml.Database{},
		Output: &types.Output{},
		Worker: &Worker{PoolSizeMultiplier: 4},
	}
	if err := yamlv3.Unmarshal(data, t); err != nil {
		return nil, err
	}

	if t.Sqlite.Filename != "" && t.Sqlite.Tablename != "" && t.Sqlite.Input {
		t.Database = t.Sqlite
	} else if t.Yaml.Filename != "" && t.Yaml.Input {
		t.Database = t.Yaml
	} else {
		return t, fmt.Errorf("no supported database format provided")
	}

	t.Worker.PoolSizeMultiplier = max(t.Worker.PoolSizeMultiplier, 1)
	t.Worker.PoolSize = t.Worker.PoolSizeMultiplier * runtime.NumCPU()

	err = t.LoadBlogs()
	return t, err
}

func (t *Task) LoadBlogs() error {
	d, err := t.Database.Read()
	if err != nil {
		return err
	}
	t.Blogs = d
	return nil
}

func (t *Task) Run(ctx context.Context) error {
	var errs []error
	if err := t.GetBlogUpdates(ctx); err != nil {
		log.Printf("error getting blog updates: %v\n", err)
		errs = append(errs, err)
	}
	if err := t.WriteDatabase(ctx); err != nil {
		log.Printf("error writing database: %v\n", err)
		errs = append(errs, err)
	}
	if err := t.OutputMarkdown(ctx); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (t *Task) GetBlogUpdates(ctx context.Context) error {
	t.FilterInputCutoff()

	var wg sync.WaitGroup
	sem := make(chan struct{}, t.Worker.PoolSize)
	echn := make(chan error, len(t.Blogs))

	for _, b := range t.Blogs {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			echn <- b.Process(ctx)
		})
	}

	wg.Wait()
	close(echn)

	slices.SortFunc(t.Blogs, func(a, b *types.Blog) int {
		return cmp.Compare(b.Post.Updated.UnixNano(), a.Post.Updated.UnixNano())
	})

	t.FilterOutputCutoff()

	var errs []error
	for err := range echn {
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (t *Task) FilterInputCutoff() {
	cutoff := t.Database.CutOff()
	filtered := t.Blogs[:0]
	for _, b := range t.Blogs {
		if b.Post.Updated.IsZero() || !b.Post.Updated.Before(cutoff) {
			filtered = append(filtered, b)
		}
	}
	t.Blogs = filtered
}

func (t *Task) FilterOutputCutoff() {
	cutoff := t.Output.Cutoff
	filtered := t.Blogs[:0]
	for _, b := range t.Blogs {
		if !b.Post.Updated.Before(cutoff) {
			filtered = append(filtered, b)
		}
	}
	t.Blogs = filtered
}

func (t *Task) WriteDatabase(ctx context.Context) error {
	var errs []error
	if t.Sqlite.Update && t.Sqlite.Filename != "" && t.Sqlite.Tablename != "" {
		errs = append(errs, t.Sqlite.Write(t.Blogs))
	}
	if t.Yaml.Update && t.Yaml.Filename != "" {
		errs = append(errs, t.Yaml.Write(t.Blogs))
	}
	return errors.Join(errs...)
}

func (t *Task) OutputMarkdown(ctx context.Context) error {
	loc, err := time.LoadLocation(t.Output.Timezone)
	if err != nil || loc == nil {
		loc = time.UTC
	}

	funcMap := template.FuncMap{
		"formatTime": func(ts time.Time) string {
			return ts.In(loc).Format("2006-01-02 15:04:05")
		},
		"safeURL": func(u string) string {
			p, err := url.Parse(u)
			if err != nil || (p.Scheme != "http" && p.Scheme != "https") {
				return "#"
			}
			return u
		},
		"safeTitle": func(s string) string {
			s = strings.ReplaceAll(s, `\`, `\\`)
			s = strings.ReplaceAll(s, `]`, `\]`)
			return s
		},
	}

	const tmpl = `---
layout: {{ .Header.Layout }}
title: {{ .Header.Title }}
permalink: {{ .Header.Permalink }}
---
{{ range .Blogs }}
[{{ safeTitle .Title }}]({{ safeURL .Url }})  
**[{{ safeTitle .Post.Title }}]({{ safeURL .Post.Url }})**  {{ formatTime .Post.Updated }}
{{ end }}
<div style="height:0;width:0;overflow:hidden;"></div>`

	t2, err := template.New("blogroll").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(t.Output.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return t2.Execute(f, t)
}
