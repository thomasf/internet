package internet

import (
	"os"
	"testing"
)

func TestParseCidrreport(t *testing.T) {
	file, err := os.Open("testdata/autnums.sample.txt")
	n := 0
	err = parseReport(file, func(*ASDescription) {
		n++
	})
	if err != nil {
		t.Error(err)
	}
	if n != 139 {
		t.Errorf("Expected 139 rows, got %d", n)
	}
}

func TestParseBrokenCidrreport(t *testing.T) {
	file, err := os.Open("testdata/autnums.invalid.sample.txt")
	err = parseReport(file, func(*ASDescription) {})
	if err == nil {
		t.Errorf("Parse error was expected")
	}
}
