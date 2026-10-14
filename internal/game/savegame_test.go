package game

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func gameDir(t *testing.T) string {
	dir := os.Getenv("U5DATA")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "Projects/u5-re/gog")
	}
	if _, err := os.Stat(filepath.Join(dir, "DATA.OVL")); err != nil {
		t.Skip("no game data")
	}
	return dir
}

// TestSaveRoundTrip reads INIT.GAM and writes it back unchanged.
func TestSaveRoundTrip(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join(gameDir(t), "INIT.GAM"))
	if err != nil { // the GOG copy keeps it with its saves
		contents, err = os.ReadFile(filepath.Join(gameDir(t), "__support", "save", "INIT.GAM"))
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := loadSave(contents); err != nil {
		t.Fatal(err)
	}
	c := &saveCursor{buf: make([]byte, saveSize), writing: true}
	visitSave(c)
	if c.at != saveSize {
		t.Fatalf("the fields take %d bytes, want %d", c.at, saveSize)
	}
	if !bytes.Equal(c.buf, contents[:saveSize]) {
		t.Fatal("INIT.GAM changed on the way through")
	}
	if party[1].nameString() == "" || partySize == 0 {
		t.Errorf("companion %q, party size %d", party[1].nameString(), partySize)
	}
}
