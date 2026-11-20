//go:build !windows

package imageutil

import "os"

func replaceFilePlatform(src, dst string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Lstat(dst); err == nil {
			return os.ErrExist
		}
	}
	return os.Rename(src, dst)
}
