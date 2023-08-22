package str

import "testing"

func TestSingleEntry(t *testing.T) {
	output := ExtractDelimitedValues("ABC", ",")
	size := len(output)
	if size != 1 {
		t.Fatalf("Actual size = %d -> expected = 1", size)
	}
	if output[0] != "ABC" {
		t.Fatalf("Actual entry = %v -> expected = ABC", output[0])
	}
}

func TestMultipleEntries(t *testing.T) {
	output := ExtractDelimitedValues("ABC, 123,XYZ ", ",")
	size := len(output)
	if size != 3 {
		t.Fatalf("Actual size = %d -> expected = 3", size)
	}
	if output[0] != "ABC" {
		t.Fatalf("Actual entry = %v -> expected = ABC", output[0])
	}
	if output[1] != "123" {
		t.Fatalf("Actual entry = %v -> expected = 123", output[1])
	}
	if output[2] != "XYZ" {
		t.Fatalf("Actual entry = %v -> expected = XYZ", output[3])
	}
}

func TestNoEntry(t *testing.T) {
	output := ExtractDelimitedValues("", ",")
	size := len(output)
	if size != 0 {
		t.Fatalf("Actual size = %d -> expected = 0", size)
	}
}
