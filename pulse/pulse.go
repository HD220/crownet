package pulse

import (
	"math" // Required for math.Abs and other operations

	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space" // For EuclideanDistance and other spatial calculations
	"crownet/synaptic"
)

// Pulse represents a signal propagating through the network space.
// It carries information from an emitting neuron to potentially affect other neurons.
type Pulse struct {
	EmittingNeuronID common.NeuronID   // ID of the neuron that emitted this pulse.
	Origin           common.Point      // N-dimensional point where the pulse originated.
	CurrentDistance  float64           // Current distance propagated from the origin.
	MaxRadius        float64           // Maximum distance this pulse can travel before dissipating.
	BaseSignalValue  common.PulseValue // Base strength/value of the pulse.
	CreationCycle    common.CycleCount // Simulation cycle when the pulse was created.
	IsActive         bool              // Flag indicating if the pulse is still active and should be processed.
}

// New creates a new Pulse instance.
//
// Parameters:
//
//	emitterID: The ID of the neuron emitting the pulse.
//	origin: The N-dimensional coordinates of the emitting neuron.
//	signal: The base signal value/strength of the pulse.
//	creationCycle: The simulation cycle in which the pulse is created.
//	maxRadius: The maximum distance the pulse can travel.
//
// Returns:
//
//	A pointer to the newly created Pulse instance.
func New(emitterID common.NeuronID,
	origin common.Point,
	signal common.PulseValue,
	creationCycle common.CycleCount,
	maxRadius float64) *Pulse {
	return &Pulse{
		EmittingNeuronID: emitterID,
		Origin:           origin,
		CurrentDistance:  0.0, // Starts at the origin
		MaxRadius:        maxRadius,
		BaseSignalValue:  signal,
		CreationCycle:    creationCycle,
		IsActive:         true,
	}
}

// Propagate advances the pulse's propagation by one step, increasing its CurrentDistance.
// It deactivates the pulse if it exceeds its MaxRadius.
// Uses propagation speed from SimParams.
func (p *Pulse) Propagate(simParams *config.SimulationParameters) {
	if !p.IsActive || simParams == nil {
		return
	}
	p.CurrentDistance += float64(simParams.General.PulsePropagationSpeed)
	if p.CurrentDistance > p.MaxRadius {
		p.IsActive = false
	}
}

// GetEffectShellForCycle determines the start and end radii of the pulse's
// effective "shell" for the current cycle, based on its propagation speed.
// This defines the spherical region where the pulse might interact with neurons.
func (p *Pulse) GetEffectShellForCycle(
	simParams *config.SimulationParameters) (shellStartRadius, shellEndRadius float64) {
	if !p.IsActive || simParams == nil {
		// If not active or params missing, shell has zero thickness at current distance
		return p.CurrentDistance, p.CurrentDistance
	}
	// Shell starts from previous distance and ends at current distance
	shellStartRadius = p.CurrentDistance - float64(simParams.General.PulsePropagationSpeed)
	if shellStartRadius < 0 {
		shellStartRadius = 0 // Cannot have negative radius
	}
	shellEndRadius = p.CurrentDistance
	return shellStartRadius, shellEndRadius
}

// PulseList manages a collection of active pulses in the network.
// REFACTOR-005: Updated to use neuronMap (map[ID]*Neuron) instead of []*Neuron.
// It also now requires access to the spatial grid for efficient neighbor lookups.
type PulseList struct { // revive:disable-line:exported Stuttering name is intended for clarity.
	Pulses       []*Pulse
	propagator   PulsePropagator
	zoneProvider PulseEffectZoneProvider
	targetSel    PulseTargetSelector
	impactCalc   PulseImpactCalculator
}

// NewPulseList creates an empty list for managing pulses with default strategies.
func NewPulseList() *PulseList {
	return &PulseList{
		Pulses:       make([]*Pulse, 0),
		propagator:   &DefaultPulsePropagator{},
		zoneProvider: &DefaultPulseEffectZoneProvider{},
		targetSel:    &DefaultPulseTargetSelector{},
		impactCalc:   &DefaultPulseImpactCalculator{},
	}
}

// Add appends a new pulse to the list.
func (pl *PulseList) Add(p *Pulse) {
	if p != nil && p.IsActive {
		pl.Pulses = append(pl.Pulses, p)
	}
}

// AddAll appends multiple new pulses to the list.
func (pl *PulseList) AddAll(pulses []*Pulse) {
	for _, p := range pulses {
		pl.Add(p) // Use Add to ensure nil/inactive checks
	}
}

// Clear removes all pulses from the list.
func (pl *PulseList) Clear() {
	pl.Pulses = make([]*Pulse, 0)
}

// GetAll returns a slice of all currently active pulses.
// This is useful for components that need to inspect all pulses, e.g., neurochemical system.
func (pl *PulseList) GetAll() []*Pulse {
	// Return a copy to prevent external modification of the slice structure,
	// though the pulses themselves are pointers and can be modified.
	activePulses := make([]*Pulse, 0, len(pl.Pulses))
	for _, p := range pl.Pulses {
		if p.IsActive {
			activePulses = append(activePulses, p)
		}
	}
	return activePulses
}

// ProcessCycle simulates one cycle of activity for all pulses in the list.
// This involves:
// 1. Propagating each active pulse.
// 2. Determining the effect zone (shell) for each pulse.
// 3. Finding target neurons within that zone using the spatial grid.
// 4. Calculating the impact of the pulse on each target neuron (potential change).
// 5. Collecting any new pulses emitted by neurons that fire as a result.
// 6. Cleaning up inactive pulses from the list.
//
// Parameters:
//
//	grid: The spatial grid for efficient querying of neuron locations.
//	weights: The network's synaptic weights, used for calculating pulse impact.
//	currentCycle: The current simulation cycle number.
//	simParams: Global simulation parameters.
//	allNeuronsMap: A map of all neurons in the network, keyed by ID.
//
// Returns:
//
//	A slice of new Pulses generated by neurons that fired due to impacts in this cycle.
//	An error if any part of the processing fails critically.
func (pl *PulseList) ProcessCycle(
	grid *space.SpatialGrid,
	weights *synaptic.NetworkWeights,
	currentCycle common.CycleCount,
	simParams *config.SimulationParameters,
	allNeuronsMap map[common.NeuronID]*neuron.Neuron,
) ([]*Pulse, error) {
	if simParams == nil {
		return nil, fmt.Errorf("ProcessCycle requires non-nil simParams")
	}

	newlyEmittedPulses := make([]*Pulse, 0)
	activePulsesNextCycle := make([]*Pulse, 0, len(pl.Pulses))

	for _, p := range pl.Pulses {
		if !p.IsActive {
			continue // Skip already inactive pulses
		}

		// 1. Propagate pulse
		pl.propagator.Propagate(p, simParams)
		if !p.IsActive { // Check if deactivated by propagation (e.g., exceeded MaxRadius)
			continue
		}

		// 2. Determine effect zone (shell)
		shellStartRadius, shellEndRadius := pl.zoneProvider.GetEffectShell(p, simParams)

		// 3. Find target neurons in the shell using the spatial grid
		// This requires grid to have a method like GetNeuronsInShell.
		// For now, using GetNeighborsWithinRadius as a proxy, assuming shellEndRadius is the key.
		// This is an oversimplification; a proper shell query is needed.
		// The target selector strategy can refine this.
		potentialTargets := pl.targetSel.SelectTargets(p, grid, allNeuronsMap, simParams, shellStartRadius, shellEndRadius)

		// 4. Calculate impact on each target neuron
		for _, targetNeuron := range potentialTargets {
			if targetNeuron.ID == p.EmittingNeuronID {
				continue // Pulse does not affect its own emitter
			}

			// DefaultPulseImpactCalculator will use weights, distance, etc.
			// It returns a new pulse if the target neuron fires.
			if newPulse := pl.impactCalc.CalculateImpact(p, targetNeuron, weights,
				currentCycle, simParams, shellStartRadius, shellEndRadius); newPulse != nil {
				newlyEmittedPulses = append(newlyEmittedPulses, newPulse)
			}
		}

		// If pulse is still active after this cycle's processing, keep it for next cycle.
		if p.IsActive {
			activePulsesNextCycle = append(activePulsesNextCycle, p)
		}
	}

	pl.Pulses = activePulsesNextCycle // Update main list with only still-active pulses
	return newlyEmittedPulses, nil
}
