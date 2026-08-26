package annotate_region

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/zerodayz7/cmdr/internal/annotate"
)

func Process(targetPath string, cfg annotate.Config, ctx context.Context) error {
	paths := make(chan string, runtime.NumCPU()*2)
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Przygotowanie lokalnej konfiguracji dla regionów
	regionCfg := Config{
		DryRun:  cfg.DryRun,
		Verbose: cfg.Verbose,
		Context: ctx,
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Go(func() {
			for path := range paths {
				select {
				case <-childCtx.Done():
					return
				default:
					// Podmiana: zamiast annotate.AnnotateFile wywołujemy AnnotateRegions
					_, err := AnnotateRegions(path, regionCfg, cfg.Profile, cfg.ProfilesConfig)

					if err != nil {
						select {
						case errChan <- fmt.Errorf("error processing %s: %w", path, err):
							cancel()
						default:
						}
						return
					}
				}
			}
		})
	}

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