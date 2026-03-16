package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"

	"gonum.org/v1/gonum/spatial/curve"
	_ "modernc.org/sqlite"
)

type BuildPhase string

const (
	BuildPhasePending  BuildPhase = "pending"     // not started
	BuildPhaseProcess  BuildPhase = "in_progress" // inserting into db
	BuildPhaseFinalize BuildPhase = "finalize"    // creating indexes
	BuildPhaseDone     BuildPhase = "done"        // complete

	insertSystemSQL = `INSERT INTO systems(id, hilbert_index, name, x, y, z, star_class) VALUES(?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`
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
	db        *sql.DB
	hilbert   curve.Hilbert3D

	transformWorkers int

	activeTransformWorkers sync.WaitGroup
	rawSystemChan          chan *RawSystem
	systemChan             chan *System
}

func NewGalaxyBuildManager(inputPath string, options *GalaxyBuildOptions) (*GalaxyBuildManager, error) {
	if options == nil {
		options = &GalaxyBuildOptions{}
	}

	m := &GalaxyBuildManager{
		inputPath:        inputPath,
		transformWorkers: clamp(options.TransformWorkers, 1, 8),
	}
	if options.TransformWorkers <= 0 {
		m.transformWorkers = defaultTransformWorkers()
	}

	dbPath := filepath.Join(DataDir, "galaxy.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	m.db = db

	hilbert, err := curve.NewHilbert3D(HilbertOrder)
	if err != nil {
		_ = m.Close()
		return nil, err
	}
	m.hilbert = hilbert
	m.activeTransformWorkers = sync.WaitGroup{}
	m.rawSystemChan = make(chan *RawSystem, 50)
	m.systemChan = make(chan *System, 100000)

	state, err := m.resolveState()
	if err != nil {
		_ = m.Close()
		return nil, err
	}
	m.state = state

	return m, nil
}

func (m *GalaxyBuildManager) Close() error {
	if m.db == nil {
		return nil
	}
	err := m.db.Close()
	m.db = nil
	return err
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
		return state, nil
	}

	if !os.IsNotExist(err) {
		return BuildState{}, err
	}

	phase, err := m.probeDatabasePhase()
	if err != nil {
		return BuildState{}, err
	}

	return BuildState{Phase: phase}, nil

}

func (m *GalaxyBuildManager) probeDatabasePhase() (BuildPhase, error) {
	tables, err := m.listTables()
	if err != nil {
		return "", err
	}

	if !slices.Contains(tables, "systems") {
		return BuildPhasePending, nil
	}

	indexes, err := m.listIndexesForTable("systems")
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

func (m *GalaxyBuildManager) listTables() ([]string, error) {
	rows, err := m.db.Query(`SELECT name FROM sqlite_master WHERE type = 'table'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (m *GalaxyBuildManager) listIndexesForTable(tableName string) ([]string, error) {
	rows, err := m.db.Query(`SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ?`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		indexes = append(indexes, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexes, nil
}

func (m *GalaxyBuildManager) saveState() error {
	return WriteJSON(BuildStatePath, m.state)
}

func (m *GalaxyBuildManager) Build() error {
	switch m.state.Phase {
	case BuildPhaseDone:
		return nil
	case BuildPhaseFinalize:
		if err := m.finalize(); err != nil {
			return err
		}
		m.state.Phase = BuildPhaseDone
		return m.saveState()
	case BuildPhasePending, BuildPhaseProcess:
		m.state.Phase = BuildPhaseProcess
		if err := m.saveState(); err != nil {
			return err
		}

		if err := m.process(); err != nil {
			return err
		}

		m.state.Phase = BuildPhaseFinalize
		if err := m.saveState(); err != nil {
			return err
		}

		if err := m.finalize(); err != nil {
			return err
		}

		m.state.Phase = BuildPhaseDone
		return m.saveState()
	default:
		return errors.New("unknown build phase")
	}
}

func (m *GalaxyBuildManager) finalize() error {
	if _, err := m.db.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_hilbert ON systems(hilbert_index)`); err != nil {
		return err
	}
	if _, err := m.db.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_name ON systems(name)`); err != nil {
		return err
	}
	return nil
}

func (m *GalaxyBuildManager) process() error {
	if err := m.ensureSystemsTable(); err != nil {
		return err
	}

	galaxyParser, err := NewGalaxyParser(m.inputPath)
	if err != nil {
		return err
	}
	defer galaxyParser.Close()

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

func (m *GalaxyBuildManager) ensureSystemsTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS systems (
			id INTEGER PRIMARY KEY,
			hilbert_index INTEGER NOT NULL,
			name TEXT NOT NULL,
			x INTEGER NOT NULL,
			y INTEGER NOT NULL,
			z INTEGER NOT NULL,
			star_class INTEGER NOT NULL
		)
	`)
	return err
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
		x, y, z := normalizeCoord(
			rawSystem.Coords.X,
			rawSystem.Coords.Y,
			rawSystem.Coords.Z,
		)
		hilbertIndex := m.hilbert.Pos([]int{int(x), int(y), int(z)})
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
			errCh <- e
			cancel()
		})
	}

	tx, err := m.db.Begin()
	if err != nil {
		propagateErr(err)
		return
	}
	stmt, err := tx.Prepare(insertSystemSQL)
	if err != nil {
		_ = tx.Rollback()
		propagateErr(err)
		return
	}

	counter := 0
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
				counter = 0
				tx, err = m.db.Begin()
				if err != nil {
					propagateErr(err)
					return
				}
				stmt, err = tx.Prepare(insertSystemSQL)
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
