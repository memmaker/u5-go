package game

import "github.com/memmaker/u5-go/internal/data"

// disk is the folder with the game files.
var disk *data.Disk

// readFile fills buffer from the file name, starting offset bytes in.
// C: ULTIMA_256e_ReadFileFromDisk
func readFile(name string, buffer []byte, offset int) {
	if err := disk.ReadAt(name, buffer, offset); err != nil {
		diskError(err)
	}
}

// writeFile writes contents to the file name.
// C: ULTIMA_25d8_WriteFileToDisk
func writeFile(name string, contents []byte) {
	if err := disk.WriteFile(name, contents); err != nil {
		diskError(err)
	}
}
