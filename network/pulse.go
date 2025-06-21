package network

import (
	"crownet/neuron"
	"math"
)

// Pulse represents a signal traveling from an emitting neuron.
type Pulse struct {
	EmittingNeuronID int
	OriginPosition   neuron.Point // Position of the emitting neuron when it fired
	Value            float64      // e.g., +0.3 for excitatory, -0.3 for inhibitory
	CreationCycle    int          // Network cycle when this pulse was created
	CurrentDistance  float64      // Distance traveled so far by this pulse
	MaxRange         float64      // Max distance this pulse can travel (e.g., SpaceMaxDimension * 2 or a smaller neuron-specific range)
}

// NewPulse creates a new pulse.
// MaxRange could be a global constant or configurable per neuron type later.
func NewPulse(neuronID int, origin neuron.Point, value float64, creationCycle int, maxRange float64) *Pulse {
	return &Pulse{
		EmittingNeuronID: neuronID,
		OriginPosition:   origin,
		Value:            value,
		CreationCycle:    creationCycle,
		CurrentDistance:  0.0, // Starts at the origin
		MaxRange:         maxRange,
	}
}

// Propagate advances the pulse's travel distance for one cycle.
// Returns true if the pulse is still active (within MaxRange).
func (p *Pulse) Propagate() bool {
	p.CurrentDistance += neuron.PulsePropagationSpeed

	// Check if the pulse has exceeded its maximum effective range or some other culling condition
	if p.CurrentDistance > p.MaxRange {
		// Consider pulse dissipation or max travel distance.
		// For now, let's use MaxRange. If MaxRange is very large, pulses could live long.
		// The README mentions "distância máxima do espaço: 8 unidades". This might be the diameter.
		// So a pulse from center might travel 4 units, or from edge to edge 8 units.
		// Let's assume MaxRange is related to the SpaceMaxDimension.
		return false // Pulse has faded or gone too far
	}
	return true
}

// GetEffectRangeForCycle returns the start and end distance shell this pulse covers in the current cycle.
func (p *Pulse) GetEffectRangeForCycle() (startDist, endDist float64) {
	// The pulse affects neurons in the shell it just entered.
	// Distance traveled up to *previous* cycle defines the inner boundary of the new shell.
	// Current distance traveled defines the outer boundary of the new shell.
	endDist = p.CurrentDistance
	startDist = math.Max(0, p.CurrentDistance-neuron.PulsePropagationSpeed) // Ensure startDist is not negative
	return startDist, endDist
}

// IsDopaminePulse checks if the pulse is from a dopaminergic neuron.
// This is a helper, as dopamine is handled differently.
// Actual dopamine pulses might need a different struct or flag.
// For now, based on value, but this is not robust.
// The neuron type of emitter is better.
func (p *Pulse) IsDopaminePulse() bool {
	// This is a placeholder. Dopamine release is not a simple +/- pulse.
	// We'd check the emitting neuron's type.
	return false // Or based on p.Value if we assign a specific value for dopamine.
}

// The README's step-by-step pulse propagation logic (1-10) is quite detailed
// and seems to imply a per-pulse, per-iteration check against all other neurons,
// using reference points. This is more complex than simple spherical expansion.
// That logic will need to be implemented in the network's pulse processing step.
// The Pulse struct here provides the basics for tracking a pulse's origin and travel.
// The "17 reference points" and "radius" in that logic needs clarification or a default assumption.
// If "radius" refers to the pulse's own maximum effect radius, it would be p.MaxRange.
// If "reference points" are fixed points in space, they need to be defined.
// The quadratic equation part (step 4) is for determining neuron positions if they are
// not directly known but inferred from distances to reference points - this seems unusual
// if neuron positions are already stored in their structs. This might be for finding
// *new* positions in synaptogenesis, or it's a misinterpretation of "calcular o vetor 16D da posição".
// It could be about calculating the *relative position vector* from emitter to receiver.

// For now, the `Propagate` and `GetEffectRangeForCycle` methods assume simple spherical expansion.
// The more complex logic from README will be in `CrowNet.processPulses`.

// PulseHit represents an event where a pulse reaches a neuron.
// This can be used to queue effects if direct application is too complex mid-loop.
// type PulseHit struct {
// 	TargetNeuronID int
// 	Value          float64
// 	ArrivalCycle   int
// }
