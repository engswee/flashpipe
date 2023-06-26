package cmd

import (
	"github.com/engswee/flashpipe/odata"
	"testing"
)

func TestFilterInactive(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "IFlow1"},
		{Id: "Mapping1"},
		{Id: "Script"},
	}
	filtered, _ := filterArtifacts(artifacts, nil, nil)
	if len(filtered) != 3 {
		t.Fatalf("Expected number of artifacts = 3, actual = %d", len(filtered))
	}
}

func TestFilterIncludeIDs(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "IFlow1"},
		{Id: "Mapping1"},
		{Id: "Script1"},
	}
	filtered, _ := filterArtifacts(artifacts, []string{"IFlow1"}, nil)
	if len(filtered) != 1 {
		t.Fatalf("Expected number of artifacts = 1, actual = %d", len(filtered))
	}
	if filtered[0].Id != "IFlow1" {
		t.Fatalf("Expected ID for first entry = IFlow1, actual = %v", filtered[0].Id)
	}
}

func TestFilterExcludeIDs(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "IFlow1"},
		{Id: "Mapping1"},
		{Id: "Script1"},
	}
	filtered, _ := filterArtifacts(artifacts, nil, []string{"IFlow1"})
	if len(filtered) != 2 {
		t.Fatalf("Expected number of artifacts = 2, actual = %d", len(filtered))
	}
	if filtered[0].Id != "Mapping1" {
		t.Fatalf("Expected ID for first entry = Mapping1, actual = %v", filtered[0].Id)
	}
	if filtered[1].Id != "Script1" {
		t.Fatalf("Expected ID for second entry = Script1, actual = %v", filtered[1].Id)
	}
}

func TestFilterIncludeInvalidID(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "IFlow1"},
		{Id: "Mapping1"},
		{Id: "Script1"},
	}
	_, err := filterArtifacts(artifacts, []string{"IFlow2"}, nil)
	errMsg := err.Error()
	if errMsg != "Artifact IFlow2 in INCLUDE_IDS does not exist" {
		t.Fatalf("Actual error returned = %s", errMsg)
	}
}

func TestFilterExcludeInvalidID(t *testing.T) {
	artifacts := []*odata.ArtifactDetails{
		{Id: "IFlow1"},
		{Id: "Mapping1"},
		{Id: "Script1"},
	}
	_, err := filterArtifacts(artifacts, nil, []string{"IFlow2"})
	errMsg := err.Error()
	if errMsg != "Artifact IFlow2 in EXCLUDE_IDS does not exist" {
		t.Fatalf("Actual error returned = %s", errMsg)
	}
}
