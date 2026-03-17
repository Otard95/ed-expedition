package services

import (
	"ed-expedition/database"
	"ed-expedition/download"
	"errors"
	"fmt"
	"path/filepath"

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

	buildManager, err := database.NewGalaxyBuildManager(s.downloadPath, s.logger, nil)
	if err != nil {
		return fmt.Errorf("failed to probe galaxy build state: %w", err)
	}

	if buildManager.Phase() == database.BuildPhaseDone {
		buildManager.Close()
		s.state = GalaxyStateReady
		return nil
	}

	s.buildManager = buildManager
	s.state = GalaxyStateBuildIncomplete

	return nil
}

func (s *GalaxyService) Stop() {
	if s.buildManager != nil {
		s.buildManager.Close()
		s.buildManager = nil
	}
	if s.downloadManager != nil {
		s.downloadManager.Close()
		s.downloadManager = nil
	}
}

func (s *GalaxyService) DownloadAndBuild() error {
	if s.state == GalaxyStateReady {
		return nil
	}

	s.running = true
	defer func() { s.running = false }()

	if s.downloadManager != nil && !s.downloadManager.IsComplete() {
		s.state = GalaxyStateDownloadIncomplete
		if err := s.downloadManager.Download(nil); err != nil {
			return fmt.Errorf("galaxy download failed: %w", err)
		}
	}

	s.state = GalaxyStateBuildIncomplete

	if s.buildManager == nil {
		buildManager, err := database.NewGalaxyBuildManager(s.downloadPath, s.logger, nil)
		if err != nil {
			return fmt.Errorf("failed to create galaxy build manager: %w", err)
		}
		s.buildManager = buildManager

	}
	if err := s.buildManager.Build(); err != nil {
		s.buildManager.Close()
		return fmt.Errorf("galaxy build failed: %w", err)
	}

	s.state = GalaxyStateReady
	return nil
}
