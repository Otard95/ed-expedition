package plotters

import (
	"ed-expedition/models"
	"testing"

	"github.com/stretchr/testify/suite"
)

// --- calculateMinFSDBoost ---

type CalculateMinFSDBoostSuite struct {
	suite.Suite
	maxRange float64
}

func TestCalculateMinFSDBoostSuite(t *testing.T) {
	suite.Run(t, new(CalculateMinFSDBoostSuite))
}

func (s *CalculateMinFSDBoostSuite) SetupSuite() {
	s.maxRange = 100.0
}

// Basic injection: just over max range up to 1.25x
func (s *CalculateMinFSDBoostSuite) TestBasicJustOverMaxRange() {
	s.Equal(models.FSDBoostInjectionBasic, calculateMinFSDBoost(100.1, s.maxRange))
}
func (s *CalculateMinFSDBoostSuite) TestBasicCeilingExact() {
	s.Equal(models.FSDBoostInjectionBasic, calculateMinFSDBoost(125.0, s.maxRange))
}

// Standard injection: just over 1.25x up to 1.50x
func (s *CalculateMinFSDBoostSuite) TestStandardJustOverBasicCeiling() {
	s.Equal(models.FSDBoostInjectionStandard, calculateMinFSDBoost(125.1, s.maxRange))
}
func (s *CalculateMinFSDBoostSuite) TestStandardCeilingExact() {
	s.Equal(models.FSDBoostInjectionStandard, calculateMinFSDBoost(150.0, s.maxRange))
}

// Premium injection: just over 1.50x up to 2.00x
func (s *CalculateMinFSDBoostSuite) TestPremiumJustOverStandardCeiling() {
	s.Equal(models.FSDBoostInjectionPremium, calculateMinFSDBoost(150.1, s.maxRange))
}
func (s *CalculateMinFSDBoostSuite) TestPremiumCeilingExact() {
	s.Equal(models.FSDBoostInjectionPremium, calculateMinFSDBoost(200.0, s.maxRange))
}
func (s *CalculateMinFSDBoostSuite) TestPremiumMidRange() {
	s.Equal(models.FSDBoostInjectionPremium, calculateMinFSDBoost(175.0, s.maxRange))
}

// Beyond premium — impossible jump, falls back to neutron
func (s *CalculateMinFSDBoostSuite) TestBeyondPremiumFallsBackToNeutron() {
	s.Equal(models.FSDBoostNeutron, calculateMinFSDBoost(200.1, s.maxRange))
}

// --- computeFSDBoostForRoute ---

type ComputeFSDBoostForRouteSuite struct {
	suite.Suite
	maxRange float64
}

func TestComputeFSDBoostForRouteSuite(t *testing.T) {
	suite.Run(t, new(ComputeFSDBoostForRouteSuite))
}

func (s *ComputeFSDBoostForRouteSuite) SetupSuite() {
	s.maxRange = 100.0
}

func (s *ComputeFSDBoostForRouteSuite) makeRoute(distances []float64, boosts []*models.FSDBoost) *models.Route {
	jumps := make([]models.RouteJump, len(distances))
	for i, d := range distances {
		jumps[i] = models.RouteJump{Distance: d, FSDBoost: boosts[i]}
	}
	return &models.Route{Jumps: jumps}
}

func (s *ComputeFSDBoostForRouteSuite) nils(n int) []*models.FSDBoost {
	return make([]*models.FSDBoost, n)
}

func (s *ComputeFSDBoostForRouteSuite) TestNoBoostWithinRange() {
	route := s.makeRoute([]float64{0, 80, 90}, s.nils(3))
	computeFSDBoostForRoute(route, s.maxRange)
	s.Nil(route.Jumps[0].FSDBoost)
	s.Nil(route.Jumps[1].FSDBoost)
	s.Nil(route.Jumps[2].FSDBoost)
}

func (s *ComputeFSDBoostForRouteSuite) TestAnnotatesOnDepartureSystem() {
	route := s.makeRoute([]float64{0, 110}, s.nils(2))
	computeFSDBoostForRoute(route, s.maxRange)
	s.Equal(models.FSDBoostInjectionBasic, *route.Jumps[0].FSDBoost)
}

func (s *ComputeFSDBoostForRouteSuite) TestLastJumpNeverAnnotated() {
	route := s.makeRoute([]float64{0, 80, 150}, s.nils(3))
	computeFSDBoostForRoute(route, s.maxRange)
	s.Nil(route.Jumps[2].FSDBoost)
}

func (s *ComputeFSDBoostForRouteSuite) TestSkipsJumpsWithBoostAlreadySet() {
	neutron := models.FSDBoostNeutron
	route := s.makeRoute([]float64{0, 150}, []*models.FSDBoost{&neutron, nil})
	computeFSDBoostForRoute(route, s.maxRange)
	s.Equal(models.FSDBoostNeutron, *route.Jumps[0].FSDBoost)
}

func (s *ComputeFSDBoostForRouteSuite) TestMixedRoute() {
	neutron := models.FSDBoostNeutron
	route := s.makeRoute(
		[]float64{0, 80, 110, 120, 175},
		[]*models.FSDBoost{nil, &neutron, nil, nil, nil},
	)
	computeFSDBoostForRoute(route, s.maxRange)
	s.Nil(route.Jumps[0].FSDBoost)                                     // next 80 — within range
	s.Equal(neutron, *route.Jumps[1].FSDBoost)                         // already neutron, untouched
	s.Equal(models.FSDBoostInjectionBasic, *route.Jumps[2].FSDBoost)   // next 120 — 1.20x
	s.Equal(models.FSDBoostInjectionPremium, *route.Jumps[3].FSDBoost) // next 175 — 1.75x
	s.Nil(route.Jumps[4].FSDBoost)                                     // last jump
}

func (s *ComputeFSDBoostForRouteSuite) TestSingleJump() {
	route := s.makeRoute([]float64{0}, s.nils(1))
	computeFSDBoostForRoute(route, s.maxRange)
	s.Nil(route.Jumps[0].FSDBoost)
}

func (s *ComputeFSDBoostForRouteSuite) TestEmptyRoute() {
	route := &models.Route{}
	computeFSDBoostForRoute(route, s.maxRange)
	s.Empty(route.Jumps)
}

func (s *ComputeFSDBoostForRouteSuite) TestImpossibleJumpFallsBackToNeutron() {
	route := s.makeRoute([]float64{0, 210}, s.nils(2))
	computeFSDBoostForRoute(route, s.maxRange)
	s.Equal(models.FSDBoostNeutron, *route.Jumps[0].FSDBoost)
}
