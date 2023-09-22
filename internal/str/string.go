package str

import (
	"strings"
)

func ExtractDelimitedValues(input string, delimiter string) []string {
	if input == "" {
		return []string{}
	} else {
		extract := strings.Split(input, delimiter)
		return TrimSlice(extract)
	}
}

func TrimSlice(input []string) []string {
	for i, s := range input {
		input[i] = strings.TrimSpace(s)
	}
	return input
}
