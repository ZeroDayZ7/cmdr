package annotate_origin

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/zerodayz7/cmdr/internal/annotate"
)

type Task struct {
	Path string
	Cfg  annotate.Config
}

type Result struct {
	Path string
	Err  error
}

func ProcessBatch(ctx context.Context, paths []string, cfg annotate.Config) []Result {
	numWorkers := min(len(paths), runtime.NumCPU())

	tasks := make(chan Task, len(paths))
	results := make(chan Result, len(paths))
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-tasks:
					if !ok {
						return
					}
					err := annotate.AnnotateFile(task.Path, task.Cfg, func(relPath string, style string) string {
						return fmt.Sprintf(style, relPath)
					})
					results <- Result{Path: task.Path, Err: err}
				}
			}
		}()
	}

UploadLoop:
	for _, p := range paths {
		select {
		case <-ctx.Done():
			break UploadLoop
		case tasks <- Task{Path: p, Cfg: cfg}:
		}
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	var finalResults []Result
	for r := range results {
		finalResults = append(finalResults, r)
	}

	return finalResults
}
