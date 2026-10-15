# u5-go

Ultima V: Warriors of Destiny (Origin, 1988), the DOS EGA version, with its program
(ULTIMA.EXE and the overlays) ported to Go and [Ebiten](https://ebitengine.org). It runs on
the original game's data files; no game data is in this repository.

## Playing

Download a build for Windows, macOS or Linux from the
[releases](https://github.com/memmaker/u5-go/releases), or build it yourself:

```sh
go run ./cmd/u5 -import /path/to/your/copy -data game   # once
go run ./cmd/u5 -data game
```

`-import SRC` finds Ultima V in whatever you hand it: a plain folder, loose files, zip
files, CD-ROM images (ISO, BIN/CUE) or the GOG installer (`setup_*.exe` with its `.bin`
slices). It copies the game files into `-data` and exits.

## Extras

- **Modern layout** (the default): the map fills the window and menus answer what the game
  is asking. F6 switches to the original screen.
- **Music**: the original PC version has none. Import the Exodus Project's *Ultima V
  Upgrade* along with the game for its MIDI score, played on a SoundFont, or use
  `-import-disk2` on the Amiga version's disks (.adf, folder or zip) for the Amiga tune.
  `-music midi|amiga|off` picks one; F2 opens the settings, Alt-V mutes the music.
- `-save DIR` keeps saved games apart from the game files; `-scale N` sets the window size.

## In the browser

The WebAssembly build (`GOOS=js GOARCH=wasm go build ./cmd/u5`) runs in
[goplay](https://ruzzoli.de/goplay/) on the player's own files.

## Releases

Pushing a `v*` tag builds Windows, macOS (universal) and Linux binaries and publishes them
as a GitHub release (`.github/workflows/release.yml`).
