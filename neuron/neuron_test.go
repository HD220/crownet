package neuron

import (
	"crownet/common"
	"crownet/config"
	"math"
	"testing"
)

func defaultTestSimParamsForNeuron() *config.SimulationParameters {
	p := config.DefaultSimulationParameters()
	p.BaseFiringThreshold = 1.0
	p.AccumulatedPulseDecayRate = 0.1
	p.AbsoluteRefractoryCycles = 2
	p.RelativeRefractoryCycles = 3
	// Neurochemical effects are not directly tested here, assume base thresholds.
	return &p
}

func TestNewNeuron(t *testing.T) {
	simParams := defaultTestSimParamsForNeuron()
	id := common.NeuronID(1)
	ntype := Excitatory
	pos := common.Point{1, 2, 3}

	n := New(id, ntype, pos, simParams)

	if n.ID != id {
		t.Errorf("NewNeuron ID = %d, want %d", n.ID, id)
	}
	if n.Type != ntype {
		t.Errorf("NewNeuron Type = %s, want %s", n.Type, ntype)
	}
	if n.Position != pos {
		t.Errorf("NewNeuron Position = %v, want %v", n.Position, pos)
	}
	if n.State != Resting {
		t.Errorf("NewNeuron State = %s, want %s", n.State, Resting)
	}
	if n.AccumulatedPulse != 0.0 {
		t.Errorf("NewNeuron AccumulatedPulse = %f, want 0.0", n.AccumulatedPulse)
	}
	if n.BaseFiringThreshold != common.Threshold(simParams.BaseFiringThreshold) {
		t.Errorf("NewNeuron BaseFiringThreshold = %f, want %f", n.BaseFiringThreshold, simParams.BaseFiringThreshold)
	}
	if n.CurrentFiringThreshold != common.Threshold(simParams.BaseFiringThreshold) {
		t.Errorf("NewNeuron CurrentFiringThreshold = %f, want %f", n.CurrentFiringThreshold, simParams.BaseFiringThreshold)
	}
	if n.LastFiredCycle != -1 {
		t.Errorf("NewNeuron LastFiredCycle = %d, want -1", n.LastFiredCycle)
	}
	if n.CyclesInCurrentState != 0 {
		t.Errorf("NewNeuron CyclesInCurrentState = %d, want 0", n.CyclesInCurrentState)
	}
	// Velocity should be zero-initialized by default by Go for the array type.
	var zeroVelocity common.Point
	if n.Velocity != zeroVelocity {
		t.Errorf("NewNeuron Velocity = %v, want %v", n.Velocity, zeroVelocity)
	}
}

func TestNeuron_IntegrateIncomingPotential(t *testing.T) {
	simParams := defaultTestSimParamsForNeuron()
	n := New(0, Excitatory, common.Point{}, simParams)
	currentCycle := common.CycleCount(10)

	t.Run("Potential below threshold", func(t *testing.T) {
		n.AccumulatedPulse = 0.0 // Reset
		n.State = Resting
		fired := n.IntegrateIncomingPotential(0.5, currentCycle)
		if fired {
			t.Error("IntegratePotential: fired unexpectedly, want false")
		}
		if n.AccumulatedPulse != 0.5 {
			t.Errorf("IntegratePotential: pulse = %f, want 0.5", n.AccumulatedPulse)
		}
		if n.State != Resting {
			t.Errorf("IntegratePotential: state = %s, want Resting", n.State)
		}
	})

	t.Run("Potential meets/exceeds threshold", func(t *testing.T) {
		n.AccumulatedPulse = 0.0
		n.State = Resting
		n.CurrentFiringThreshold = 1.0
		fired := n.IntegrateIncomingPotential(1.0, currentCycle) // Meets threshold
		if !fired {
			t.Error("IntegratePotential: did not fire, want true")
		}
		if n.AccumulatedPulse != 1.0 { // Pulse still accumulates before state change for this cycle
			t.Errorf("IntegratePotential: pulse = %f, want 1.0", n.AccumulatedPulse)
		}
		// State change to Firing is handled by AdvanceState based on this firing event.
		// IntegrateIncomingPotential itself just returns true. The test for state change
		// should be in TestNeuronStateMachine or by calling AdvanceState after this.
		// For now, let's assume IntegratePotential sets it to Firing *if* conditions allow.
		// The current Neuron.IntegrateIncomingPotential directly sets state to Firing.
		if n.State != Firing {
			t.Errorf("IntegratePotential: state = %s, want Firing", n.State)
		}
		if n.LastFiredCycle != currentCycle {
			t.Errorf("IntegratePotential: LastFiredCycle = %d, want %d", n.LastFiredCycle, currentCycle)
		}
		if n.CyclesInCurrentState != 0 { // Should reset on state change
			t.Errorf("IntegratePotential: CyclesInCurrentState = %d, want 0", n.CyclesInCurrentState)
		}
	})

	t.Run("Attempt to fire while in AbsoluteRefractory", func(t *testing.T) {
		n.State = AbsoluteRefractory
		n.AccumulatedPulse = 0.0
		fired := n.IntegrateIncomingPotential(2.0, currentCycle+1) // High potential
		if fired {
			t.Error("IntegratePotential: fired while in AbsoluteRefractory, want false")
		}
		if n.AccumulatedPulse != 2.0 { // Potential should still accumulate
			t.Errorf("IntegratePotential: pulse = %f, want 2.0", n.AccumulatedPulse)
		}
		if n.State != AbsoluteRefractory { // State should not change
			t.Errorf("IntegratePotential: state changed from AbsoluteRefractory, want AbsoluteRefractory state = %s", n.State)
		}
	})
}

func TestNeuron_AdvanceState(t *testing.T) {
	simParams := defaultTestSimParamsForNeuron()
	n := New(0, Excitatory, common.Point{}, simParams)

	t.Run("From Firing to AbsoluteRefractory", func(t *testing.T) {
		n.State = Firing
		n.CyclesInCurrentState = 0
		n.AccumulatedPulse = simParams.BaseFiringThreshold // Assume it just fired

		n.AdvanceState(1, simParams) // Advance to next cycle

		if n.State != AbsoluteRefractory {
			t.Errorf("AdvanceState: from Firing, state = %s, want AbsoluteRefractory", n.State)
		}
		if n.CyclesInCurrentState != 0 { // Reset for new state
			t.Errorf("AdvanceState: CyclesInCurrentState = %d, want 0", n.CyclesInCurrentState)
		}
		if n.AccumulatedPulse != 0.0 { // Potential should reset after firing
			t.Errorf("AdvanceState: AccumulatedPulse after firing = %f, want 0.0", n.AccumulatedPulse)
		}
	})

	t.Run("Through AbsoluteRefractory", func(t *testing.T) {
		n.State = AbsoluteRefractory
		n.CyclesInCurrentState = 0
		for i := 0; i < int(simParams.AbsoluteRefractoryCycles)-1; i++ {
			n.AdvanceState(common.CycleCount(2+i), simParams)
			if n.State != AbsoluteRefractory {
				t.Fatalf("AdvanceState: in AbsoluteRefractory, state changed prematurely to %s at cycle %d", n.State, i)
			}
			if n.CyclesInCurrentState != common.CycleCount(i+1) {
				t.Errorf("AdvanceState: CyclesInCurrentState in AbsRefr = %d, want %d", n.CyclesInCurrentState, i+1)
			}
		}
		// Next AdvanceState should transition it
		n.AdvanceState(common.CycleCount(2+int(simParams.AbsoluteRefractoryCycles)-1), simParams)
		if n.State != RelativeRefractory {
			t.Errorf("AdvanceState: from Absolute to Relative, state = %s, want RelativeRefractory", n.State)
		}
		if n.CyclesInCurrentState != 0 { // Reset for new state
			t.Errorf("AdvanceState: CyclesInCurrentState for new Relative = %d, want 0", n.CyclesInCurrentState)
		}
	})

	t.Run("Through RelativeRefractory", func(t *testing.T) {
		n.State = RelativeRefractory
		n.CyclesInCurrentState = 0
		// CurrentFiringThreshold should be elevated during RelativeRefractory, but this is handled by neurochemical.ApplyEffects.
		// Here we just test state transition timing.
		for i := 0; i < int(simParams.RelativeRefractoryCycles)-1; i++ {
			n.AdvanceState(common.CycleCount(10+i), simParams) // Use arbitrary base cycle
			if n.State != RelativeRefractory {
				t.Fatalf("AdvanceState: in RelativeRefractory, state changed prematurely to %s at cycle %d", n.State, i)
			}
		}
		n.AdvanceState(common.CycleCount(10+int(simParams.RelativeRefractoryCycles)-1), simParams)
		if n.State != Resting {
			t.Errorf("AdvanceState: from Relative to Resting, state = %s, want Resting", n.State)
		}
		if n.CyclesInCurrentState != 0 {
			t.Errorf("AdvanceState: CyclesInCurrentState for new Resting = %d, want 0", n.CyclesInCurrentState)
		}
	})
}

func TestNeuron_DecayPotential(t *testing.T) {
	simParams := defaultTestSimParamsForNeuron()
	n := New(0, Excitatory, common.Point{}, simParams)

	n.AccumulatedPulse = 1.0
	n.DecayPotential(simParams)
	expected := 1.0 * (1.0 - simParams.AccumulatedPulseDecayRate)
	if math.Abs(float64(n.AccumulatedPulse)-expected) > 1e-9 {
		t.Errorf("DecayPotential: pulse = %f, want %f", n.AccumulatedPulse, expected)
	}

	n.AccumulatedPulse = 0.05 // Below decay amount if decay is > potential
	n.DecayPotential(simParams)
	if n.AccumulatedPulse != 0.0 { // Should clamp to 0, not go negative from decay
		t.Errorf("DecayPotential: pulse decayed below zero to %f, want 0.0", n.AccumulatedPulse)
	}
}

func TestEmittedPulseSign(t *testing.T) {
	simParams := defaultTestSimParamsForNeuron()
	tests := []struct {
		name string
		ntype Type
		want common.PulseValue
	}{
		{"Excitatory", Excitatory, 1.0},
		{"Inhibitory", Inhibitory, -1.0},
		{"Dopaminergic", Dopaminergic, 0.0}, // Dopaminergic effect is chemical, not standard pulse
		{"Input", Input, 1.0},
		{"Output", Output, 1.0}, // Assuming output neurons also emit excitatory-like pulses if they fire
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := New(0, tt.ntype, common.Point{}, simParams)
			if got := n.EmittedPulseSign(); got != tt.want {
				t.Errorf("EmittedPulseSign() for type %s = %v, want %v", tt.ntype, got, tt.want)
			}
		})
	}
}
