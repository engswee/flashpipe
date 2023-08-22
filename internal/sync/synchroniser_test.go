package sync

import (
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterInactive(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "DummyIFlow"},
		{Id: "DummyMapping"},
		{Id: "DummyScript"},
	}
	filtered, _ := filterArtifacts(artifacts, nil, nil)
	if len(filtered) != 3 {
		t.Fatalf("Expected number of artifacts = 3, actual = %d", len(filtered))
	}
}

func TestFilterIncludeIDs(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "DummyIFlow"},
		{Id: "DummyMapping"},
		{Id: "DummyScript"},
	}
	filtered, _ := filterArtifacts(artifacts, []string{"DummyIFlow"}, nil)
	assert.Equal(t, 1, len(filtered), "Expected number of artifacts = 1")
	assert.Equal(t, "DummyIFlow", filtered[0].Id, "Expected ID for first entry = DummyIFlow")
}

func TestFilterExcludeIDs(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "DummyIFlow"},
		{Id: "DummyMapping"},
		{Id: "DummyScript"},
	}
	filtered, _ := filterArtifacts(artifacts, nil, []string{"DummyIFlow"})
	assert.Equal(t, 2, len(filtered), "Expected number of artifacts = 2")
	assert.Equal(t, "DummyMapping", filtered[0].Id, "Expected ID for first entry = DummyMapping")
	assert.Equal(t, "DummyScript", filtered[1].Id, "Expected ID for second entry = DummyScript")
}

func TestFilterIncludeInvalidID(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "DummyIFlow"},
		{Id: "DummyMapping"},
		{Id: "DummyScript"},
	}
	_, err := filterArtifacts(artifacts, []string{"DummyIFlow2"}, nil)
	assert.Equal(t, "Artifact DummyIFlow2 in INCLUDE_IDS does not exist", err.Error(), "Incorrect error message")
}

func TestFilterExcludeInvalidID(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "DummyIFlow"},
		{Id: "DummyMapping"},
		{Id: "DummyScript"},
	}
	_, err := filterArtifacts(artifacts, nil, []string{"DummyIFlow2"})
	assert.Equal(t, "Artifact DummyIFlow2 in EXCLUDE_IDS does not exist", err.Error(), "Incorrect error message")
}
