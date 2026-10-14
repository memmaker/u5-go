// Package adf reads the files off Ultima V's Amiga disks. Their format is the game's own,
// not AmigaDOS: a directory of 32-byte entries at a fixed place (the name, NUL-padded, the
// size, and the track and sector of the first sector, big-endian), and files as chains of
// 512-byte sectors, each starting with the track and sector of the next.
package adf

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	sectorBytes     = 512
	sectorsPerTrack = 11
	tracks          = 160 // both sides of a double-density disk
	diskBytes       = tracks * sectorsPerTrack * sectorBytes
	directoryAt     = 6144
	entryBytes      = 32
	nameBytes       = 16
	linkBytes       = 4
	maxEntries      = 128
	firstPrintable  = ' '
	lastPrintable   = '~'
)

// ErrNotU5Disk is the error for an image that is no Ultima V Amiga disk.
var ErrNotU5Disk = errors.New("adf: not an Ultima V Amiga disk")

// Files reads the files of the disk image b, by their names on the disk.
func Files(b []byte) (map[string][]byte, error) {
	if len(b) != diskBytes {
		return nil, ErrNotU5Disk
	}
	files := map[string][]byte{}
	for i := range maxEntries {
		e := b[directoryAt+i*entryBytes:]
		if len(e) < entryBytes {
			break
		}
		name := string(bytes.TrimRight(e[:nameBytes], "\x00"))
		if name == "" || !printable(name) {
			break
		}
		size := int(binary.BigEndian.Uint32(e[nameBytes:]))
		track := int(binary.BigEndian.Uint16(e[nameBytes+4:]))
		sector := int(binary.BigEndian.Uint16(e[nameBytes+6:]))
		data := make([]byte, 0, size)
		for len(data) < size {
			at := (track*sectorsPerTrack + sector) * sectorBytes
			if at < 0 || at+sectorBytes > len(b) || cap(data) > diskBytes {
				return nil, ErrNotU5Disk
			}
			track = int(binary.BigEndian.Uint16(b[at:]))
			sector = int(binary.BigEndian.Uint16(b[at+2:]))
			data = append(data, b[at+linkBytes:at+sectorBytes]...)
		}
		files[name] = data[:size]
	}
	if len(files) == 0 {
		return nil, ErrNotU5Disk
	}
	return files, nil
}

func printable(s string) bool {
	for i := range len(s) {
		if s[i] < firstPrintable || s[i] > lastPrintable {
			return false
		}
	}
	return true
}
