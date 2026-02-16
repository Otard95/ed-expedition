package database

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// Hilbert curve: order 20 for ~0.1 ly precision (fits in 60 bits)
	HilbertOrder  = 20
	HilbertBits   = 60
	MaxHilbertKey = uint64(1) << HilbertBits

	// Galaxy coordinate origin and scaling
	OriginX    = -42213.8
	OriginY    = -29359.8
	OriginZ    = -23405.0
	CoordScale = 10 // 0.1 ly precision
)

// Binary file header format (8 bytes)
const (
	HeaderSize    = 8
	HeaderMagic   = "EDEX"
	HeaderVersion = 1
)

type FileType string

const (
	FileTypeSystemsBin FileType = "SB"
	FileTypeSystemsIdx FileType = "SI"
	FileTypeNamesBin   FileType = "NB"
	FileTypeNamesTrie  FileType = "NT"
)

func WriteHeader(buf []byte, fileType FileType) error {
	if len(buf) < HeaderSize {
		return errors.New("buffer too small for header")
	}
	copy(buf[0:4], HeaderMagic)
	copy(buf[4:6], fileType)
	binary.LittleEndian.PutUint16(buf[6:8], HeaderVersion)
	return nil
}

func ReadHeader(buf []byte) (FileType, uint16, error) {
	if len(buf) < HeaderSize {
		return "", 0, errors.New("buffer too small for header")
	}
	if string(buf[0:4]) != HeaderMagic {
		return "", 0, fmt.Errorf("invalid magic: expected %q, got %q", HeaderMagic, string(buf[0:4]))
	}
	fileType := FileType(buf[4:6])
	switch fileType {
	case FileTypeSystemsBin, FileTypeSystemsIdx, FileTypeNamesBin, FileTypeNamesTrie:
		// valid
	default:
		return "", 0, fmt.Errorf("invalid file type: %q", fileType)
	}
	version := binary.LittleEndian.Uint16(buf[6:8])
	return fileType, version, nil
}
