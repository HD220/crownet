package pulse

import (
	"math"
	"math/rand"
	"sort" // Added for sorting neuron IDs if needed for stable test comparisons
	"testing"

	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"crownet/synaptic"
)

// Helper to get default sim params for tests
func getDefaultSimParamsForPulseTest() *config.SimulationParameters {
	p := config.DefaultSimulationParameters()
	p.General.PulsePropagationSpeed = 1.0 // Standard speed for easier calculations
	p.NeuronBehavior.BaseFiringThreshold = 1.0
	return &p
}

func TestNewPulse(t *testing.T) {
	origin := common.Point{0, 0}
	p := New(1, origin, 1.0, 0, 10.0)

	if p == nil {
		t.Fatal("NewPulse returned nil")
	}
	if p.EmittingNeuronID != 1 {
		t.Errorf("NewPulse EmittingNeuronID = %d, want 1", p.EmittingNeuronID)
	}
	if p.CurrentDistance != 0.0 {
		t.Errorf("NewPulse CurrentDistance = %f, want 0.0", p.CurrentDistance)
	}
	if !p.IsActive {
		t.Errorf("NewPulse IsActive = %v, want true", p.IsActive)
	}
}

func TestPulse_Propagate(t *testing.T) {
	simParams := getDefaultSimParamsForPulseTest()
	p := New(1, common.Point{0, 0}, 1.0, 0, 5.0) // MaxRadius 5.0
	p.Propagate(simParams)                      // Dist = 1.0
	p.Propagate(simParams)                      // Dist = 2.0

	if p.CurrentDistance != 2.0 {
		t.Errorf("Propagate() CurrentDistance = %f, want 2.0", p.CurrentDistance)
	}
	if !p.IsActive {
		t.Errorf("Propagate() IsActive should be true, got false")
	}

	p.Propagate(simParams) // Dist = 3.0
	p.Propagate(simParams) // Dist = 4.0
	p.Propagate(simParams) // Dist = 5.0
	p.Propagate(simParams) // Dist = 6.0, should become inactive

	if p.CurrentDistance != 6.0 { // Distance still updates
		t.Errorf("Propagate() CurrentDistance after exceeding max = %f, want 6.0", p.CurrentDistance)
	}
	if p.IsActive {
		t.Errorf("Propagate() IsActive should be false after exceeding MaxRadius, got true")
	}

	// Test with different propagation speed
	simParams.General.PulsePropagationSpeed = 2.0
	p2 := New(2, common.Point{0, 0}, 1.0, 0, 5.0)
	p2.Propagate(simParams) // Dist = 2.0
	if p2.CurrentDistance != 2.0 {
		t.Errorf("Propagate() with speed 2.0, CurrentDistance = %f, want 2.0", p2.CurrentDistance)
	}
	p2.Propagate(simParams) // Dist = 4.0
	p2.Propagate(simParams) // Dist = 6.0, inactive
	if !p2.IsActive {
		// Corrected: IsActive should be false
		if p2.IsActive {
			t.Errorf("Propagate() with speed 2.0, IsActive should be false, got true")
		}
	} else {
		// This branch should not be hit if logic is correct.
		// If it is, it means IsActive was true when it should have been false.
		t.Errorf("Propagate() with speed 2.0, IsActive was unexpectedly true after exceeding MaxRadius")
	}

	// Test propagation when simParams is nil (should not change)
	p3 := New(3, common.Point{0, 0}, 1.0, 0, 5.0)
	p3.CurrentDistance = 1.0
	p3.Propagate(nil) // Pass nil simParams
	if p3.CurrentDistance != 1.0 || !p3.IsActive {
		t.Errorf("Propagate() with nil simParams should not change pulse state, got dist %f, active %v",
			p3.CurrentDistance, p.IsActive)
	}
}

func TestPulse_GetEffectShellForCycle(t *testing.T) {
	simParams := getDefaultSimParamsForPulseTest() // Speed = 1.0
	p := New(1, common.Point{0, 0}, 1.0, 0, 10.0)

	// Cycle 1: Propagate once
	p.Propagate(simParams) // CurrentDistance = 1.0
	s, e := p.GetEffectShellForCycle(simParams)
	if s != 0.0 || e != 1.0 {
		t.Errorf("GetEffectShellForCycle() after 1 step: got (%f,%f), want (0.0,1.0)", s, e)
	}

	// Cycle 2: Propagate again
	p.Propagate(simParams) // CurrentDistance = 2.0
	s, e = p.GetEffectShellForCycle(simParams)
	if s != 1.0 || e != 2.0 {
		t.Errorf("GetEffectShellForCycle() after 2 steps: got (%f,%f), want (1.0,2.0)", s, e)
	}

	// Test with nil simParams (should return current dist for both)
	pNilParams := New(1, common.Point{0, 0}, 1.0, 0, 10.0)
	pNilParams.CurrentDistance = 3.5
	sNil, eNil := pNilParams.GetEffectShellForCycle(nil)
	if sNil != pNilParams.CurrentDistance || eNil != pNilParams.CurrentDistance {
		t.Errorf("GetEffectShellForCycle() with nil simParams, got (%f,%f) want (%f,%f)",
			sNil, eNil, pNilParams.CurrentDistance, pNilParams.CurrentDistance)
	}
}

func TestPulseList_AddAndClear(t *testing.T) {
	pl := NewPulseList()
	p1 := New(1, common.Point{0, 0}, 1.0, 0, 10.0)
	p2 := New(2, common.Point{0, 0}, 1.0, 0, 10.0)

	pl.Add(p1)
	if len(pl.Pulses) != 1 {
		t.Fatalf("Add() failed, len = %d, want 1", len(pl.Pulses))
	}
	pl.Add(p2)
	if len(pl.Pulses) != 2 {
		t.Fatalf("Add() failed, len = %d, want 2", len(pl.Pulses))
	}

	pl.Clear()
	if len(pl.Pulses) != 0 {
		t.Errorf("Clear() failed, len = %d, want 0", len(pl.Pulses))
	}

	// Test AddAll
	pulsesToAdd := []*Pulse{p1, p2, New(3, common.Point{0, 0}, 1.0, 0, 10.0)}
	pl.AddAll(pulsesToAdd)
	if len(pl.Pulses) != 3 {
		t.Errorf("AddAll() failed, len = %d, want 3", len(pl.Pulses))
	}
}

func TestPulseList_ProcessCycle(t *testing.T) {
	simParams := getDefaultSimParamsForPulseTest()
	rng := rand.New(rand.NewSource(0)) // For NetworkWeights if it uses RNG
	// Initialize NetworkWeights, check for error
	sw, err := synaptic.NewNetworkWeights(simParams, rng)
	if err != nil {
		t.Fatalf("Failed to create NetworkWeights: %v", err)
	}

	// Setup neurons
	n1 := neuron.New(1, neuron.Excitatory, common.Point{0, 0}, simParams) // Emitter
	n2 := neuron.New(2, neuron.Excitatory, common.Point{0.5, 0}, simParams) // Target, very close
	n3 := neuron.New(3, neuron.Excitatory, common.Point{10, 10}, simParams) // Target, far away
	allNeuronsMap := map[common.NeuronID]*neuron.Neuron{
		n1.ID: n1,
		n2.ID: n2,
		n3.ID: n3,
	}
	// Set a weight for testing impact calculation
	sw.SetWeight(n1.ID, n2.ID, 1.0) // Strong connection to ensure n2 fires

	// Setup spatial grid
	gridCellSize := 1.0
	gridMinBound := common.Point{-20, -20} // Example bounds
	// Initialize SpatialGrid, check for error
	grid, err := space.NewSpatialGrid(gridCellSize, common.PointDimension, gridMinBound)
	if err != nil {
		t.Fatalf("Failed to create SpatialGrid: %v", err)
	}
	grid.AddNeuron(n1) // Error check omitted for brevity in test setup
	grid.AddNeuron(n2)
	grid.AddNeuron(n3)

	pl := NewPulseList()
	p := New(n1.ID, n1.Position, 1.0, 0, 10.0) // Pulse from n1
	pl.Add(p)

	// Cycle 1: Pulse propagates, should hit n2 and cause it to fire
	// n2's BaseFiringThreshold is 1.0. Pulse BaseSignalValue is 1.0.
	// DefaultPulseImpactCalculator: impact = pulse.BaseSignalValue * weight
	// Impact on n2 = 1.0 * 1.0 = 1.0. This should make n2 fire.
	newPulses, err := pl.ProcessCycle(grid, sw, 1, simParams, allNeuronsMap)
	if err != nil {
		t.Fatalf("ProcessCycle() error: %v", err)
	}

	if len(pl.Pulses) != 1 { // Original pulse p should still be active and in list
		t.Errorf("ProcessCycle() active pulses len = %d, want 1", len(pl.Pulses))
	}
	if !p.IsActive || p.CurrentDistance != simParams.General.PulsePropagationSpeed {
		t.Errorf("ProcessCycle() original pulse state: active=%v, dist=%f. Want true, %f",
			p.IsActive, p.CurrentDistance, simParams.General.PulsePropagationSpeed)
	}

	if len(newPulses) != 1 {
		t.Fatalf("ProcessCycle() new pulses len = %d, want 1 (from n2 firing)", len(newPulses))
	}
	if newPulses[0].EmittingNeuronID != n2.ID {
		t.Errorf("ProcessCycle() new pulse emitter = %d, want %d (n2.ID)",
			newPulses[0].EmittingNeuronID, n2.ID)
	}

	// Check n2's state (should have fired and be in AbsoluteRefractory)
	// Note: PulseList.ProcessCycle itself doesn't advance neuron states beyond potential accumulation.
	// The firing and state transition (Firing -> AbsoluteRefractory) is handled by
	// the neuron's AdvanceState method, which is called by the main network simulation loop,
	// *after* IntegrateIncomingPotential (called by PulseImpactCalculator).
	// So, n2.CurrentState might still be Resting or Integrating here, but its potential is high.
	// The DefaultPulseImpactCalculator *does* set neuron.CurrentState = Firing and returns a new pulse.
	// Let's verify n2's state after impact.
	if allNeuronsMap[n2.ID].CurrentState != neuron.Firing {
		t.Errorf("ProcessCycle: n2 state after impact = %v, want Firing", allNeuronsMap[n2.ID].CurrentState)
	}

	// Cycle 2: Original pulse continues, new pulse from n2 starts
	pl.AddAll(newPulses) // Add the new pulse to the list for next cycle processing
	newPulsesFromCycle2, err := pl.ProcessCycle(grid, sw, 2, simParams, allNeuronsMap)
	if err != nil {
		t.Fatalf("ProcessCycle() cycle 2 error: %v", err)
	}
	// Expect 2 active pulses in pl.Pulses now (original p, and pulse from n2)
	// Both should have propagated.
	if len(pl.Pulses) != 2 {
		t.Errorf("ProcessCycle() cycle 2 active pulses len = %d, want 2. Got: %v", len(pl.Pulses), pl.Pulses)
	}
	// No new firings expected as n2 is refractory and n3 is too far.
	if len(newPulsesFromCycle2) != 0 {
		t.Errorf("ProcessCycle() cycle 2 new pulses len = %d, want 0. Got: %v",
			len(newPulsesFromCycle2), newPulsesFromCycle2)
	}
}

// Unused functions, remove or use them.
// func sortNeuronIDsPulseTest(ids []common.NeuronID) {
// 	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
// }

// func getNeuronIDsFromPulses(pulses []*Pulse) []common.NeuronID {
// 	ids := make([]common.NeuronID, len(pulses))
// 	for i, p := range pulses {
// 		ids[i] = p.EmittingNeuronID
// 	}
// 	return ids
// }
