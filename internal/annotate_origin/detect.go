package annotate_origin

import (
	"bufio"
	"regexp"
	"strings"
)

var annotationRegex = regexp.MustCompile(`(?m)^(//|#|--|/\*|<!--)\s*cmdr:`)

func HasAnnotation(content string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))

	lineCount := 0

	for scanner.Scan() {
		lineCount++

		if lineCount > 10 {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		if annotationRegex.MatchString(line) {
			return true
		}
	}

	return false
}
