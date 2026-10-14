package data

import (
	"os"
	"path/filepath"
	"testing"
)

// gameDir is the original game files (set U5DATA, default the u5-re checkout's GOG copy).
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

func TestCompressedTiles(t *testing.T) {
	disk := NewDisk(gameDir(t))
	ovl, err := disk.ReadFile("DATA.OVL")
	if err != nil {
		t.Fatal(err)
	}
	const compressedFileTable, compressedFiles = 9706, 29
	names := NewSegment(ovl).Strings(compressedFileTable, compressedFiles)
	if names[3] != "TILES.16" {
		t.Fatalf("compressed file 3 is %q", names[3])
	}
	disk.SetCompressed(names)
	tiles, err := disk.LoadResource("tiles.16")
	if err != nil {
		t.Fatal(err)
	}
	const tileCount, tileBytes = 512, 128
	if len(tiles) != tileCount*tileBytes {
		t.Fatalf("TILES.16 unpacked to %d bytes", len(tiles))
	}
}
