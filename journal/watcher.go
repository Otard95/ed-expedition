package journal

import (
	"ed-expedition/lib/channels"
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"encoding/json"
	"fmt"
	"hash/fnv"
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
	dir         string
	watcher     *fsnotify.Watcher
	currentFile string
	seek        int64
	started     bool
	logger      wailsLogger.Logger

	Loadout   *channels.FanoutChannel[*LoadoutEvent]
	FSDJump   *channels.FanoutChannel[*FSDJumpEvent]
	FSDTarget *channels.FanoutChannel[*FSDTargetEvent]
	Location  *channels.FanoutChannel[*LocationEvent]
	StartJump *channels.FanoutChannel[*StartJumpEvent]
	SyncState *channels.FanoutChannel[models.JournalSync]

	// Status
	Scooping            *channels.FanoutChannel[bool]
	Fuel                *channels.FanoutChannel[*FuelStatus]
	FsdCharging         *channels.FanoutChannel[bool]
	prevFsdChargingFlag bool
	prevHyperdriveCFlag bool

	statusDebounce *time.Timer
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
		SyncState: channels.NewFanoutChannel[models.JournalSync]("SyncState", 1, FanoutChannelTimeout, logger),

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
				// Elite writes Status.json frequently and non-atomically, often
				// firing several events per write. Debounce so we only read once
				// the write has settled, avoiding mid-write reads.
				if jw.statusDebounce != nil {
					jw.statusDebounce.Stop()
				}
				jw.statusDebounce = time.AfterFunc(10*time.Millisecond, jw.handleStatusUpdate)
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
	if jw.statusDebounce != nil {
		jw.statusDebounce.Stop()
	}
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

	lines, err := jw.parseLines(buf)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse journal data: %v", err))
	}

	if len(lines) == 0 {
		return nil
	}

	jw.publishSyncState(lines[len(lines)-1])
	jw.dispatch(lines)

	return nil
}

type parsedLine struct {
	Raw       []byte
	Timestamp time.Time
	Event     EventType
}

func (jw *Watcher) parseLines(data []byte) ([]parsedLine, error) {
	rawLines := slice.Split(data, '\n')
	jw.logger.Trace(fmt.Sprintf("[parseLines] Processing %d lines", len(rawLines)))

	parsed := make([]parsedLine, 0, len(rawLines))

	for i, line := range rawLines {
		if len(line) == 0 {
			continue
		}

		jw.logger.Trace(fmt.Sprintf("[parseLines] Line %d: %s", i, string(line)))

		var base struct {
			Timestamp time.Time `json:"timestamp"`
			Event     EventType `json:"event"`
		}
		if err := json.Unmarshal(line, &base); err != nil {
			jw.logger.Trace(fmt.Sprintf("[parseLines] Line %d: unmarshal error: %v", i, err))
			return nil, err
		}

		jw.logger.Trace(fmt.Sprintf("[parseLines] Line %d: event=%s timestamp=%v", i, base.Event, base.Timestamp))

		parsed = append(parsed, parsedLine{
			Raw:       line,
			Timestamp: base.Timestamp,
			Event:     base.Event,
		})
	}

	return parsed, nil
}

func (jw *Watcher) filterSyncBoundary(lines []parsedLine, syncState *models.JournalSync) []parsedLine {
	filtered := make([]parsedLine, 0, len(lines))
	pastBoundary := syncState.EventHash == ""

	for _, line := range lines {
		if line.Timestamp.Before(syncState.Timestamp) {
			jw.logger.Trace(fmt.Sprintf("[filterSync] timestamp %v before %v, skipping", line.Timestamp, syncState.Timestamp))
			continue
		}

		if !pastBoundary {
			if line.Timestamp.After(syncState.Timestamp) {
				pastBoundary = true
			} else {
				// At the boundary timestamp: skip everything up to and
				// including the hash match
				pastBoundary = hashLine(line.Raw) == syncState.EventHash
				continue
			}
		}

		filtered = append(filtered, line)
	}

	return filtered
}

func hashLine(raw []byte) string {
	h := fnv.New32a()
	h.Write(raw)
	return fmt.Sprintf("%x", h.Sum32())
}

func (jw *Watcher) publishSyncState(last parsedLine) {
	jw.SyncState.Publish(models.JournalSync{
		Timestamp: last.Timestamp,
		EventHash: hashLine(last.Raw),
	})
}

func (jw *Watcher) dispatch(lines []parsedLine) {
	for _, line := range lines {
		switch line.Event {
		case Loadout:
			var event LoadoutEvent
			if err := json.Unmarshal(line.Raw, &event); err == nil {
				jw.logger.Trace("[dispatch] Publishing Loadout")
				jw.Loadout.Publish(&event)
			}
		case FSDJump:
			var event FSDJumpEvent
			if err := json.Unmarshal(line.Raw, &event); err == nil {
				jw.logger.Trace(fmt.Sprintf("[FSD_TIMING] FSDJump event: system=%s, timestamp=%v, fuelLevel=%.2f, fuelUsed=%.2f",
					event.StarSystem, event.Timestamp, event.FuelLevel, event.FuelUsed))
				jw.logger.Trace("[dispatch] Publishing FSDJump")
				jw.FSDJump.Publish(&event)
			}
		case FSDTarget:
			var event FSDTargetEvent
			if err := json.Unmarshal(line.Raw, &event); err == nil {
				jw.logger.Trace(fmt.Sprintf("[dispatch] Publishing FSDTarget: %s", event.Name))
				jw.FSDTarget.Publish(&event)
			} else {
				jw.logger.Trace(fmt.Sprintf("[dispatch] FSDTarget unmarshal error: %v", err))
			}
		case Location:
			var event LocationEvent
			if err := json.Unmarshal(line.Raw, &event); err == nil {
				jw.logger.Trace("[dispatch] Publishing Location")
				jw.Location.Publish(&event)
			}
		case StartJump:
			var event StartJumpEvent
			if err := json.Unmarshal(line.Raw, &event); err == nil {
				starSystem := "<none>"
				if event.StarSystem != nil {
					starSystem = *event.StarSystem
				}
				jw.logger.Trace(fmt.Sprintf("[FSD_TIMING] StartJump event: type=%s, system=%s, timestamp=%v",
					event.JumpType, starSystem, event.Timestamp))
				jw.logger.Trace(fmt.Sprintf("[dispatch] Publishing StartJump: %s to %s", event.JumpType, starSystem))
				jw.StartJump.Publish(&event)
			}
		}
	}
}
