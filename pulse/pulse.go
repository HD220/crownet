// Package pulse defines the Pulse type, representing signals that propagate
// through the neural network, and PulseList, a manager for collections of
// active pulses. It handles pulse creation, propagation, and interaction
// with neurons.
package pulse

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"crownet/synaptic"
)

// defaultPulseMaxTravelRadiusFactor is multiplied by SimParams.SpaceMaxDimension
// to set the MaxTravelRadius for newly created pulses. A factor of 2.0 implies
// pulses can travel across the diameter of the defined space.
const defaultPulseMaxTravelRadiusFactor = 2.0

// Pulse represents an individual signal propagating through the neural network.
// It carries information about its origin, strength, and current propagation status.
type Pulse struct {
	EmittingNeuronID common.NeuronID   // ID of the neuron that emitted this pulse.
	OriginPosition   common.Point      // Position in space where the pulse originated.
	BaseSignalValue  common.PulseValue // Base strength/value of the pulse (e.g., +1.0 for excitatory, -1.0 for inhibitory).
	CreationCycle    common.CycleCount // Simulation cycle in which this pulse was created.
	CurrentDistance  float64           // Current distance the pulse has traveled from its origin.
	MaxTravelRadius  float64           // Maximum distance this pulse can travel before becoming inactive.
}

// New creates and returns a new Pulse instance, initialized with the provided parameters.
// Parameters:
//   emitterID: ID of the neuron emitting the pulse.
//   origin: The spatial position where the pulse starts.
//   signal: The base signal value of the pulse.
//   creationCycle: The simulation cycle when the pulse is generated.
//   maxRadius: The maximum distance the pulse can travel before becoming inactive.
func New(emitterID common.NeuronID, origin common.Point, signal common.PulseValue, creationCycle common.CycleCount, maxRadius float64) *Pulse {
	return &Pulse{
		EmittingNeuronID: emitterID,
		OriginPosition:   origin,
		BaseSignalValue:  signal,
		CreationCycle:    creationCycle,
		CurrentDistance:  0.0,
		MaxTravelRadius:  maxRadius,
	}
}

// Propagate advances the pulse's CurrentDistance by the PulsePropagationSpeed defined in simParams.
// It returns true if the pulse is still active (i.e., its CurrentDistance is less than its MaxTravelRadius),
// and false otherwise.
// A defensive check for nil simParams is included.
func (p *Pulse) Propagate(simParams *config.SimulationParameters) (isActive bool) {
	if simParams == nil {
		// Log error or handle as critical if SimParams can ever be nil here.
		// For now, assume pulse cannot propagate without SimParams.
		return false
	}
	p.CurrentDistance += simParams.PulsePropagationSpeed
	return p.CurrentDistance < p.MaxTravelRadius
}

// GetEffectShellForCycle calculates the inner (shellStartRadius) and outer (shellEndRadius)
// boundaries of the spherical shell where this pulse exerts its influence during the current cycle.
// The shell's thickness is determined by PulsePropagationSpeed.
// shellStartRadius is clamped to be non-negative.
// A defensive check for nil simParams is included.
func (p *Pulse) GetEffectShellForCycle(simParams *config.SimulationParameters) (shellStartRadius, shellEndRadius float64) {
	if simParams == nil {
		// Return a zero-width shell if SimParams is nil, effectively making the pulse have no area of effect.
		return p.CurrentDistance, p.CurrentDistance
	}
	shellEndRadius = p.CurrentDistance
	shellStartRadius = p.CurrentDistance - simParams.PulsePropagationSpeed
	if shellStartRadius < 0 {
		shellStartRadius = 0
	}
	return shellStartRadius, shellEndRadius
}

// PulseList manages a collection of active pulses within the neural network.
// It provides methods to add, clear, and access these pulses.
// Its main responsibility is to orchestrate the processing of all pulses during a simulation cycle,
// using configurable strategies for propagation, target selection, and impact calculation.
type PulseList struct {
	pulses         []*Pulse // Internal slice holding the active pulses.
	propagator     PulsePropagator
	zoneProvider   PulseEffectZoneProvider
	targetSelector PulseTargetSelector
	impactCalc     PulseImpactCalculator
}

// NewPulseList creates and returns an empty PulseList, ready to store pulses.
// It initializes with default strategy implementations.
func NewPulseList() *PulseList {
	return &PulseList{
		pulses:         make([]*Pulse, 0),
		propagator:     &DefaultPulsePropagator{},
		zoneProvider:   &DefaultPulseEffectZoneProvider{},
		targetSelector: &DefaultPulseTargetSelector{},
		impactCalc:     &DefaultPulseImpactCalculator{},
	}
}

// SetStrategies allows for overriding the default pulse processing strategies.
// This is useful for testing or for implementing alternative simulation mechanics.
func (pl *PulseList) SetStrategies(
	propagator PulsePropagator,
	zoneProvider PulseEffectZoneProvider,
	targetSelector PulseTargetSelector,
	impactCalc PulseImpactCalculator,
) {
	if propagator != nil {
		pl.propagator = propagator
	}
	if zoneProvider != nil {
		pl.zoneProvider = zoneProvider
	}
	if targetSelector != nil {
		pl.targetSelector = targetSelector
	}
	if impactCalc != nil {
		pl.impactCalc = impactCalc
	}
}

// Add appends a single pulse to the list of active pulses.
func (pl *PulseList) Add(p *Pulse) {
	if p == nil { // Avoid adding nil pulses
		return
	}
	pl.pulses = append(pl.pulses, p)
}

// AddAll appends a slice of new pulses to the list of active pulses.
// It filters out any nil pulses from the input slice.
func (pl *PulseList) AddAll(newPulses []*Pulse) {
	for _, p := range newPulses {
		if p != nil { // Add only non-nil pulses
			pl.pulses = append(pl.pulses, p)
		}
	}
}

// Clear removes all pulses from the list, effectively resetting it to an empty state.
func (pl *PulseList) Clear() {
	pl.pulses = make([]*Pulse, 0) // Replace with a new empty slice.
}

// GetAll returns a slice containing all pulses currently managed by the PulseList.
// Note: This returns a reference to the internal slice; modifications to the returned slice
// will affect the PulseList's internal state. Consider returning a copy if immutability is required.
func (pl *PulseList) GetAll() []*Pulse {
    return pl.pulses
}

// Count returns the current number of active pulses in the list.
func (pl *PulseList) Count() int {
    return len(pl.pulses)
}

// processSinglePulseOnTargetNeuron processes the effect of a given pulse 'p' on a 'targetNeuron'.
// It checks if the target neuron is within the pulse's current spherical shell of effect.
// If it is, the pulse's potential (modulated by synaptic weight) is integrated by the target neuron.
// If the target neuron fires as a result, a new pulse is generated and returned.
//
// Parameters:
//   - p: The active pulse being processed.
//   - targetNeuron: The neuron being checked for an effect from pulse 'p'.
//   - weights: The network's synaptic weights, used to modulate pulse potential.
//   - currentCycle: The current simulation cycle, used for new pulse creation time.
//   - simParams: Simulation parameters, used for pulse speed, effect shell, and new pulse properties.
//   - shellStartRadius: The inner radius of the pulse's current effect shell.
//   - shellEndRadius: The outer radius of the pulse's current effect shell.
//
// Returns:
// REFACTOR-005: processSinglePulseOnTargetNeuron logic is now handled by implementations
// of the PulseImpactCalculator interface (e.g., DefaultPulseImpactCalculator).
// The original function has been removed.

// ProcessCycle advances the state of all pulses in the PulseList by one simulation cycle.
// It involves several steps:
// 1. Propagate each pulse: Update its CurrentDistance. Inactive pulses (exceeding MaxTravelRadius) are removed.
// 2. For each remaining active pulse, determine its spherical shell of effect for the current cycle.
// 3. For each active pulse, iterate through all neurons in the network:
//    - If a neuron falls within the pulse's effect shell, process the interaction using
//      `processSinglePulseOnTargetNeuron`. This may result in the neuron firing.
// 4. Collect all newly generated pulses (from neurons that fired in step 3).
// 5. Update the PulseList to contain only pulses that are still active after propagation.
//
// Parameters:
//   - spatialGrid: The spatial grid index used to efficiently find candidate neurons
//     that might be affected by the pulses. This replaces a brute-force iteration over all neurons.
//   - weights: The network's synaptic weights, used by `processSinglePulseOnTargetNeuron`.
//     Note: This is expected to be `*synaptic.NetworkWeights` after its refactoring.
//   - currentCycle: The current simulation cycle number.
//   - simParams: Global simulation parameters.
//   - allNeurons: A slice of all neurons in the network, potentially used by some target selectors.
//
// Returns:
//   - []*Pulse: A slice of all pulses newly generated during this cycle.
func (pl *PulseList) ProcessCycle(
	spatialGrid *space.SpatialGrid, // Can be nil if targetSelector doesn't use it
	weights *synaptic.NetworkWeights,
	currentCycle common.CycleCount,
	simParams *config.SimulationParameters,
	allNeurons []*neuron.Neuron, // For target selectors that might not use spatial grid
) (newlyGeneratedPulses []*Pulse) {

	// Defensive nil checks for strategies and critical parameters.
	if pl.propagator == nil || pl.zoneProvider == nil || pl.targetSelector == nil || pl.impactCalc == nil {
		// This indicates PulseList was not properly initialized or strategies were nilled.
		// Log a critical error or panic. For now, return empty.
		// log.Printf("Critical error: PulseList strategies not initialized in ProcessCycle")
		return make([]*Pulse, 0)
	}
	if weights == nil || simParams == nil || allNeurons == nil {
		// log.Printf("Critical error: ProcessCycle called with nil weights, simParams, or allNeurons")
		return make([]*Pulse, 0)
	}

	remainingActivePulses := make([]*Pulse, 0, len(pl.pulses))
	newlyGeneratedPulses = make([]*Pulse, 0)

	for _, p := range pl.pulses {
		// 1. Propagate pulse using the strategy
		if !pl.propagator.Propagate(p, simParams) {
			continue // Pulse is no longer active
		}
		remainingActivePulses = append(remainingActivePulses, p)

		// 2. Determine effect zone using the strategy
		shellStartRadius, shellEndRadius := pl.zoneProvider.GetEffectShell(p, simParams)

		// 3. Select candidate targets using the strategy
		candidateNeurons := pl.targetSelector.GetCandidateTargets(p, shellEndRadius, spatialGrid, allNeurons, simParams)

		// 4. Calculate impact on each candidate target using the strategy
		for _, targetNeuron := range candidateNeurons {
			if newPulse := pl.impactCalc.CalculateImpact(p, targetNeuron, weights, currentCycle, simParams, shellStartRadius, shellEndRadius); newPulse != nil {
				newlyGeneratedPulses = append(newlyGeneratedPulses, newPulse)
			}
		}
	}

	pl.pulses = remainingActivePulses
	return newlyGeneratedPulses
}
