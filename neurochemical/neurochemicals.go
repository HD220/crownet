// Package neurochemical manages the simulation of neurochemical substances like
// cortisol and dopamine, their production, decay, and modulatory effects on
// neuron behavior, learning rates, and synaptogenesis.
package neurochemical

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/pulse"
	"crownet/space"
	"math"
)

const (
	// minFiringThresholdValue defines the absolute minimum value a neuron's firing threshold can take.
	minFiringThresholdValue = 0.01
)

// getNormalizedLevel calculates the normalized level of a chemical (e.g., cortisol, dopamine)
// by dividing its current level by its maximum possible level.
// The result is clamped to the range [0.0, 1.0].
// Returns 0.0 if maxLevel is not positive to prevent division by zero or undefined behavior.
func getNormalizedLevel(level common.Level, maxLevel float64) float64 {
	if maxLevel <= 0 {
		return 0.0
	}
	normalized := float64(level) / maxLevel
	return math.Min(1.0, math.Max(0.0, normalized))
}

// applyChemicalInfluence applies the modulatory effect of a normalized chemical level
// on a given target value (e.g., a learning rate factor, a threshold).
// The effect is multiplicative: newValue = currentValue * (1 + influenceStrength * normalizedLevel).
// - currentValue: The value to be modified.
// - chemicalInfluenceStrength: The parameter defining how strongly and in what direction the chemical affects the target value.
// - normalizedChemicalLevel: The chemical's current level, normalized to [0,1].
// Returns the new, modified value. If the normalized chemical level is zero or less, the original value is returned unchanged.
func applyChemicalInfluence(currentValue float64, chemicalInfluenceStrength common.Factor, normalizedChemicalLevel float64) float64 {
	if normalizedChemicalLevel > 0 { // Only apply influence if the chemical is present at some effective level.
		return currentValue * (1.0 + float64(chemicalInfluenceStrength)*normalizedChemicalLevel)
	}
	return currentValue // No chemical presence (or non-positive normalized level), no change to the value.
}

// Environment encapsulates the state of neurochemicals (Cortisol and Dopamine) and their
// modulatory effects on learning and synaptogenesis within the neural network.
type Environment struct {
	CortisolLevel common.Level // Current level of cortisol in the environment.
	DopamineLevel common.Level // Current level of dopamine in the environment.

	// SynaptogenesisModulationFactor is influenced by chemical levels and affects neuron movement.
	SynaptogenesisModulationFactor common.Factor
	// LearningRateModulationFactor is influenced by chemical levels and affects Hebbian learning.
	LearningRateModulationFactor   common.Factor
}

// calculateCortisolStimulation determines the amount of cortisol to be produced in the current cycle.
// Production is triggered by "strong" (BaseSignalValue > 0) pulses hitting the cortisol gland's position.
// Each such pulse contributes an amount defined by SimParams.CortisolProductionPerHit.
// cortisolGlandPosition is the fixed location of the gland, sourced from SimParams.
func calculateCortisolStimulation(activePulses []*pulse.Pulse, cortisolGlandPosition common.Point, simParams *config.SimulationParameters) float64 {
	// REFACTOR-007: Add nil check for simParams
	if simParams == nil {
		// Consider logging: log.Println("Warning: calculateCortisolStimulation called with nil simParams")
		return 0.0
	}
	if simParams.Neurochemical.CortisolProductionPerHit <= 0 {
		return 0.0 // No production if the per-hit amount is zero or negative.
	}

	pulsesHittingGland := 0
	for _, p := range activePulses {
		// Consider only pulses with positive base signal value as potentially "stressful" or significant enough
		// to trigger cortisol release. This is a simplifying assumption.
		if p.BaseSignalValue > 0 {
			// GetEffectShellForCycle defines the spherical region where the pulse has an effect in the current cycle.
			// This depends on pulse speed and how many cycles it has already propagated.
			shellStart, shellEnd := p.GetEffectShellForCycle(simParams)
			distToGland := space.EuclideanDistance(p.OriginPosition, cortisolGlandPosition)

			// Check if the gland's fixed position falls within the pulse's expanding spherical shell of influence.
			if distToGland >= shellStart && distToGland < shellEnd {
				pulsesHittingGland++
			}
		}
	}
	return float64(pulsesHittingGland) * float64(simParams.Neurochemical.CortisolProductionPerHit) // Total cortisol produced from pulse hits.
}

// calculateDopamineProduction determines the amount of dopamine to be produced in the current cycle.
// Production is triggered by dopaminergic neurons that are currently firing.
// Each firing dopaminergic neuron contributes an amount defined by SimParams.DopamineProductionPerEvent.
func calculateDopamineProduction(neurons []*neuron.Neuron, simParams *config.SimulationParameters) float64 {
	// REFACTOR-007: Add nil check for simParams
	if simParams == nil {
		// Consider logging: log.Println("Warning: calculateDopamineProduction called with nil simParams")
		return 0.0
	}
	if simParams.Neurochemical.DopamineProductionPerEvent <= 0 {
		return 0.0 // No production if the per-event amount is zero or negative.
	}

	dopamineProducedThisCycle := 0.0
	for _, n := range neurons {
		// Accumulate dopamine if a neuron is of Dopaminergic type and is in the Firing state.
		if n.Type == neuron.Dopaminergic && n.CurrentState == neuron.Firing {
			dopamineProducedThisCycle += float64(simParams.Neurochemical.DopamineProductionPerEvent)
		}
	}
	return dopamineProducedThisCycle
}

// updateChemicalLevel calculates the new level of a chemical after considering
// production in the current cycle and natural decay.
// The level is clamped to be non-negative and not to exceed maxLevel.
// - currentLevel: The chemical's level at the start of the cycle.
// - decayRate: The fraction of the current level that decays per cycle.
// - productionThisCycle: The amount of chemical produced from events in this cycle.
// - maxLevel: The maximum permissible level for this chemical.
func updateChemicalLevel(currentLevel common.Level, decayRate common.Rate, productionThisCycle float64, maxLevel common.Level) common.Level {
	level := float64(currentLevel) // Work with float64 for calculations

	// Add event-based production.
	level += productionThisCycle // Correctly adds production amount

	// Apply decay (proportional to current level).
	level -= level * float64(decayRate)

	// Clamp levels to ensure they are within valid physiological/simulation bounds.
	if level < 0 {
		level = 0
	}
	if maxLevel > 0 && level > float64(maxLevel) { // Only clamp by maxLevel if it's positive.
		level = float64(maxLevel)
	} else if level > 0 && maxLevel <= 0 { // If maxLevel is invalid (e.g. 0 or negative), but level is positive, this is an issue.
                                          // For robustness, perhaps cap at a very large number or log warning.
                                          // Here, we just ensure it doesn't exceed a positive maxLevel.
                                          // If maxLevel is 0, level will be clamped to 0 if it was positive.
    }
	return common.Level(level)
}

// NewEnvironment creates a new neurochemical environment with initial default values.
// Chemical levels start at 0, and modulation factors start at 1.0 (no effect).
func NewEnvironment() *Environment {
	return &Environment{
		CortisolLevel:                  0.0,
		DopamineLevel:                  0.0,
		SynaptogenesisModulationFactor: 1.0, // Default to no modulation initially
		LearningRateModulationFactor:   1.0, // Default to no modulation initially
	}
}

// UpdateLevels recalculates the levels of cortisol and dopamine based on network activity
// (pulse events, dopaminergic neuron firings) and applies decay.
// After updating levels, it recalculates the overall modulation factors for learning and synaptogenesis.
// - neurons: Current state of all neurons in the network.
// - activePulses: All pulses active in the current cycle.
// - cortisolGlandPosition: The fixed position of the cortisol gland, from SimParams.
// - simParams: Global simulation parameters containing rates, factors, and max levels.
func (env *Environment) UpdateLevels(
	neurons []*neuron.Neuron,
	activePulses []*pulse.Pulse,
	cortisolGlandPosition common.Point, // Explicitly passed, though now part of simParams.
	simParams *config.SimulationParameters,
) {
	// REFACTOR-007: Add nil check for simParams
	if simParams == nil {
		// Consider logging: log.Println("Warning: Environment.UpdateLevels called with nil simParams. Levels will not update.")
		return
	}

	// Calculate production amounts for this cycle based on network events.
	cortisolProduction := calculateCortisolStimulation(activePulses, cortisolGlandPosition, simParams) // cortisolGlandPosition is from simParams too.
	dopamineProduction := calculateDopamineProduction(neurons, simParams)

	// Update chemical levels considering production and decay.
	env.CortisolLevel = updateChemicalLevel(
		env.CortisolLevel,
		common.Rate(simParams.Neurochemical.CortisolDecayRate),
		cortisolProduction,
		common.Level(simParams.Neurochemical.CortisolMaxLevel),
	)

	env.DopamineLevel = updateChemicalLevel(
		env.DopamineLevel,
		common.Rate(simParams.Neurochemical.DopamineDecayRate),
		dopamineProduction,
		common.Level(simParams.Neurochemical.DopamineMaxLevel),
	)

	// After levels are updated, recalculate their effects on modulation factors
	env.recalculateModulationFactors(simParams)
}

// recalculateModulationFactors updates the LearningRateModulationFactor and SynaptogenesisModulationFactor
// based on the current (updated) levels of Cortisol and Dopamine.
// These factors are then used by other systems (e.g., Hebbian learning, synaptogenesis)
// to scale their respective processes.
// Effects are multiplicative and sequential (dopamine first, then cortisol).
func (env *Environment) recalculateModulationFactors(simParams *config.SimulationParameters) {
	// REFACTOR-007: Add nil check for simParams
	if simParams == nil {
		// Consider logging: log.Println("Warning: Environment.recalculateModulationFactors called with nil simParams. Factors remain default.")
		env.LearningRateModulationFactor = 1.0
		env.SynaptogenesisModulationFactor = 1.0
		return
	}
	lrFactor := 1.0 // Start with a base factor of 1.0 (no modulation)
	normalizedDopamine := getNormalizedLevel(env.DopamineLevel, float64(simParams.Neurochemical.DopamineMaxLevel))
	normalizedCortisol := getNormalizedLevel(env.CortisolLevel, float64(simParams.Neurochemical.CortisolMaxLevel))

	// Apply Dopamine effect on Learning Rate
	lrFactor = applyChemicalInfluence(lrFactor, simParams.Neurochemical.DopamineInfluenceOnLR, normalizedDopamine)
	// Apply Cortisol effect on Learning Rate
	lrFactor = applyChemicalInfluence(lrFactor, simParams.Neurochemical.CortisolInfluenceOnLR, normalizedCortisol)

	// Ensure learning rate factor does not fall below the minimum defined in SimParams.
	// MinLearningRateFactor is in the Learning sub-struct of SimParams.
	lrFactor = math.Max(float64(simParams.Learning.MinLearningRateFactor), lrFactor)
	env.LearningRateModulationFactor = common.Factor(lrFactor)

	synFactor := 1.0
	// Apply Dopamine effect on Synaptogenesis Factor
	synFactor = applyChemicalInfluence(synFactor, simParams.Neurochemical.DopamineInfluenceOnSynapto, normalizedDopamine)
	// Apply Cortisol effect on Synaptogenesis Factor
	synFactor = applyChemicalInfluence(synFactor, simParams.Neurochemical.CortisolInfluenceOnSynapto, normalizedCortisol)

	// Ensure synaptogenesis factor is not negative.
	env.SynaptogenesisModulationFactor = common.Factor(math.Max(0.0, synFactor)) // Ensure factor is not negative.
}

// ApplyEffectsToNeurons adjusts the CurrentFiringThreshold of each neuron in the network
// based on the current levels of Cortisol and Dopamine.
// The effects are multiplicative on the neuron's BaseFiringThreshold.
// Cortisol and Dopamine effects are applied sequentially.
// The final threshold is clamped by minFiringThresholdValue.
func (env *Environment) ApplyEffectsToNeurons(neurons []*neuron.Neuron, simParams *config.SimulationParameters) {
	// REFACTOR-007: Add nil check for simParams
	if simParams == nil {
		// Consider logging: log.Println("Warning: Environment.ApplyEffectsToNeurons called with nil simParams. Neuron thresholds not changed.")
		// Ensure thresholds are at least base if this happens, though they should be if not modified.
		// This path implies an issue, but we avoid panic.
		for _, n := range neurons {
			if n != nil { // Defensive check for nil neuron in slice
				n.CurrentFiringThreshold = n.BaseFiringThreshold
			}
		}
		return
	}
	normalizedCortisol := getNormalizedLevel(env.CortisolLevel, float64(simParams.Neurochemical.CortisolMaxLevel))
	normalizedDopamine := getNormalizedLevel(env.DopamineLevel, float64(simParams.Neurochemical.DopamineMaxLevel))

	for _, n := range neurons {
		baseThreshold := float64(n.BaseFiringThreshold)
		modifiedThreshold := baseThreshold

		// Apply Cortisol effect on Firing Threshold first.
		modifiedThreshold = applyChemicalInfluence(modifiedThreshold, simParams.Neurochemical.FiringThresholdIncreaseOnCort, normalizedCortisol)

		// Then apply Dopamine effect on the (potentially cortisol-modified) threshold.
		modifiedThreshold = applyChemicalInfluence(modifiedThreshold, simParams.Neurochemical.FiringThresholdIncreaseOnDopa, normalizedDopamine)

		// Ensure the final threshold does not fall below a defined minimum positive value.
		n.CurrentFiringThreshold = common.Threshold(math.Max(minFiringThresholdValue, modifiedThreshold))
	}
}
