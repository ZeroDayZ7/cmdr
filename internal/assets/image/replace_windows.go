//go:build windows

package imageutil

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func replaceFilePlatform(src, dst string, overwrite bool) error {
	srcPtr, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return fmt.Errorf("convert source path %s to UTF-16: %w", src, err)
	}
	dstPtr, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return fmt.Errorf("convert destination path %s to UTF-16: %w", dst, err)
	}
	flags := uint32(windows.MOVEFILE_WRITE_THROUGH)
	if overwrite {
		flags |= windows.MOVEFILE_REPLACE_EXISTING
	}
	return windows.MoveFileEx(srcPtr, dstPtr, flags)
}
