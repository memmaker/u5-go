# Ultima V Go port — checklist

Reference C: ~/Projects/u5-re/ref/u5d/src. All routines ported; internal/game/pending.go keeps only notPorted.

## Build
- [x] cast2.go (spell helpers, shrines, Codex, save game): compiles, wired into table loading
- [x] `go build ./... && go test ./...` green

## Modules still to port (pending.go stubs)
- [x] cast.c: Cast command, Use command
- [x] shoppes.c: armoury, stables, herbalist, guild, healer
- [x] shoppes2.c / shoppes3.c: tavern, shipwright, inn
- [x] dungeon.c: dungeon main loop, attack, klimb, drawing
- [x] dnglook.c: dungeon look, gem view, rooms
- [x] blckthrn.c: death, capture by Blackthorn
- [x] endgame.c: ending
- [x] intro.c INTRO_132a: transfer from Ultima IV

## Playability
- [x] New game from the intro reaches the overworld
- [x] Walk, talk, buy in a shop, fight, cast, save + load
- [x] Enter a dungeon and a dungeon room
- [x] Random-key fuzz run without panics
- [x] `-import` works on the GOG setup_*.exe

## Modernization (as in u4-go)
- [x] `-save DIR`: saved games apart from the game files
- [x] No direction echo for every step (kept back, printed only when more text follows; F2 Steps restores it)
- [x] Message history (F5)
- [x] Remove emulation layers: pixel framebuffer, no T1K.DRV emulation, real-time pacing, note-based sound
- [x] MIDI music of the Exodus Ultima V upgrade (player-supplied XMI, found by -import; no VGA patch exists for U5), -nomusic, Alt-V, Alt-+/-
- [x] goplay volume faders (music / sfx)
- [x] Modern layout: full-window map, toggleable overlays with the original border pieces (F6)
- [x] Wider map view by the original's terrain, NPC and line-of-sight rules
- [x] Context menus + mouse + gamepad (menus type original keys), conversation keyword menu
- [x] Menu polish: held arrows repeat, white selection, only applicable commands, clear save replies
- [ ] 3D dungeon view in the modern layout (skipped: U5's own dungeon view is already first-person)
- [x] Amiga extras on drive B (`-import-disk2`, DISK2/): Protracker theme + harp sample (tiles and pictures are identical to the PC's, so unused)
- [x] Settings menu (F2): music source (off / MIDI / Amiga) ✓, volume ✓, effects ✓, instruments (harp) ✓, steps ✓, layout ✓

## goplay
- [x] js/wasm build
- [x] Register u5 in goplay games.js (disk2, sf2) + deploy.sh + README tables
- [x] Box art (fetch-boxes.sh u5)
- [x] Deploy
