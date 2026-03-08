package main

import (
	"context"
	"flag"
	"log"

	"github.com/blebon/blogroll-md/internal/tasks"
)

func main() {
	inputPtr := flag.String("input", "task.yml", "task input file")
	flag.Parse()

	t, err := tasks.NewTask(*inputPtr)
	if err != nil {
		log.Fatalf("error creating task: %v", err)
	}

	ctx := context.Background()
	err = t.Run(ctx)
	if err != nil {
		log.Fatalf("error running task: %v", err)
	}
}
