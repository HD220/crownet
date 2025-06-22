package neurochemical

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron" // Para acessar neuron.Type e modificar neuron.CurrentFiringThreshold
	"crownet/pulse"  // Para verificar pulsos atingindo a glândula de cortisol
	"crownet/space"  // Para EuclideanDistance
	"math"
)

// Environment representa o estado dos neuroquímicos na rede.
type Environment struct {
	CortisolLevel common.Level
	DopamineLevel common.Level

	// Fatores de modulação calculados que afetam outras partes da simulação.
	// Estes são atualizados a cada ciclo com base nos níveis químicos.
	SynaptogenesisModulationFactor common.Factor
	LearningRateModulationFactor   common.Factor
}

// NewEnvironment cria um novo ambiente neuroquímico com níveis iniciais.
func NewEnvironment() *Environment {
	return &Environment{
		CortisolLevel: 0.0,
		DopamineLevel: 0.0,
		// Fatores iniciam em 1.0 (sem modulação)
		SynaptogenesisModulationFactor: 1.0,
		LearningRateModulationFactor:   1.0,
	}
}

// UpdateLevels atualiza os níveis de cortisol e dopamina com base na atividade da rede e decaimento.
func (env *Environment) UpdateLevels(
	neurons []*neuron.Neuron,
	activePulses []*pulse.Pulse, // Alterado de *pulse.PulseList para []*pulse.Pulse para corresponder ao uso em network.go
	cortisolGlandPosition common.Point,
	simParams *config.SimulationParameters,
) {
	// --- Atualização do Cortisol ---
	pulsesHittingGland := 0
	if simParams.CortisolProductionPerHit > 0 { // Só calcular se houver produção
		for _, p := range activePulses {
			if p.BaseSignalValue > 0 { // Considerar apenas pulsos excitatórios
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

	// --- Atualização da Dopamina ---
	dopamineProducedThisCycle := 0.0
	if simParams.DopamineProductionPerEvent > 0 { // Só calcular se houver produção
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

// recalculateModulationFactors atualiza os fatores de aprendizado e sinaptogênese
// com base nos níveis atuais de cortisol e dopamina.
func (env *Environment) recalculateModulationFactors(simParams *config.SimulationParameters) {
	// Modulação da Taxa de Aprendizado
	lrFactor := 1.0
	// Efeito da Dopamina: (1 + Influence) * LR, onde Influence é -1 a +inf.
	// Se simParams.DopamineInfluenceOnLR = 0.8, fator é 1.8. Se -0.5, fator é 0.5.
	if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 { // Evitar divisão por zero e processamento desnecessário
		normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
		lrFactor *= (1.0 + simParams.DopamineInfluenceOnLR*normalizedDopamine)
	}

	// Efeito do Cortisol: (1 + Influence) * LR
	if simParams.CortisolMaxLevel > 0 && env.CortisolLevel > 0 {
		normalizedCortisol := math.Min(1.0, float64(env.CortisolLevel)/simParams.CortisolMaxLevel)
		// CortisolInfluenceOnLR é esperado ser negativo (ex: -0.5) para suprimir
		lrFactor *= (1.0 + simParams.CortisolInfluenceOnLR*normalizedCortisol)
	}
	// O modelo anterior de CortisolLearningSuppressionFactor e CortisolHighEffectThreshold foi substituído
	// por um CortisolInfluenceOnLR mais simples, aplicado proporcionalmente.
	// Se for necessário o modelo de limiar, ele precisaria ser reimplementado.
	// Por agora, usamos o simParams.CortisolInfluenceOnLR.

	lrFactor = math.Max(simParams.MinLearningRateFactor, lrFactor) // Garante mínimo
	env.LearningRateModulationFactor = common.Factor(lrFactor)

	// Modulação da Sinaptogênese
	synFactor := 1.0
	// Efeito da Dopamina: (1 + Influence) * SynRate
	if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 {
		normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
		synFactor *= (1.0 + simParams.DopamineInfluenceOnSynapto*normalizedDopamine)
	}
	// Efeito do Cortisol: (1 + Influence) * SynRate
	if simParams.CortisolMaxLevel > 0 && env.CortisolLevel > 0 {
		normalizedCortisol := math.Min(1.0, float64(env.CortisolLevel)/simParams.CortisolMaxLevel)
		synFactor *= (1.0 + simParams.CortisolInfluenceOnSynapto*normalizedCortisol)
	}
	// O modelo anterior de SynaptogenesisReductionFactor foi substituído.

	env.SynaptogenesisModulationFactor = common.Factor(math.Max(0.0, synFactor)) // Garantir não negativo
}

// ApplyEffectsToNeurons modifica os limiares de disparo dos neurônios.
// Usa FiringThresholdIncreaseOnCort e FiringThresholdIncreaseOnDopa de simParams.
func (env *Environment) ApplyEffectsToNeurons(neurons []*neuron.Neuron, simParams *config.SimulationParameters) {
	for _, n := range neurons {
		baseThreshold := float64(n.BaseFiringThreshold)
		modifiedThreshold := baseThreshold

		// Efeito do Cortisol
		// FiringThresholdIncreaseOnCort: positivo aumenta, negativo diminui.
		// Ex: base * (1 + FiringThresholdIncreaseOnCort * normalizedCortisol)
		if simParams.CortisolMaxLevel > 0 && env.CortisolLevel > 0 {
			normalizedCortisol := math.Min(1.0, float64(env.CortisolLevel)/simParams.CortisolMaxLevel)
			modifiedThreshold *= (1.0 + simParams.FiringThresholdIncreaseOnCort*normalizedCortisol)
		}

		// Efeito da Dopamina (aplicado sobre o limiar já modificado pelo cortisol)
		// FiringThresholdIncreaseOnDopa: positivo aumenta, negativo diminui.
		// Ex: current_modified_threshold * (1 + FiringThresholdIncreaseOnDopa * normalizedDopamine)
		if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 {
			normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
			modifiedThreshold *= (1.0 + simParams.FiringThresholdIncreaseOnDopa*normalizedDopamine)
		}

		// A lógica U-shape complexa anterior foi removida em favor dos fatores diretos.
		// Se a lógica U-shape for um requisito, precisará ser reimplementada usando
		// os parâmetros CortisolMinEffectThreshold, OptimalLow, OptimalHigh, MaxReduction, HighIncrease.

		n.CurrentFiringThreshold = common.Threshold(math.Max(0.01, modifiedThreshold)) // Limiar mínimo absoluto
	}
}
```
