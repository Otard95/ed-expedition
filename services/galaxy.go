package services

import (
	"ed-expedition/database"
	"ed-expedition/download"
	"fmt"
	"path"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type GalaxyService struct {
	logger wailsLogger.Logger

	downloadManager *download.Manager
}

func NewGalaxyService(logger wailsLogger.Logger) (*GalaxyService, error) {
	cacheDir, err := database.GetCacheDir()
	if err != nil {
		return nil, fmt.Errorf("Failed to create GalaxyService: Could not get cache dir: %s", err.Error())
	}

	dlm, err := download.NewManager(
		"https://downloads.spansh.co.uk/systems.json.gz",
		path.Join(cacheDir, "systems.json.gz"),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create GalaxyService: Could not create DownloadManager %s", err.Error())
	}

	return &GalaxyService{
		logger:          logger,
		downloadManager: dlm,
	}, nil
}

func (s *GalaxyService) Start() {
	if s.downloadManager.DownloadedBytes() == 0 {
	}
}
