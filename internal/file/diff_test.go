package file

import (
	"testing"
)

func TestCompareDirectorySameIgnoringOrigin(t *testing.T) {
	contentDiffer := DiffDirectories("../../test/testdata/DiffComparison/Dir1/", "../../test/testdata/DiffComparison/Dir2/")
	if contentDiffer == true {
		t.Fatalf("Directory contents differ")
	}
}

func TestCompareDirectoryDifferent(t *testing.T) {
	contentDiffer := DiffDirectories("../../test/testdata/DiffComparison/Dir1/", "../../test/testdata/DiffComparison/Dir3/")
	if contentDiffer == false {
		t.Fatalf("Directory contents do not differ")
	}
}
