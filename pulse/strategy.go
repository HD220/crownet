// Package pulse defines strategies for pulse propagation and interaction.
package pulse

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"crownet/synaptic"
)

// PulsePropagator defines how a pulse moves and determines if it's still active.
type PulsePropagator interface {
	// Propagate updates the pulse's state (e.g., CurrentDistance) based on simParams.
	// It returns true if the pulse is still active, false otherwise.
	// The updated CurrentDistance should be reflected in the passed pulse object.
	Propagate(pulse *Pulse, simParams *config.SimulationParameters) (isActive bool)
}

// PulseEffectZoneProvider determines the spherical shell of a pulse's influence.
type PulseEffectZoneProvider interface {
	// GetEffectShell returns the start and end radius of the pulse's effect zone
	// for the current cycle, based on its state and simParams.
	GetEffectShell(pulse *Pulse, simParams *config.SimulationParameters) (shellStartRadius, shellEndRadius float64)
}

// PulseTargetSelector identifies candidate neurons that might be affected by a pulse.
type PulseTargetSelector interface {
	// GetCandidateTargets returns a slice of neurons that are potential targets
	// for the given pulse, considering its current effect zone.
	// allNeurons might be needed if no spatialGrid is used by an implementation.
	GetCandidateTargets(
		pulse *Pulse,
		shellEndRadius float64, // Provided by PulseEffectZoneProvider
		spatialGrid *space.SpatialGrid, // Grid can be nil if an implementation doesn't use it
		allNeurons []*neuron.Neuron, // Fallback or primary source if no grid
		simParams *config.SimulationParameters,
	) []*neuron.Neuron
}

// PulseImpactCalculator determines the effect of a pulse on a target neuron
// and whether the target neuron fires, generating a new pulse.
type PulseImpactCalculator interface {
	// CalculateImpact processes the interaction between a pulse and a targetNeuron.
	// It must verify if the targetNeuron is within the pulse's precise effect shell
	// (defined by shellStartRadius and shellEndRadius) before applying effects.
	// It returns a new Pulse if targetNeuron fires as a result, otherwise nil.
	CalculateImpact(
		pulse *Pulse,
		targetNeuron *neuron.Neuron,
		weights *synaptic.NetworkWeights,
		currentCycle common.CycleCount,
		simParams *config.SimulationParameters,
		shellStartRadius, shellEndRadius float64, // To perform precise in-shell check
	) (newlyGeneratedPulse *Pulse)
}

// --- Default Strategy Implementations ---

// DefaultPulsePropagator implements the PulsePropagator interface using
// the original logic from Pulse.Propagate.
type DefaultPulsePropagator struct{}

// Propagate advances the pulse's CurrentDistance by the PulsePropagationSpeed
// defined in simParams.General.PulsePropagationSpeed.
// It updates p.CurrentDistance directly.
// Returns true if the pulse is still active (CurrentDistance < MaxTravelRadius).
func (dpp *DefaultPulsePropagator) Propagate(p *Pulse, simParams *config.SimulationParameters) (isActive bool) {
	if simParams == nil {
		return false // Cannot propagate without SimParams
	}
	if p == nil {
		return false // Cannot propagate nil pulse
	}
	p.CurrentDistance += simParams.General.PulsePropagationSpeed // Assumes SimParams is grouped
	return p.CurrentDistance < p.MaxTravelRadius
}

// DefaultPulseEffectZoneProvider implements the PulseEffectZoneProvider interface
// using the original logic from Pulse.GetEffectShellForCycle.
type DefaultPulseEffectZoneProvider struct{}

// GetEffectShell calculates the inner and outer boundaries of the pulse's spherical
// effect shell for the current cycle, based on its CurrentDistance and
// simParams.General.PulsePropagationSpeed.
func (dpzp *DefaultPulseEffectZoneProvider) GetEffectShell(p *Pulse, simParams *config.SimulationParameters) (shellStartRadius, shellEndRadius float64) {
	if simParams == nil || p == nil {
		if p != nil {
			return p.CurrentDistance, p.CurrentDistance // Zero-width shell
		}
		return 0, 0 // Or some other appropriate default for nil pulse
	}
	shellEndRadius = p.CurrentDistance
	shellStartRadius = p.CurrentDistance - simParams.General.PulsePropagationSpeed // Assumes SimParams is grouped
	if shellStartRadius < 0 {
		shellStartRadius = 0
	}
	return shellStartRadius, shellEndRadius
}

// DefaultPulseTargetSelector implements the PulseTargetSelector interface.
// It uses the SpatialGrid to find candidate neurons if a grid is provided;
// otherwise, it would consider all neurons (though the current default relies on the grid).
type DefaultPulseTargetSelector struct{}

// GetCandidateTargets uses the spatialGrid to query for neurons within the
// pulse's outer effect radius (shellEndRadius).
// If spatialGrid is nil, this default implementation returns an empty slice,
// as it relies on the grid for efficient candidate selection.
// A more robust fallback could iterate allNeurons, but that's less performant.
func (dpts *DefaultPulseTargetSelector) GetCandidateTargets(
	pulse *Pulse,
	shellEndRadius float64,
	spatialGrid *space.SpatialGrid,
	allNeurons []*neuron.Neuron, // Parameter provided for alternative strategies
	simParams *config.SimulationParameters,
) []*neuron.Neuron {
	if pulse == nil {
		return []*neuron.Neuron{}
	}
	if spatialGrid != nil {
		return spatialGrid.QuerySphereForCandidates(pulse.OriginPosition, shellEndRadius)
	}
	// Fallback: If no spatial grid, a simple strategy might return all neurons,
	// and PulseImpactCalculator would do all distance checks.
	// However, for the default, we assume grid usage if available.
	// If not, an empty list means no candidates found by this strategy.
	// For a production system, one might log a warning if grid is nil but expected.
	// Or, could iterate allNeurons here and do a coarse distance check:
	// var candidates []*neuron.Neuron
	// for _, n := range allNeurons {
	//    if space.EuclideanDistance(pulse.OriginPosition, n.Position) <= shellEndRadius {
	//        candidates = append(candidates, n)
	//    }
	// }
	// return candidates
	return []*neuron.Neuron{} // Default to no candidates if no grid
}

// DefaultPulseImpactCalculator implements the PulseImpactCalculator interface
// using the logic from the original processSinglePulseOnTargetNeuron function.
type DefaultPulseImpactCalculator struct{}

// CalculateImpact processes the interaction between a pulse and a targetNeuron.
// It first checks if the targetNeuron is within the pulse's effect shell
// (defined by shellStartRadius and shellEndRadius). If so, it calculates the
// effective potential based on synaptic weight and applies it to the targetNeuron.
// If the targetNeuron fires, a new Pulse is created and returned.
func (dpic *DefaultPulseImpactCalculator) CalculateImpact(
	p *Pulse,
	targetNeuron *neuron.Neuron,
	weights *synaptic.NetworkWeights,
	currentCycle common.CycleCount,
	simParams *config.SimulationParameters,
	shellStartRadius, shellEndRadius float64,
) (newlyGeneratedPulse *Pulse) {

	if p == nil || targetNeuron == nil || weights == nil || simParams == nil {
		return nil
	}
	if targetNeuron.ID == p.EmittingNeuronID {
		return nil // Neuron cannot be affected by its own pulse this way
	}

	distanceToTarget := space.EuclideanDistance(p.OriginPosition, targetNeuron.Position)

	if distanceToTarget >= shellStartRadius && distanceToTarget < shellEndRadius {
		weight := weights.GetWeight(p.EmittingNeuronID, targetNeuron.ID)
		effectivePotential := p.BaseSignalValue * common.PulseValue(weight)

		if effectivePotential == 0 {
			return nil
		}

		if targetNeuron.IntegrateIncomingPotential(effectivePotential, currentCycle) {
			emittedSignal := targetNeuron.EmittedPulseSign()
			if emittedSignal != 0 {
				// Accessing defaultPulseMaxTravelRadiusFactor:
				// This constant was in pulse.go. It should ideally be part of simParams
				// or passed to NewPulse or this calculator if it can vary.
				// For now, let's assume it's accessible or re-define it if local to strategy.
				// To avoid direct dependency on a package-level const from another file if this
				// were in a different package, it's better if such factors are in simParams.
				// For now, using the existing constant.
				// TODO: Consider moving defaultPulseMaxTravelRadiusFactor to SimParams for better configurability.
				// As SimParams is grouped, this would be SimParams.General.DefaultPulseMaxTravelRadiusFactor
				// For now, directly using the existing constant if it's still in pulse.go or make it local.
				// Let's assume it's still accessible via pulse.defaultPulseMaxTravelRadiusFactor or defined locally.
				// For this refactor, I'll use the existing package-level const from pulse.go
				newPulseMaxRadius := simParams.General.SpaceMaxDimension * defaultPulseMaxTravelRadiusFactor
				return New( // New is from pulse.go
					targetNeuron.ID,
					targetNeuron.Position,
					emittedSignal,
					currentCycle,
					newPulseMaxRadius,
				)
			}
		}
	}
	return nil
}
