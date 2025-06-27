package pulse

import (
	"math" // For math.Abs
	"math/rand"
	"sort"
	"testing"

	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"crownet/synaptic"
	// "reflect" // Unused import
)

// Helper to create a neuron for testing (copied from space/spatial_grid_test.go and adapted).
func newTestNeuron(id common.NeuronID, pos common.Point) *neuron.Neuron {
	dummySimParams := config.DefaultSimulationParameters()
	// Ensure critical fields for neuron.New are set.
	dummySimParams.NeuronBehavior.BaseFiringThreshold = 1.0
	// Other NeuronBehavior fields like AbsoluteRefractoryCycles, etc., will use defaults
	// from DefaultSimulationParameters(), which should be fine for basic neuron creation.
	return neuron.New(id, neuron.Excitatory, pos, &dummySimParams)
}

func defaultTestSimParamsForPulse() *config.SimulationParameters {
	p := config.DefaultSimulationParameters()
	p.General.PulsePropagationSpeed = 1.0
	p.General.SpaceMaxDimension = 10.0 // Used for default new pulse MaxTravelRadius
	p.NeuronBehavior.BaseFiringThreshold = 1.0
	// For PulseImpactCalculator tests (if they use these neuron props)
	p.NeuronBehavior.AbsoluteRefractoryCycles = 2
	p.NeuronBehavior.RelativeRefractoryCycles = 3
	p.NeuronBehavior.AccumulatedPulseDecayRate = 0.1
	return &p
}

// Helper for creating synaptic weights for tests
func newTestSynapticWeights(simParams *config.SimulationParameters, rng *rand.Rand) *synaptic.NetworkWeights {
	sw, _ := synaptic.NewNetworkWeights(simParams, rng)
	return sw
}

func TestNewPulse(t *testing.T) {
	emitterID := common.NeuronID(1)
	origin := common.Point{1, 2}
	signal := common.PulseValue(1.0)
	creationCycle := common.CycleCount(10)
	maxRadius := 20.0

	p := New(emitterID, origin, signal, creationCycle, maxRadius)

	if p.EmittingNeuronID != emitterID {
		t.Errorf("NewPulse EmittingNeuronID got %d, want %d", p.EmittingNeuronID, emitterID)
	}
	if p.OriginPosition != origin {
		t.Errorf("NewPulse OriginPosition got %v, want %v", p.OriginPosition, origin)
	}
	if p.BaseSignalValue != signal {
		t.Errorf("NewPulse BaseSignalValue got %f, want %f", p.BaseSignalValue, signal)
	}
	if p.CreationCycle != creationCycle {
		t.Errorf("NewPulse CreationCycle got %d, want %d", p.CreationCycle, creationCycle)
	}
	if p.MaxTravelRadius != maxRadius {
		t.Errorf("NewPulse MaxTravelRadius got %f, want %f", p.MaxTravelRadius, maxRadius)
	}
	if p.CurrentDistance != 0.0 {
		t.Errorf("NewPulse CurrentDistance got %f, want 0.0", p.CurrentDistance)
	}
}

func TestPulse_Propagate(t *testing.T) {
	simParams := defaultTestSimParamsForPulse()
	p := New(0, common.Point{}, 1.0, 0, 5.0) // MaxTravelRadius = 5.0
	// simParams.General.PulsePropagationSpeed is already 1.0 from defaultTestSimParamsForPulse

	tests := []struct {
		name            string
		propagations    int
		wantActive      bool
		wantCurrentDist float64
	}{
		{"propagate once", 1, true, 1.0},
		{"propagate to boundary", 5, false, 5.0}, // CurrentDistance == MaxTravelRadius -> inactive
		{"propagate past boundary", 6, false, 6.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.CurrentDistance = 0 // Reset
			active := true
			for i := 0; i < tt.propagations; i++ {
				active = p.Propagate(simParams)
			}
			if active != tt.wantActive {
				t.Errorf("Propagate() active = %v, want %v after %d step(s)", active, tt.wantActive, tt.propagations)
			}
			if math.Abs(p.CurrentDistance-tt.wantCurrentDist) > 1e-9 {
				t.Errorf("Propagate() CurrentDistance = %f, want %f after %d step(s)", p.CurrentDistance, tt.wantCurrentDist, tt.propagations)
			}
		})
	}
	t.Run("Nil simParams", func(t *testing.T) {
		p.CurrentDistance = 0
		active := p.Propagate(nil)
		if active {
			t.Error("Propagate() with nil simParams should return false (inactive)")
		}
	})
}

func TestPulse_GetEffectShellForCycle(t *testing.T) {
	simParams := defaultTestSimParamsForPulse()
	p := New(0, common.Point{}, 1.0, 0, 20.0)
	simParams.General.PulsePropagationSpeed = 2.0 // Override default for this test

	p.CurrentDistance = 0.0 // Before first propagation
	s, e := p.GetEffectShellForCycle(simParams)
	if s != 0.0 || e != 0.0 {
		t.Errorf("Initial shell: got (%f,%f), want (0,0)", s, e)
	}

	p.CurrentDistance = 5.0 // After some propagation
	s, e = p.GetEffectShellForCycle(simParams)
	// shellEnd = CurrentDistance = 5.0
	// shellStart = CurrentDistance - Speed = 5.0 - 2.0 = 3.0
	if math.Abs(s-3.0) > 1e-9 || math.Abs(e-5.0) > 1e-9 {
		t.Errorf("Shell after dist 5.0: got (%f,%f), want (3.0,5.0)", s, e)
	}

	p.CurrentDistance = 1.0 // Distance < Speed
	s, e = p.GetEffectShellForCycle(simParams)
	// shellEnd = CurrentDistance = 1.0
	// shellStart = CurrentDistance - Speed = 1.0 - 2.0 = -1.0, clamped to 0
	if math.Abs(s-0.0) > 1e-9 || math.Abs(e-1.0) > 1e-9 {
		t.Errorf("Shell after dist 1.0 (dist < speed): got (%f,%f), want (0.0,1.0)", s, e)
	}

	t.Run("Nil simParams", func(t *testing.T) {
		p.CurrentDistance = 5.0
		s, e := p.GetEffectShellForCycle(nil)
		if s != p.CurrentDistance || e != p.CurrentDistance {
			t.Errorf("GetEffectShellForCycle() with nil simParams, got (%f,%f) want (%f,%f)", s, e, p.CurrentDistance, p.CurrentDistance)
		}
	})
}

func TestPulseList_AddClearCount(t *testing.T) {
	pl := NewPulseList()
	if pl.Count() != 0 {
		t.Errorf("New PulseList count = %d, want 0", pl.Count())
	}

	p1 := New(0, common.Point{}, 1.0, 0, 10)
	pl.Add(p1)
	if pl.Count() != 1 {
		t.Errorf("After Add(p1), count = %d, want 1", pl.Count())
	}

	pl.Add(nil) // Should not add nil
	if pl.Count() != 1 {
		t.Errorf("After Add(nil), count = %d, want 1", pl.Count())
	}

	p2 := New(1, common.Point{}, 1.0, 0, 10)
	p3 := New(2, common.Point{}, 1.0, 0, 10)
	pl.AddAll([]*Pulse{p2, nil, p3}) // Add a slice with a nil
	if pl.Count() != 3 {
		t.Errorf("After AddAll({p2,nil,p3}), count = %d, want 3", pl.Count())
	}

	all := pl.GetAll()
	if len(all) != 3 {
		t.Errorf("GetAll() len = %d, want 3", len(all))
	}

	pl.Clear()
	if pl.Count() != 0 {
		t.Errorf("After Clear(), count = %d, want 0", pl.Count())
	}
}

// TestProcessSinglePulseOnTargetNeuron has been removed as the function
// processSinglePulseOnTargetNeuron was refactored out into strategies
// (REFACTOR-005). The logic is now tested indirectly via TestPulseList_ProcessCycle
// which uses DefaultPulseImpactCalculator.

// TestPulseList_ProcessCycle is more of an integration test.
// For true unit tests, dependencies like 'weights' and 'neurons' (now 'spatialGrid')
// would ideally be mocked. Here, we use real instances with simple setups.
func TestPulseList_ProcessCycle(t *testing.T) {
	simParams := defaultTestSimParamsForPulse()
	rng := rand.New(rand.NewSource(42))
	weights := newTestSynapticWeights(simParams, rng) // Real weights instance

	// Neurons for the grid
	n0 := newTestNeuron(0, common.Point{0, 0})   // Emitter
	n1 := newTestNeuron(1, common.Point{1.5, 0}) // Target, dist 1.5
	n2 := newTestNeuron(2, common.Point{5, 0})   // Target, dist 5 (too far for initial pulse)
	allNeurons := []*neuron.Neuron{n0, n1, n2}

	// Setup SpatialGrid
	var gridMinBound common.Point
	for i := range gridMinBound {
		gridMinBound[i] = common.Coordinate(-simParams.General.SpaceMaxDimension)
	} // Use General substruct and cast
	// Ensure space.pointDimension is accessible or use a literal if not.
	// Assuming space.pointDimension is available from space/geometry.go
	gridCellSize := float64(simParams.General.PulsePropagationSpeed) * 2.0
	// Ensure space.pointDimension is used correctly. It's a const in space package.
	grid, _ := space.NewSpatialGrid(gridCellSize, common.PointDimension, gridMinBound) // Use common.PointDimension
	grid.Build(allNeurons)

	// Setup PulseList
	pl := NewPulseList()
	// Pulse from n0, travels at 1.0 units/cycle. MaxTravelRadius 3.0
	// Shell for cycle 1: [0,1), cycle 2: [1,2), cycle 3: [2,3)
	initialPulse := New(n0.ID, n0.Position, 1.0, 0, 3.0)
	pl.Add(initialPulse)

	weights.SetWeight(n0.ID, n1.ID, 1.0) // n0 -> n1, weight 1.0
	n1.CurrentFiringThreshold = 0.5      // n1 will fire if hit

	// Cycle 1
	// Pulse initialPulse: CurrentDistance becomes 1.0. Shell [0,1). n1 (dist 1.5) not hit.
	newPulses1 := pl.ProcessCycle(grid, weights, 1, simParams, allNeurons)
	if len(newPulses1) != 0 {
		t.Errorf("Cycle 1: Expected 0 new pulses, got %d", len(newPulses1))
	}
	if pl.Count() != 1 { // initialPulse should still be active
		t.Errorf("Cycle 1: Expected 1 active pulse, got %d", pl.Count())
	}
	if math.Abs(initialPulse.CurrentDistance-1.0) > 1e-9 {
		t.Errorf("Cycle 1: initialPulse distance %f, want 1.0", initialPulse.CurrentDistance)
	}

	// Cycle 2
	// Pulse initialPulse: CurrentDistance becomes 2.0. Shell [1,2). n1 (dist 1.5) IS hit. n1 fires.
	newPulses2 := pl.ProcessCycle(grid, weights, 2, simParams, allNeurons)
	if len(newPulses2) != 1 {
		t.Fatalf("Cycle 2: Expected 1 new pulse, got %d", len(newPulses2))
	}
	if newPulses2[0].EmittingNeuronID != n1.ID {
		t.Errorf("Cycle 2: New pulse emitter ID %d, want %d (n1.ID)", newPulses2[0].EmittingNeuronID, n1.ID)
	}
	if pl.Count() != 1 { // initialPulse should still be active
		t.Errorf("Cycle 2: Expected 1 active pulse (original), got %d", pl.Count())
	}
	if math.Abs(initialPulse.CurrentDistance-2.0) > 1e-9 {
		t.Errorf("Cycle 2: initialPulse distance %f, want 2.0", initialPulse.CurrentDistance)
	}
	// Add the new pulse for next cycle (normally CrowNet would do this)
	pl.AddAll(newPulses2)
	if pl.Count() != 2 {
		t.Errorf("Cycle 2: After adding new, expected 2 active pulses, got %d", pl.Count())
	}

	// Cycle 3
	// Pulse initialPulse: CurrentDistance becomes 3.0. Shell [2,3). n1 (dist 1.5) not in shell. Becomes inactive.
	// Pulse from n1: CurrentDistance becomes 1.0. Shell [0,1). No one hit.
	newPulses3 := pl.ProcessCycle(grid, weights, 3, simParams, allNeurons)
	if len(newPulses3) != 0 {
		t.Errorf("Cycle 3: Expected 0 new pulses, got %d", len(newPulses3))
	}
	// initialPulse (dist 3.0) is now inactive (MaxTravelRadius 3.0, so CurrentDistance < MaxTravelRadius is false)
	// pulse from n1 (dist 1.0) is active
	if pl.Count() != 1 {
		t.Errorf("Cycle 3: Expected 1 active pulse (from n1), got %d", pl.Count())
		for _, p := range pl.GetAll() {
			t.Logf("Active pulse in C3: ID %d, dist %f", p.EmittingNeuronID, p.CurrentDistance)
		}
	}
	if pl.Count() == 1 && pl.GetAll()[0].EmittingNeuronID != n1.ID {
		t.Errorf("Cycle 3: Remaining active pulse emitter ID %d, want %d (n1.ID)", pl.GetAll()[0].EmittingNeuronID, n1.ID)
	}
}

// Helper to sort neuron IDs for comparison
func sortNeuronIDsPulseTest(ids []common.NeuronID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
}

// Helper to extract IDs for comparison
func getNeuronIDsFromPulses(pulses []*Pulse) []common.NeuronID {
	ids := make([]common.NeuronID, len(pulses))
	seen := make(map[common.NeuronID]bool)
	count := 0
	for _, p := range pulses {
		if !seen[p.EmittingNeuronID] {
			ids[count] = p.EmittingNeuronID
			seen[p.EmittingNeuronID] = true
			count++
		}
	}
	return ids[:count]
}
