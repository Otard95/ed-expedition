package journal

import (
	"ed-expedition/lib/channels"
	"ed-expedition/lib/slice"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
)

var journalFilePattern = regexp.MustCompile(`Journal\.(\d{4}-\d{2}-\d{2}T\d{6})\.(\d+)\.json`)

type JournalWatcher struct {
	dir           string
	watcher       *fsnotify.Watcher
	currentFile   string
	seek          int64
	lastTimestamp time.Time

	Loadout   *channels.FanoutChannel[*LoadoutEvent]
	FSDJump   *channels.FanoutChannel[*FSDJumpEvent]
	FSDTarget *channels.FanoutChannel[*FSDTargetEvent]
}

func NewJournalWatcher(dir string, lastTimestamp time.Time) (*JournalWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	return &JournalWatcher{
		dir:           dir,
		watcher:       watcher,
		currentFile:   "",
		lastTimestamp: lastTimestamp,

		Loadout:   channels.NewFanoutChannel[*LoadoutEvent](32),
		FSDJump:   channels.NewFanoutChannel[*FSDJumpEvent](32),
		FSDTarget: channels.NewFanoutChannel[*FSDTargetEvent](32),
	}, nil
}

func (jw *JournalWatcher) Start() {
	go func() {
		for e := range jw.watcher.Events {
			file := path.Base(e.Name)

			if file == jw.currentFile {
				jw.handleJournalUpdate()
				continue
			}

			if !journalFilePattern.MatchString(file) {
				continue
			}

			jw.currentFile = file
			jw.seek = 0
			jw.handleJournalUpdate()
		}
	}()
}

func (jw *JournalWatcher) Close() {
	jw.watcher.Close()
}

func (jw *JournalWatcher) handleJournalUpdate() error {
	file, err := os.Open(path.Join(jw.dir, jw.currentFile))
	if err != nil {
		return err
	}
	defer file.Close()

	pos, err := file.Seek(jw.seek, 0)
	if err != nil {
		panic(err)
	}
	if pos != jw.seek {
		panic(fmt.Sprintf("Failed to seek! got to %d aimed for %d", pos, jw.seek))
	}

	buf := []byte{}
	for {
		b := make([]byte, 50000)
		n, err := file.Read(b)
		jw.seek += int64(n)
		if len(buf) == 0 {
			buf = b
		} else {
			buf = append(buf, b...)
		}

		if err == io.EOF {
			break
		}
	}

	for _, line := range slice.Split(buf, '\n') {
		if len(line) == 0 {
			break
		}

		var base struct {
			Timestamp time.Time `json:"timestamp"`
			Event     EventType `json:"event"`
		}
		json.Unmarshal(line, &base)

		if base.Timestamp.Before(jw.lastTimestamp) {
			continue
		}
		jw.lastTimestamp = base.Timestamp

		switch base.Event {
		case Loadout:
			var event LoadoutEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.Loadout.Publish(&event)
			}
		case FSDJump:
			var event FSDJumpEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.FSDJump.Publish(&event)
			}
		case FSDTarget:
			var event FSDTargetEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.FSDTarget.Publish(&event)
			}
		}
	}

	return nil
}
