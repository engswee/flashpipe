package str

import (
	"github.com/rs/zerolog/log"
	"slices"
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

func TrimManifestField(field string, width int) string {
	// If the length of the artifact name is longer than the allowed width of
	// MANIFEST.MF (72 characters), then it will flow over to the next line.
	// According to specification, the next line starts with a space
	// so remove it if is a space
	// Example below:
	// Bundle-Name: ALVO 1308-S Microsoft SharePoint Download Drive Item Conten
	// t
	index := width - 13
	if len(field) > index && field[index] == ' ' {
		field = field[:index] + field[index+1:]
	}
	return field
}

func FilterIDs(id string, includedIds []string, excludedIds []string) bool {
	// Filter in/out IDs
	if len(includedIds) > 0 {
		if !slices.Contains(includedIds, id) {
			log.Warn().Msgf("Skipping %v as it is not in --ids-include", id)
			return true
		}
	}
	if len(excludedIds) > 0 {
		if slices.Contains(excludedIds, id) {
			log.Warn().Msgf("Skipping %v as it is in --ids-exclude", id)
			return true
		}
	}
	return false
}
