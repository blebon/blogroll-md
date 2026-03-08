package example

import (
	"context"
	"testing"

	"github.com/blebon/blogroll-md/internal/tasks"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	task, err := tasks.NewTask("task.yml")
	assert.NoErrorf(t, err, "unexpected error unmarshaling task yaml")
	task.Run(context.Background())
}
