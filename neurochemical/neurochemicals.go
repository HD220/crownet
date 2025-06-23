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
	minFiringThresholdValue = 0.01
)

// getNormalizedLevel calculates the normalized level of a chemical, ensuring it's between 0.0 and 1.0.
// Returns 0 if maxLevel is not positive.
func getNormalizedLevel(level common.Level, maxLevel float64) float64 {
	if maxLevel <= 0 {
		return 0.0 // Avoid division by zero or negative max level issues
	}
	normalized := float64(level) / maxLevel
	return math.Min(1.0, math.Max(0.0, normalized)) // Ensure it's clamped between 0 and 1
}

type Environment struct {
	CortisolLevel common.Level
	DopamineLevel common.Level
	SynaptogenesisModulationFactor common.Factor
	LearningRateModulationFactor   common.Factor
}

// calculateCortisolStimulation calculates cortisol production based on pulses hitting a gland.
func calculateCortisolStimulation(activePulses []*pulse.Pulse, cortisolGlandPosition common.Point, simParams *config.SimulationParameters) float64 {
	if simParams.CortisolProductionPerHit <= 0 {
		return 0.0
	}

	pulsesHittingGland := 0
	for _, p := range activePulses {
		// Consider only active pulses that have a positive base signal value,
		// implying they are excitatory or have some "strength".
		if p.BaseSignalValue > 0 {
			shellStart, shellEnd := p.GetEffectShellForCycle(simParams)
			distToGland := space.EuclideanDistance(p.OriginPosition, cortisolGlandPosition)
			// Check if the gland is within the pulse's spherical shell of effect for the current cycle.
			if distToGland >= shellStart && distToGland < shellEnd {
				pulsesHittingGland++
			}
		}
	}
	return float64(pulsesHittingGland) * simParams.CortisolProductionPerHit
}

// calculateDopamineProduction calculates dopamine production based on firing dopaminergic neurons.
func calculateDopamineProduction(neurons []*neuron.Neuron, simParams *config.SimulationParameters) float64 {
	if simParams.DopamineProductionPerEvent <= 0 {
		return 0.0
	}

	dopamineProducedThisCycle := 0.0
	for _, n := range neurons {
		if n.Type == neuron.Dopaminergic && n.CurrentState == neuron.Firing {
			dopamineProducedThisCycle += simParams.DopamineProductionPerEvent
		}
	}
	return dopamineProducedThisCycle
}

// updateChemicalLevel applies decay and event-based production to a chemical level, ensuring it stays within bounds.
func updateChemicalLevel(currentLevel common.Level, decayRate common.Rate, productionThisCycle float64, maxLevel common.Level) common.Level {
	level := currentLevel

	// Add event-based production
	level += common.Level(productionThisCycle)

	// Apply decay (proportional to current level)
	level -= level * common.Level(decayRate)

	// Clamp levels
	if level < 0 {
		level = 0
	}
	if level > maxLevel { // Assumes maxLevel is appropriately set (e.g., non-negative)
		level = maxLevel
	}
	return level
}

func NewEnvironment() *Environment {
	return &Environment{
		CortisolLevel: 0.0,
		DopamineLevel: 0.0,
		SynaptogenesisModulationFactor: 1.0,
		LearningRateModulationFactor:   1.0,
	}
}

func (env *Environment) UpdateLevels(
	neurons []*neuron.Neuron,
	activePulses []*pulse.Pulse,
	cortisolGlandPosition common.Point,
	simParams *config.SimulationParameters,
) {
	// Calculate production for this cycle
	cortisolProduction := calculateCortisolStimulation(activePulses, cortisolGlandPosition, simParams)
	dopamineProduction := calculateDopamineProduction(neurons, simParams)

	// Update levels using the helper function
	env.CortisolLevel = updateChemicalLevel(
		env.CortisolLevel,
		common.Rate(simParams.CortisolDecayRate),
		cortisolProduction,
		common.Level(simParams.CortisolMaxLevel),
	)

	env.DopamineLevel = updateChemicalLevel(
		env.DopamineLevel,
		common.Rate(simParams.DopamineDecayRate),
		dopamineProduction,
		common.Level(simParams.DopamineMaxLevel),
	)

	// After levels are updated, recalculate their effects on modulation factors
	env.recalculateModulationFactors(simParams)
}

func (env *Environment) recalculateModulationFactors(simParams *config.SimulationParameters) {
	lrFactor := 1.0
	// Apply Dopamine effect on Learning Rate
	if env.DopamineLevel > 0 { // Only apply if there's some dopamine
		normalizedDopamine := getNormalizedLevel(env.DopamineLevel, simParams.DopamineMaxLevel)
		if normalizedDopamine > 0 { // Ensure normalized level is positive before applying factor
			lrFactor *= (1.0 + simParams.DopamineInfluenceOnLR*normalizedDopamine)
		}
	}

	// Apply Cortisol effect on Learning Rate
	if env.CortisolLevel > 0 { // Only apply if there's some cortisol
		normalizedCortisol := getNormalizedLevel(env.CortisolLevel, simParams.CortisolMaxLevel)
		if normalizedCortisol > 0 { // Ensure normalized level is positive
			lrFactor *= (1.0 + simParams.CortisolInfluenceOnLR*normalizedCortisol)
		}
	}

	lrFactor = math.Max(simParams.MinLearningRateFactor, lrFactor)
	env.LearningRateModulationFactor = common.Factor(lrFactor)

	synFactor := 1.0
	// Apply Dopamine effect on Synaptogenesis Factor
	if env.DopamineLevel > 0 {
		normalizedDopamine := getNormalizedLevel(env.DopamineLevel, simParams.DopamineMaxLevel)
		if normalizedDopamine > 0 {
			synFactor *= (1.0 + simParams.DopamineInfluenceOnSynapto*normalizedDopamine)
		}
	}

	// Apply Cortisol effect on Synaptogenesis Factor
	if env.CortisolLevel > 0 {
		normalizedCortisol := getNormalizedLevel(env.CortisolLevel, simParams.CortisolMaxLevel)
		if normalizedCortisol > 0 {
			synFactor *= (1.0 + simParams.CortisolInfluenceOnSynapto*normalizedCortisol)
		}
	}

	env.SynaptogenesisModulationFactor = common.Factor(math.Max(0.0, synFactor)) // Synaptogenesis factor cannot be negative
}

func (env *Environment) ApplyEffectsToNeurons(neurons []*neuron.Neuron, simParams *config.SimulationParameters) {
	for _, n := range neurons {
		baseThreshold := float64(n.BaseFiringThreshold)
		modifiedThreshold := baseThreshold

		// Apply Cortisol effect on Firing Threshold
		if env.CortisolLevel > 0 {
			normalizedCortisol := getNormalizedLevel(env.CortisolLevel, simParams.CortisolMaxLevel)
			if normalizedCortisol > 0 {
				modifiedThreshold *= (1.0 + simParams.FiringThresholdIncreaseOnCort*normalizedCortisol)
			}
		}

		// Apply Dopamine effect on Firing Threshold
		if env.DopamineLevel > 0 {
			normalizedDopamine := getNormalizedLevel(env.DopamineLevel, simParams.DopamineMaxLevel)
			if normalizedDopamine > 0 {
				modifiedThreshold *= (1.0 + simParams.FiringThresholdIncreaseOnDopa*normalizedDopamine)
			}
		}
		n.CurrentFiringThreshold = common.Threshold(math.Max(minFiringThresholdValue, modifiedThreshold))
	}
}
