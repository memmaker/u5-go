package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// writeU4Save writes a PARTY.SAV holding an Avatar fighter named Shamino.
func writeU4Save(t *testing.T, dir string) {
	t.Helper()
	save := make([]byte, u4VirtuesAt+u4VirtuesBytes)
	hero := save[u4HeroAt:]
	for i, v := range []uint16{u4Hp: 700, u4MaxHp: 800, u4Exp: 5000, u4Str: 45, u4Dex: 25, u4Int: 8, u4Mag: 8} {
		binary.LittleEndian.PutUint16(hero[i*2:], v)
	}
	copy(hero[u4NameAt:], "Shamino")
	hero[u4GenderAt] = maleGender
	hero[u4ClassAt] = 2 // fighter
	if err := os.WriteFile(filepath.Join(dir, u4SaveFile), save, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestTransfer converts an Ultima IV Avatar and checks the saved hero.
func TestTransfer(t *testing.T) {
	dir := copyGame(t)
	writeU4Save(t, dir)
	if err := Install(dir); err != nil {
		t.Fatal(err)
	}
	host.StartHeadless(Run)
	runTicks(100)
	toMenu(t)
	press("T")
	runTicks(50)
	snapshot(t, "transfer1")
	press(" ")
	runTicks(50)
	snapshot(t, "transfer2")
	press("Y")
	runTicks(20)
	press("Y")
	for range 6 {
		runTicks(20)
		press(" ")
	}
	runTicks(50)
	snapshot(t, "transfer3")
	contents, err := os.ReadFile(filepath.Join(dir, "SAVED.GAM"))
	if err != nil {
		t.Fatal(err)
	}
	if err := loadSave(contents); err != nil {
		t.Fatal(err)
	}
	h := party[0]
	t.Logf("%s class %c level %d exp %d str %d dex %d int %d hp %d", h.nameString(), h.class, h.level, h.exp, h.str, h.dex, h.intel, h.hp)
	if h.nameString() != "Shamino" || h.class != classAvatar || h.level != 4 || h.exp != 500 ||
		h.str != 23 || h.dex != 18 || h.intel != 8 || h.maxHP != 120 {
		t.Fatal("wrong conversion")
	}
}

// TestTransferMissing returns to the menu when there is no Ultima IV save.
func TestTransferMissing(t *testing.T) {
	startGame(t)
	runTicks(100)
	toMenu(t)
	press("T")
	runTicks(50)
	press(" ")
	toMenu(t)
}
