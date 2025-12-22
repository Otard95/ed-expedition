package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	CannotRewindError = errors.New("Cannot rewind")
)

type Transaction struct {
	id        string
	canRewind bool
	actions   []TransactionAction
	mu        sync.Mutex
}
type TransactionAction struct {
	target  string
	tmpFile string
}

func NewTransaction(name string) *Transaction {
	return &Transaction{
		id:        name,
		canRewind: true,
	}
}

func (t *Transaction) WriteJSON(path string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + "." + t.id + "." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".tmp"
	err = os.WriteFile(tmpPath, content, 0644)
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.actions = append(t.actions, TransactionAction{target: path, tmpFile: tmpPath})
	t.mu.Unlock()

	return nil
}

func (t *Transaction) Rewind() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.canRewind {
		return CannotRewindError
	}

	t.canRewind = false

	var errs []string
	for _, a := range t.actions {
		if err := os.Remove(a.tmpFile); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", a.tmpFile, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to remove temp files: %s", strings.Join(errs, "; "))
	}

	return nil
}

func (t *Transaction) Apply() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.canRewind = false

	for _, a := range t.actions {
		if err := os.Rename(a.tmpFile, a.target); err != nil {
			time.Sleep(time.Millisecond)
			if err := os.Rename(a.tmpFile, a.target); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result T
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func WriteJSON(path string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + "." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".tmp"
	err = os.WriteFile(tmpPath, content, 0644)
	if err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
