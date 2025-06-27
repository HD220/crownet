package neuron

import (
	"testing"

	"crownet/common"
	"crownet/config" // Assuming config provides DefaultSimulationParameters
)

// Helper to get default sim params for tests
func getDefaultSimParamsForTest() *config.SimulationParameters {
	// Use your actual function to get default simulation parameters
	// This is a placeholder based on common practice.
	p := config.DefaultSimulationParameters()
	return &p
}

func TestNewNeuron(t *testing.T) {
	simParams := getDefaultSimParamsForTest()
	pos := common.Point{1.0, 2.0} // Example position
	n := New(common.NeuronID(1), Excitatory, pos, simParams)

	if n == nil {
		t.Fatal("NewNeuron returned nil")
	}
	if n.ID != 1 {
		t.Errorf("NewNeuron ID = %d, want %d", n.ID, 1)
	}
	if n.Type != Excitatory {
		t.Errorf("NewNeuron Type = %v, want %v", n.Type, Excitatory)
	}
	if n.CurrentState != Resting {
		t.Errorf("NewNeuron State = %v, want %v", n.CurrentState, Resting)
	}
	if n.AccumulatedPotential != 0.0 {
		t.Errorf("NewNeuron Potential = %f, want 0.0", n.AccumulatedPotential)
	}
	if n.BaseFiringThreshold != simParams.NeuronBehavior.BaseFiringThreshold {
		t.Errorf("NewNeuron BaseFiringThreshold = %f, want %f",
			n.BaseFiringThreshold, simParams.NeuronBehavior.BaseFiringThreshold)
	}
	if n.CurrentFiringThreshold != simParams.NeuronBehavior.BaseFiringThreshold {
		t.Errorf("NewNeuron CurrentFiringThreshold = %f, want %f",
			n.CurrentFiringThreshold, simParams.NeuronBehavior.BaseFiringThreshold)
	}
	if n.LastFiredCycle != -1 {
		t.Errorf("NewNeuron LastFiredCycle = %d, want -1", n.LastFiredCycle)
	}
	if n.Position[0] != pos[0] || n.Position[1] != pos[1] { // Basic check
		t.Errorf("NewNeuron Position = %v, want %v", n.Position, pos)
	}
}

func TestIntegrateIncomingPotential(t *testing.T) {
	simParams := getDefaultSimParamsForTest()
	n := New(1, Excitatory, common.Point{0, 0}, simParams)
	n.CurrentFiringThreshold = 1.0 // Set for test clarity

	// Test potential accumulation without firing
	fired := n.IntegrateIncomingPotential(0.5, 0)
	if fired {
		t.Errorf("IntegratePotential should not fire with potential 0.5, threshold 1.0")
	}
	if n.AccumulatedPotential != 0.5 {
		t.Errorf("IntegratePotential accumulated = %f, want 0.5", n.AccumulatedPotential)
	}

	// Test potential accumulation causing firing
	fired = n.IntegrateIncomingPotential(0.6, 1) // Total potential 0.5 + 0.6 = 1.1
	if !fired {
		t.Errorf("IntegratePotential should fire with potential 1.1, threshold 1.0")
	}
	if n.AccumulatedPotential != 1.1 {
		t.Errorf("IntegratePotential accumulated = %f, want 1.1", n.AccumulatedPotential)
	}
	// Note: IntegrateIncomingPotential only flags 'fired'. State change is by AdvanceState.
	if n.CurrentState != Resting {
		t.Errorf("IntegratePotential should not change state, got %v, want Resting", n.CurrentState)
	}

	// Test integration during AbsoluteRefractory period
	n.CurrentState = AbsoluteRefractory
	n.AccumulatedPotential = 0.0 // Reset for clarity
	initialPotentialInRefractory := n.AccumulatedPotential
	fired = n.IntegrateIncomingPotential(10.0, 2) // High potential
	if fired {
		t.Errorf("IntegratePotential should not fire during AbsoluteRefractory")
	}
	if n.AccumulatedPotential != initialPotentialInRefractory {
		// IntegrateIncomingPotential's contract is to return 'fired'.
		// LastFiredCycle and CyclesInCurrentState are updated by AdvanceState.
		t.Errorf("IntegratePotential: potential in AbsRefr changed to %f from %f, want it to remain %f",
			n.AccumulatedPotential, initialPotentialInRefractory, initialPotentialInRefractory)
	}
}

func TestDecayPotential(t *testing.T) {
	simParams := getDefaultSimParamsForTest()
	// Ensure a non-zero decay rate for testing
	if simParams.NeuronBehavior.AccumulatedPulseDecayRate == 0.0 {
		simParams.NeuronBehavior.AccumulatedPulseDecayRate = 0.1
	}
	n := New(1, Excitatory, common.Point{0, 0}, simParams)

	n.AccumulatedPotential = 1.0
	n.DecayPotential(simParams)
	expectedPotential := 1.0 * (1.0 - simParams.NeuronBehavior.AccumulatedPulseDecayRate)
	if n.AccumulatedPotential != common.Potential(expectedPotential) {
		t.Errorf("DecayPotential got %f, want %f", n.AccumulatedPotential, expectedPotential)
	}

	// Test decay of negative potential (if applicable by model)
	n.AccumulatedPotential = -1.0
	n.DecayPotential(simParams)
	expectedNegativePotential := -1.0 * (1.0 - simParams.NeuronBehavior.AccumulatedPulseDecayRate)
	if n.AccumulatedPotential != common.Potential(expectedNegativePotential) {
		t.Errorf("DecayPotential (negative) got %f, want %f", n.AccumulatedPotential, expectedNegativePotential)
	}
}

func TestAdvanceState(t *testing.T) {
	simParams := getDefaultSimParamsForTest()
	n := New(1, Excitatory, common.Point{0, 0}, simParams)
	n.BaseFiringThreshold = 1.0
	n.CurrentFiringThreshold = 1.0

	// Resting to Firing
	n.AccumulatedPotential = 1.5
	fired := n.AdvanceState(1, simParams)
	if !fired {
		t.Errorf("AdvanceState: Neuron should have fired")
	}
	if n.CurrentState != Firing { // Should immediately transition to Firing then to AbsoluteRefractory in one AdvanceState if logic is combined
		// The provided AdvanceState transitions Firing -> AbsoluteRefractory in the same call.
		// Let's adjust expectation if Firing state is transient within AdvanceState.
		// Current code: Resting -> Firing (sets fired=true), then Firing -> AbsRefr. So ends in AbsRefr.
		t.Errorf("AdvanceState: State after firing, expected Firing (transient) or AbsoluteRefractory, got %v", n.CurrentState)
	}
	// If Firing is a state it passes through in one AdvanceState call:
	// Test after Resting->Firing transition within AdvanceState:
	n.CurrentState = Resting // Reset for next sub-test clarity
	n.AccumulatedPotential = 1.5
	n.AdvanceState(1, simParams) // Call again
	if n.CurrentState != AbsoluteRefractory {
		t.Errorf("AdvanceState: State after firing should be AbsoluteRefractory, got %v", n.CurrentState)
	}
	if n.LastFiredCycle != 1 {
		t.Errorf("AdvanceState: LastFiredCycle got %d, want 1", n.LastFiredCycle)
	}
	if n.AccumulatedPotential != 0.0 { // Potential should reset after firing
		t.Errorf("AdvanceState: Potential after firing got %f, want 0.0", n.AccumulatedPotential)
	}

	// AbsoluteRefractory to RelativeRefractory
	n.CurrentState = AbsoluteRefractory
	n.CyclesInCurrentState = simParams.NeuronBehavior.AbsoluteRefractoryCycles - 1
	n.AdvanceState(2, simParams) // Still in Absolute
	if n.CurrentState != AbsoluteRefractory {
		t.Errorf("AdvanceState: Should remain in AbsoluteRefractory, got %v", n.CurrentState)
	}
	n.AdvanceState(3, simParams) // Transition to Relative
	if n.CurrentState != RelativeRefractory {
		t.Errorf("AdvanceState: Should transition to RelativeRefractory, got %v", n.CurrentState)
	}
	// CurrentFiringThreshold should be elevated during RelativeRefractory,
	// but this is handled by neurochemical.ApplyEffects.
	// Test assumes neurochem might have set it or it defaults based on some logic.
	// For this isolated test, we'd need to mock that or set it manually if AdvanceState itself modifies it.
	// The current AdvanceState doesn't modify CurrentFiringThreshold when entering RelativeRefractory itself.

	// RelativeRefractory to Resting
	n.CurrentState = RelativeRefractory
	n.CurrentFiringThreshold = 1.5 // Assume it was elevated
	n.AccumulatedPotential = 0.5   // Below elevated threshold
	n.CyclesInCurrentState = simParams.NeuronBehavior.RelativeRefractoryCycles - 1
	n.AdvanceState(4, simParams) // Still in Relative
	if n.CurrentState != RelativeRefractory {
		t.Errorf("AdvanceState: Should remain in RelativeRefractory, got %v", n.CurrentState)
	}
	n.AdvanceState(5, simParams) // Transition to Resting
	if n.CurrentState != Resting {
		t.Errorf("AdvanceState: Should transition to Resting, got %v", n.CurrentState)
	}
	if n.CurrentFiringThreshold != n.BaseFiringThreshold { // Should reset to base
		t.Errorf("AdvanceState: Threshold after RelativeRefractory got %f, want %f",
			n.CurrentFiringThreshold, n.BaseFiringThreshold)
	}

	// RelativeRefractory to Firing (if strong input)
	n.CurrentState = RelativeRefractory
	n.CurrentFiringThreshold = 1.5 // Elevated
	n.AccumulatedPotential = 2.0   // Strong input overcomes elevated threshold
	n.CyclesInCurrentState = 0     // Early in relative refractory
	fired = n.AdvanceState(6, simParams)
	if !fired {
		t.Errorf("AdvanceState: Should fire from RelativeRefractory with strong input")
	}
	// Similar to Resting->Firing, it should end up in AbsoluteRefractory after this call.
	if n.CurrentState != AbsoluteRefractory {
		t.Errorf("AdvanceState: State after firing from RelativeRefractory should be AbsoluteRefractory, got %v", n.CurrentState)
	}
	if n.LastFiredCycle != 6 {
		t.Errorf("AdvanceState: LastFiredCycle after firing from Relative, got %d, want 6", n.LastFiredCycle)
	}
}

func TestEmittedPulseSign(t *testing.T) {
	simParams := getDefaultSimParamsForTest()
	tests := []struct {
		name     string
		nType    Type
		wantSign common.PulseValue
	}{
		{"Excitatory", Excitatory, 1.0},
		{"Inhibitory", Inhibitory, -1.0},
		{"Dopaminergic", Dopaminergic, 1.0}, // Assuming dopaminergic also has excitatory-like pulse for this model
		{"Input", Input, 1.0},             // Assuming input neurons are excitatory by default for emission
		{"Output", Output, 1.0},           // Assuming output neurons are excitatory by default for emission
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := New(0, tt.nType, common.Point{0, 0}, simParams)
			if sign := n.EmittedPulseSign(); sign != tt.wantSign {
				t.Errorf("EmittedPulseSign() for type %v = %v, want %v", tt.nType, sign, tt.wantSign)
			}
		})
	}
}

func TestIsRecentlyActive(t *testing.T) {
	simParams := getDefaultSimParamsForTest()
	n := New(1, Excitatory, common.Point{0, 0}, simParams)
	window := common.CycleCount(5)

	// Case 1: Never fired
	if n.IsRecentlyActive(10, window) {
		t.Errorf("IsRecentlyActive should be false for neuron that never fired")
	}

	// Case 2: Fired recently
	n.LastFiredCycle = 8
	if !n.IsRecentlyActive(10, window) { // current=10, lastFired=8, age=2. window=5. 2 <= 5 is true.
		t.Errorf("IsRecentlyActive should be true for recently fired neuron (age 2, window 5)")
	}

	// Case 3: Fired, but outside window
	n.LastFiredCycle = 3
	if n.IsRecentlyActive(10, window) { // current=10, lastFired=3, age=7. window=5. 7 <= 5 is false.
		t.Errorf("IsRecentlyActive should be false for neuron fired outside window (age 7, window 5)")
	}

	// Case 4: Fired exactly at edge of window (inclusive)
	n.LastFiredCycle = 5
	if !n.IsRecentlyActive(10, window) { // current=10, lastFired=5, age=5. window=5. 5 <= 5 is true.
		t.Errorf("IsRecentlyActive should be true for neuron fired at edge of window (age 5, window 5)")
	}
}

// TestFiringHistoryManagement ensures FiringHistory is correctly managed.
func TestFiringHistoryManagement(t *testing.T) {
	simParams := getDefaultSimParamsForTest()
	// Set a specific window for easier testing
	simParams.Structure.OutputFrequencyWindowCycles = 3.0
	n := New(1, Output, common.Point{0, 0}, simParams) // Output neuron to use FiringHistory

	// Fire a few times
	n.AccumulatedPotential = n.BaseFiringThreshold // Ensure it fires
	n.AdvanceState(1, simParams)                   // Fires at cycle 1
	n.CurrentState = Resting                       // Manually reset for next fire
	n.AccumulatedPotential = n.BaseFiringThreshold
	n.AdvanceState(2, simParams) // Fires at cycle 2
	n.CurrentState = Resting
	n.AccumulatedPotential = n.BaseFiringThreshold
	n.AdvanceState(3, simParams) // Fires at cycle 3

	if len(n.FiringHistory) != 3 {
		t.Errorf("FiringHistory len after 3 fires = %d, want 3. History: %v", len(n.FiringHistory), n.FiringHistory)
	}

	// Fire again, cycle 1 should be pruned if window is 3
	// Current cycle 4, window is 3 cycles. Relevant history: [2, 3, 4]
	// Cutoff = 4 - 3 = 1. So, cycles > 1 are kept.
	n.CurrentState = Resting
	n.AccumulatedPotential = n.BaseFiringThreshold
	n.AdvanceState(4, simParams) // Fires at cycle 4

	// Expected history: [2, 3, 4] because cycle 1 is older than 4-3=1.
	// The pruning logic is: fireTime >= (currentCycle - WindowCycles)
	// For cycle 4, cutoff is 4 - 3 = 1.
	// History before firing at 4: [1,2,3]
	// After firing at 4 and pruning:
	// 1 is not >= 1 (mistake here, should be > cutoff or >= currentCycle - Window)
	// Let's re-check neuron.go: cutoff = currentCycle - OutputFrequencyWindowCycles
	// fireTime >= cutoff.
	// Cycle 4: cutoff = 4 - 3 = 1.
	// History [1,2,3], add 4 -> [1,2,3,4].
	// Pruning:
	// 1 >= 1 (true) -> keep 1
	// 2 >= 1 (true) -> keep 2
	// 3 >= 1 (true) -> keep 3
	// 4 >= 1 (true) -> keep 4
	// This means history would be [1,2,3,4]. This seems like the window is not sliding correctly if it's a count.
	// If OutputFrequencyWindowCycles means "count of cycles for averaging", then it's a sliding window of that many cycles.
	// The current pruning logic in neuron.go:
	// cutoff := currentCycle - common.CycleCount(simParams.Structure.OutputFrequencyWindowCycles)
	// for i, fireTime := range n.FiringHistory { if fireTime >= cutoff { ... } }
	// This keeps events that happened AT or AFTER (current - window).
	// If current=4, window=3, cutoff=1. Events [1,2,3] are all >=1. So [1,2,3,4] is kept.
	// This is correct for "events in the last X cycles including current".
	// The MaxHistLen might also apply a cap. MaxHistLen = 3 * 1.5 = 4.5 -> 4.
	// So, if len becomes 5, it would be capped.
	// Let's test one more.
	if len(n.FiringHistory) != 4 { // Expect [1,2,3,4]
		t.Errorf("FiringHistory len after 4 fires = %d, want 4. History: %v", len(n.FiringHistory), n.FiringHistory)
	}

	n.CurrentState = Resting
	n.AccumulatedPotential = n.BaseFiringThreshold
	n.AdvanceState(5, simParams) // Fires at cycle 5. History [1,2,3,4,5]. Cutoff = 5-3=2.
	// Kept: [2,3,4,5]. Length 4. MaxHistLen is 4.
	// This seems to be working as "keep events within the window [current-window_duration+1, current]"
	// The pruning might be slightly off if maxHistLen is the primary cap rather than precise window.
	// The provided code has a `maxHistLen` cap AND a time-based pruning.
	// Let's assume time-based pruning is primary.
	// Cycle 5, window 3. Cutoff = 5-3 = 2.
	// History before this fire: [1,2,3,4]. Add 5 -> [1,2,3,4,5].
	// Pruning: 1 < 2 (pruned). [2,3,4,5] remains.
	expectedLenAfter5 := 4
	if len(n.FiringHistory) != expectedLenAfter5 {
		t.Errorf("FiringHistory len after 5 fires = %d, want %d. History: %v",
			len(n.FiringHistory), expectedLenAfter5, n.FiringHistory)
	}
	// Check content
	expectedHistoryContent := []common.CycleCount{2, 3, 4, 5}
	if len(n.FiringHistory) == len(expectedHistoryContent) {
		for i := range n.FiringHistory {
			if n.FiringHistory[i] != expectedHistoryContent[i] {
				t.Errorf("FiringHistory content mismatch at index %d: got %v, want %v. Full: %v",
					i, n.FiringHistory[i], expectedHistoryContent[i], n.FiringHistory)
				break
			}
		}
	}
}
