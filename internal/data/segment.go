package data

import "encoding/binary"

// Segment is the original program's data segment, which DATA.OVL holds as it is in
// memory: the tables, the names and the messages the game prints. Its addresses are the
// data segment's offsets, as the program used them; pointers inside it are such
// addresses too.
type Segment struct {
	image []byte
}

// dataSegmentFileOffset is where address 0 of the data segment lies in DATA.OVL (the
// file starts 16 bytes before the segment).
const dataSegmentFileOffset = 16

// NewSegment wraps the contents of DATA.OVL.
func NewSegment(dataOvl []byte) *Segment { return &Segment{dataOvl} }

func (s *Segment) at(address int) int { return address + dataSegmentFileOffset }

// Byte reads the byte at address.
func (s *Segment) Byte(address int) byte { return s.image[s.at(address)] }

// Word reads the 16-bit little-endian word at address.
func (s *Segment) Word(address int) uint16 {
	return binary.LittleEndian.Uint16(s.image[s.at(address):])
}

// Bytes copies count bytes from address.
func (s *Segment) Bytes(address, count int) []byte {
	out := make([]byte, count)
	copy(out, s.image[s.at(address):])
	return out
}

// Words reads count words from address.
func (s *Segment) Words(address, count int) []uint16 {
	out := make([]uint16, count)
	for i := range out {
		out[i] = s.Word(address + 2*i)
	}
	return out
}

// SignedWords reads count signed words from address.
func (s *Segment) SignedWords(address, count int) []int {
	out := make([]int, count)
	for i := range out {
		out[i] = int(int16(s.Word(address + 2*i)))
	}
	return out
}

// String reads the zero-terminated string at address.
func (s *Segment) String(address int) string {
	start := s.at(address)
	end := start
	for end < len(s.image) && s.image[end] != 0 {
		end++
	}
	return string(s.image[start:end])
}

// Strings reads count string pointers from address and the strings they point at.
func (s *Segment) Strings(address, count int) []string {
	out := make([]string, count)
	for i := range out {
		out[i] = s.String(int(s.Word(address + 2*i)))
	}
	return out
}
