package download

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Manager struct {
	srcUrl      string
	destPath    string
	partialPath string
	file        *os.File
	totalBytes  int64
	cursor      int64
}

// TotalBytes returns the total size of the file being downloaded.
func (m *Manager) TotalBytes() int64 {
	return m.totalBytes
}

// DownloadedBytes returns the number of bytes already downloaded (for resume).
func (m *Manager) DownloadedBytes() int64 {
	return m.cursor
}

// IsComplete returns true if the download is already complete.
func (m *Manager) IsComplete() bool {
	return m.cursor >= m.totalBytes
}

func NewManager(srcUrl, destPath string) (*Manager, error) {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	info, err := getRemoteInfo(srcUrl)
	if err != nil {
		return nil, fmt.Errorf("URL does not support resumable downloading: %w", err)
	}

	hash := computeHash(info.etag, srcUrl)
	partialPath := destPath + "." + hash + ".partial"

	cursor := int64(0)
	flags := os.O_WRONLY | os.O_CREATE

	stat, err := os.Stat(partialPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("could not stat partial file: %w", err)
		}
	} else {
		cursor = stat.Size()
	}

	file, err := os.OpenFile(partialPath, flags, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open partial file: %w", err)
	}

	if cursor > 0 {
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			file.Close()
			return nil, fmt.Errorf("could not seek to end of partial file: %w", err)
		}
	}

	return &Manager{
		srcUrl:      srcUrl,
		destPath:    destPath,
		partialPath: partialPath,
		file:        file,
		totalBytes:  info.size,
		cursor:      cursor,
	}, nil
}

// Download performs the HTTP download with range resume support.
// It writes data to the partial file and renames to destPath on completion.
// The provided callback is called periodically with the number of bytes downloaded so far.
func (m *Manager) Download(progressFn func(downloaded int64)) error {
	if m.IsComplete() {
		if err := m.finalize(); err != nil {
			return err
		}
		return nil
	}

	req, err := http.NewRequest("GET", m.srcUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if m.cursor > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", m.cursor))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if m.cursor > 0 && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server did not honor range request, got status: %s", resp.Status)
	}
	if m.cursor == 0 && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			written, writeErr := m.file.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}
			m.cursor += int64(written)
			if progressFn != nil {
				progressFn(m.cursor)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("failed to read response body: %w", readErr)
		}
	}

	if err := m.finalize(); err != nil {
		return err
	}

	return nil
}

func (m *Manager) finalize() error {
	if err := m.file.Close(); err != nil {
		return fmt.Errorf("failed to close partial file: %w", err)
	}
	m.file = nil

	if err := os.Rename(m.partialPath, m.destPath); err != nil {
		return fmt.Errorf("failed to rename partial file to destination: %w", err)
	}

	return nil
}

// Close closes the file handle without finalizing. Use this for cleanup on error.
func (m *Manager) Close() error {
	if m.file != nil {
		err := m.file.Close()
		m.file = nil
		return err
	}
	return nil
}

func computeHash(etag, url string) string {
	h := sha256.New()
	h.Write([]byte(etag))
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

type remoteInfo struct {
	etag string
	size int64
}

func getRemoteInfo(url string) (*remoteInfo, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code: %s", resp.Status)
	}

	etag := resp.Header.Get("ETag")
	sizeStr := resp.Header.Get("Content-Length")
	acceptRanges := resp.Header.Get("Accept-Ranges")

	if etag == "" {
		return nil, errors.New("missing ETag header")
	}
	if sizeStr == "" {
		return nil, errors.New("missing Content-Length header")
	}
	if acceptRanges != "bytes" {
		return nil, errors.New("server does not support byte range requests")
	}

	etag = strings.Trim(etag, "\"")

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Content-Length is not a valid number: %w", err)
	}

	return &remoteInfo{etag: etag, size: size}, nil
}
