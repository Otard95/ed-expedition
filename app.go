package main

import (
	"context"
	"ed-expedition/journal"
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

type App struct {
	ctx               context.Context
	logger            wailsLogger.Logger
	journalDir        string
	journalWatcher    *journal.Watcher
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

// startup is called by Wails. We save the context to enable runtime method calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := a.initServices(); err != nil {
		a.logger.Error(err.Error())
		os.Exit(1)
	}
	if err := a.startServices(); err != nil {
		a.logger.Error(err.Error())
		os.Exit(1)
	}

	a.initAvailablePlotters()

	a.jumpHistoryChan = a.expeditionService.JumpHistory.Subscribe()
	go func() {
		for event := range a.jumpHistoryChan {
			runtime.EventsEmit(ctx, "JumpHistory", *event)
			if next := a.expeditionService.GetNextSystemName(); next != nil {
				runtime.ClipboardSetText(ctx, *next)
			}
		}
	}()

	a.targetChan = a.journalWatcher.FSDTarget.Subscribe()
	go func() {
		for event := range a.targetChan {
			runtime.EventsEmit(ctx, "Target", *event)
		}
	}()

	a.completeExpeditionChan = a.expeditionService.CompleteExpedition.Subscribe()
	go func() {
		for event := range a.completeExpeditionChan {
			runtime.EventsEmit(ctx, "CompleteExpedition", *event)
		}
	}()

	a.currentJumpChan = a.expeditionService.CurrentJump.Subscribe()
	go func() {
		for event := range a.currentJumpChan {
			runtime.EventsEmit(ctx, "CurrentJump", *event)
		}
	}()

	a.fuelAlertChan = a.expeditionService.FuelAlert.Subscribe()
	go func() {
		for event := range a.fuelAlertChan {
			runtime.EventsEmit(ctx, "FuelAlert", *event)
		}
	}()

	a.jobStatusChan = a.jobService.JobStatus.Subscribe()
	go func() {
		for status := range a.jobStatusChan {
			runtime.EventsEmit(ctx, "job:"+status.ID, *status)
		}
	}()
}

func (a *App) initServices() error {
	a.logger.Info(fmt.Sprintf("[ed-expedition] starting watcher in '%s'", a.journalDir))

	journalWatcher, err := journal.NewWatcher(a.journalDir, a.logger)
	if err != nil {
		return fmt.Errorf("failed to watch journal directory: %w", err)
	}
	a.journalWatcher = journalWatcher

	stateService := services.NewAppStateService(journalWatcher, a.logger)
	a.stateService = stateService

	var lastKnownLocation int64
	if stateService.State.LastKnownLocation != nil {
		lastKnownLocation = stateService.State.LastKnownLocation.SystemID
	}

	a.expeditionService = services.NewExpeditionService(journalWatcher, a.logger, lastKnownLocation)

	a.galaxyService = services.NewGalaxyService(a.logger)
	a.jobService = services.NewJobService(a.logger)

	return nil
}

func (a *App) startServices() error {
	a.expeditionService.Start()
	a.stateService.Start()
	a.jobService.Start()

	if err := a.galaxyService.Start(); err != nil {
		return fmt.Errorf("failed to start galaxy service: %w", err)
	}

	if a.expeditionService.Index.ActiveExpeditionID != nil && a.stateService.State.LastKnownLocation != nil {
		err := a.journalWatcher.Sync(a.stateService.State.LastKnownLocation.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to sync journal: %w", err)
		}
	}

	a.logger.Info("[app.go] start journalWatcher")
	a.journalWatcher.Start()

	return nil
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
	if a.jumpHistoryChan != nil && a.expeditionService != nil {
		a.expeditionService.JumpHistory.Unsubscribe(a.jumpHistoryChan)
	}
	if a.targetChan != nil && a.journalWatcher != nil {
		a.journalWatcher.FSDTarget.Unsubscribe(a.targetChan)
	}
	if a.completeExpeditionChan != nil && a.expeditionService != nil {
		a.expeditionService.CompleteExpedition.Unsubscribe(a.completeExpeditionChan)
	}
	if a.currentJumpChan != nil && a.expeditionService != nil {
		a.expeditionService.CurrentJump.Unsubscribe(a.currentJumpChan)
	}
	if a.fuelAlertChan != nil && a.expeditionService != nil {
		a.expeditionService.FuelAlert.Unsubscribe(a.fuelAlertChan)
	}
	if a.jobStatusChan != nil && a.jobService != nil {
		a.jobService.JobStatus.Unsubscribe(a.jobStatusChan)
	}

	if a.jobService != nil {
		a.jobService.Stop()
	}
	if a.galaxyService != nil {
		a.galaxyService.Stop()
	}
	if a.stateService != nil {
		a.stateService.Stop()
	}
	if a.expeditionService != nil {
		a.expeditionService.Stop()
	}
	if a.journalWatcher != nil {
		a.journalWatcher.Close()
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

	switch a.stateService.State.GalaxyDecision {
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
	if err := a.stateService.AcceptGalaxy(); err != nil {
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

func (a *App) DeclineGalaxy() error {
	return a.stateService.DeclineGalaxy()
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

func (a *App) GetPlotterInputConfig(plotterId string) (plotters.PlotterInputConfig, error) {
	if plotter, ok := a.availablePlotters[plotterId]; ok {
		return plotter.InputConfig(), nil
	}

	return plotters.PlotterInputConfig{}, fmt.Errorf("Unknown plotter id '%s'", plotterId)
}

type plotRouteCtx struct {
	Route *models.Route
}

func (a *App) PlotRoute(expeditionId, plotterId, from, to string, inputs plotters.PlotterInputs) (string, error) {
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
