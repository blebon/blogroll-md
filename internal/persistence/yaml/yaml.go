package yaml

import (
	"os"
	"time"

	"github.com/blebon/blogroll-md/internal/types"
	"gopkg.in/yaml.v3"
)

type Database struct {
	Filename string    `yaml:"filename"`
	Cutoff   time.Time `yaml:"cutoff"`
	Input    bool      `yaml:"input"`
	Update   bool      `yaml:"update"`
}

func (d *Database) Read() (types.Blogs, error) {
	data, err := os.ReadFile(d.Filename)
	if err != nil {
		return nil, err
	}

	var blogs types.Blogs
	if err := yaml.Unmarshal(data, &blogs); err != nil {
		return nil, err
	}

	return blogs, nil
}

func (d *Database) Write(blogs types.Blogs) error {
	data, err := yaml.Marshal(blogs)
	if err != nil {
		return err
	}

	return os.WriteFile(d.Filename, data, 0644)
}

func (d *Database) CutOff() time.Time {
	return d.Cutoff
}
