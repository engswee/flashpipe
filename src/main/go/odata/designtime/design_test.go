package designtime

import (
	"fmt"
	"github.com/engswee/flashpipe/file"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createUpdateDeployDelete(id string, name string, packageId string, dt DesigntimeArtifact, t *testing.T) {
	encoded, err := file.ZipDirToBase64(fmt.Sprintf("../../testdata/artifacts/create/%v", id))
	if err != nil {
		t.Fatalf("Error converting directory to base64 - %v", err)
	}
	err = dt.Create(id, name, packageId, encoded)
	if err != nil {
		t.Fatalf("Create failed with error - %v", err)
	}

	exists, err := dt.Exists(id, "active")
	if err != nil {
		t.Fatalf("Exists failed with error - %v", err)
	}
	assert.Equalf(t, true, exists, "Expected exists = true")

	encoded, err = file.ZipDirToBase64(fmt.Sprintf("../../testdata/artifacts/update/%v", id))
	if err != nil {
		t.Fatalf("Error converting directory to base64 - %v", err)
	}
	err = dt.Update(id, name, packageId, encoded)
	if err != nil {
		t.Fatalf("Update failed with error - %v", err)
	}

	version, err := dt.GetVersion(id, "active")
	if err != nil {
		t.Fatalf("GetVersion failed with error - %v", err)
	}
	assert.Equal(t, "1.0.1", version, "Expected version = 1.0.1")

	err = dt.Deploy(id)
	if err != nil {
		t.Fatalf("Deploy failed with error - %v", err)
	}

	content, err := dt.GetContent(id, "active")
	if err != nil {
		t.Fatalf("GetContent failed with error - %v", err)
	}
	assert.Greater(t, len(content), 0, "Expected len(content) > 0")

	err = dt.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed with error - %v", err)
	}
}
