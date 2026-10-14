package adf

import (
	"os"
	"testing"
)

// TestDisk reads the Amiga disk named by U5_AMIGA_ADF.
func TestDisk(t *testing.T) {
	path := os.Getenv("U5_AMIGA_ADF")
	if path == "" {
		t.Skip("U5_AMIGA_ADF not set")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	files, err := Files(b)
	if err != nil {
		t.Fatal(err)
	}
	if m := files["u5music"]; len(m) < 1084 || string(m[1080:1084]) != "M.K." {
		t.Fatalf("no module among %d files", len(files))
	}
}

func TestNotADisk(t *testing.T) {
	if _, err := Files(make([]byte, diskBytes)); err == nil {
		t.Fatal("blank disk read")
	}
}
