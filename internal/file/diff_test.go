package file

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiffDirectories_SameIgnoringOrigin(t *testing.T) {
	contentDiffer := DiffDirectories("../../test/testdata/DiffComparison/Dir1/", "../../test/testdata/DiffComparison/Dir2/")

	assert.False(t, contentDiffer, "Directory contents differ")
}

func TestDiffDirectories_Different(t *testing.T) {
	contentDiffer := DiffDirectories("../../test/testdata/DiffComparison/Dir1/", "../../test/testdata/DiffComparison/Dir3/")

	assert.True(t, contentDiffer, "Directory contents do not differ")
}

func TestDiffFile_Different(t *testing.T) {
	fileDiffer := DiffFile("../../test/testdata/DiffComparison/Dir1/MANIFEST.MF", "../../test/testdata/DiffComparison/Dir3/MANIFEST.MF")

	assert.True(t, fileDiffer, "File contents do not differ")
}
