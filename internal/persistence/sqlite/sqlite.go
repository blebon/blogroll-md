package sqlite

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/blebon/blogroll-md/internal/types"
	_ "modernc.org/sqlite"
)

var validIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (d *Database) validateTablename() error {
	if !validIdentifier.MatchString(d.Tablename) {
		return fmt.Errorf("invalid table name %q", d.Tablename)
	}
	return nil
}

type Database struct {
	Filename  string    `yaml:"filename"`
	Tablename string    `yaml:"tablename"`
	Input     bool      `yaml:"input"`
	Cutoff    time.Time `yaml:"cutoff"`
	Update    bool      `yaml:"update"`
}

func unixToTime(ts float64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(int64(ts), 0)
}

func (d *Database) Read() (types.Blogs, error) {
	if err := d.validateTablename(); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", d.Filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf(
		`SELECT nameBlog, urlBlog, titlePost, urlPost, datePost, xmlBlog FROM %s`,
		d.Tablename,
	)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs types.Blogs
	for rows.Next() {
		var (
			nameBlog, urlBlog, titlePost, urlPost, xmlBlog string
			datePost                                       float64
		)
		if err := rows.Scan(&nameBlog, &urlBlog, &titlePost, &urlPost, &datePost, &xmlBlog); err != nil {
			return nil, err
		}
		blog := &types.Blog{
			Title: nameBlog,
			Url:   urlBlog,
			Feed:  xmlBlog,
			Post: types.Post{
				Title:   titlePost,
				Url:     urlPost,
				Updated: unixToTime(datePost),
			},
		}
		blogs = append(blogs, blog)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blogs, nil
}

func (d *Database) Write(blogs types.Blogs) error {
	if err := d.validateTablename(); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", d.Filename)
	if err != nil {
		return err
	}
	defer db.Close()

	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			nameBlog  text,
			urlBlog   text,
			titlePost text,
			urlPost   text,
			datePost  float,
			xmlBlog   text,
			UNIQUE (nameBlog, urlBlog)
		)`, d.Tablename)
	if _, err := db.Exec(createSQL); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	upsertSQL := fmt.Sprintf(`
		INSERT INTO %s (nameBlog, urlBlog, titlePost, urlPost, datePost, xmlBlog)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (nameBlog, urlBlog) DO UPDATE SET
			xmlBlog   = excluded.xmlBlog,
			titlePost = excluded.titlePost,
			urlPost   = excluded.urlPost,
			datePost  = excluded.datePost`, d.Tablename)

	for _, b := range blogs {
		if _, err = tx.Exec(upsertSQL,
			b.Title,
			b.Url,
			b.Post.Title,
			b.Post.Url,
			float64(b.Post.Updated.Unix()),
			b.Feed,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *Database) CutOff() time.Time {
	return d.Cutoff
}
