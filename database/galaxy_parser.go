package database

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type RawSystem struct {
	ID64     uint64    `json:"id64"`
	Name     string    `json:"name"`
	MainStar string    `json:"mainStar"`
	Coords   RawCoords `json:"coords"`
}

type RawCoords struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type GalaxyParser struct {
	file       *os.File
	gzipReader *gzip.Reader
	bufReader  *bufio.Reader
	decoder    *json.Decoder

	inArray   bool
	bytesRead int64
}

func NewGalaxyParser(path string) (*GalaxyParser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	bufReader := bufio.NewReaderSize(gzipReader, 256*1024)
	decoder := json.NewDecoder(bufReader)

	return &GalaxyParser{
		file:       file,
		gzipReader: gzipReader,
		bufReader:  bufReader,
		decoder:    decoder,
	}, nil
}

func (p *GalaxyParser) Close() error {
	var errs []error
	if p.gzipReader != nil {
		if err := p.gzipReader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if p.file != nil {
		if err := p.file.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (p *GalaxyParser) Next() (*RawSystem, error) {
	if !p.inArray {
		tok, err := p.decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to read opening token: %w", err)
		}
		if delim, ok := tok.(json.Delim); !ok || delim != '[' {
			return nil, fmt.Errorf("expected '[' at start of JSON, got %v", tok)
		}
		p.inArray = true
	}

	if !p.decoder.More() {
		return nil, io.EOF
	}

	var sys RawSystem
	if err := p.decoder.Decode(&sys); err != nil {
		return nil, fmt.Errorf("failed to decode system: %w", err)
	}

	return &sys, nil
}

func NormalizeCoord(x, y, z float64) (nx uint32, ny uint32, nz uint32) {
	nx = uint32((x - OriginX) * CoordScale)
	ny = uint32((y - OriginY) * CoordScale)
	nz = uint32((z - OriginZ) * CoordScale)
	return nx, ny, nz
}

var starClassMap = map[string]uint8{
	// Main Sequence
	"O (Blue-White) Star":    0x01,
	"B (Blue-White) Star":    0x02,
	"A (Blue-White) Star":    0x03,
	"F (White) Star":         0x04,
	"G (White-Yellow) Star":  0x05,
	"K (Yellow-Orange) Star": 0x06,
	"M (Red dwarf) Star":     0x07,

	// Giants & Supergiants
	"K (Yellow-Orange giant) Star":      0x10,
	"M (Red giant) Star":                0x11,
	"M (Red super giant) Star":          0x12,
	"A (Blue-White super giant) Star":   0x13,
	"B (Blue-White super giant) Star":   0x14,
	"F (White super giant) Star":        0x15,
	"G (White-Yellow super giant) Star": 0x16,

	// Brown Dwarfs
	"L (Brown dwarf) Star": 0x20,
	"T (Brown dwarf) Star": 0x21,
	"Y (Brown dwarf) Star": 0x22,

	// Carbon Stars
	"C Star":       0x30,
	"CN Star":      0x31,
	"CJ Star":      0x32,
	"MS-type Star": 0x33,
	"S-type Star":  0x34,

	// White Dwarfs
	"White Dwarf (D) Star":   0x40,
	"White Dwarf (DA) Star":  0x41,
	"White Dwarf (DAB) Star": 0x42,
	"White Dwarf (DAV) Star": 0x43,
	"White Dwarf (DAZ) Star": 0x44,
	"White Dwarf (DB) Star":  0x45,
	"White Dwarf (DBV) Star": 0x46,
	"White Dwarf (DBZ) Star": 0x47,
	"White Dwarf (DC) Star":  0x48,
	"White Dwarf (DCV) Star": 0x49,
	"White Dwarf (DQ) Star":  0x4A,

	// Wolf-Rayet
	"Wolf-Rayet Star":    0x60,
	"Wolf-Rayet C Star":  0x61,
	"Wolf-Rayet N Star":  0x62,
	"Wolf-Rayet NC Star": 0x63,
	"Wolf-Rayet O Star":  0x64,

	// Proto Stars
	"T Tauri Star":      0x70,
	"Herbig Ae/Be Star": 0x71,

	// Compact Objects
	"Neutron Star":            0x80,
	"Black Hole":              0x81,
	"Supermassive Black Hole": 0x82,
}

func ParseStarClass(starType string) uint8 {
	if starType == "" {
		return 0x00
	}
	if class, ok := starClassMap[starType]; ok {
		return class
	}
	if strings.Contains(starType, "White Dwarf") {
		return 0x40
	}
	if strings.Contains(starType, "Wolf-Rayet") {
		return 0x60
	}
	return 0x00
}
