package database

import (
	"ed-expedition/lib/filepool"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type BuildPhase string

const (
	BuildPhasePending  BuildPhase = "pending"
	BuildPhaseProcess  BuildPhase = "process"
	BuildPhaseCompile  BuildPhase = "compile"
	BuildPhaseIndex    BuildPhase = "index"
	BuildPhaseFinalize BuildPhase = "finalize"
	BuildPhaseDone     BuildPhase = "done"

	NumBuckets      = 1000
	BucketGroupSize = 50
)

type BuildState struct {
	Phase              BuildPhase `json:"phase"`
	InputSize          int64      `json:"input_size,omitempty"`
	LastSystemID64     uint64     `json:"last_system_id64,omitempty"`
	NamesBinSize       int64      `json:"names_bin_size,omitempty"`
	SortedBuckets      []int      `json:"sorted_buckets,omitempty"`
	SystemsBinComplete bool       `json:"systems_bin_complete,omitempty"`
}

type GalaxyBuildManager struct {
	inputPath string
	state     BuildState
}

func NewGalaxyBuildManager(inputPath string) (*GalaxyBuildManager, error) {
	m := &GalaxyBuildManager{
		inputPath: inputPath,
	}

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
	statePath, err := BuildStatePath()
	if err != nil {
		return BuildState{}, err
	}

	data, err := os.ReadFile(statePath)
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

	systemsBinPath, err := SystemsBinPath()
	if err != nil {
		return BuildState{}, err
	}
	if _, err := os.Stat(systemsBinPath); err == nil {
		return BuildState{Phase: BuildPhaseDone}, nil
	}

	return BuildState{Phase: BuildPhasePending}, nil
}

func (m *GalaxyBuildManager) saveState() error {
	statePath, err := BuildStatePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(m.state)
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

func (m *GalaxyBuildManager) Start() error {
	if m.state.Phase != BuildPhasePending {
		return errors.New("cannot start: build already in progress or complete")
	}

	inputSize, err := m.getInputSize()
	if err != nil {
		return err
	}
	m.state.InputSize = inputSize
	m.state.Phase = BuildPhaseProcess

	if err := m.saveState(); err != nil {
		return err
	}

	return m.process()
}

func (m *GalaxyBuildManager) Continue() error {
	switch m.state.Phase {
	case BuildPhasePending:
		return errors.New("cannot continue: no build in progress")
	case BuildPhaseDone:
		return errors.New("cannot continue: build already complete")
	}

	if err := m.validateInput(); err != nil {
		return err
	}

	switch m.state.Phase {
	case BuildPhaseProcess:
		return m.process()
	case BuildPhaseCompile:
		return m.compile()
	case BuildPhaseIndex:
		return m.index()
	case BuildPhaseFinalize:
		return m.finalize()
	}

	return nil
}

func (m *GalaxyBuildManager) process() error {
	if err := m.ensureBucketDirs(); err != nil {
		return err
	}

	pool, err := filepool.NewFilePool(100)
	if err != nil {
		return err
	}
	defer pool.CloseAll()

	buckets := make([]*filepool.PooledFile, NumBuckets)
	for bucket := range NumBuckets {
		buckets[bucket] = pool.NewFile(
			m.bucketPath(bucket),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
	}

	namesBinPath, err := m.cacheNamesBinPath()
	if err != nil {
		return err
	}
	namesFile := pool.NewFile(namesBinPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	// TODO: implement process phase - parse JSON, write to buckets and names.bin
	_ = buckets
	_ = namesFile

	return nil
}

func (m *GalaxyBuildManager) ensureBucketDirs() error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	numGroups := NumBuckets / BucketGroupSize
	for g := range numGroups {
		dir := filepath.Join(cacheDir, "buckets", fmt.Sprintf("%02d", g))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (m *GalaxyBuildManager) bucketPath(index int) string {
	cacheDir, _ := GetCacheDir()
	group := index / BucketGroupSize
	return filepath.Join(cacheDir, "buckets", fmt.Sprintf("%02d", group), fmt.Sprintf("%03d.bin", index))
}

func (m *GalaxyBuildManager) cacheNamesBinPath() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "names.bin"), nil
}

func (m *GalaxyBuildManager) compile() error {
	// TODO: implement compile phase
	return nil
}

func (m *GalaxyBuildManager) index() error {
	// TODO: implement index phase
	return nil
}

func (m *GalaxyBuildManager) finalize() error {
	// TODO: implement finalize phase
	// Move all output files from cache to data dir:
	// - systems.bin, systems.idx, names.bin, names.trie
	// Delete build state file
	return nil
}

func (m *GalaxyBuildManager) getInputSize() (int64, error) {
	info, err := os.Stat(m.inputPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (m *GalaxyBuildManager) validateInput() error {
	inputSize, err := m.getInputSize()
	if err != nil {
		return err
	}
	if inputSize != m.state.InputSize {
		return errors.New("input file size mismatch: file may have changed since build started")
	}
	return nil
}
