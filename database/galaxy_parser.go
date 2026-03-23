package database

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
