package neuron

import (
	"math" // For math.Abs if needed for potential comparisons, or other math functions

	"crownet/common"
	"crownet/config"
)

// Neuron represents a single computational unit in the neural network.
// It maintains its state, properties, and connections.
// REFACTOR-005: Added Position, LastMovedCycle, Velocity.
// REFACTOR-CONFIG-001: Updated to use new SimParams structure.
type Neuron struct {
	ID                     common.NeuronID      // Unique identifier for the neuron.
	Type                   Type                 // Functional type of the neuron (e.g., Excitatory, Input).
	CurrentState           State                // Current operational state (e.g., Resting, Firing).
	AccumulatedPotential   common.Potential     // Current membrane potential.
	BaseFiringThreshold    common.Threshold     // Base threshold for firing.
	CurrentFiringThreshold common.Threshold     // Effective firing threshold, possibly modulated by neurochemicals or refractory state.
	LastFiredCycle         common.CycleCount    // Simulation cycle in which the neuron last fired.
	CyclesInCurrentState   common.CycleCount    // Tracks how many cycles the neuron has been in its current state.
	Position               common.Point         // N-dimensional coordinates of the neuron in the simulation space.
	LastMovedCycle         common.CycleCount    // Simulation cycle in which the neuron last moved.
	Velocity               common.Point         // Current velocity vector of the neuron (for synaptogenesis).
	FiringHistory          []common.CycleCount  // Records recent firing cycles for frequency calculation.
	SimParams              *config.SimulationParameters // Reference to global simulation parameters.
}

// New creates and initializes a new Neuron instance.
//
// Parameters:
//
//	id: The unique identifier for the neuron.
//	neuronType: The functional type of the neuron.
//	initialPosition: The initial N-dimensional coordinates of the neuron.
//	simParams: A pointer to the global SimulationParameters.
//
// Returns:
//
//	A pointer to the newly created Neuron.
//	Returns nil if simParams is nil, as it's essential for initialization.
func New(id common.NeuronID,
	neuronType Type,
	initialPosition common.Point,
	simParams *config.SimulationParameters) *Neuron {
	if simParams == nil {
		// Log or handle error: simParams are crucial.
		return nil // Or panic, depending on error handling strategy.
	}

	baseThreshold := simParams.NeuronBehavior.BaseFiringThreshold
	// Ensure FiringHistory has a capacity related to OutputFrequencyWindowCycles for efficient appends.
	// This helps in avoiding frequent reallocations if the history is used for frequency calculation.
	// Max capacity could be OutputFrequencyWindowCycles + a small buffer.
	// For simplicity, let's use a fixed initial capacity or one based on OutputFrequencyWindowCycles.
	// If OutputFrequencyWindowCycles is very large, this might be memory intensive.
	// Consider if FiringHistory should be capped or managed differently.
	historyCapacity := int(math.Max(10, simParams.Structure.OutputFrequencyWindowCycles*1.2)) // Example capacity

	return &Neuron{
		ID:                     id,
		Type:                   neuronType,
		CurrentState:           Resting,
		AccumulatedPotential:   0.0,
		BaseFiringThreshold:    baseThreshold,
		CurrentFiringThreshold: baseThreshold, // Initially same as base
		LastFiredCycle:         -1,            // -1 indicates never fired
		CyclesInCurrentState:   0,
		Position:               initialPosition,
		LastMovedCycle:         -1, // -1 indicates never moved
		Velocity:               make(common.Point, common.PointDimension), // Initialize velocity to zero vector
		FiringHistory:          make([]common.CycleCount, 0, historyCapacity),
		SimParams:              simParams, // Store the reference
	}
}

// IntegrateIncomingPotential updates the neuron's accumulated potential based on
// an incoming pulse's value. It checks if this new potential crosses the firing
// threshold and returns true if the neuron fires as a result.
// The actual state transition to Firing is handled by AdvanceState.
// This method primarily focuses on the potential accumulation and threshold check.
// Parameter currentCycle is renamed to _ if not used.
func (n *Neuron) IntegrateIncomingPotential(potential common.PulseValue, _ common.CycleCount) (fired bool) {
	if n.CurrentState == AbsoluteRefractory {
		return false // Cannot integrate or fire during absolute refractory period.
	}

	n.AccumulatedPotential += common.Potential(potential)

	// Check for firing if not in relative refractory (where threshold is higher but can still fire)
	// or even if in relative, if potential overcomes the elevated threshold.
	if n.AccumulatedPotential >= common.Potential(n.CurrentFiringThreshold) {
		// Note: Actual firing (state change, pulse emission) is handled by AdvanceState.
		// This function just signals that threshold was met.
		return true
	}
	return false
}

// DecayPotential reduces the neuron's accumulated potential over time, simulating leakage.
// The decay rate is defined in SimulationParameters.
func (n *Neuron) DecayPotential(simParams *config.SimulationParameters) {
	if simParams == nil {
		// This check might be redundant if n.SimParams is always expected to be set.
		// However, defensive programming suggests keeping it if simParams can be nil.
		// If n.SimParams is used, ensure it's not nil at Neuron creation or before calling this.
		// Assuming n.SimParams is reliable:
		if n.SimParams == nil {
			return // Or log error
		}
		simParams = n.SimParams // Use neuron's own SimParams if argument is nil
	}

	decayRate := simParams.NeuronBehavior.AccumulatedPulseDecayRate
	n.AccumulatedPotential *= (1.0 - common.Potential(decayRate))

	// Ensure potential doesn't overshoot towards negative infinity due to strong inhibition + decay.
	// Clamp to a reasonable minimum if necessary, e.g., -BaseFiringThreshold or 0 if only positive potential matters.
	// For now, simple decay. If potentials can be negative, this is fine.
	// If potential should not go below 0 (or some other resting potential), add clamping here.
	// Example: if n.AccumulatedPotential < 0 { n.AccumulatedPotential = 0 } (if resting potential is 0)
}

// AdvanceState updates the neuron's state based on its current condition and rules
// for refractory periods, firing, and returning to resting state.
// It should be called once per simulation cycle for each neuron.
//
// Returns:
//
//	True if the neuron fires in this cycle, false otherwise.
func (n *Neuron) AdvanceState(currentCycle common.CycleCount, simParams *config.SimulationParameters) (fired bool) {
	if simParams == nil && n.SimParams == nil {
		return false // Critical params missing
	}
	if simParams == nil {
		simParams = n.SimParams // Use neuron's own SimParams
	}

	n.CyclesInCurrentState++
	fired = false

	switch n.CurrentState {
	case Resting:
		// If potential crosses threshold (e.g., due to IntegrateIncomingPotential call just before this), transition to Firing.
		// This check is somewhat redundant if IntegrateIncomingPotential already determined firing potential.
		// However, direct manipulation or other effects might also change potential.
		if n.AccumulatedPotential >= common.Potential(n.CurrentFiringThreshold) {
			n.CurrentState = Firing
			n.CyclesInCurrentState = 0 // Reset for new state
			// Firing actions (pulse emission, LastFiredCycle update) handled below
		}
	case Firing: // Neuron just fired or was forced to fire.
		fired = true // Confirm firing for this cycle.
		n.LastFiredCycle = currentCycle
		n.AccumulatedPotential = 0.0 // Reset potential after firing.

		// Record firing event for frequency calculation, ensure history doesn't grow indefinitely.
		n.FiringHistory = append(n.FiringHistory, currentCycle)
		// Pruning logic for FiringHistory: Keep only relevant window.
		// Window size from SimParams.Structure.OutputFrequencyWindowCycles.
		// This ensures FiringHistory doesn't grow unbounded.
		// Max history length could be OutputFrequencyWindowCycles + a small buffer.
		maxHistLen := int(simParams.Structure.OutputFrequencyWindowCycles * 1.5) // Example buffer
		if len(n.FiringHistory) > maxHistLen {
			// More sophisticated pruning: remove elements older than currentCycle - OutputFrequencyWindowCycles.
			// Simple approach: keep last N elements if maxHistLen is a hard cap.
			// If OutputFrequencyWindowCycles is the exact window, then we need to trim precisely.
			cutoff := currentCycle - common.CycleCount(simParams.Structure.OutputFrequencyWindowCycles)
			validIndex := 0
			for i, fireTime := range n.FiringHistory {
				if fireTime >= cutoff {
					n.FiringHistory[validIndex] = n.FiringHistory[i]
					validIndex++
				}
			}
			n.FiringHistory = n.FiringHistory[:validIndex]
		}

		// Transition to AbsoluteRefractory.
		n.CurrentState = AbsoluteRefractory
		n.CyclesInCurrentState = 0
		n.CurrentFiringThreshold = n.BaseFiringThreshold * 1000 // Effectively infinite during absolute
	case AbsoluteRefractory:
		if n.CyclesInCurrentState >= simParams.NeuronBehavior.AbsoluteRefractoryCycles {
			n.CurrentState = RelativeRefractory
			n.CyclesInCurrentState = 0
			// Set threshold for relative refractory period (e.g., higher than base).
			// This is often handled by neurochemical modulation or a fixed factor.
			// For now, assume CurrentFiringThreshold is updated by neurochem system
			// or a separate mechanism. If not, it should be set here.
			// Example: n.CurrentFiringThreshold = n.BaseFiringThreshold * 1.5
			// Current model: Neurochemical system updates CurrentFiringThreshold.
			// If no chemical system, it should revert towards base or a relative refractory value.
			// For simplicity, let's assume the neurochemical system will adjust it.
			// If it's not adjusted by neurochem, it might remain very high.
			// Let's ensure it's at least reset towards base if no other system manages it.
			// This part is tricky: if neurochem is off, what should threshold be?
			// Assuming BaseFiringThreshold if no other effects.
			// This logic should align with ApplyEffectsToNeurons in neurochemical pkg.
			// For now, just transition state. Threshold managed elsewhere or by default decay.
		}
	case RelativeRefractory:
		// Threshold is elevated. Neuron can still fire if strong input arrives.
		// Check if potential crossed the (elevated) CurrentFiringThreshold.
		if n.AccumulatedPotential >= common.Potential(n.CurrentFiringThreshold) {
			n.CurrentState = Firing
			n.CyclesInCurrentState = 0
			// Firing actions handled in next cycle's Firing case.
		} else if n.CyclesInCurrentState >= simParams.NeuronBehavior.RelativeRefractoryCycles {
			// Refractory period ended, return to Resting.
			n.CurrentState = Resting
			n.CyclesInCurrentState = 0
			n.CurrentFiringThreshold = n.BaseFiringThreshold // Reset to base threshold.
		}
	}
	return fired
}

// EmittedPulseSign determines the sign (+1 or -1) of the pulse emitted by this neuron.
// Excitatory and Dopaminergic neurons emit positive pulses.
// Inhibitory neurons emit negative pulses.
// Input/Output types typically rely on their underlying nature if not specified,
// or could be neutral (0) if they don't directly emit standard pulses.
// For now, assume Input/Output neurons are excitatory by default for pulse emission.
func (n *Neuron) EmittedPulseSign() common.PulseValue {
	switch n.Type {
	case Excitatory, Dopaminergic, Input, Output: // Assuming Input/Output are excitatory for emission
		return 1.0
	case Inhibitory:
		return -1.0
	default:
		return 0.0 // Should not happen for known types
	}
}

// IsRecentlyActive checks if the neuron has fired within a given number of cycles (window).
// This is a helper that might be used by learning or synaptogenesis rules.
func (n *Neuron) IsRecentlyActive(currentCycle common.CycleCount, window common.CycleCount) bool {
	if n.LastFiredCycle == -1 { // Never fired
		return false
	}
	return (currentCycle - n.LastFiredCycle) <= window
}
