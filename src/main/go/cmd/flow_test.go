package cmd

import "testing"

func TestCompareDirectorySameIgnoringOrigin(t *testing.T) {
	contentDiffer := diffDirectories("/Users/engswee/Development/Git/CPI/flashpipe/src/test/resources/test-data/DiffComparison/Dir1/", "/Users/engswee/Development/Git/CPI/flashpipe/src/test/resources/test-data/DiffComparison/Dir2/")
	if contentDiffer == true {
		t.Fatalf("Directory contents differ")
	}
}

func TestCompareDirectoryDifferent(t *testing.T) {
	contentDiffer := diffDirectories("/Users/engswee/Development/Git/CPI/flashpipe/src/test/resources/test-data/DiffComparison/Dir1/", "/Users/engswee/Development/Git/CPI/flashpipe/src/test/resources/test-data/DiffComparison/Dir3/")
	if contentDiffer == false {
		t.Fatalf("Directory contents do not differ")
	}
}
