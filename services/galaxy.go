package services

import (
	"context"
	"ed-expedition/database"
	"ed-expedition/download"
	"ed-expedition/lib/job"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type GalaxyState string

const (
	GalaxyStateNone               GalaxyState = "none"
	GalaxyStateDownloadIncomplete GalaxyState = "download_incomplete"
	GalaxyStateBuildIncomplete    GalaxyState = "build_incomplete"
	GalaxyStateReady              GalaxyState = "ready"

	spanshDownloadURL = "https://downloads.spansh.co.uk/systems.json.gz"
)

type GalaxyService struct {
	logger wailsLogger.Logger

	db              *database.GalaxyDB
	downloadManager *download.Manager
	buildManager    *database.GalaxyBuildManager
	downloadPath    string
	state           GalaxyState
	running         bool
}

func NewGalaxyService(logger wailsLogger.Logger) *GalaxyService {
	downloadPath := filepath.Join(database.CacheDir, "systems.json.gz")

	return &GalaxyService{
		logger:       logger,
		downloadPath: downloadPath,
		state:        GalaxyStateNone,
	}
}

func (s *GalaxyService) State() GalaxyState {
	return s.state
}

func (s *GalaxyService) Running() bool {
	return s.running
}

func (s *GalaxyService) openDB() error {
	if s.db != nil {
		return nil
	}
	db, err := database.OpenGalaxyDB()
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *GalaxyService) Start() error {
	var err error
	s.downloadManager, err = download.NewManager(spanshDownloadURL, s.downloadPath)
	if err != nil {
		if !errors.Is(err, download.ErrDestinationExists) {
			return fmt.Errorf("failed to create GalaxyService: %w", err)
		}
	} else if !s.downloadManager.IsComplete() {
		if s.downloadManager.DownloadedBytes() > 0 {
			s.state = GalaxyStateDownloadIncomplete
		}
		return nil
	}

	if err := s.openDB(); err != nil {
		return fmt.Errorf("failed to open galaxy database: %w", err)
	}

	buildManager, err := database.NewGalaxyBuildManager(s.db, s.downloadPath, s.logger, nil)
	if err != nil {
		return fmt.Errorf("failed to probe galaxy build state: %w", err)
	}

	if buildManager.Phase() == database.BuildPhaseDone {
		s.state = GalaxyStateReady
		return nil
	}

	s.buildManager = buildManager
	s.state = GalaxyStateBuildIncomplete

	return nil
}

func (s *GalaxyService) Stop() {
	s.buildManager = nil
	if s.downloadManager != nil {
		s.downloadManager.Close()
		s.downloadManager = nil
	}
	if s.db != nil {
		s.db.Close()
		s.db = nil
	}
}

func (s *GalaxyService) BuildJob() *job.Job[job.NoCtx, any] {
	if s.state == GalaxyStateReady || s.running {
		return nil
	}

	s.running = true

	j := job.New(
		"Galaxy",
		job.NoCtx{},
		[]job.PhaseConfig[job.NoCtx]{
			{
				Name:     "download",
				Label:    "Download",
				Type:     job.PhaseTypeObservable,
				Callback: s.doDownload,
			},
			{
				Name:  "insert",
				Label: "Build",
				Type:  job.PhaseTypeObservable,
				Callback: func(ctx context.Context, state *job.NoCtx, tracker *job.ProgressTracker) error {
					if err := s.setupBuildManager(); err != nil {
						return err
					}
					if s.buildManager.Phase() == database.BuildPhaseFinalize {
						return nil
					}
					return s.buildManager.Process(ctx, tracker)
				},
			},
			{
				Name:  "index",
				Label: "Optimize",
				Type:  job.PhaseTypeEstimated,
				EstimateCallback: func(completed map[string]time.Duration) time.Duration {
					if duration, ok := completed["insert"]; ok {
						return duration * 3 / 2
					}
					return 40 * time.Minute
				},
				Callback: func(ctx context.Context, state *job.NoCtx, tracker *job.ProgressTracker) error {
					return s.buildManager.Finalize(ctx, tracker)
				},
			},
		},
		func(state job.NoCtx) (any, error) {
			s.state = GalaxyStateReady
			s.running = false
			return nil, nil
		},
		s.logger,
	)

	go func() {
		for status := range j.StatusChange().Subscribe() {
			if status.State == job.JobStateError {
				s.running = false
			}
		}
	}()

	return j
}

func (s *GalaxyService) doDownload(ctx context.Context, _state *job.NoCtx, tracker *job.ProgressTracker) error {
	if s.downloadManager != nil && !s.downloadManager.IsComplete() {
		s.state = GalaxyStateDownloadIncomplete
		tracker.SetTotal(float64(s.downloadManager.TotalBytes()))
		err := s.downloadManager.Download(func(downloaded int64) {
			tracker.SetProgress(float64(downloaded))
		})
		if err != nil {
			return fmt.Errorf("galaxy download failed: %w", err)
		}
	}
	return nil
}

func (s *GalaxyService) setupBuildManager() error {
	s.state = GalaxyStateBuildIncomplete

	if s.buildManager == nil {
		if err := s.openDB(); err != nil {
			return fmt.Errorf("failed to open galaxy database: %w", err)
		}

		buildManager, err := database.NewGalaxyBuildManager(s.db, s.downloadPath, s.logger, nil)
		if err != nil {
			return fmt.Errorf("failed to create galaxy build manager: %w", err)
		}
		s.buildManager = buildManager
	}
	return nil
}
