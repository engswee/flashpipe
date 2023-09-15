package str

import (
	"strings"
)

func ExtractDelimitedValues(input string, delimiter string) []string {
	if input == "" {
		return []string{}
	} else {
		extract := strings.Split(input, delimiter)
		for i, s := range extract {
			extract[i] = strings.TrimSpace(s)
		}
		return extract
	}
}
