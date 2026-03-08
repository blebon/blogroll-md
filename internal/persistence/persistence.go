package persistence

import (
	"time"

	"github.com/blebon/blogroll-md/internal/types"
)

//mockery:generate: true
type Persistence interface {
	Read() (types.Blogs, error)
	Write(types.Blogs) error
	CutOff() time.Time
}
