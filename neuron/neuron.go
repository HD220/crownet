// Package neuron defines the core Neuron type, its operational states, behaviors,
// and related constants used in the CrowNet simulation.
package neuron

import (
	"crownet/common"
	"crownet/config"
	"math" // Added for math.Abs in DecayPotential, was missing from original import list if nearZeroThreshold was used.
)

// Constants for emitted pulse signals define the base value of a pulse
// emitted by neurons of certain types.
const (
	// ExcitatoryPulseSignal is the value for pulses that have an excitatory effect.
	ExcitatoryPulseSignal common.PulseValue = 1.0
	// InhibitoryPulseSignal is the value for pulses that have an inhibitory effect.
	InhibitoryPulseSignal common.PulseValue = -1.0
	// NeutralPulseSignal is the value for neutral or non-standard pulses,
	// e.g., from Dopaminergic neurons whose primary effect is modulatory.
	NeutralPulseSignal common.PulseValue = 0.0
)

// nearZeroThreshold is a small value used to clamp accumulated potential to zero
// if it's very close, helping to manage floating-point inaccuracies.
// helping to manage floating-point inaccuracies.
const nearZeroThreshold = 1e-5

// Neuron representa uma unidade computacional individual na rede neural.
// Contém o estado e propriedades de um neurônio, incluindo sua posição, tipo, estado de disparo,
// electrical potential, thresholds, and recent activity information.
type Neuron struct {
	ID common.NeuronID // Global unique identifier for the neuron.
	// Type defines the functional role of the neuron in the network (e.g., Excitatory, Input).
	Type Type
	// Position represents the neuron's coordinates in the N-dimensional simulation space.
	Position common.Point
	// CurrentState indicates the neuron's current operational state (e.g., Resting, Firing).
	CurrentState State
	// AccumulatedPotential is the current electrical potential accumulated by the neuron from received pulses.
	AccumulatedPotential common.PulseValue
	// BaseFiringThreshold is the base potential threshold the neuron must reach to fire.
	BaseFiringThreshold common.Threshold
	// CurrentFiringThreshold is the current firing threshold, which can be modulated by neurochemicals.
	CurrentFiringThreshold common.Threshold
	// LastFiredCycle records the simulation cycle in which the neuron last fired (-1 if never).
	LastFiredCycle common.CycleCount
	// CyclesInCurrentState tracks how many cycles the neuron has been in its current state (useful for refractory periods).
	CyclesInCurrentState common.CycleCount
	// Velocity represents the neuron's velocity vector, used for the synaptogenesis (movement) mechanism.
	Velocity common.Vector
}

// New creates and initializes a new Neuron with the provided parameters.
// The initial state is Resting, accumulated potential is 0, and LastFiredCycle is -1.
// CurrentFiringThreshold is initialized with the BaseFiringThreshold from simParams.
// The initial velocity is a zero vector.
// It is assumed simParams is not nil; callers should ensure this.
func New(id common.NeuronID, neuronType Type, initialPosition common.Point, simParams *config.SimulationParameters) *Neuron {
	if simParams == nil {
		// This case should ideally be handled by the caller or result in a panic
		// if simParams are essential for neuron initialization (which they are for BaseFiringThreshold).
		// For now, proceeding with the assumption that simParams is always valid as per original code structure.
		// Consider returning an error: return nil, fmt.Errorf("NewNeuron: simParams cannot be nil")
		// Or panicking: panic("NewNeuron: simParams cannot be nil")
		// Depending on desired error handling strategy.
		// The current code will panic if simParams is nil due to direct dereference for BaseFiringThreshold.
	}
	n := &Neuron{
		ID:                     id,
		Type:                   neuronType,
		// as BaseFiringThreshold cannot be determined.
		// For now, assume simParams is always provided correctly by the caller (NewCrowNet).
		// Consider adding error return if this assumption might be violated.
	}
	n := &Neuron{
		ID:                     id,
		Type:                   neuronType,
		Position:               initialPosition,
		CurrentState:           Resting,
		AccumulatedPotential:   0.0,
		BaseFiringThreshold:    common.Threshold(simParams.BaseFiringThreshold),
		CurrentFiringThreshold: common.Threshold(simParams.BaseFiringThreshold),
		LastFiredCycle:         -1,
		CyclesInCurrentState:   0,
		Velocity:               common.Vector{},
	}
	return n
}

// IntegrateIncomingPotential atualiza o potencial acumulado do neurônio com um pulso recebido
// e determina se o neurônio dispara.
// IntegrateIncomingPotential updates the neuron's accumulated potential with an incoming pulse value
// and determines if the neuron fires.
// If the neuron is in an AbsoluteRefractory state, it cannot integrate potential or fire.
// If the accumulated potential exceeds the CurrentFiringThreshold, the neuron enters the Firing state.
// It returns true if the neuron fired as a result of this integration, false otherwise.
func (n *Neuron) IntegrateIncomingPotential(potential common.PulseValue, currentCycle common.CycleCount) (fired bool) {
	// Neurons in absolute refractory period cannot integrate new potentials or fire.
	if n.CurrentState == AbsoluteRefractory {
		return false
	}

	n.AccumulatedPotential += potential

	// Check if the neuron fires
	if n.AccumulatedPotential < n.CurrentFiringThreshold { // Direct comparison as both are float64 underlying types
		return false
	}
	n.CurrentState = Firing
	n.CyclesInCurrentState = 0
	return true
}

// AdvanceState atualiza o estado do neurônio com base em sua lógica de máquina de estados interna
// e nos parâmetros de simulação (e.g., durações dos períodos refratários).
// and simulation parameters (e.g., refractory period durations).
//
// State transitions are:
//   Firing             -> AbsoluteRefractory
//   AbsoluteRefractory -> RelativeRefractory (after simParams.AbsoluteRefractoryCycles)
//   RelativeRefractory -> Resting          (after simParams.RelativeRefractoryCycles)
// Accumulated potential is reset when entering AbsoluteRefractory state.
func (n *Neuron) AdvanceState(currentCycle common.CycleCount, simParams *config.SimulationParameters) {
	n.CyclesInCurrentState++ // Increment cycles spent in the current state.

	switch n.CurrentState {
	case Firing:
		// After firing, neuron enters absolute refractory period.
		n.CurrentState = AbsoluteRefractory
		n.CyclesInCurrentState = 0 // Reset counter for the new state.
		n.LastFiredCycle = currentCycle // Record the cycle of this firing event.
		n.AccumulatedPotential = 0.0    // Reset potential after firing.
	case AbsoluteRefractory:
		// If absolute refractory period has ended, transition to relative refractory.
		if simParams != nil && n.CyclesInCurrentState >= simParams.AbsoluteRefractoryCycles {
			n.CurrentState = RelativeRefractory
			n.CyclesInCurrentState = 0
		}
	case RelativeRefractory:
		// If relative refractory period has ended, transition back to resting.
		if simParams != nil && n.CyclesInCurrentState >= simParams.RelativeRefractoryCycles {
			n.CurrentState = Resting
			n.CyclesInCurrentState = 0
		}
	case Resting:
		// No state change based on time alone when resting; stays resting until potential causes firing.
		// CyclesInCurrentState will continue to increment, which is fine.
	}
}

// DecayPotential applies exponential decay to the neuron's accumulated potential.
// The potential decays towards zero at a rate defined by simParams.AccumulatedPulseDecayRate.
// If the resulting potential is very close to zero, it is clamped to zero to manage
// floating-point inaccuracies.
func (n *Neuron) DecayPotential(simParams *config.SimulationParameters) {
	if simParams == nil { // Defensive check
		return
	}
	decayRate := simParams.AccumulatedPulseDecayRate
	if decayRate <= 0 { // No decay if rate is zero or negative.
		return
	}
	if decayRate >= 1.0 { // Full decay if rate is 1.0 or more.
		n.AccumulatedPotential = 0.0
		return
	}

	// Potential decays towards zero by the decayRate factor.
	n.AccumulatedPotential *= common.PulseValue(1.0 - float64(decayRate)) // Ensure decayRate is float64 for calculation

	// Clamp to zero if very close, to avoid floating point inaccuracies.
	if math.Abs(float64(n.AccumulatedPotential)) < nearZeroThreshold {
		n.AccumulatedPotential = 0.0
	}
}

// EmittedPulseSign returns the base signal value (+1.0, -1.0, or 0.0) of the pulse
// this neuron emits upon firing, based on its type.
// Input and Output neurons are treated as Excitatory by default for pulse emission.
// Dopaminergic neurons emit a NeutralPulseSignal, as their effect is typically
// via chemical modulation rather than direct synaptic pulses in this model.
func (n *Neuron) EmittedPulseSign() common.PulseValue {
	switch n.Type {
	case Excitatory, Input, Output: // Input/Output neurons treated as excitatory for pulse emission.
		return ExcitatoryPulseSignal
	case Inhibitory:
		return InhibitoryPulseSignal
	case Dopaminergic:
		// Dopaminergic neurons primarily act via chemical modulation,
		// not direct synaptic pulses in this simplified model.
		return NeutralPulseSignal
	default:
		// Unknown neuron types also emit a neutral signal by default.
		return NeutralPulseSignal
	}
}

// UpdatePosition updates the neuron's position based on its current velocity.
// The update assumes a time step of 1 cycle: new_position = old_position + velocity.
func (n *Neuron) UpdatePosition() {
	for i := range n.Position { // Iterate over dimensions using range
		n.Position[i] += n.Velocity[i] // Both are common.Coordinate, direct addition is fine
	}
}
