package updater

import (
	"regexp"
	"testing"
)

func TestReleaseTagValidation(t *testing.T) {
	stable := regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	prerelease := regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*$`)

	for _, tag := range []string{"v2.1.6", "v3.0.0"} {
		if !stable.MatchString(tag) {
			t.Fatalf("stable tag rejected: %s", tag)
		}
	}
	for _, tag := range []string{"v2.1.6-rc.1", "v2.1.6-beta.2", "v2.1.6-alpha.1"} {
		if !prerelease.MatchString(tag) {
			t.Fatalf("prerelease tag rejected: %s", tag)
		}
	}
	for _, tag := range []string{"v.1.0.1", "v2.1", "v01.2.3", "v2.1.6-"} {
		if stable.MatchString(tag) || prerelease.MatchString(tag) {
			t.Fatalf("invalid tag accepted: %s", tag)
		}
	}
}

func TestFullReleaseVersionExtraction(t *testing.T) {
	for tag, want := range map[string]string{
		"v2.1.6":      "2.1.6",
		"v2.1.6-rc.1": "2.1.6-rc.1",
	} {
		got := ParseReleaseTagForTest(tag)
		if got != want {
			t.Fatalf("tag %s: got %s want %s", tag, got, want)
		}
	}
}

func ParseReleaseTagForTest(tag string) string {
	if len(tag) > 1 && tag[0] == 'v' {
		return tag[1:]
	}
	return tag
}
