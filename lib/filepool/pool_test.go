package filepool

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFile struct {
	calls []string
}

func (m *mockFile) Write(data []byte) (int, error) {
	m.calls = append(m.calls, fmt.Sprintf("Write:%d", len(data)))
	return len(data), nil
}

func (m *mockFile) Read(data []byte) (int, error) {
	m.calls = append(m.calls, fmt.Sprintf("Read:%d", len(data)))
	return 0, nil
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	m.calls = append(m.calls, fmt.Sprintf("Seek:%d:%d", offset, whence))
	return offset, nil
}

func (m *mockFile) Sync() error {
	m.calls = append(m.calls, "Sync")
	return nil
}

func (m *mockFile) Truncate(size int64) error {
	m.calls = append(m.calls, fmt.Sprintf("Truncate:%d", size))
	return nil
}

func (m *mockFile) Close() error {
	m.calls = append(m.calls, "Close")
	return nil
}

func (m *mockFile) hasCalls(expected ...string) bool {
	for _, e := range expected {
		found := false
		for _, c := range m.calls {
			if c == e {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestNewFilePool(t *testing.T) {
	pool, err := NewFilePool(10)
	require.NoError(t, err)
	assert.NotNil(t, pool)
}

func TestNewFileReturnsPooledFile(t *testing.T) {
	pool, _ := NewFilePool(10)

	pf := pool.NewFile("/tmp/test.bin", os.O_CREATE|os.O_RDWR, 0644)

	assert.NotNil(t, pf)
	assert.Equal(t, "/tmp/test.bin", pf.path)
}

func TestWriteOpensFileAndWritesData(t *testing.T) {
	var opened []*mockFile
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		f := &mockFile{}
		opened = append(opened, f)
		return f, nil
	}

	pool, _ := NewFilePoolWithOpener(10, opener)
	pf := pool.NewFile("test.bin", os.O_CREATE|os.O_RDWR, 0644)

	n, err := pf.Write([]byte("hello"))

	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Len(t, opened, 1)
	assert.True(t, opened[0].hasCalls("Write:5"))
}

func TestSizeTracking(t *testing.T) {
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		return &mockFile{}, nil
	}

	pool, _ := NewFilePoolWithOpener(10, opener)
	pf := pool.NewFile("test.bin", os.O_CREATE|os.O_RDWR, 0644)

	assert.Equal(t, int64(0), pf.Size)

	pf.Write([]byte("hello"))
	assert.Equal(t, int64(5), pf.Size)

	pf.Write([]byte(" world"))
	assert.Equal(t, int64(11), pf.Size)
}

func TestTruncateSetsSize(t *testing.T) {
	pool, _ := NewFilePool(10)
	pf := pool.NewFile("test.bin", os.O_CREATE|os.O_RDWR, 0644)

	pf.Truncate(100)

	assert.Equal(t, int64(100), pf.Size)
}

func TestTruncateCalledOnOpen(t *testing.T) {
	var opened []*mockFile
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		f := &mockFile{}
		opened = append(opened, f)
		return f, nil
	}

	pool, _ := NewFilePoolWithOpener(10, opener)
	pf := pool.NewFile("test.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf.Truncate(50)

	pf.Write([]byte("x"))

	assert.True(t, opened[0].hasCalls("Truncate:50"))
}

func TestEvictionClosesFile(t *testing.T) {
	var opened []*mockFile
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		f := &mockFile{}
		opened = append(opened, f)
		return f, nil
	}

	pool, _ := NewFilePoolWithOpener(2, opener)

	pf1 := pool.NewFile("1.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf2 := pool.NewFile("2.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf3 := pool.NewFile("3.bin", os.O_CREATE|os.O_RDWR, 0644)

	pf1.Write([]byte("one"))
	pf2.Write([]byte("two"))
	pf3.Write([]byte("three")) // should evict pf1

	assert.Len(t, opened, 3)
	assert.True(t, opened[0].hasCalls("Close"), "expected first file to be closed after eviction")
	assert.False(t, opened[1].hasCalls("Close"))
	assert.False(t, opened[2].hasCalls("Close"))
}

func TestEvictionSyncsBeforeClose(t *testing.T) {
	var opened []*mockFile
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		f := &mockFile{}
		opened = append(opened, f)
		return f, nil
	}

	pool, _ := NewFilePoolWithOpener(2, opener)

	pf1 := pool.NewFile("1.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf2 := pool.NewFile("2.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf3 := pool.NewFile("3.bin", os.O_CREATE|os.O_RDWR, 0644)

	pf1.Write([]byte("one"))
	pf2.Write([]byte("two"))
	pf3.Write([]byte("three")) // should evict pf1

	assert.True(t, opened[0].hasCalls("Sync", "Close"), "expected first file to be synced and closed")
}

func TestReOpenAfterEviction(t *testing.T) {
	openCount := make(map[string]int)
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		openCount[path]++
		return &mockFile{}, nil
	}

	pool, _ := NewFilePoolWithOpener(2, opener)

	pf1 := pool.NewFile("1.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf2 := pool.NewFile("2.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf3 := pool.NewFile("3.bin", os.O_CREATE|os.O_RDWR, 0644)

	pf1.Write([]byte("one"))   // opens 1.bin
	pf2.Write([]byte("two"))   // opens 2.bin
	pf3.Write([]byte("three")) // opens 3.bin, evicts 1.bin
	pf1.Write([]byte("again")) // re-opens 1.bin

	assert.Equal(t, 2, openCount["1.bin"], "expected 1.bin to be opened twice")
}

func TestSync(t *testing.T) {
	var opened []*mockFile
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		f := &mockFile{}
		opened = append(opened, f)
		return f, nil
	}

	pool, _ := NewFilePoolWithOpener(10, opener)

	pf1 := pool.NewFile("1.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf2 := pool.NewFile("2.bin", os.O_CREATE|os.O_RDWR, 0644)

	pf1.Write([]byte("one"))
	pf2.Write([]byte("two"))

	err := pool.Sync()

	require.NoError(t, err)
	assert.True(t, opened[0].hasCalls("Sync"))
	assert.True(t, opened[1].hasCalls("Sync"))
}

func TestCloseAll(t *testing.T) {
	var opened []*mockFile
	opener := func(path string, flag int, perm os.FileMode) (File, error) {
		f := &mockFile{}
		opened = append(opened, f)
		return f, nil
	}

	pool, _ := NewFilePoolWithOpener(10, opener)

	pf1 := pool.NewFile("1.bin", os.O_CREATE|os.O_RDWR, 0644)
	pf2 := pool.NewFile("2.bin", os.O_CREATE|os.O_RDWR, 0644)

	pf1.Write([]byte("one"))
	pf2.Write([]byte("two"))

	pool.CloseAll()

	assert.True(t, opened[0].hasCalls("Close"))
	assert.True(t, opened[1].hasCalls("Close"))
}

func TestIntegrationWithRealFiles(t *testing.T) {
	dir := t.TempDir()
	pool, _ := NewFilePool(2)
	defer pool.CloseAll()

	path := filepath.Join(dir, "test.bin")
	pf := pool.NewFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	pf.Write([]byte("hello"))
	pf.Write([]byte(" world"))

	pool.Sync()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(data))
}

func TestIntegrationTruncateOnResume(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")

	os.WriteFile(path, []byte("initial data that should be truncated"), 0644)

	pool, _ := NewFilePool(2)
	defer pool.CloseAll()

	pf := pool.NewFile(path, os.O_CREATE|os.O_RDWR, 0644)
	pf.Truncate(5)

	pf.Write([]byte("XY"))

	pool.Sync()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "initiXY", string(data))
}
