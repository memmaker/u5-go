package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveDir(t *testing.T) {
	game, save := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(game, "saved.gam"), []byte("start"), 0o644)
	d := NewDisk(game)
	d.SaveDir = save
	if b, _ := d.ReadFile("SAVED.GAM"); string(b) != "start" {
		t.Fatalf("read %q before saving", b)
	}
	if err := d.WriteFile("SAVED.GAM", []byte("saved")); err != nil {
		t.Fatal(err)
	}
	if b, _ := d.ReadFile("SAVED.GAM"); string(b) != "saved" {
		t.Fatalf("read %q after saving", b)
	}
	if b, _ := os.ReadFile(filepath.Join(game, "saved.gam")); string(b) != "start" {
		t.Fatalf("game folder changed to %q", b)
	}
}
