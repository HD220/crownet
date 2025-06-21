package network

import (
	"crownet/neuron"
	"math"
	"testing"
)

func TestNewPulse(t *testing.T) {
	origin := neuron.Point{1, 2, 3}
	p := NewPulse(1, origin, 1.0, 5, 10.0)

	if p.EmittingNeuronID != 1 {
		t.Errorf("NewPulse EmittingNeuronID incorrect. Expected 1, got %d", p.EmittingNeuronID)
	}
	if p.OriginPosition[0] != 1.0 { // Quick check of position copy
		t.Errorf("NewPulse OriginPosition incorrect.")
	}
	if p.Value != 1.0 {
		t.Errorf("NewPulse Value incorrect. Expected 1.0, got %.2f", p.Value)
	}
	if p.CreationCycle != 5 {
		t.Errorf("NewPulse CreationCycle incorrect. Expected 5, got %d", p.CreationCycle)
	}
	if p.CurrentDistance != 0.0 {
		t.Errorf("NewPulse CurrentDistance incorrect. Expected 0.0, got %.2f", p.CurrentDistance)
	}
	if p.MaxRange != 10.0 {
		t.Errorf("NewPulse MaxRange incorrect. Expected 10.0, got %.2f", p.MaxRange)
	}
}

func TestPulsePropagate(t *testing.T) {
	p := NewPulse(1, neuron.Point{}, 1.0, 0, 2.0) // MaxRange = 2.0

	// Cycle 1
	active := p.Propagate() // CurrentDistance = 0.6
	if !active {
		t.Errorf("Pulse should be active after 1st propagation")
	}
	if math.Abs(p.CurrentDistance-neuron.PulsePropagationSpeed) > 1e-9 {
		t.Errorf("CurrentDistance after 1 prop: Expected %.2f, got %.2f", neuron.PulsePropagationSpeed, p.CurrentDistance)
	}

	// Cycle 2
	active = p.Propagate() // CurrentDistance = 1.2
	if !active {
		t.Errorf("Pulse should be active after 2nd propagation")
	}
	if math.Abs(p.CurrentDistance-2*neuron.PulsePropagationSpeed) > 1e-9 {
		t.Errorf("CurrentDistance after 2 prop: Expected %.2f, got %.2f", 2*neuron.PulsePropagationSpeed, p.CurrentDistance)
	}

	// Cycle 3
	active = p.Propagate() // CurrentDistance = 1.8
	if !active {
		t.Errorf("Pulse should be active after 3rd propagation")
	}
	if math.Abs(p.CurrentDistance-3*neuron.PulsePropagationSpeed) > 1e-9 {
		t.Errorf("CurrentDistance after 3 prop: Expected %.2f, got %.2f", 3*neuron.PulsePropagationSpeed, p.CurrentDistance)
	}

	// Cycle 4 - Pulse should dissipate as 4*0.6 = 2.4 > MaxRange 2.0
	active = p.Propagate() // CurrentDistance = 2.4
	if active {
		t.Errorf("Pulse should be inactive after exceeding MaxRange")
	}
}

func TestGetEffectRangeForCycle(t *testing.T) {
	p := NewPulse(1, neuron.Point{}, 1.0, 0, 10.0)

	// Initial state (before first propagation)
	start, end := p.GetEffectRangeForCycle()
	if !(start == 0 && end == 0) {
		t.Errorf("Initial GetEffectRangeForCycle incorrect. Expected [0,0), got [%.2f, %.2f)", start, end)
	}

	p.Propagate() // CurrentDistance = 0.6
	start, end = p.GetEffectRangeForCycle()
	if !(math.Abs(start-0.0) < 1e-9 && math.Abs(end-0.6) < 1e-9) {
		t.Errorf("GetEffectRangeForCycle after 1 prop incorrect. Expected [0,0.6), got [%.2f, %.2f)", start, end)
	}

	p.Propagate() // CurrentDistance = 1.2
	start, end = p.GetEffectRangeForCycle()
	expectedStart := neuron.PulsePropagationSpeed   // 0.6
	expectedEnd := 2 * neuron.PulsePropagationSpeed // 1.2
	if !(math.Abs(start-expectedStart) < 1e-9 && math.Abs(end-expectedEnd) < 1e-9) {
		t.Errorf("GetEffectRangeForCycle after 2 prop incorrect. Expected [%.2f,%.2f), got [%.2f, %.2f)", expectedStart, expectedEnd, start, end)
	}
}
