package types

import "time"

type Header struct {
	Layout    string `yaml:"layout"`
	Title     string `yaml:"title"`
	Permalink string `yaml:"permalink"`
}

type Output struct {
	Filename string    `yaml:"filename"`
	Cutoff   time.Time `yaml:"cutoff"`
	Timezone string    `yaml:"timezone"`
}
