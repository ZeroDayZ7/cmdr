package updater

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseVersion(raw string) ([3]int, string, error) {
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "v")
	parts := strings.SplitN(text, "-", 2)
	core := parts[0]
	pieces := strings.Split(core, ".")
	if len(pieces) != 3 {
		return [3]int{}, "", fmt.Errorf("invalid semantic version %q", raw)
	}
	var out [3]int
	for i, piece := range pieces {
		if piece == "" || (len(piece) > 1 && piece[0] == '0') {
			return [3]int{}, "", fmt.Errorf("invalid semantic version %q", raw)
		}
		value, err := strconv.Atoi(piece)
		if err != nil {
			return [3]int{}, "", fmt.Errorf("invalid semantic version %q", raw)
		}
		out[i] = value
	}
	prerelease := ""
	if len(parts) == 2 {
		prerelease = parts[1]
		if prerelease == "" {
			return [3]int{}, "", fmt.Errorf("invalid semantic version %q", raw)
		}
	}
	return out, prerelease, nil
}

func CompareVersions(a, b string) (int, error) {
	av, apr, err := ParseVersion(a)
	if err != nil {
		return 0, err
	}
	bv, bpr, err := ParseVersion(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < 3; i++ {
		if av[i] < bv[i] {
			return -1, nil
		}
		if av[i] > bv[i] {
			return 1, nil
		}
	}
	if apr == "" && bpr != "" {
		return 1, nil
	}
	if apr != "" && bpr == "" {
		return -1, nil
	}
	if apr < bpr {
		return -1, nil
	}
	if apr > bpr {
		return 1, nil
	}
	return 0, nil
}
