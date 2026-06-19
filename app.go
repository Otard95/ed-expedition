package main

import (
	"context"
	"ed-expedition/journal"
	"ed-expedition/lib/form"
	"ed-expedition/lib/fs"
	"ed-expedition/lib/job"
	"ed-expedition/lib/vec"
	"ed-expedition/models"
	"ed-expedition/plotters"
	"ed-expedition/services"
	"fmt"
	"os"
	"strings"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type settingDef struct {
	Config form.InputFieldConfig
	Get    func() string
	Apply  func(value string) error
}

type App struct {
	ctx               context.Context
	logger            wailsLogger.Logger
	journalDir        string
	journalWatcher    *journal.Watcher
	settings          *models.Settings
	settingDefs       []settingDef
	stateService      *services.AppStateService
	expeditionService *services.ExpeditionService
	galaxyService     *services.GalaxyService
	jobService        *services.JobService
	availablePlotters map[string]plotters.Plotter

	targetChan             chan *journal.FSDTargetEvent
	jumpHistoryChan        chan *models.JumpHistoryEntry
	completeExpeditionChan chan *models.Expedition
	currentJumpChan        chan *models.JumpHistoryEntry
	fuelAlertChan          chan *services.FuelAlert
	jobStatusChan          chan *job.JobStatus
}

func NewApp(
	logger wailsLogger.Logger,
	journalDir string,
) *App {
	return &App{logger: logger, journalDir: journalDir}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	settings, err := models.LoadSettings()
	if err != nil {
		a.logger.Error(fmt.Sprintf("failed to load settings: %v", err))
		os.Exit(1)
	}
	a.settings = settings
	a.journalDir = a.resolveJournalDir()

	if err := a.initCoreServices(); err != nil {
		a.logger.Error(err.Error())
		os.Exit(1)
	}
	if err := a.startCoreServices(); err != nil {
		a.logger.Error(err.Error())
		os.Exit(1)
	}

	a.initAvailablePlotters()

	if a.journalDir != "" {
		if err := a.startupJournalServices(); err != nil {
			a.logger.Error(err.Error())
			os.Exit(1)
		}
	}
}

func (a *App) resolveJournalDir() string {
	if a.journalDir != "" {
		// -j was provided; save to settings only if not already set
		if a.settings.JournalDir == nil {
			a.settings.JournalDir = &a.journalDir
			_ = models.SaveSettings(a.settings)
		}
		return a.journalDir
	}

	if a.settings.JournalDir != nil {
		return *a.settings.JournalDir
	}

	detected := journal.DetectJournalDir()
	if detected != "" {
		a.settings.JournalDir = &detected
		_ = models.SaveSettings(a.settings)
		return detected
	}

	return ""
}

func (a *App) getSettingDefs() []settingDef {
	if a.settingDefs != nil {
		return a.settingDefs
	}
	a.settingDefs = []settingDef{
		{
			Config: form.InputFieldConfig{
				Name:    "journal_dir",
				Label:   "Journal Directory",
				Type:    form.DirectoryInput,
				Section: "General",
				Info:    "Path to the Elite Dangerous journal directory.",
			},
			Get: func() string {
				if a.settings.JournalDir != nil {
					return *a.settings.JournalDir
				}
				return ""
			},
			Apply: func(value string) error {
				return a.SetJournalDir(value)
			},
		},
		{
			Config: form.InputFieldConfig{
				Name:    "galaxy_decision",
				Label:   "Galaxy Database",
				Type:    form.StringInput,
				Section: "General",
				Info:    "Whether to download the full galaxy database for the built-in plotter.",
				Options: []form.InputOption{
					{Value: string(models.GalaxyAccepted), Label: "Accepted"},
					{Value: string(models.GalaxyDeclined), Label: "Declined"},
				},
			},
			Get: func() string {
				return string(a.settings.GalaxyDecision)
			},
			Apply: func(value string) error {
				switch models.GalaxyDecision(value) {
				case models.GalaxyAccepted, models.GalaxyDeclined:
					prev := a.settings.GalaxyDecision
					a.settings.GalaxyDecision = models.GalaxyDecision(value)
					if err := models.SaveSettings(a.settings); err != nil {
						return err
					}
					if prev != models.GalaxyDecision(value) {
						runtime.WindowReloadApp(a.ctx)
					}
					return nil
				default:
					return fmt.Errorf("invalid galaxy decision: %s", value)
				}
			},
		},
		{
			Config: form.InputFieldConfig{
				Name:    "debug",
				Label:   "Debug Mode",
				Type:    form.BoolInput,
				Section: "Advanced",
				Info:    "Show additional debug information such as route and jump metadata.",
			},
			Get: func() string {
				return form.EncodeBool(a.settings.Debug)
			},
			Apply: func(value string) error {
				a.settings.Debug = value == "1"
				return models.SaveSettings(a.settings)
			},
		},
	}
	return a.settingDefs
}

func (a *App) GetSettingsConfig() []form.InputFieldConfig {
	defs := a.getSettingDefs()
	configs := make([]form.InputFieldConfig, len(defs))
	for i, def := range defs {
		configs[i] = def.Config
		configs[i].Default = def.Get()
	}
	return configs
}

func (a *App) UpdateSetting(key, value string) error {
	for _, def := range a.getSettingDefs() {
		if def.Config.Name == key {
			return def.Apply(value)
		}
	}
	return fmt.Errorf("unknown setting: %s", key)
}

func (a *App) initCoreServices() error {
	a.stateService = services.NewAppStateService(a.logger)

	var lastKnownLocation int64
	if a.stateService.State.LastKnownLocation != nil {
		lastKnownLocation = a.stateService.State.LastKnownLocation.SystemID
	}
	a.expeditionService = services.NewExpeditionService(a.logger, lastKnownLocation)

	a.galaxyService = services.NewGalaxyService(a.logger)
	a.jobService = services.NewJobService(a.logger)

	return nil
}

func (a *App) startCoreServices() error {
	a.jobService.Start()

	if err := a.galaxyService.Start(); err != nil {
		return fmt.Errorf("failed to start galaxy service: %w", err)
	}

	a.jobStatusChan = a.jobService.JobStatus.Subscribe()
	go func() {
		for status := range a.jobStatusChan {
			runtime.EventsEmit(a.ctx, "job:"+status.ID, *status)
		}
	}()

	return nil
}

func (a *App) startupJournalServices() error {
	a.logger.Info(fmt.Sprintf("[ed-expedition] watching journals in '%s'", a.journalDir))

	watcher, err := journal.NewWatcher(a.journalDir, a.logger)
	if err != nil {
		return fmt.Errorf("failed to watch journal directory: %w", err)
	}
	a.journalWatcher = watcher

	a.stateService.SetWatcher(watcher)
	a.stateService.Start()

	a.expeditionService.SetWatcher(watcher)
	a.expeditionService.Start()

	a.jumpHistoryChan = a.expeditionService.JumpHistory.Subscribe()
	go func() {
		for event := range a.jumpHistoryChan {
			runtime.EventsEmit(a.ctx, "JumpHistory", *event)
			if next := a.expeditionService.GetNextSystemName(); next != nil {
				runtime.ClipboardSetText(a.ctx, *next)
			}
		}
	}()

	a.targetChan = watcher.FSDTarget.Subscribe()
	go func() {
		for event := range a.targetChan {
			runtime.EventsEmit(a.ctx, "Target", *event)
		}
	}()

	a.completeExpeditionChan = a.expeditionService.CompleteExpedition.Subscribe()
	go func() {
		for event := range a.completeExpeditionChan {
			runtime.EventsEmit(a.ctx, "CompleteExpedition", *event)
		}
	}()

	a.currentJumpChan = a.expeditionService.CurrentJump.Subscribe()
	go func() {
		for event := range a.currentJumpChan {
			runtime.EventsEmit(a.ctx, "CurrentJump", *event)
		}
	}()

	a.fuelAlertChan = a.expeditionService.FuelAlert.Subscribe()
	go func() {
		for event := range a.fuelAlertChan {
			runtime.EventsEmit(a.ctx, "FuelAlert", *event)
		}
	}()

	if a.expeditionService.Index.ActiveExpeditionID != nil && a.stateService.State.JournalSync != nil {
		if err := watcher.Sync(*a.stateService.State.JournalSync); err != nil {
			return fmt.Errorf("failed to sync journal: %w", err)
		}
	}

	a.logger.Info("[app.go] start journalWatcher")
	watcher.Start()

	return nil
}

func (a *App) teardownJournalServices() {
	if a.journalWatcher == nil {
		return
	}

	if a.jumpHistoryChan != nil {
		a.expeditionService.JumpHistory.Unsubscribe(a.jumpHistoryChan)
		a.jumpHistoryChan = nil
	}
	if a.targetChan != nil {
		a.journalWatcher.FSDTarget.Unsubscribe(a.targetChan)
		a.targetChan = nil
	}
	if a.completeExpeditionChan != nil {
		a.expeditionService.CompleteExpedition.Unsubscribe(a.completeExpeditionChan)
		a.completeExpeditionChan = nil
	}
	if a.currentJumpChan != nil {
		a.expeditionService.CurrentJump.Unsubscribe(a.currentJumpChan)
		a.currentJumpChan = nil
	}
	if a.fuelAlertChan != nil {
		a.expeditionService.FuelAlert.Unsubscribe(a.fuelAlertChan)
		a.fuelAlertChan = nil
	}

	a.stateService.Stop()
	a.expeditionService.Stop()

	a.journalWatcher.Close()
	a.journalWatcher = nil
}

func (a *App) initAvailablePlotters() {
	a.availablePlotters = map[string]plotters.Plotter{
		"spansh_galaxy_plotter": plotters.SpanshGalaxyPlotter{},
	}

	if a.galaxyService.State() == services.GalaxyStateReady {
		a.availablePlotters["basic_plotter"] = plotters.BasicPlotter{GalaxyQuerier: a.galaxyService}
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.teardownJournalServices()

	if a.jobStatusChan != nil && a.jobService != nil {
		a.jobService.JobStatus.Unsubscribe(a.jobStatusChan)
	}
	if a.jobService != nil {
		a.jobService.Stop()
	}
	if a.galaxyService != nil {
		a.galaxyService.Stop()
	}
}

type GalaxyStatus string

const (
	GalaxyStatusPrompt         GalaxyStatus = "prompt"
	GalaxyStatusPromptContinue GalaxyStatus = "prompt_continue"
	GalaxyStatusUnavailable    GalaxyStatus = "unavailable"
	GalaxyStatusInProgress     GalaxyStatus = "in_progress"
	GalaxyStatusReady          GalaxyStatus = "ready"
)

var AllGalaxyStatus = []struct {
	Value  GalaxyStatus
	TSName string
}{
	{GalaxyStatusPrompt, "PROMPT"},
	{GalaxyStatusPromptContinue, "PROMPT_CONTINUE"},
	{GalaxyStatusUnavailable, "UNAVAILABLE"},
	{GalaxyStatusInProgress, "IN_PROGRESS"},
	{GalaxyStatusReady, "READY"},
}

func (a *App) GetGalaxyState() GalaxyStatus {
	if a.galaxyService.State() == services.GalaxyStateReady {
		return GalaxyStatusReady
	}

	switch a.settings.GalaxyDecision {
	case models.GalaxyNotAsked:
		return GalaxyStatusPrompt
	case models.GalaxyDeclined:
		return GalaxyStatusUnavailable
	case models.GalaxyAccepted:
		if a.galaxyService.Running() {
			return GalaxyStatusInProgress
		}
		return GalaxyStatusPromptContinue
	default:
		return GalaxyStatusPrompt
	}
}

func (a *App) AcceptGalaxy() (string, error) {
	a.settings.GalaxyDecision = models.GalaxyAccepted
	if err := models.SaveSettings(a.settings); err != nil {
		return "", err
	}

	now := time.Now()
	a.stateService.State.GalaxyDownloadedAt = &now
	if err := models.SaveAppState(a.stateService.State); err != nil {
		return "", err
	}

	return a.runGalaxyBuild(), nil
}

func (a *App) ContinueGalaxyBuild() string {
	return a.runGalaxyBuild()
}

func (a *App) runGalaxyBuild() string {
	j := a.galaxyService.BuildJob()
	if j == nil {
		return ""
	}
	a.jobService.RegisterAndRun(j, a.ctx)
	return j.Id()
}

func (a *App) DeclineGalaxy() error {
	a.settings.GalaxyDecision = models.GalaxyDeclined
	return models.SaveSettings(a.settings)
}

func (a *App) MockJob(durationSeconds int) string {
	phaseDuration := time.Duration(durationSeconds) * time.Second / 3

	j := job.New(
		"Mock Job",
		job.NoCtx{},
		[]job.PhaseConfig[job.NoCtx]{
			{
				Name:  "download",
				Label: "Downloading",
				Type:  job.PhaseTypeObservable,
				Callback: func(ctx context.Context, state *job.NoCtx, tracker *job.ProgressTracker) error {
					steps := 100
					tracker.SetTotal(float64(steps))
					for i := range steps {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-time.After(phaseDuration / time.Duration(steps)):
							tracker.SetProgress(float64(i + 1))
						}
					}
					return nil
				},
			},
			{
				Name:  "process",
				Label: "Processing",
				Type:  job.PhaseTypeObservable,
				Callback: func(ctx context.Context, state *job.NoCtx, tracker *job.ProgressTracker) error {
					steps := 50
					tracker.SetTotal(float64(steps))
					for i := range steps {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-time.After(phaseDuration / time.Duration(steps)):
							tracker.SetProgress(float64(i + 1))
						}
					}
					return nil
				},
			},
			{
				Name:  "finalize",
				Label: "Finalizing",
				Type:  job.PhaseTypeEstimated,
				EstimateCallback: func(completed map[string]time.Duration) time.Duration {
					if d, ok := completed["process"]; ok {
						return d
					}
					return phaseDuration
				},
				Callback: func(ctx context.Context, state *job.NoCtx, tracker *job.ProgressTracker) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(phaseDuration):
						return nil
					}
				},
			},
		},
		func(state job.NoCtx) (any, error) { return "mock complete", nil },
		a.logger,
	)

	a.jobService.RegisterAndRun(j, a.ctx)
	return j.Id()
}

type SystemValidation struct {
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
}

func (a *App) ValidateSystemName(name string) SystemValidation {
	canonical, valid, err := a.galaxyService.ValidateSystemName(name)
	if err != nil {
		return SystemValidation{}
	}
	return SystemValidation{Name: canonical, Valid: valid}
}

func (a *App) AutocompleteSystems(prefix string) []string {
	names, err := a.galaxyService.AutocompleteSystems(prefix, 10)
	if err != nil {
		return nil
	}
	return names
}

func (a *App) DebugHilbertGroups(x, y, z, radius float64, useParallelQueries bool) *services.HilbertGroupDebug {
	return a.galaxyService.DebugHilbertGroups(vec.NewVec3(x, y, z), radius, useParallelQueries)
}

func (a *App) GetJournalDirStatus() bool {
	return a.journalWatcher != nil
}

func (a *App) BrowseDirectory(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

func (a *App) SetJournalDir(path string) error {
	if !fs.IsDir(path) {
		return fmt.Errorf("invalid directory: %s", path)
	}

	a.settings.JournalDir = &path
	if err := models.SaveSettings(a.settings); err != nil {
		return fmt.Errorf("failed to save journal dir: %w", err)
	}

	a.journalDir = path
	a.teardownJournalServices()
	return a.startupJournalServices()
}

func (a *App) GetExpeditionSummaries() []models.ExpeditionSummary {
	return a.expeditionService.Index.Expeditions
}

func (a *App) CreateExpedition() (string, error) {
	return a.expeditionService.CreateExpedition()
}

func (a *App) LoadExpedition(id string) (*models.Expedition, error) {
	return models.LoadExpedition(id)
}

func (a *App) LoadRoutes(expeditionId string) ([]*models.Route, error) {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return nil, err
	}

	routes, err := expedition.LoadRoutes()
	if err != nil {
		return nil, err
	}

	return routes, nil
}

func (a *App) GetPlotterOptions() map[string]string {
	options := make(map[string]string, len(a.availablePlotters))

	for k, v := range a.availablePlotters {
		options[k] = v.String()
	}

	return options
}

func (a *App) GetPlotterInputConfig(plotterId string) (form.InputConfig, error) {
	if plotter, ok := a.availablePlotters[plotterId]; ok {
		return plotter.InputConfig(), nil
	}

	return form.InputConfig{}, fmt.Errorf("Unknown plotter id '%s'", plotterId)
}

func (a *App) GetLoadout() *models.Loadout {
	return a.stateService.State.LastKnownLoadout
}

type plotRouteCtx struct {
	Route *models.Route
}

func (a *App) PlotRoute(expeditionId, plotterId, from, to string, inputs form.InputValues) (string, error) {
	plotter, ok := a.availablePlotters[plotterId]
	if !ok {
		return "", fmt.Errorf("Unknown plotter id '%s'", plotterId)
	}

	loadout := a.stateService.State.LastKnownLoadout
	if loadout == nil {
		return "", fmt.Errorf("No ship loadout available - please load game first")
	}

	if a.galaxyService.State() == services.GalaxyStateReady {
		canonicalFrom, validFrom, _ := a.galaxyService.ValidateSystemName(from)
		canonicalTo, validTo, _ := a.galaxyService.ValidateSystemName(to)
		if validFrom {
			from = canonicalFrom
		}
		if validTo {
			to = canonicalTo
		}

		if !validFrom || !validTo {
			var invalid []string
			if !validFrom {
				invalid = append(invalid, fmt.Sprintf("'%s'", from))
			}
			if !validTo {
				invalid = append(invalid, fmt.Sprintf("'%s'", to))
			}
			return "", fmt.Errorf("unknown system(s): %s", strings.Join(invalid, ", "))
		}
	}

	j := job.New("Plot Route", plotRouteCtx{}, []job.PhaseConfig[plotRouteCtx]{
		{
			Name:  "plot",
			Label: fmt.Sprintf("%s → %s", from, to),
			Type:  plotter.ProgressType(),
			Callback: func(ctx context.Context, state *plotRouteCtx, tracker *job.ProgressTracker) error {
				route, err := plotter.Plot(from, to, inputs, loadout, a.logger, tracker)
				if err != nil {
					return err
				}
				state.Route = route
				return nil
			},
		},
	}, func(state plotRouteCtx) (*models.Route, error) {
		if err := a.expeditionService.AddRouteToExpedition(expeditionId, state.Route); err != nil {
			return nil, fmt.Errorf("failed to add route to expedition: %w", err)
		}
		return state.Route, nil
	}, a.logger)

	a.jobService.RegisterAndRun(j, a.ctx)
	return j.Id(), nil
}

func (a *App) DeleteExpedition(id string) error {
	return a.expeditionService.DeleteExpedition(id)
}

func (a *App) RenameExpedition(id, name string) error {
	return a.expeditionService.RenameExpedition(id, name)
}

func (a *App) RemoveRouteFromExpedition(expeditionId, routeId string) error {
	return a.expeditionService.RemoveRouteFromExpedition(expeditionId, routeId)
}

func (a *App) CreateLink(expeditionId string, from, to models.RoutePosition) error {
	return a.expeditionService.CreateLink(expeditionId, from, to)
}

func (a *App) DeleteLink(expeditionId, linkId string) error {
	return a.expeditionService.DeleteLink(expeditionId, linkId)
}

func (a *App) StartExpedition(expeditionId string) error {
	var currentSystemId *int64
	if a.stateService.State.LastKnownLocation != nil {
		currentSystemId = &a.stateService.State.LastKnownLocation.SystemID
	}
	return a.expeditionService.StartExpedition(expeditionId, currentSystemId)
}

func (a *App) EndActiveExpedition() error {
	return a.expeditionService.EndActiveExpedition(nil)
}

type LoadActiveExpeditionPayload struct {
	Expedition *models.Expedition
	BakedRoute *models.Route
}

func (a *App) LoadActiveExpedition() (*LoadActiveExpeditionPayload, error) {
	expedition, err := a.expeditionService.Index.LoadActiveExpedition()
	if err != nil {
		return nil, err
	}
	if expedition == nil {
		return nil, fmt.Errorf("no active expedition")
	}

	bakedRoute, err := expedition.LoadBaked()
	if err != nil {
		return nil, err
	}

	return &LoadActiveExpeditionPayload{
		Expedition: expedition,
		BakedRoute: bakedRoute,
	}, nil
}
