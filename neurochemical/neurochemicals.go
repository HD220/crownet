package neurochemical

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/pulse"
	"crownet/space"
	"math"
)

type Environment struct {
	CortisolLevel common.Level
	DopamineLevel common.Level
	SynaptogenesisModulationFactor common.Factor
	LearningRateModulationFactor   common.Factor
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
	pulsesHittingGland := 0
	if simParams.CortisolProductionPerHit > 0 {
		for _, p := range activePulses {
			if p.BaseSignalValue > 0 {
				shellStart, shellEnd := p.GetEffectShellForCycle(simParams)
				distToGland := space.EuclideanDistance(p.OriginPosition, cortisolGlandPosition)
				if distToGland >= shellStart && distToGland < shellEnd {
					pulsesHittingGland++
				}
			}
		}
		if pulsesHittingGland > 0 {
			production := float64(pulsesHittingGland) * simParams.CortisolProductionPerHit
			env.CortisolLevel += common.Level(production)
		}
	}

	env.CortisolLevel -= env.CortisolLevel * common.Level(simParams.CortisolDecayRate)
	if env.CortisolLevel < 0 {
		env.CortisolLevel = 0
	}
	if env.CortisolLevel > common.Level(simParams.CortisolMaxLevel) {
		env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	}

	dopamineProducedThisCycle := 0.0
	if simParams.DopamineProductionPerEvent > 0 {
		for _, n := range neurons {
			if n.Type == neuron.Dopaminergic && n.CurrentState == neuron.Firing {
				dopamineProducedThisCycle += simParams.DopamineProductionPerEvent
			}
		}
		if dopamineProducedThisCycle > 0 {
			env.DopamineLevel += common.Level(dopamineProducedThisCycle)
		}
	}

	env.DopamineLevel -= env.DopamineLevel * common.Level(simParams.DopamineDecayRate)
	if env.DopamineLevel < 0 {
		env.DopamineLevel = 0
	}
	if env.DopamineLevel > common.Level(simParams.DopamineMaxLevel) {
		env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	}

	env.recalculateModulationFactors(simParams)
}

func (env *Environment) recalculateModulationFactors(simParams *config.SimulationParameters) {
	lrFactor := 1.0
	if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 {
		normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
		lrFactor *= (1.0 + simParams.DopamineInfluenceOnLR*normalizedDopamine)
	}

	if simParams.CortisolMaxLevel > 0 && env.CortisolLevel > 0 {
		normalizedCortisol := math.Min(1.0, float64(env.CortisolLevel)/simParams.CortisolMaxLevel)
		lrFactor *= (1.0 + simParams.CortisolInfluenceOnLR*normalizedCortisol)
	}

	lrFactor = math.Max(simParams.MinLearningRateFactor, lrFactor)
	env.LearningRateModulationFactor = common.Factor(lrFactor)

	synFactor := 1.0
	if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 {
		normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
		synFactor *= (1.0 + simParams.DopamineInfluenceOnSynapto*normalizedDopamine)
	}
	if simParams.CortisolMaxLevel > 0 && env.CortisolLevel > 0 {
		normalizedCortisol := math.Min(1.0, float64(env.CortisolLevel)/simParams.CortisolMaxLevel)
		synFactor *= (1.0 + simParams.CortisolInfluenceOnSynapto*normalizedCortisol)
	}

	env.SynaptogenesisModulationFactor = common.Factor(math.Max(0.0, synFactor))
}

func (env *Environment) ApplyEffectsToNeurons(neurons []*neuron.Neuron, simParams *config.SimulationParameters) {
	for _, n := range neurons {
		baseThreshold := float64(n.BaseFiringThreshold)
		modifiedThreshold := baseThreshold

		if simParams.CortisolMaxLevel > 0 && env.CortisolLevel > 0 {
			normalizedCortisol := math.Min(1.0, float64(env.CortisolLevel)/simParams.CortisolMaxLevel)
			modifiedThreshold *= (1.0 + simParams.FiringThresholdIncreaseOnCort*normalizedCortisol)
		}

		if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 {
			normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
			modifiedThreshold *= (1.0 + simParams.FiringThresholdIncreaseOnDopa*normalizedDopamine)
		}

		n.CurrentFiringThreshold = common.Threshold(math.Max(0.01, modifiedThreshold))
	}
}
