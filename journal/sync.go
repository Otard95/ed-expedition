package journal

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"time"
)

// Looks over all logs at and after the provided timestamp and emits configured
// events from these
func (jw *Watcher) Sync(since time.Time) error {
	jw.logger.Trace(fmt.Sprintf("[Sync] Called with since=%v", since))
	if jw.started {
		return errors.New("Cannot call journal.Watcher.Sync() after the watcher has been started")
	}

	entries, err := os.ReadDir(jw.dir)
	if err != nil {
		return err
	}

	journals := make([]*JournalName, 0, 16)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !journalFilePattern.MatchString(entry.Name()) {
			continue
		}
		name, err := parseJournalName(entry.Name())
		if err != nil {
			return err
		}
		journals = append(journals, name)
	}

	slices.SortFunc(journals, func(a, b *JournalName) int {
		if timeDiff := a.time.Compare(b.time); timeDiff != 0 {
			return timeDiff
		}
		return a.part - b.part
	})

	jw.logger.Trace(fmt.Sprintf("[Sync] Found %d journals", len(journals)))
	for i, j := range journals {
		jw.logger.Trace(fmt.Sprintf("[Sync]   %d: %s", i, j.String()))
	}

	if len(journals) == 0 {
		jw.logger.Trace("[Sync] No journals found, returning")
		return nil
	}

	// TODO: Look for optimization potential
	//  - Maybe [Heading entry](https://elite-journal.readthedocs.io/en/latest/File%20Format.html#heading-entry)
	//
	// Since the timestamp in the journal's filename is from when it was started,
	// and we get the first one AFTER the provided timestamp 'since', we need to
	// go back to the journals with next recent timestamp, as these might include
	// event that are after our provided 'since'
	// Regardless, since each timestamp of every event handled by 'processData'
	// is checked against 'since', we'll never process event's we should not.
	cutoff := slices.IndexFunc(journals, func(j *JournalName) bool { return j.time.After(since) })
	jw.logger.Trace(fmt.Sprintf("[Sync] IndexFunc returned: %d", cutoff))
	if cutoff < 0 {
		cutoff = len(journals) - 1
		jw.logger.Trace(fmt.Sprintf("[Sync] No journal after since, cutoff set to last: %d", cutoff))
	} else if cutoff > 0 {
		cutoff--
		jw.logger.Trace(fmt.Sprintf("[Sync] Decremented cutoff to: %d", cutoff))
	}
	jw.logger.Trace(fmt.Sprintf("[Sync] Before part adjustment, cutoff=%d, journal part=%d", cutoff, journals[cutoff].part))
	cutoff -= journals[cutoff].part - 1
	jw.logger.Trace(fmt.Sprintf("[Sync] After part adjustment, cutoff=%d", cutoff))
	// If journal files have been deleted the first journal might be a part > 1
	if cutoff < 0 {
		jw.logger.Trace("[Sync] Cutoff negative, setting to 0")
		cutoff = 0
	}

	journals = journals[cutoff:]
	jw.logger.Trace(fmt.Sprintf("[Sync] Processing %d journals starting from index %d", len(journals), cutoff))

	jw.lastTimestamp = since
	for i, journal := range journals {
		jw.logger.Trace(fmt.Sprintf("[Sync] Processing journal %d: %s", i, journal.name))
		content, err := os.ReadFile(path.Join(jw.dir, journal.name))
		if err != nil {
			return err
		}
		jw.logger.Trace(fmt.Sprintf("[Sync] Read %d bytes from %s", len(content), journal.name))
		err = jw.processData(content)
		if err != nil {
			return err
		}
	}

	jw.logger.Trace("[Sync] Complete")
	return nil
}

type JournalName struct {
	name string
	time time.Time
	part int
}

func (j *JournalName) String() string {
	return fmt.Sprintf("%s (time: %s, part: %d)", j.name, j.time.Format("2006-01-02T15:04:05"), j.part)
}

func parseJournalName(name string) (*JournalName, error) {
	matches := journalFilePattern.FindStringSubmatch(name)
	if matches == nil {
		return nil, errors.New("File name is not a journal name")
	}

	timestamp, err := time.Parse("2006-01-02T150405", matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	part, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse part number: %w", err)
	}

	return &JournalName{
		name: name,
		time: timestamp,
		part: part,
	}, nil
}
