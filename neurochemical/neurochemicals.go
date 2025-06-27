package neurochemical

import (
	"math"

	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/pulse"
	"crownet/space"
)

// Environment manages the levels and effects of neurochemicals in the network.
type Environment struct {
	CortisolLevel                  common.Level // Current global cortisol level.
	DopamineLevel                  common.Level // Current global dopamine level.
	LearningRateModulationFactor   common.Factor
	SynaptogenesisModulationFactor common.Factor
	NeurochemicalParams            *config.NeurochemicalParams // Reference to configuration
}

// NewEnvironment creates a new neurochemical environment with default levels.
// It requires NeurochemicalParams from the main simulation configuration.
func NewEnvironment(params *config.NeurochemicalParams) *Environment {
	return &Environment{
		CortisolLevel:                  0.0,
		DopamineLevel:                  0.0,
		LearningRateModulationFactor:   1.0, // Default: no modulation
		SynaptogenesisModulationFactor: 1.0, // Default: no modulation
		NeurochemicalParams:            params,
	}
}

// applyChemicalInfluence calculates the modified value of a target parameter
// based on the level of a specific neurochemical and its influence strength.
//
// Parameters:
//   - currentValue: The original value of the parameter to be modified.
//   - chemicalInfluenceStrength: The parameter defining how strongly and in what direction
//     the chemical affects the target value.
//   - normalizedChemicalLevel: The current level of the chemical, normalized (e.g., 0 to 1).
//
// Returns the new, modified value. If the normalized chemical level is zero or less,
// the original value is returned unchanged.
func applyChemicalInfluence(currentValue float64,
	chemicalInfluenceStrength common.Factor,
	normalizedChemicalLevel float64) float64 {
	if normalizedChemicalLevel <= 0 {
		return currentValue // No chemical, no effect
	}
	// Modulation: currentValue * (1 + (influence * level))
	// Example: If influence is +0.5 and level is 1.0, value increases by 50%.
	// If influence is -0.2 and level is 0.5, value decreases by 10%.
	return currentValue * (1.0 + (float64(chemicalInfluenceStrength) * normalizedChemicalLevel))
}

// calculateCortisolStimulation determines the amount of cortisol to be produced
// based on network activity, specifically active pulses near the cortisol gland.
// This is a simplified model where pulses "hitting" the gland region trigger production.
func calculateCortisolStimulation(activePulses []*pulse.Pulse,
	cortisolGlandPosition common.Point,
	simParams *config.SimulationParameters) float64 {
	if simParams == nil || activePulses == nil {
		return 0.0
	}

	// For simplicity, assume the gland is a point. Pulses "hit" if they pass within a small radius.
	// This radius could be a fixed value or part of simParams.
	// Let's use a small, fixed radius for now, e.g., 1.0 spatial units.
	// Or, more realistically, it could be related to pulse effective radius or interaction range.
	// Using a simplified "hit" if pulse's current sphere overlaps gland position.
	// This is a placeholder for a more robust geometric check.
	const glandHitRadius = 0.5 // Arbitrary small radius for gland sensitivity
	pulsesHittingGland := 0

	for _, p := range activePulses {
		// Check if the pulse's current spherical shell (defined by its CurrentDistance and some thickness)
		// intersects with the cortisolGlandPosition.
		// This is complex. Simplified: if distance from pulse origin to gland is roughly pulse current distance.
		// A more accurate check would involve sphere-point intersection.
		// For now, let's use a simpler check: if the gland is within the pulse's current effective radius.
		// This assumes pulses expand and affect a region.
		// A very simple model: if distance(pulse.Origin, glandPosition) is "close enough"
		// and pulse is "active enough" (e.g. not decayed too much).
		// This needs to be consistent with how pulses affect neurons.
		// Let's assume a pulse has an effect up to its CurrentDistance + some interaction width.
		// If gland is within this, it's a "hit".

		// Simpler: If pulse is expanding and its current radius is near the gland.
		// This is still not quite right. Let's count pulses whose *current position*
		// (if they had one, they are areas) is close to the gland.
		// For now, let's assume any active pulse has a chance to stimulate if gland is "globally" sensitive.
		// This function might need access to the spatial grid or more detailed pulse geometry.

		// Simplest model: any active pulse contributes if it's near the gland.
		// This needs a more robust geometric model.
		// Placeholder: if distance from pulse origin to gland is within a certain range.
		// This is not ideal as it doesn't account for pulse propagation.
		// Using a very naive approach for now: count active pulses as potential stressors.
		// This should be refined based on how "stress" is defined in the simulation.
		// For now, let's count pulses whose origin is near the gland.
		// This is still not quite right.
		// A better simplified model: if gland is within max pulse radius and pulse is active.
		distToGland := space.EuclideanDistance(p.Origin, cortisolGlandPosition)
		if distToGland < p.MaxRadius { // If gland is within potential reach of this pulse
			// This is still a placeholder. A proper model would check if the pulse's
			// current expanding shell intersects the gland's sensitive volume.
			// For now, any active pulse whose origin is "somewhat" near the gland might contribute.
			// This needs significant refinement.
			// Using a very simple "global stress" model for now:
			// Every active pulse contributes a small amount to potential cortisol production.
			// This is not using cortisolGlandPosition effectively.
			// Let's assume pulses need to be near the gland.
			// This requires a concept of pulse "current location" or effective zone.
			// For now, this function is a placeholder for a more meaningful calculation.
			// Let's assume if a pulse is active, it contributes to a global "stress" level
			// that then stimulates cortisol. This bypasses the gland position for now.
			// This is not a good model.
			// Reverting to a slightly more plausible placeholder:
			// Count pulses whose *current effective sphere* (center at origin, radius CurrentDistance)
			// would overlap with the gland position. This is still an oversimplification.
			// For true point pulses, this logic is flawed.
			// Let's assume the gland is sensitive to pulses passing nearby.
			// This needs a proper geometric check (e.g., line segment to sphere intersection for pulse path).
			// Given current pulse model (expanding shell), if gland is within CurrentDistance of origin, it's "hit".
			if distToGland <= p.CurrentDistance+glandHitRadius && distToGland >= p.CurrentDistance-glandHitRadius {
				pulsesHittingGland++
			}
		}
	}
	// Total cortisol produced from pulse hits.
	return float64(pulsesHittingGland) * float64(simParams.Neurochemical.CortisolProductionPerHit)
}

// updateChemicalLevel calculates the new level of a neurochemical after one cycle,
// considering its current level, decay rate, production in this cycle, and maximum level.
func updateChemicalLevel(currentLevel common.Level,
	decayRate common.Rate,
	productionThisCycle float64,
	maxLevel common.Level) common.Level {
	// Apply decay: newLevel = currentLevel * (1 - decayRate)
	levelAfterDecay := currentLevel * (1.0 - common.Level(decayRate))

	// Add production: newLevel = levelAfterDecay + production
	levelAfterProduction := levelAfterDecay + common.Level(productionThisCycle)

	// Clamp to [0, maxLevel]
	finalLevel := levelAfterProduction
	if finalLevel < 0 {
		finalLevel = 0
	}
	if maxLevel > 0 && finalLevel > maxLevel { // Only clamp if maxLevel is positive
		finalLevel = maxLevel
	} else if levelAfterProduction > 0 && maxLevel <= 0 {
		// If maxLevel is invalid (e.g. 0 or negative), but level is positive, this is an issue.
		// For robustness, perhaps cap at a very large number or log warning.
		// Here, we just ensure it doesn't exceed a positive maxLevel.
		// If maxLevel is 0, level will be clamped to 0 if it was positive.
	}
	return finalLevel
}

// UpdateLevels recalculates the levels of all neurochemicals based on decay,
// production (e.g., from network activity or specific events), and applies modulation effects.
// `neurons` and `activePulses` are used to determine activity-dependent production.
// `cortisolGlandPosition` is used for cortisol production calculation.
func (env *Environment) UpdateLevels(
	neurons map[common.NeuronID]*neuron.Neuron, // Changed from slice to map
	activePulses []*pulse.Pulse,
	cortisolGlandPosition common.Point, // Passed directly
	simParams *config.SimulationParameters, // Passed directly
) {
	if env.NeurochemicalParams == nil || simParams == nil {
		// Consider logging: log.Println("Warning: Environment.UpdateLevels called with nil params. Levels not updated.")
		return
	}

	// Cortisol update
	// Production can be base rate + activity-dependent (e.g., pulses near gland)
	cortisolProduction := float64(env.NeurochemicalParams.CortisolProductionRate)
	// cortisolGlandPosition is from simParams too.
	cortisolProduction += calculateCortisolStimulation(activePulses, cortisolGlandPosition, simParams)
	env.CortisolLevel = updateChemicalLevel(
		env.CortisolLevel,
		env.NeurochemicalParams.CortisolDecayRate,
		cortisolProduction,
		env.NeurochemicalParams.CortisolMaxLevel,
	)

	// Dopamine update
	// Production can be base rate + event-dependent (e.g., successful task, output neuron firing)
	dopamineProduction := float64(env.NeurochemicalParams.DopamineProductionRate)
	// Event-based production (e.g., if certain output neurons fire) needs to be triggered from CrowNet.
	// For now, only base production is handled here.
	// Add env.ProduceDopamineEvent(amount) to be called by CrowNet when appropriate.
	env.DopamineLevel = updateChemicalLevel(
		env.DopamineLevel,
		env.NeurochemicalParams.DopamineDecayRate,
		dopamineProduction, // This should include event-based production summed up before this call
		env.NeurochemicalParams.DopamineMaxLevel,
	)

	// After updating levels, recalculate modulation factors
	env.recalculateModulationFactors(simParams)
}

// recalculateModulationFactors updates the learning rate and synaptogenesis modulation factors
// based on the current cortisol and dopamine levels.
func (env *Environment) recalculateModulationFactors(simParams *config.SimulationParameters) {
	if env.NeurochemicalParams == nil || simParams == nil {
		// Consider logging: log.Println("Warning: Environment.recalculateModulationFactors called with nil params. Factors remain default.")
		return
	}

	// Normalize levels (0 to 1) if params define max levels > 0
	normCortisol := 0.0
	if env.NeurochemicalParams.CortisolMaxLevel > 0 {
		normCortisol = float64(env.CortisolLevel / env.NeurochemicalParams.CortisolMaxLevel)
	}
	normDopamine := 0.0
	if env.NeurochemicalParams.DopamineMaxLevel > 0 {
		normDopamine = float64(env.DopamineLevel / env.NeurochemicalParams.DopamineMaxLevel)
	}

	// Learning Rate Modulation
	// LR_mod = 1.0 (base) + (CortisolEffect * normCortisol) + (DopamineEffect * normDopamine)
	// Effects can be positive or negative factors from NeurochemicalParams.
	lrMod := 1.0
	lrMod += float64(env.NeurochemicalParams.CortisolInfluenceOnLR) * normCortisol
	lrMod += float64(env.NeurochemicalParams.DopamineInfluenceOnLR) * normDopamine
	// Clamp modulation factor (e.g., to prevent negative or excessively high rates)
	// simParams.Learning.MinLearningRateFactor provides a floor.
	if lrMod < float64(simParams.Learning.MinLearningRateFactor) {
		lrMod = float64(simParams.Learning.MinLearningRateFactor)
	}
	env.LearningRateModulationFactor = common.Factor(lrMod)

	// Synaptogenesis Modulation (similar logic)
	synaptoMod := 1.0
	synaptoMod += float64(env.NeurochemicalParams.CortisolInfluenceOnSynapto) * normCortisol
	synaptoMod += float64(env.NeurochemicalParams.DopamineInfluenceOnSynapto) * normDopamine
	// Clamping for synaptogenesis factor (e.g., must be non-negative)
	if synaptoMod < 0 {
		synaptoMod = 0
	}
	env.SynaptogenesisModulationFactor = common.Factor(synaptoMod)
}

// ApplyEffectsToNeurons modifies properties of all neurons in the network
// based on current neurochemical levels (e.g., adjusting firing thresholds).
// This is typically called once per simulation cycle after levels are updated.
func (env *Environment) ApplyEffectsToNeurons(
	neurons map[common.NeuronID]*neuron.Neuron, // Changed from slice to map
	simParams *config.SimulationParameters,
) {
	if env.NeurochemicalParams == nil || simParams == nil {
		// Consider logging: log.Println("Warning: Environment.ApplyEffectsToNeurons called with nil params. Neuron thresholds not changed.")
		return
	}

	normCortisol := 0.0
	if env.NeurochemicalParams.CortisolMaxLevel > 0 {
		normCortisol = float64(env.CortisolLevel / env.NeurochemicalParams.CortisolMaxLevel)
	}
	normDopamine := 0.0
	if env.NeurochemicalParams.DopamineMaxLevel > 0 {
		normDopamine = float64(env.DopamineLevel / env.NeurochemicalParams.DopamineMaxLevel)
	}

	for _, n := range neurons {
		// Reset to base threshold first
		modifiedThreshold := float64(n.BaseFiringThreshold)

		// Apply cortisol effect on threshold
		modifiedThreshold = applyChemicalInfluence(modifiedThreshold,
			env.NeurochemicalParams.FiringThresholdIncreaseOnCort, normCortisol)

		// Apply dopamine effect on threshold (potentially on the already cortisol-modified threshold)
		modifiedThreshold = applyChemicalInfluence(modifiedThreshold,
			env.NeurochemicalParams.FiringThresholdIncreaseOnDopa, normDopamine)

		// Ensure threshold doesn't go below a minimum (e.g., 0 or a small positive value)
		// This minimum could be part of SimParams if needed. For now, ensure non-negative.
		if modifiedThreshold < 0 {
			modifiedThreshold = 0 // Or some other floor like neuron.MinFiringThreshold
		}
		n.CurrentFiringThreshold = common.Threshold(modifiedThreshold)
	}
}

// ApplyEffectsToSingleNeuron modifies properties of a single neuron.
// Factorized out for use when advancing state of individual neurons if preferred.
func (env *Environment) ApplyEffectsToSingleNeuron(
	n *neuron.Neuron,
	simParams *config.SimulationParameters,
) {
	if env.NeurochemicalParams == nil || simParams == nil || n == nil {
		return
	}
	// This function body is identical to the loop body in ApplyEffectsToNeurons.
	// It's duplicated for clarity or if ApplyEffectsToNeurons is removed.
	// For DRY, ApplyEffectsToNeurons could just iterate and call this.
	normCortisol := 0.0
	if env.NeurochemicalParams.CortisolMaxLevel > 0 {
		normCortisol = float64(env.CortisolLevel / env.NeurochemicalParams.CortisolMaxLevel)
	}
	normDopamine := 0.0
	if env.NeurochemicalParams.DopamineMaxLevel > 0 {
		normDopamine = float64(env.DopamineLevel / env.NeurochemicalParams.DopamineMaxLevel)
	}

	modifiedThreshold := float64(n.BaseFiringThreshold)
	modifiedThreshold = applyChemicalInfluence(modifiedThreshold,
		env.NeurochemicalParams.FiringThresholdIncreaseOnCort, normCortisol)
	modifiedThreshold = applyChemicalInfluence(modifiedThreshold,
		env.NeurochemicalParams.FiringThresholdIncreaseOnDopa, normDopamine)

	if modifiedThreshold < 0 {
		modifiedThreshold = 0
	}
	n.CurrentFiringThreshold = common.Threshold(modifiedThreshold)
}

// GetModulationFactor provides a generic way to get a combined modulation factor.
// This can be used by other systems (like learning, synaptogenesis) to query
// how their base rates should be modulated by the current chemical state.
// minFactor ensures the modulation doesn't reduce the effect below a certain floor.
func (env *Environment) GetModulationFactor(
	cortisolInfluence common.Factor,
	dopamineInfluence common.Factor,
	minFactor common.Factor,
) common.Factor {
	if env.NeurochemicalParams == nil {
		return 1.0 // Default if params not set
	}
	normCortisol := 0.0
	if env.NeurochemicalParams.CortisolMaxLevel > 0 {
		normCortisol = float64(env.CortisolLevel / env.NeurochemicalParams.CortisolMaxLevel)
	}
	normDopamine := 0.0
	if env.NeurochemicalParams.DopamineMaxLevel > 0 {
		normDopamine = float64(env.DopamineLevel / env.NeurochemicalParams.DopamineMaxLevel)
	}

	modFactor := 1.0
	modFactor += float64(cortisolInfluence) * normCortisol
	modFactor += float64(dopamineInfluence) * normDopamine

	if modFactor < float64(minFactor) {
		modFactor = float64(minFactor)
	}
	return common.Factor(modFactor)
}

// ProduceDopamineEvent is called to simulate a dopamine release event.
// It increases the dopamine level by a specified amount, considering production parameters.
// This is typically called by CrowNet based on simulation events (e.g., reward).
func (env *Environment) ProduceDopamineEvent(eventMagnitude float64, simParams *config.SimulationParameters) {
	if env.NeurochemicalParams == nil || simParams == nil {
		return
	}
	// Production is eventMagnitude * DopamineProductionPerEvent
	// This amount is then added to the existing dopamine level in the next UpdateLevels call,
	// or handled immediately here if UpdateLevels is not called frequently enough.
	// For simplicity, let's assume this directly adds to a temporary "producedThisCycle" buffer
	// which UpdateLevels will then use.
	// If direct update is needed:
	productionAmount := eventMagnitude * float64(env.NeurochemicalParams.DopamineProductionPerEvent)
	env.DopamineLevel += common.Level(productionAmount)
	// Clamp immediately if direct update, or let UpdateLevels handle clamping.
	if env.NeurochemicalParams.DopamineMaxLevel > 0 &&
		env.DopamineLevel > env.NeurochemicalParams.DopamineMaxLevel {
		env.DopamineLevel = env.NeurochemicalParams.DopamineMaxLevel
	}
	if env.DopamineLevel < 0 { // Should not happen if productionAmount is positive
		env.DopamineLevel = 0
	}
	// After direct update, factors might need immediate recalculation if effects are immediate.
	// env.recalculateModulationFactors(simParams) // Potentially call this if effects are instant.
}
