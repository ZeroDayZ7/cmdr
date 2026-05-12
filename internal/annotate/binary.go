package annotate

import "bytes"

func IsBinary(data []byte) bool {

	limit := min(len(data), 8192)
	return bytes.IndexByte(data[:limit], 0) != -1
}
