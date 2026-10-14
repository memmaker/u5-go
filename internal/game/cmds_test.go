package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// startInBritain starts a game from INIT.GAM standing before Britain and enters it.
func startInBritain(t *testing.T) {
	t.Helper()
	dir := copyGame(t)
	if err := Install(dir); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(dir, "INIT.GAM"))
	if err != nil {
		t.Fatal(err)
	}
	if err := loadSave(contents); err != nil {
		t.Fatal(err)
	}
	copy(party[0].name[:], "Avatar\x00")
	const britain = 2
	mapID, level = 0, 0
	mapX, mapY = placeX[britain-1], placeY[britain-1]
	if err := os.WriteFile(filepath.Join(dir, "SAVED.GAM"), encodeSave(), 0o644); err != nil {
		t.Fatal(err)
	}
	startGame2(t, dir, "J")
	runTicks(100)
	press("E")
	runTicks(200)
	if mapID != britain {
		t.Fatalf("map %d", mapID)
	}
}

// step runs keys and fails when the game ends.
func step(t *testing.T, what string, keys ...host.Key) {
	t.Helper()
	for _, k := range keys {
		host.KeyQueue = append(host.KeyQueue, k)
		if !runTicks(20) {
			t.Fatalf("the game ended during %s", what)
		}
	}
	if !runTicks(40) {
		t.Fatalf("the game ended after %s", what)
	}
}

func scan(code byte) host.Key { return host.Key{Scan: code} }

// TestCommands tries the general commands in Britain.
func TestCommands(t *testing.T) {
	startInBritain(t)
	char := func(c byte) host.Key {
		host.PushChar(c)
		k := host.KeyQueue[len(host.KeyQueue)-1]
		host.KeyQueue = host.KeyQueue[:len(host.KeyQueue)-1]
		return k
	}
	step(t, "look", char('l'), scan(host.ScanUp))
	snapshot(t, "look")
	step(t, "search", char('s'), scan(host.ScanUp), char('1'), char('\r'))
	step(t, "get", char('g'), scan(host.ScanUp))
	step(t, "open", char('o'), scan(host.ScanUp))
	step(t, "jimmy", char('j'), scan(host.ScanUp), char('1'), char('\r'))
	step(t, "push", char('p'), scan(host.ScanUp))
	step(t, "zstats", char('z'), char('1'), char('\r'))
	for range 17 {
		step(t, "zstats page", scan(host.ScanRight))
		snapshot(t, "zstats")
	}
	step(t, "zstats done", char(' '))
	step(t, "ready", char('r'), char('1'), char('\r'))
	snapshot(t, "ready")
	step(t, "ready item", char('\r'))
	step(t, "ready done", char(27))
	step(t, "new order", char('n'), char('1'), char('\r'))
	step(t, "ignite", char('i'))
	step(t, "yell", char('y'), char('h'), char('i'), char('\r'))
	snapshot(t, "cmds")
}
