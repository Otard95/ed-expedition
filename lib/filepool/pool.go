package filepool

import (
	"ed-expedition/lib/lru"
	"os"
)

type File interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Seek(int64, int) (int64, error)
	Sync() error
	Truncate(int64) error
	Close() error
}

type FileOpener func(path string, flag int, perm os.FileMode) (File, error)

func defaultOpener(path string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(path, flag, perm)
}

type FilePool struct {
	pooledFiles []*PooledFile
	lru         *lru.LRU[string, File]
	opener      FileOpener
}

func NewFilePool(maxOpenHandles int) (*FilePool, error) {
	return NewFilePoolWithOpener(maxOpenHandles, defaultOpener)
}

func NewFilePoolWithOpener(maxOpenHandles int, opener FileOpener) (*FilePool, error) {
	lru, err := lru.NewWithEvict(maxOpenHandles, func(_ string, fd File) {
		fd.Sync()
		fd.Close()
	})
	if err != nil {
		return nil, err
	}

	return &FilePool{lru: lru, opener: opener}, nil
}

func (fp *FilePool) getFile(pf *PooledFile) (File, error) {
	if fd, ok := fp.lru.Get(pf.path); ok {
		return fd, nil
	}

	fd, err := fp.opener(pf.path, pf.flag, pf.perm)
	if err != nil {
		return nil, err
	}

	if err := fd.Truncate(pf.Size); err != nil {
		fd.Close()
		return nil, err
	}

	if _, err := fd.Seek(pf.Size, 0); err != nil {
		fd.Close()
		return nil, err
	}

	fp.lru.Set(pf.path, fd)

	return fd, nil
}

func (fp *FilePool) NewFile(path string, flag int, perm os.FileMode) *PooledFile {
	fd := &PooledFile{path: path, flag: flag, perm: perm, pool: fp}
	fp.pooledFiles = append(fp.pooledFiles, fd)
	return fd
}

func (fp *FilePool) Sync() error {
	for _, pf := range fp.pooledFiles {
		if fd, ok := fp.lru.Peek(pf.path); ok {
			if err := fd.Sync(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fp *FilePool) CloseAll() {
	for _, pf := range fp.pooledFiles {
		fp.lru.Delete(pf.path)
	}
	fp.pooledFiles = nil
}

type PooledFile struct {
	path string
	flag int
	perm os.FileMode
	pool *FilePool
	Size int64
}

func (pf *PooledFile) Truncate(size int64) {
	pf.Size = size
}

func (pf *PooledFile) Write(data []byte) (n int, err error) {
	file, err := pf.pool.getFile(pf)
	if err != nil {
		return 0, err
	}

	n, err = file.Write(data)
	pf.Size += int64(n)
	return n, err
}

func (pf *PooledFile) Read(data []byte) (n int, err error) {
	file, err := pf.pool.getFile(pf)
	if err != nil {
		return 0, err
	}

	return file.Read(data)
}
