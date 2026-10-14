package game

import (
	"errors"

	"github.com/memmaker/u5-go/internal/data"
	"github.com/memmaker/u5-go/internal/ega"
	"github.com/memmaker/u5-go/internal/host"
)

// resourceFiles are the files the program loads whole (D_25ea); they are the ones stored
// packed.
var resourceFiles = []string{
	proportionalFont, britishFile, titleFile, tileFile, "ITEMS.16",
	"DNG1.16", "DNG2.16", "DNG3.16",
	"MON0.16", "MON1.16", "MON2.16", "MON3.16", "MON4.16", "MON5.16", "MON6.16", "MON7.16",
	"CREATE.16", textPictures,
	"STORY1.16", "STORY2.16", "STORY3.16", "STORY4.16", "STORY5.16", "STORY6.16",
	"END1.16", "END2.16", ultimaPictures, creditPictures, "ENDSC.16",
}

// SaveDir, when set, keeps the saved games apart from the game files (-save).
var SaveDir string

// Install opens the game files in dir and sets up the video driver and the tables.
func Install(dir string) error {
	// a game abandoned while drawing the cursor (in the tests) leaves the cursor still
	advanceCursor, underline, inverse, centred = true, false, false, false
	disk = data.NewDisk(dir)
	disk.SaveDir = SaveDir
	disk.SetCompressed(resourceFiles)
	dataOvl, err := disk.ReadFile("DATA.OVL")
	if err != nil {
		return err
	}
	driver, err := disk.ReadFile("T1K.DRV")
	if err != nil {
		return err
	}
	if screen, err = ega.New(driver); err != nil {
		return err
	}
	loadTables(dataOvl)
	loadClockTables()
	loadViewTables()
	loadCombatTables()
	loadFontTables()
	loadIntroTables()
	loadOutdoorTables()
	loadTownTables()
	loadTalkTables()
	loadCommandTables()
	loadSjogTables()
	loadLookTables()
	loadStatsTables()
	loadFightTables()
	loadCastTables()
	loadSpellCastTables()
	loadShopTables()
	loadTavernTables()
	loadInnTables()
	loadDungeonTables()
	loadDungeonLookTables()
	loadBlackthornTables()
	loadEndgameTables()
	loadTransferTables()
	initTextWindows()
	host.ShowFrame = screen.Frame
	return nil
}

// Run plays the game: the title screen, then the world, a settlement or a dungeon in turn.
// C: main
func Run() {
	defer func() {
		if r := recover(); r != nil && r != errQuit {
			panic(r)
		}
	}()
	keyBuffering = false
	introMain()
	followMap()
	showStatus()
	fromDungeon := false
	for {
		fromWorld := false
		if mapID == 0 {
			outdoorsMain()
			fromWorld = true
			fromDungeon = false
		}
		if mapID == 0 {
			// Leaving the world without entering anything ends the game.
			return
		}
		if mapID < firstDungeonMap {
			enterSettlement(fromWorld || fromDungeon)
			townMainLoop()
			fromDungeon = false
		} else {
			dungeonMainLoop(fromWorld)
			fromDungeon = true
		}
		contents := make([]byte, actorSlots*actorRecordSize)
		readFile(worldSaveFile(), contents, 0)
		decodeActors(contents, &actors)
		if mapID == 0 && level != 0 {
			writeFile(worldSaveFile(), contents)
		}
	}
}

// errQuit ends the game from deep inside it (Ctrl-E, Quit at the end).
var errQuit = errors.New("quit")

// exitGame leaves the game.
// C: ULTIMA_0878_RestoreVideoMode, exit
func exitGame() { panic(errQuit) }
