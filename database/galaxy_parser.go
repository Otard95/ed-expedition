package database

import (
	"bufio"
	"compress/gzip"
	"ed-expedition/lib/vec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type RawSystem struct {
	ID64     uint64   `json:"id64"`
	Name     string   `json:"name"`
	MainStar string   `json:"mainStar"`
	Coords   vec.Vec3 `json:"coords"`
}

type RawCoords struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type countingReader struct {
	r         io.Reader
	BytesRead chan int64
	count     int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.count += int64(n)
	select {
	case c.BytesRead <- c.count:
	default:
	}
	return n, err
}

func (c *countingReader) Close() {
	close(c.BytesRead)
}

type GalaxyParser struct {
	file       *os.File
	counter    *countingReader
	gzipReader *gzip.Reader
	decoder    *json.Decoder

	inArray bool
}

func NewGalaxyParser(path string) (*GalaxyParser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	counter := &countingReader{
		r:         file,
		BytesRead: make(chan int64, 1),
	}
	bufReader := bufio.NewReaderSize(counter, 1024*1024)

	gzipReader, err := gzip.NewReader(bufReader)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	decoder := json.NewDecoder(gzipReader)

	return &GalaxyParser{
		file:       file,
		counter:    counter,
		gzipReader: gzipReader,
		decoder:    decoder,
	}, nil
}

func (p *GalaxyParser) BytesRead() <-chan int64 {
	return p.counter.BytesRead
}

func (p *GalaxyParser) TotalBytes() (int64, error) {
	info, err := p.file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
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
	if p.counter != nil {
		p.counter.Close()
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
