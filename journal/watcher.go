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
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

var journalFilePattern = regexp.MustCompile(`Journal\.(\d{4}-\d{2}-\d{2}T\d{6})\.(\d+)\.log`)

const FanoutChannelTimeout = 200 * time.Millisecond

type Watcher struct {
	dir           string
	watcher       *fsnotify.Watcher
	currentFile   string
	seek          int64
	lastTimestamp time.Time
	started       bool
	logger        wailsLogger.Logger

	Loadout   *channels.FanoutChannel[*LoadoutEvent]
	FSDJump   *channels.FanoutChannel[*FSDJumpEvent]
	FSDTarget *channels.FanoutChannel[*FSDTargetEvent]
	Location  *channels.FanoutChannel[*LocationEvent]
	StartJump *channels.FanoutChannel[*StartJumpEvent]

	// Status
	Scooping            *channels.FanoutChannel[bool]
	Fuel                *channels.FanoutChannel[*FuelStatus]
	FsdCharging         *channels.FanoutChannel[bool]
	prevFsdChargingFlag bool
	prevHyperdriveCFlag bool
}

func NewWatcher(dir string, logger wailsLogger.Logger) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	return &Watcher{
		dir:         dir,
		watcher:     watcher,
		currentFile: "",
		logger:      logger,

		Loadout:   channels.NewFanoutChannel[*LoadoutEvent]("Loadout", 32, FanoutChannelTimeout, logger),
		FSDJump:   channels.NewFanoutChannel[*FSDJumpEvent]("FSDJump", 32, FanoutChannelTimeout, logger),
		FSDTarget: channels.NewFanoutChannel[*FSDTargetEvent]("FSDTarget", 32, FanoutChannelTimeout, logger),
		Location:  channels.NewFanoutChannel[*LocationEvent]("Location", 32, FanoutChannelTimeout, logger),
		StartJump: channels.NewFanoutChannel[*StartJumpEvent]("StartJump", 32, FanoutChannelTimeout, logger),

		Scooping:    channels.NewFanoutChannel[bool]("Scooping", 0, 5*time.Millisecond, logger),
		Fuel:        channels.NewFanoutChannel[*FuelStatus]("Fuel", 0, 5*time.Millisecond, logger),
		FsdCharging: channels.NewFanoutChannel[bool]("FsdCharging", 0, 5*time.Millisecond, logger),
	}, nil
}

func (jw *Watcher) Start() {
	jw.started = true
	go func() {
		for e := range jw.watcher.Events {
			file := path.Base(e.Name)

			if file == "Status.json" {
				go jw.handleStatusUpdate()
				continue
			}

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

func (jw *Watcher) Close() {
	jw.watcher.Close()
}

func (jw *Watcher) handleJournalUpdate() error {
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
			buf = b[:n]
		} else {
			buf = append(buf, b[:n]...)
		}

		if err == io.EOF {
			break
		}
	}

	err = jw.processData(buf)
	if err != nil {
		panic(fmt.Sprintf("Failed to process journal data: %v", err))
	}

	return nil
}

func (jw *Watcher) processData(data []byte) error {
	lines := slice.Split(data, '\n')
	jw.logger.Trace(fmt.Sprintf("[processData] Processing %d lines, lastTimestamp=%v", len(lines), jw.lastTimestamp))

	for i, line := range lines {
		if len(line) == 0 {
			jw.logger.Trace(fmt.Sprintf("[processData] Line %d: empty, skipping", i))
			continue
		}

		jw.logger.Trace(fmt.Sprintf("[processData] Line %d: %s", i, string(line)))

		var base struct {
			Timestamp time.Time `json:"timestamp"`
			Event     EventType `json:"event"`
		}
		err := json.Unmarshal(line, &base)
		if err != nil {
			jw.logger.Trace(fmt.Sprintf("[processData] Line %d: unmarshal error: %v", i, err))
			return err
		}

		jw.logger.Trace(fmt.Sprintf("[processData] Line %d: event=%s timestamp=%v", i, base.Event, base.Timestamp))

		if base.Timestamp.Before(jw.lastTimestamp) {
			jw.logger.Trace(fmt.Sprintf("[processData] Line %d: timestamp %v not after lastTimestamp %v, skipping", i, base.Timestamp, jw.lastTimestamp))
			continue
		}
		jw.lastTimestamp = base.Timestamp

		switch base.Event {
		case Loadout:
			var event LoadoutEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace("[processData] Publishing Loadout")
				jw.Loadout.Publish(&event)
			}
		case FSDJump:
			var event FSDJumpEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace(fmt.Sprintf("[FSD_TIMING] FSDJump event: system=%s, timestamp=%v, fuelLevel=%.2f, fuelUsed=%.2f",
					event.StarSystem, event.Timestamp, event.FuelLevel, event.FuelUsed))
				jw.logger.Trace("[processData] Publishing FSDJump")
				jw.FSDJump.Publish(&event)
			}
		case FSDTarget:
			var event FSDTargetEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace(fmt.Sprintf("[processData] Publishing FSDTarget: %s", event.Name))
				jw.FSDTarget.Publish(&event)
			} else {
				jw.logger.Trace(fmt.Sprintf("[processData] FSDTarget unmarshal error: %v", err))
			}
		case Location:
			var event LocationEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace("[processData] Publishing Location")
				jw.Location.Publish(&event)
			}
		case StartJump:
			var event StartJumpEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				starSystem := "<none>"
				if event.StarSystem != nil {
					starSystem = *event.StarSystem
				}
				jw.logger.Trace(fmt.Sprintf("[FSD_TIMING] StartJump event: type=%s, system=%s, timestamp=%v",
					event.JumpType, starSystem, event.Timestamp))
				jw.logger.Trace(fmt.Sprintf("[processData] Publishing StartJump: %s to %s", event.JumpType, starSystem))
				jw.StartJump.Publish(&event)
			}
		}
	}

	return nil
}
