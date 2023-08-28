package str

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractDelimitedValues_SingleEntry(t *testing.T) {
	output := ExtractDelimitedValues("ABC", ",")

	assert.Equal(t, 1, len(output), "Expected size = 1")
	assert.Equal(t, "ABC", output[0], "Expected entry = ABC")
}

func TestExtractDelimitedValues_MultipleEntries(t *testing.T) {
	output := ExtractDelimitedValues("ABC, 123,XYZ ", ",")

	assert.Equal(t, 3, len(output), "Expected size = 1")
	assert.Equal(t, "ABC", output[0], "Expected entry = ABC")
	assert.Equal(t, "123", output[1], "Expected entry = 123")
	assert.Equal(t, "XYZ", output[2], "Expected entry = XYZ")
}

func TestExtractDelimitedValues_NoEntry(t *testing.T) {
	output := ExtractDelimitedValues("", ",")

	assert.Equal(t, 0, len(output), "Expected size = ")
}

func TestContains_True(t *testing.T) {
	entries := []string{"ABC", "123", "XYZ"}
	output := Contains("ABC", entries)

	assert.True(t, output, "ABC not found in list")
}

func TestContains_False(t *testing.T) {
	entries := []string{"123", "XYZ"}
	output := Contains("ABC", entries)

	assert.False(t, output, "ABC found in list")
}
