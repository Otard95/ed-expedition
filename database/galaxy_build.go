package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"sync"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type BuildPhase string

const (
	BuildPhasePending  BuildPhase = "pending"     // not started
	BuildPhaseProcess  BuildPhase = "in_progress" // inserting into db
	BuildPhaseFinalize BuildPhase = "finalize"    // creating indexes
	BuildPhaseDone     BuildPhase = "done"        // complete
)

type BuildState struct {
	Phase     BuildPhase `json:"phase"`
	InputSize int64      `json:"input_size,omitempty"`
}

type GalaxyBuildOptions struct {
	TransformWorkers int
}

type GalaxyBuildManager struct {
	inputPath string
	state     BuildState
	db        *GalaxyDB
	logger    wailsLogger.Logger

	transformWorkers int

	activeTransformWorkers sync.WaitGroup
	rawSystemChan          chan *RawSystem
	systemChan             chan *System
}

func NewGalaxyBuildManager(db *GalaxyDB, inputPath string, logger wailsLogger.Logger, options *GalaxyBuildOptions) (*GalaxyBuildManager, error) {
	if options == nil {
		options = &GalaxyBuildOptions{}
	}

	m := &GalaxyBuildManager{
		inputPath:        inputPath,
		db:               db,
		logger:           logger,
		transformWorkers: clamp(options.TransformWorkers, 1, 8),
	}
	if options.TransformWorkers <= 0 {
		m.transformWorkers = defaultTransformWorkers()
	}

	m.activeTransformWorkers = sync.WaitGroup{}
	m.rawSystemChan = make(chan *RawSystem, 50)
	m.systemChan = make(chan *System, 100000)

	state, err := m.resolveState()
	if err != nil {
		return nil, err
	}
	m.state = state

	return m, nil
}

func (m *GalaxyBuildManager) Phase() BuildPhase {
	return m.state.Phase
}

func (m *GalaxyBuildManager) resolveState() (BuildState, error) {
	data, err := os.ReadFile(BuildStatePath)
	if err == nil {
		var state BuildState
		if err := json.Unmarshal(data, &state); err != nil {
			return BuildState{}, err
		}
		m.logger.Debug(fmt.Sprintf("[GalaxyBuild] Resolved state from file: phase=%s", state.Phase))
		return state, nil
	}

	if !os.IsNotExist(err) {
		return BuildState{}, err
	}

	m.logger.Debug("[GalaxyBuild] No state file found, probing database")
	phase, err := m.probeDatabasePhase()
	if err != nil {
		return BuildState{}, err
	}

	m.logger.Debug(fmt.Sprintf("[GalaxyBuild] Probed database phase: %s", phase))
	return BuildState{Phase: phase}, nil

}

func (m *GalaxyBuildManager) probeDatabasePhase() (BuildPhase, error) {
	tables, err := m.db.ListTables()
	if err != nil {
		return "", err
	}

	if !slices.Contains(tables, "systems") {
		return BuildPhasePending, nil
	}

	indexes, err := m.db.ListIndexesForTable("systems")
	if err != nil {
		return "", err
	}

	hasHilbert := slices.Contains(indexes, "idx_systems_hilbert")
	hasName := slices.Contains(indexes, "idx_systems_name")

	if hasHilbert && hasName {
		return BuildPhaseDone, nil
	}

	return BuildPhaseFinalize, nil

}

func (m *GalaxyBuildManager) saveState() error {
	return WriteJSON(BuildStatePath, m.state)
}

func (m *GalaxyBuildManager) Build() error {
	if m.state.Phase == BuildPhaseDone {
		m.logger.Debug("[GalaxyBuild] Build already complete, skipping")
		return nil
	}

	m.logger.Info(fmt.Sprintf("[GalaxyBuild] Build starting at phase=%s", m.state.Phase))
	buildStart := time.Now()

	var err error
	switch m.state.Phase {
	case BuildPhaseFinalize:
		err = m.runFinalize()
	case BuildPhasePending, BuildPhaseProcess:
		err = m.runFullBuild()
	default:
		err = fmt.Errorf("unknown build phase: %s", m.state.Phase)
	}

	if err != nil {
		m.logger.Error(fmt.Sprintf("[GalaxyBuild] Build failed after %s: %v", time.Since(buildStart), err))
		return err
	}

	m.logger.Info(fmt.Sprintf("[GalaxyBuild] Build completed in %s", time.Since(buildStart)))
	return nil
}

func (m *GalaxyBuildManager) runFinalize() error {
	finalizeStart := time.Now()
	m.logger.Debug("[GalaxyBuild] Creating indexes")

	if err := m.db.EnsureSystemsIndexes(); err != nil {
		return err
	}

	m.logger.Debug(fmt.Sprintf("[GalaxyBuild] Indexes created in %s", time.Since(finalizeStart)))
	m.state.Phase = BuildPhaseDone
	return m.saveState()
}

func (m *GalaxyBuildManager) runFullBuild() error {
	m.state.Phase = BuildPhaseProcess
	if err := m.saveState(); err != nil {
		return err
	}

	processStart := time.Now()
	if err := m.process(); err != nil {
		return err
	}
	m.logger.Debug(fmt.Sprintf("[GalaxyBuild] Processing completed in %s", time.Since(processStart)))

	m.state.Phase = BuildPhaseFinalize
	if err := m.saveState(); err != nil {
		return err
	}

	return m.runFinalize()
}

func (m *GalaxyBuildManager) process() error {
	if err := m.db.EnsureSystemsTable(); err != nil {
		return err
	}

	galaxyParser, err := NewGalaxyParser(m.inputPath)
	if err != nil {
		return err
	}
	defer galaxyParser.Close()

	m.logger.Debug(fmt.Sprintf("[GalaxyBuild] Starting pipeline: %d transform workers", m.transformWorkers))

	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	errCh := make(chan error, 1)

	go m.writeSystems(ctx, cancel, doneCh, errCh)

	for range m.transformWorkers {
		m.activeTransformWorkers.Add(1)
		go m.transformRawSystem(ctx)
	}

	readErr := m.readInput(ctx, galaxyParser)

	close(m.rawSystemChan)
	m.activeTransformWorkers.Wait()
	close(m.systemChan)

	<-doneCh

	var writeErr error
	select {
	case writeErr = <-errCh:
	default:
	}

	errText := "Build failed because of:\n"
	if readErr != nil {
		errText += " - Reader: " + readErr.Error() + "\n"
	}
	if writeErr != nil {
		errText += " - Writer: " + writeErr.Error() + "\n"
	}
	if writeErr != nil || readErr != nil {
		return errors.New(errText)
	}
	return nil
}

func (m *GalaxyBuildManager) readInput(ctx context.Context, galaxyParser *GalaxyParser) error {
	for {
		rawSystem, err := galaxyParser.Next()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return errors.New("stopped by cancellation")
		case m.rawSystemChan <- rawSystem:
		}
	}
}

func (m *GalaxyBuildManager) transformRawSystem(ctx context.Context) {
	defer m.activeTransformWorkers.Done()

	for {
		rawSystem, ok := <-m.rawSystemChan
		if !ok {
			return
		}
		x, y, z := NormalizeCoord(
			rawSystem.Coords.X,
			rawSystem.Coords.Y,
			rawSystem.Coords.Z,
		)
		hilbertIndex := Hilbert(x, y, z)
		starClass := parseStarClass(rawSystem.MainStar)

		sys := &System{
			rawSystem.ID64,
			uint64(hilbertIndex),
			rawSystem.Name,
			x, y, z,
			starClass,
		}
		select {
		case <-ctx.Done():
			return
		case m.systemChan <- sys:
		}
	}
}

func (m *GalaxyBuildManager) writeSystems(
	ctx context.Context,
	cancel context.CancelFunc,
	doneCh chan<- struct{},
	errCh chan<- error,
) {
	defer func() {
		close(doneCh)
		close(errCh)
	}()

	errorOnce := sync.Once{}
	propagateErr := func(e error) {
		errorOnce.Do(func() {
			m.logger.Error(fmt.Sprintf("[GalaxyBuild] Writer error: %v", e))
			errCh <- e
			cancel()
		})
	}

	tx, err := m.db.Begin()
	if err != nil {
		propagateErr(err)
		return
	}
	stmt, err := tx.PrepareSystemInsert()
	if err != nil {
		_ = tx.Rollback()
		propagateErr(err)
		return
	}

	counter := 0
	totalInserted := 0
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case s, ok := <-m.systemChan:
			if !ok {
				break loop
			}

			if _, err := stmt.Exec(s.Id, s.hilbertKey, s.Name, s.X, s.Y, s.Z, s.StarClass); err != nil {
				propagateErr(err)
				break loop
			}
			counter++

			if counter >= 100000 {
				if err := stmt.Close(); err != nil {
					propagateErr(err)
					break loop
				}
				if err := tx.Commit(); err != nil {
					propagateErr(err)
					break loop
				}
				totalInserted += counter
				m.logger.Debug(fmt.Sprintf("[GalaxyBuild] Committed batch, %d systems inserted so far", totalInserted))
				counter = 0
				tx, err = m.db.Begin()
				if err != nil {
					propagateErr(err)
					return
				}
				stmt, err = tx.PrepareSystemInsert()
				if err != nil {
					_ = tx.Rollback()
					propagateErr(err)
					return
				}
			}
		}
	}

	if err = stmt.Close(); err != nil {
		propagateErr(err)
	}
	if counter > 0 {
		if err = tx.Commit(); err != nil {
			propagateErr(err)
		}
		totalInserted += counter
		m.logger.Debug(fmt.Sprintf("[GalaxyBuild] Final batch committed, %d systems inserted total", totalInserted))
		return
	}
	_ = tx.Rollback()
}

func defaultTransformWorkers() int {
	workers := runtime.GOMAXPROCS(0) / 4
	return clamp(workers, 1, 8)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
