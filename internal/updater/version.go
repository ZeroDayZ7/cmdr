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
	if apr == "" && bpr == "" {
		return 0, nil
	}

	ai := strings.Split(apr, ".")
	bi := strings.Split(bpr, ".")
	for i := 0; i < len(ai) && i < len(bi); i++ {
		if ai[i] == bi[i] {
			continue
		}
		an, aerr := strconv.Atoi(ai[i])
		bn, berr := strconv.Atoi(bi[i])
		if aerr == nil && berr == nil {
			if an < bn {
				return -1, nil
			}
			if an > bn {
				return 1, nil
			}
			continue
		}
		if aerr == nil {
			return -1, nil
		} // numeric identifiers have lower precedence than non-numeric
		if berr == nil {
			return 1, nil
		}
		if ai[i] < bi[i] {
			return -1, nil
		}
		if ai[i] > bi[i] {
			return 1, nil
		}
	}
	if len(ai) < len(bi) {
		return -1, nil
	}
	if len(ai) > len(bi) {
		return 1, nil
	}
	return 0, nil
}

func ParseReleaseTag(raw string) (string, bool, error) {
	text := strings.TrimSpace(raw)
	if text == "" || strings.HasPrefix(text, "v.") || !strings.HasPrefix(text, "v") {
		return "", false, fmt.Errorf("invalid release tag %q", raw)
	}
	version := strings.TrimPrefix(text, "v")
	_, prerelease, err := ParseVersion(version)
	if err != nil {
		return "", false, fmt.Errorf("invalid release tag %q", raw)
	}
	return version, prerelease != "", nil
}
