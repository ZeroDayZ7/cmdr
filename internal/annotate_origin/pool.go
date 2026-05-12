package annotate

import (
	"context"
	"runtime"
	"sync"
)

type Task struct {
	Path string
	Cfg  Config
}

type Result struct {
	Path string
	Err  error
}

func ProcessBatch(ctx context.Context, paths []string, cfg Config) []Result {
	numWorkers := min(runtime.NumCPU(), len(paths))

	tasks := make(chan Task, len(paths))
	results := make(chan Result, len(paths))
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-tasks:
					if !ok {
						return
					}
					err := AnnotateFile(task.Path, task.Cfg)
					results <- Result{Path: task.Path, Err: err}
				}
			}
		})
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
