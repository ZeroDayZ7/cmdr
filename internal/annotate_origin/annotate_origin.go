package annotate_origin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/zerodayz7/cmdr/internal/annotate" // Importujemy shared
)

func Process(targetPath string, cfg annotate.Config, ctx context.Context) error {
	paths := make(chan string, runtime.NumCPU()*2)
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Worker Pool
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				select {
				case <-childCtx.Done():
					return
				default:
					err := annotate.AnnotateFile(path, cfg, func(relPath string, style string) string {
						return fmt.Sprintf(style, relPath)
					})

					if err != nil {
						select {
						case errChan <- err:
							cancel()
						default:
						}
						return
					}
				}
			}
		}()
	}

	// Producer
	go func() {
		defer close(paths)
		err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if annotate.ShouldIgnore(path, cfg) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				return nil
			}

			ext := filepath.Ext(path)
			// POPRAWKA: annotate.IsAllowedExtension (z dużej litery)
			if annotate.IsAllowedExtension(ext, cfg) {
				select {
				case <-childCtx.Done():
					return childCtx.Err()
				case paths <- path:
				}
			}
			return nil
		})

		if err != nil && err != context.Canceled {
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	wg.Wait()
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
