package download

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

type Manager struct {
	srcUrl, destPath string
	file             *os.File
	size             int64
	cursor           int64
}

func NewManager(srcUrl, destPath string) (*Manager, error) {
	if err := os.MkdirAll(path.Dir(destPath), 0744); err != nil {
		return nil, fmt.Errorf("Failed to create directory: %s", err.Error())
	}

	info, err := headInfo(srcUrl)
	if err != nil {
		return nil, fmt.Errorf("Url does not support incremental downloading: %s", err.Error())
	}

	filePath := destPath + ".partial"
	flags := os.O_WRONLY
	cursor := int64(0)

	stat, err := os.Stat(filePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("Could not stat file: %s", err.Error())
		}
		flags = flags|os.O_CREATE|os.O_TRUNC
	} else if info.lastModified.After(stat.ModTime()) {
		flags = flags|os.O_CREATE|os.O_TRUNC
	} else {
		flags = flags|os.O_APPEND
		cursor = stat.Size()
	}

	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return nil, err
	}

	return &Manager{
		srcUrl: srcUrl,
		destPath: destPath,
		file: file,
		size: info.size,
		cursor: cursor,
	}, nil
}

type HeadInfo struct {
	lastModified time.Time
	size         int64
	acceptRanges string
}

func headInfo(url string) (*HeadInfo, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got status code: %s", resp.Status)
	}

	lastModifiedStr := resp.Header.Get("Last-Modified")
	sizeStr := resp.Header.Get("Content-Length")
	var size int64
	acceptRanges := resp.Header.Get("Accept-Ranges")

	if lastModifiedStr == "" {
		return nil, errors.New("Missing lastModifiedStr header")
	}
	if sizeStr == "" {
		return nil, errors.New("Missing Content-Length header")
	}
	if acceptRanges == "" {
		return nil, errors.New("Missing Accept-Ranges header")
	}

	lastModified, err := time.Parse(time.RFC1123, lastModifiedStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse Last-Modified header: %w", err)
	}

	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return nil, errors.New("Content-Length not number")
	}

	if acceptRanges != "bytes" {
		return nil, errors.New("Accept ranges is not 'bytes'")
	}

	return &HeadInfo{lastModified: lastModified, size: size, acceptRanges: acceptRanges}, nil
}
