package updater

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

func VerifySHA256(r io.Reader, expected string) error {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, strings.TrimSpace(expected)) {
		return fmt.Errorf("checksum mismatch: got %s, want %s", actual, strings.TrimSpace(expected))
	}
	return nil
}

func ParseChecksums(r io.Reader) (map[string]string, error) {
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid checksum line: %q", line)
		}
		sum := strings.Trim(fields[0], "*")
		name := strings.TrimSpace(strings.Join(fields[1:], " "))
		if len(sum) != 64 {
			return nil, fmt.Errorf("invalid SHA-256 checksum for %s", name)
		}
		checksums[name] = sum
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return checksums, nil
}
