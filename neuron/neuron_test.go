package neuron

import (
	"math"
	"testing"
)

func TestNewNeuron(t *testing.T) {
	pos := Point{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	n := NewNeuron(1, pos, ExcitatoryNeuron, 1.0)
	if n.ID != 1 || n.Type != ExcitatoryNeuron || n.BaseFiringThreshold != 1.0 {
		t.Errorf("NewNeuron() failed to initialize fields correctly. Got ID %d, Type %d, Threshold %.2f", n.ID, n.Type, n.BaseFiringThreshold)
	}
	if n.State != RestingState {
		t.Errorf("NewNeuron() initial state incorrect. Expected RestingState, got %d", n.State)
	}
	if n.AccumulatedPulse != 0.0 {
		t.Errorf("NewNeuron() initial accumulated pulse incorrect. Expected 0.0, got %.2f", n.AccumulatedPulse)
	}
}

func TestNeuronStateTransitions(t *testing.T) {
	n := NewNeuron(1, Point{}, ExcitatoryNeuron, 1.0)
	currentCycle := 0

	// Fire the neuron
	n.State = FiringState
	n.UpdateState(currentCycle) // Cycle 0: Firing -> Absolute Refractory
	if n.State != AbsoluteRefractoryState {
		t.Errorf("State should be AbsoluteRefractory after firing, got %d", n.State)
	}
	if n.LastFiredCycle != currentCycle {
		t.Errorf("LastFiredCycle not updated. Expected %d, got %d", currentCycle, n.LastFiredCycle)
	}
	if n.AccumulatedPulse != 0.0 { // Assuming reset after firing
		t.Errorf("AccumulatedPulse should reset after firing. Got %.2f", n.AccumulatedPulse)
	}

	// Go through Absolute Refractory
	for i := 0; i < AbsoluteRefractoryCycles; i++ {
		currentCycle++
		n.UpdateState(currentCycle)
		if i < AbsoluteRefractoryCycles-1 && n.State != AbsoluteRefractoryState {
			t.Errorf("Should remain in AbsoluteRefractory. Cycle %d, State %d", currentCycle, n.State)
		}
	}
	if n.State != RelativeRefractoryState {
		t.Errorf("State should be RelativeRefractory after Absolute. Got %d", n.State)
	}

	// Go through Relative Refractory
	// n.CurrentFiringThreshold = n.BaseFiringThreshold // Reset for this test part
	for i := 0; i < RelativeRefractoryCycles; i++ {
		currentCycle++
		n.UpdateState(currentCycle)
		if i < RelativeRefractoryCycles-1 && n.State != RelativeRefractoryState {
			t.Errorf("Should remain in RelativeRefractory. Cycle %d, State %d", currentCycle, n.State)
		}
	}
	if n.State != RestingState {
		t.Errorf("State should be Resting after Relative. Got %d", n.State)
	}
	if n.CurrentFiringThreshold != n.BaseFiringThreshold {
		t.Errorf("Firing threshold should reset to base after refractory period. Got %.2f, Base %.2f", n.CurrentFiringThreshold, n.BaseFiringThreshold)
	}
}

func TestPulseAccumulationAndDecay(t *testing.T) {
	n := NewNeuron(1, Point{}, ExcitatoryNeuron, 1.0)

	// Receive some pulses
	n.ReceivePulse(0.3, 0) // acc = 0.3
	n.ReceivePulse(0.3, 0) // acc = 0.6
	if math.Abs(n.AccumulatedPulse-0.6) > 1e-9 {
		t.Errorf("AccumulatedPulse incorrect. Expected 0.6, got %.2f", n.AccumulatedPulse)
	}

	// Decay
	n.DecayPulseAccumulation() // 0.6 - 0.06 = 0.54
	if math.Abs(n.AccumulatedPulse-0.54) > 1e-9 {
		t.Errorf("Decay incorrect. Expected 0.54, got %.2f", n.AccumulatedPulse)
	}

	n.DecayPulseAccumulation() // 0.54 - 0.054 = 0.486
	if math.Abs(n.AccumulatedPulse-0.486) > 1e-9 {
		t.Errorf("Decay incorrect. Expected 0.486, got %.2f", n.AccumulatedPulse)
	}
}

func TestNeuronFiring(t *testing.T) {
	n := NewNeuron(1, Point{}, ExcitatoryNeuron, 1.0)
	fired := n.ReceivePulse(0.5, 0) // acc = 0.5, not enough to fire
	if fired {
		t.Errorf("Neuron fired prematurely.")
	}
	if n.State == FiringState {
		t.Errorf("Neuron state changed to Firing prematurely.")
	}

	fired = n.ReceivePulse(0.5, 0) // acc = 1.0, should fire
	if !fired {
		t.Errorf("Neuron did not fire when threshold was met.")
	}
	if n.State != FiringState {
		t.Errorf("Neuron state did not change to FiringState upon firing.")
	}
}

func TestGetBasePulseSign(t *testing.T) {
	excitatory := NewNeuron(1, Point{}, ExcitatoryNeuron, 1.0)
	inhibitory := NewNeuron(2, Point{}, InhibitoryNeuron, 1.0)
	dopaminergic := NewNeuron(3, Point{}, DopaminergicNeuron, 1.0)
	inputN := NewNeuron(4, Point{}, InputNeuron, 1.0)
	outputN := NewNeuron(5, Point{}, OutputNeuron, 1.0)

	if excitatory.GetBasePulseSign() != 1.0 {
		t.Errorf("Excitatory base signal incorrect. Expected 1.0, got %.2f", excitatory.GetBasePulseSign())
	}
	if inhibitory.GetBasePulseSign() != -1.0 {
		t.Errorf("Inhibitory base signal incorrect. Expected -1.0, got %.2f", inhibitory.GetBasePulseSign())
	}
	if dopaminergic.GetBasePulseSign() != 0.0 {
		t.Errorf("Dopaminergic base signal incorrect. Expected 0.0, got %.2f", dopaminergic.GetBasePulseSign())
	}
	if inputN.GetBasePulseSign() != 1.0 {
		t.Errorf("InputNeuron base signal incorrect. Expected 1.0, got %.2f", inputN.GetBasePulseSign())
	}
	if outputN.GetBasePulseSign() != 1.0 {
		t.Errorf("OutputNeuron base signal incorrect. Expected 1.0, got %.2f", outputN.GetBasePulseSign())
	}
}

func TestFiringInAbsoluteRefractory(t *testing.T) {
	n := NewNeuron(1, Point{}, ExcitatoryNeuron, 1.0)
	n.State = AbsoluteRefractoryState
	n.CyclesInCurrentState = 0

	fired := n.ReceivePulse(2.0, 0) // Stimulus well above threshold
	if fired {
		t.Errorf("Neuron fired while in AbsoluteRefractoryState.")
	}
	if n.State == FiringState {
		t.Errorf("Neuron state changed to FiringState while in AbsoluteRefractoryState.")
	}
}
