package data

import (
	"os"
	"path/filepath"
	"strings"
)

// Disk is the folder with the game files. DOS file names are case-insensitive, so a name
// is looked up in any case; files the game writes keep the upper-case DOS name.
type Disk struct {
	Dir string
	// SaveDir, when set, keeps the files the game writes (the saved games): they are written
	// there, and read from there before Dir.
	SaveDir string
	// compressed are the names of the LZW-packed files (DATA.OVL's compressed-file table).
	compressed []string
}

// NewDisk opens the game folder dir.
func NewDisk(dir string) *Disk { return &Disk{Dir: dir} }

// path finds name in the folders in any case, the save folder first; a name in neither
// maps to its upper-case path in the save folder (or the game folder without one).
func (d *Disk) path(name string) string {
	if d.SaveDir != "" {
		if p, ok := find(d.SaveDir, name); ok {
			return p
		}
	}
	if p, ok := find(d.Dir, name); ok {
		return p
	}
	return filepath.Join(d.writeDir(), strings.ToUpper(name))
}

// writeDir is where written files go.
func (d *Disk) writeDir() string {
	if d.SaveDir != "" {
		return d.SaveDir
	}
	return d.Dir
}

// find looks for name in dir in any case.
func find(dir, name string) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name()), true
		}
	}
	return "", false
}

// Exists says whether the folder has the file name.
func (d *Disk) Exists(name string) bool {
	_, err := os.Stat(d.path(name))
	return err == nil
}

// ReadFile returns the whole file name as it is on disk.
func (d *Disk) ReadFile(name string) ([]byte, error) { return os.ReadFile(d.path(name)) }

// ReadAt fills buffer from the file name, starting offset bytes in. A file shorter than
// that leaves the rest of buffer as it was, as DOS's read did.
// C: ULTIMA_7234_ReadFile
func (d *Disk) ReadAt(name string, buffer []byte, offset int) error {
	contents, err := d.ReadFile(name)
	if err != nil {
		return err
	}
	if offset < len(contents) {
		copy(buffer, contents[offset:])
	}
	return nil
}

// WriteFile replaces the file name with contents.
// C: ULTIMA_7296_WriteFile
func (d *Disk) WriteFile(name string, contents []byte) error {
	const fileMode = 0o644
	p := filepath.Join(d.writeDir(), strings.ToUpper(name))
	if found, ok := find(d.writeDir(), name); ok {
		p = found
	}
	return os.WriteFile(p, contents, fileMode)
}

// SetCompressed sets the names of the files LoadResource unpacks.
func (d *Disk) SetCompressed(names []string) { d.compressed = names }

// isCompressed says whether name starts with one of the compressed names (in any case),
// the way the original matched them.
// C: ULTIMA_1588_IsFileCompressed
func (d *Disk) isCompressed(name string) bool {
	for _, packed := range d.compressed {
		if len(name) >= len(packed) && strings.EqualFold(name[:len(packed)], packed) {
			return true
		}
	}
	return false
}

// LoadResource reads a whole file, unpacking it when it is one of the compressed ones.
// C: ULTIMA_125d_LoadResourceFileImpl
func (d *Disk) LoadResource(name string) ([]byte, error) {
	contents, err := d.ReadFile(name)
	if err != nil {
		return nil, err
	}
	if d.isCompressed(name) {
		return Decompress(contents)
	}
	return contents, nil
}
