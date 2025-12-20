package services

import (
	"ed-expedition/journal"
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TestLogger implements wails Logger interface for testing
type TestLogger struct{}

func (l *TestLogger) Print(message string)   {}
func (l *TestLogger) Trace(message string)   {}
func (l *TestLogger) Debug(message string)   {}
func (l *TestLogger) Info(message string)    {}
func (l *TestLogger) Warning(message string) {}
func (l *TestLogger) Error(message string)   {}
func (l *TestLogger) Fatal(message string)   {}

// RecordingLogger implements wails Logger interface and records trace messages
type RecordingLogger struct {
	Messages []string
}

func (l *RecordingLogger) Print(message string)   {}
func (l *RecordingLogger) Trace(message string)   { l.Messages = append(l.Messages, message) }
func (l *RecordingLogger) Debug(message string)   {}
func (l *RecordingLogger) Info(message string)    {}
func (l *RecordingLogger) Warning(message string) {}
func (l *RecordingLogger) Error(message string)   {}
func (l *RecordingLogger) Fatal(message string)   {}

type ExpeditionServiceTestSuite struct {
	suite.Suite
	tmpDir     string
	watcher    *journal.Watcher
	service    *ExpeditionService
	distance   float64
	fuelUsed   float64
	fuelLevel  float64
	bakedIndex int
}

func (s *ExpeditionServiceTestSuite) SetupTest() {
	var err error
	s.tmpDir, err = os.MkdirTemp("", "journal-test-*")
	if err != nil {
		s.T().Fatalf("Failed to create temp dir: %v", err)
	}
	os.Setenv("ED_EXPEDITION_DATA_DIR", s.tmpDir)

	createActiveExpedition(s.T(), s.tmpDir, []Jump{
		{name: "Sol", id: 1},
		{name: "Alpha Centauri", id: 2},
		{name: "Bernard's Star", id: 3},
		{name: "Luhman 16", id: 4},
	})

	s.watcher, err = journal.NewWatcher(s.tmpDir, &TestLogger{})
	if err != nil {
		s.T().Fatalf("Failed to create watcher: %v", err)
	}

	s.service = NewExpeditionService(s.watcher, &TestLogger{})
	s.service.Start()
	s.watcher.Start()

	s.distance = 20
	s.fuelUsed = 2
	s.fuelLevel = 14
	s.bakedIndex = 0
}

func (s *ExpeditionServiceTestSuite) TearDownTest() {
	if s.service != nil {
		s.service.Stop()
	}
	if s.watcher != nil {
		s.watcher.Close()
	}
	if s.tmpDir != "" {
		os.RemoveAll(s.tmpDir)
	}
}

func (s *ExpeditionServiceTestSuite) TestCorrectlyRecordOnRouteJump() {
	// Assert initial state
	assert.NotNil(s.T(), s.service)
	assert.NotNil(s.T(), s.service.activeExpedition)
	assert.Len(s.T(), s.service.activeExpedition.JumpHistory, 0)
	assert.NotNil(s.T(), s.service.Index.ActiveExpeditionID)
	assert.Equal(s.T(), "active", *s.service.Index.ActiveExpeditionID)
	assert.NotNil(s.T(), s.service.bakedRoute)
	assert.Len(s.T(), s.service.bakedRoute.Jumps, 4)

	// Simulate journal updates
	s.bakedIndex = 1
	jumpTime := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	simulateJump(s.T(), s.tmpDir, Jump{name: "Alpha Centauri", id: 2, distance: &s.distance, fuelUsed: &s.fuelUsed, fuelLevel: &s.fuelLevel}, jumpTime)
	time.Sleep(10 * time.Millisecond)

	// Assert updates
	assert.Len(s.T(), s.service.activeExpedition.JumpHistory, 1)
	assert.EqualValues(s.T(), models.JumpHistoryEntry{
		Timestamp:  jumpTime,
		SystemName: "Alpha Centauri",
		SystemID:   2,
		BakedIndex: &s.bakedIndex,

		Distance:  s.distance,
		FuelUsed:  s.fuelUsed,
		FuelLevel: s.fuelLevel,

		Expected:  true,
		Synthetic: false,
	}, s.service.activeExpedition.JumpHistory[0])
}

func (s *ExpeditionServiceTestSuite) TestCorrectlyRecordJumpToStartSystem() {
	// Assert initial state
	assert.NotNil(s.T(), s.service)
	assert.NotNil(s.T(), s.service.activeExpedition)
	assert.Len(s.T(), s.service.activeExpedition.JumpHistory, 0)
	assert.NotNil(s.T(), s.service.Index.ActiveExpeditionID)
	assert.Equal(s.T(), "active", *s.service.Index.ActiveExpeditionID)
	assert.NotNil(s.T(), s.service.bakedRoute)
	assert.Len(s.T(), s.service.bakedRoute.Jumps, 4)

	// Simulate journal updates
	jumpTime := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	simulateJump(s.T(), s.tmpDir, Jump{name: "Sol", id: 1, distance: &s.distance, fuelUsed: &s.fuelUsed, fuelLevel: &s.fuelLevel}, jumpTime)
	time.Sleep(10 * time.Millisecond)

	// Assert updates
	assert.Len(s.T(), s.service.activeExpedition.JumpHistory, 1)
	assert.EqualValues(s.T(), models.JumpHistoryEntry{
		Timestamp:  jumpTime,
		SystemName: "Sol",
		SystemID:   1,
		BakedIndex: &s.bakedIndex,

		Distance:  s.distance,
		FuelUsed:  s.fuelUsed,
		FuelLevel: s.fuelLevel,

		Expected:  true,
		Synthetic: false,
	}, s.service.activeExpedition.JumpHistory[0])
}

func (s *ExpeditionServiceTestSuite) TestCorrectlyRecordDetour() {
	// Jump to a system that's not in the route at all
	jumpTime := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	simulateJump(s.T(), s.tmpDir, Jump{name: "Betelgeuse", id: 999, distance: &s.distance, fuelUsed: &s.fuelUsed, fuelLevel: &s.fuelLevel}, jumpTime)
	time.Sleep(10 * time.Millisecond)

	// Assert detour was recorded
	assert.Len(s.T(), s.service.activeExpedition.JumpHistory, 1)
	jump := s.service.activeExpedition.JumpHistory[0]
	assert.Equal(s.T(), "Betelgeuse", jump.SystemName)
	assert.Equal(s.T(), int64(999), jump.SystemID)
	assert.Nil(s.T(), jump.BakedIndex)
	assert.Equal(s.T(), false, jump.Expected)
	assert.Equal(s.T(), false, jump.Synthetic)
	assert.Equal(s.T(), 0, s.service.activeExpedition.CurrentBakedIndex)
}

func (s *ExpeditionServiceTestSuite) TestCorrectlyRecordDetourThenExpected() {
	// First jump: detour to system not in route
	jumpTime1 := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	simulateJump(s.T(), s.tmpDir, Jump{name: "Betelgeuse", id: 999, distance: &s.distance, fuelUsed: &s.fuelUsed, fuelLevel: &s.fuelLevel}, jumpTime1)
	time.Sleep(10 * time.Millisecond)

	// Second jump: expected system (Alpha Centauri)
	s.bakedIndex = 1
	jumpTime2 := time.Date(2025, 12, 20, 10, 5, 0, 0, time.UTC)
	simulateJump(s.T(), s.tmpDir, Jump{name: "Alpha Centauri", id: 2, distance: &s.distance, fuelUsed: &s.fuelUsed, fuelLevel: &s.fuelLevel}, jumpTime2)
	time.Sleep(10 * time.Millisecond)

	// Assert both jumps recorded correctly
	assert.Len(s.T(), s.service.activeExpedition.JumpHistory, 2)

	// First jump: detour
	assert.Equal(s.T(), "Betelgeuse", s.service.activeExpedition.JumpHistory[0].SystemName)
	assert.Equal(s.T(), false, s.service.activeExpedition.JumpHistory[0].Expected)
	assert.Nil(s.T(), s.service.activeExpedition.JumpHistory[0].BakedIndex)

	// Second jump: expected
	assert.EqualValues(s.T(), models.JumpHistoryEntry{
		Timestamp:  jumpTime2,
		SystemName: "Alpha Centauri",
		SystemID:   2,
		BakedIndex: &s.bakedIndex,
		Distance:   s.distance,
		FuelUsed:   s.fuelUsed,
		FuelLevel:  s.fuelLevel,
		Expected:   true,
		Synthetic:  false,
	}, s.service.activeExpedition.JumpHistory[1])
	assert.Equal(s.T(), 1, s.service.activeExpedition.CurrentBakedIndex)
}

func (s *ExpeditionServiceTestSuite) TestCorrectlyRecordDetourToOnRouteButNotExpected() {
	// Jump to Bernard's Star (index 2), skipping Alpha Centauri (index 1)
	s.bakedIndex = 2
	jumpTime := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	simulateJump(s.T(), s.tmpDir, Jump{name: "Bernard's Star", id: 3, distance: &s.distance, fuelUsed: &s.fuelUsed, fuelLevel: &s.fuelLevel}, jumpTime)
	time.Sleep(10 * time.Millisecond)

	// Assert jump recorded as on-route but not expected
	assert.Len(s.T(), s.service.activeExpedition.JumpHistory, 1)
	assert.EqualValues(s.T(), models.JumpHistoryEntry{
		Timestamp:  jumpTime,
		SystemName: "Bernard's Star",
		SystemID:   3,
		BakedIndex: &s.bakedIndex,
		Distance:   s.distance,
		FuelUsed:   s.fuelUsed,
		FuelLevel:  s.fuelLevel,
		Expected:   false,
		Synthetic:  false,
	}, s.service.activeExpedition.JumpHistory[0])
	assert.Equal(s.T(), 2, s.service.activeExpedition.CurrentBakedIndex)
}

func (s *ExpeditionServiceTestSuite) TestAutoCompleteWhenReachingLastJump() {
	// Jump through all systems in order
	jumps := []Jump{
		{name: "Alpha Centauri", id: 2},
		{name: "Bernard's Star", id: 3},
		{name: "Luhman 16", id: 4},
	}

	// Store expedition pointer before last jump (it will be cleared after completion)
	var expedition *models.Expedition

	baseTime := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	for i, jump := range jumps {
		jumpTime := baseTime.Add(time.Duration(i*5) * time.Minute)

		// Capture expedition pointer before last jump
		if i == len(jumps)-1 {
			expedition = s.service.activeExpedition
		}

		simulateJump(s.T(), s.tmpDir, Jump{name: jump.name, id: jump.id, distance: &s.distance, fuelUsed: &s.fuelUsed, fuelLevel: &s.fuelLevel}, jumpTime)
		time.Sleep(10 * time.Millisecond)
	}

	// Verify all jumps recorded
	assert.Len(s.T(), expedition.JumpHistory, 3)

	// Verify expedition auto-completed
	assert.Equal(s.T(), models.StatusCompleted, expedition.Status)
	assert.Equal(s.T(), 3, expedition.CurrentBakedIndex)

	// Verify active expedition cleared from service
	assert.Nil(s.T(), s.service.activeExpedition)

	// Verify active expedition cleared from index
	assert.Nil(s.T(), s.service.Index.ActiveExpeditionID)

	// Verify expedition summary in index is updated to completed
	expeditionInIndex := slice.Find(s.service.Index.Expeditions, func(e models.ExpeditionSummary) bool {
		return e.ID == "active"
	})
	assert.NotNil(s.T(), expeditionInIndex)
	assert.Equal(s.T(), models.StatusCompleted, expeditionInIndex.Status)
}

func TestExpeditionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ExpeditionServiceTestSuite))
}

type Jump struct {
	name      string
	id        int64
	distance  *float64
	fuelUsed  *float64
	fuelLevel *float64
}

func createActiveExpedition(t *testing.T, dir string, jumps []Jump) {
	t.Helper()

	os.MkdirAll(path.Join(dir, "expeditions"), 0700)
	os.MkdirAll(path.Join(dir, "routes"), 0700)

	filePath := filepath.Join(dir, "routes", "baked-route.json")
	if err := os.WriteFile(filePath, []byte(buildRouteJson(t, "baked-route", jumps)), 0600); err != nil {
		t.Fatalf("Failed to create route file %s: %v", filePath, err)
	}

	filePath = filepath.Join(dir, "expeditions", "active.json")
	if err := os.WriteFile(filePath, []byte(buildExpeditionJson(t, "active", "baked-route")), 0600); err != nil {
		t.Fatalf("Failed to create expedition file %s: %v", filePath, err)
	}

	filePath = filepath.Join(dir, "index.json")
	if err := os.WriteFile(filePath, []byte(buildIndexJson(t, "active")), 0600); err != nil {
		t.Fatalf("Failed to create index file %s: %v", filePath, err)
	}
}

func buildIndexJson(t *testing.T, id string) string {
	t.Helper()
	return `{
  "active_expedition_id": "` + id + `",
  "expeditions": [{
    "id": "` + id + `",
    "name": "Active",
    "status": "active",
    "created_at": "2025-12-15T16:14:10.296462334+01:00",
    "last_updated": "2025-12-18T09:46:01.027582275+01:00"
  }]
}`
}
func buildExpeditionJson(t *testing.T, id, bakedRouteId string) string {
	t.Helper()
	return `{
  "id": "` + id + `",
  "name": "Active",
  "created_at": "2025-12-12T10:00:00Z",
  "last_updated": "2025-12-16T13:44:55.493970202+01:00",
  "status": "active",
  "start": {
    "route_id": "route-001-colonia-highway-north",
    "jump_index": 0
  },
  "routes": [ "route-001-colonia-highway-north" ],
  "links": [ ],
  "baked_route_id": "` + bakedRouteId + `",
  "current_baked_index": 0,
  "current_baked_index": 0,
  "jump_history": []
}`
}
func buildRouteJson(t *testing.T, id string, jumps []Jump) string {
	t.Helper()
	return `{
  "id": "` + id + `",
  "name": "Route",
  "plotter": "spansh",
  "plotter_parameters": {},
  "plotter_metadata": {},
  "jumps": [` + strings.Join(slice.Map(
		jumps,
		func(j *Jump) string {
			return `{"system_name": "` + j.name + `", "system_id":` + strconv.FormatInt(j.id, 10) + `, "scoopable": true, "must_refuel": false, "distance": 0, "position": {"x": -1111.56, "y": -134.22, "z": 65269.75}}`
		},
	), ",") + `],
  "created_at": "2025-12-11T11:30:00Z"
}`
}

func simulateJump(t *testing.T, dir string, jump Jump, timestamp time.Time) {
	t.Helper()

	journalFile := filepath.Join(dir, "Journal.2025-12-20T100000.01.json")

	distance := 0.0
	if jump.distance != nil {
		distance = *jump.distance
	}
	fuelUsed := 0.0
	if jump.fuelUsed != nil {
		fuelUsed = *jump.fuelUsed
	}
	fuelLevel := 0.0
	if jump.fuelLevel != nil {
		fuelLevel = *jump.fuelLevel
	}

	event := `{"timestamp":"` + timestamp.UTC().Format(time.RFC3339) + `","event":"FSDJump","Taxi":false,"Multicrew":false,"StarSystem":"` + jump.name + `","SystemAddress":` + strconv.FormatInt(jump.id, 10) + `,"StarPos":[0,0,0],"SystemAllegiance":"","SystemEconomy":"","SystemEconomy_Localised":"","SystemSecondEconomy":"","SystemSecondEconomy_Localised":"","SystemGovernment":"","SystemGovernment_Localised":"","SystemSecurity":"","SystemSecurity_Localised":"","Population":0,"Body":"","BodyID":0,"BodyType":"","JumpDist":` + strconv.FormatFloat(distance, 'f', 2, 64) + `,"FuelUsed":` + strconv.FormatFloat(fuelUsed, 'f', 2, 64) + `,"FuelLevel":` + strconv.FormatFloat(fuelLevel, 'f', 2, 64) + `}` + "\n"

	file, err := os.OpenFile(journalFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open journal file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(event); err != nil {
		t.Fatalf("Failed to write to journal file: %v", err)
	}
}
