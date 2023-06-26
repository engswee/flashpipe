package str

import (
	"fmt"
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

func Normalise(input string, normaliseAction string, normalisePrefixOrSuffix string) string {
	switch normaliseAction {
	case "ADD_PREFIX":
		return fmt.Sprintf("%v%v", normalisePrefixOrSuffix, input)
	case "ADD_SUFFIX":
		return fmt.Sprintf("%v%v", input, normalisePrefixOrSuffix)
	case "DELETE_PREFIX":
		return strings.TrimPrefix(input, normalisePrefixOrSuffix)
	case "DELETE_SUFFIX":
		return strings.TrimSuffix(input, normalisePrefixOrSuffix)
	default:
		return input
	}
}
