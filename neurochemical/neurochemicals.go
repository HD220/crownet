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
	activePulses []*pulse.Pulse,
	cortisolGlandPosition common.Point,
	simParams *config.SimulationParameters,
) {
	// --- Atualização do Cortisol ---
	pulsesHittingGland := 0
	for _, p := range activePulses {
		// Considerar apenas pulsos excitatórios para produção de cortisol
		if p.BaseSignalValue > 0 { // Assumindo que BaseSignalValue > 0 é excitatório
			// Verificar se o pulso está na "casca de efeito" correta para atingir a glândula
			shellStart, shellEnd := p.GetEffectShellForCycle(simParams)
			distToGland := space.EuclideanDistance(p.OriginPosition, cortisolGlandPosition)

			// O pulso atinge a glândula se a posição da glândula estiver dentro da casca de efeito do pulso
			// E também dentro do raio de sensibilidade da própria glândula.
			// A segunda condição (distToGland <= simParams.CortisolGlandRadius) é mais um verificador de proximidade geral
			// do que a casca do pulso em si. O correto é se a casca do pulso INTERSECTA a esfera da glândula.
			// Para simplificar, vamos assumir que se a casca do pulso contém o *ponto central* da glândula, conta como hit.
			if distToGland >= shellStart && distToGland < shellEnd {
				pulsesHittingGland++
			}
		}
	}

	if pulsesHittingGland > 0 {
		production := float64(pulsesHittingGland) * simParams.CortisolProductionPerHit
		env.CortisolLevel += common.Level(production)
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
	for _, n := range neurons {
		if n.Type == neuron.Dopaminergic && n.CurrentState == neuron.Firing {
			dopamineProducedThisCycle += simParams.DopamineProductionPerEvent
		}
	}

	if dopamineProducedThisCycle > 0 {
		env.DopamineLevel += common.Level(dopamineProducedThisCycle)
	}

	env.DopamineLevel -= env.DopamineLevel * common.Level(simParams.DopamineDecayRate)
	if env.DopamineLevel < 0 {
		env.DopamineLevel = 0
	}
	if env.DopamineLevel > common.Level(simParams.DopamineMaxLevel) {
		env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	}

	// Após atualizar os níveis, recalcular os fatores de modulação.
	env.recalculateModulationFactors(simParams)
}

// recalculateModulationFactors atualiza os fatores de aprendizado e sinaptogênese
// com base nos níveis atuais de cortisol e dopamina.
func (env *Environment) recalculateModulationFactors(simParams *config.SimulationParameters) {
	// Modulação da Taxa de Aprendizado
	lrFactor := 1.0
	// Efeito da Dopamina na Taxa de Aprendizado
	if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 {
		normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
		// Modelo: LR_Multiplier = 1 + (MaxMultiplier - 1) * normalized_chem
		// Se MaxMultiplier = 2, então vai de 1 (sem dopamina) a 2 (max dopamina)
		dopamineEffectOnLR := 1.0 + (simParams.MaxDopamineLearningMultiplier-1.0)*normalizedDopamine
		lrFactor *= dopamineEffectOnLR
	}
	// Efeito do Cortisol na Taxa de Aprendizado
	if env.CortisolLevel >= common.Level(simParams.CortisolHighEffectThreshold) {
		// Modelo: LR_Multiplier diminui de 1 (no CortisolHighEffectThreshold)
		// até CortisolLearningSuppressionFactor (no CortisolMaxLevel)
		suppressionRange := simParams.CortisolMaxLevel - simParams.CortisolHighEffectThreshold
		cortisolSuppressionFactor := simParams.CortisolLearningSuppressionFactor
		if suppressionRange > 0 { // Evitar divisão por zero se HighEffectThreshold == MaxLevel
			// quanto do caminho de HighEffectThreshold para MaxLevel o cortisol atual percorreu (0 a 1)
			t := (float64(env.CortisolLevel) - simParams.CortisolHighEffectThreshold) / suppressionRange
			t = math.Max(0, math.Min(1, t)) // clamp t entre 0 e 1
			// Interpola linearmente entre 1.0 (em HighEffectThreshold) e CortisolLearningSuppressionFactor (em MaxLevel)
			effectiveSuppression := 1.0 - t*(1.0-cortisolSuppressionFactor)
			lrFactor *= effectiveSuppression
		} else { // Cortisol está em HighEffectThreshold ou acima, e HighEffectThreshold == MaxLevel
			lrFactor *= cortisolSuppressionFactor
		}
	}
	// Garantir que o fator de aprendizado não seja menor que um mínimo.
	lrFactor = math.Max(simParams.MinLearningRateFactor, lrFactor)
	env.LearningRateModulationFactor = common.Factor(lrFactor)

	// Modulação da Sinaptogênese
	synFactor := 1.0
	// Efeito do Cortisol na Sinaptogênese (redução em níveis altos)
	if env.CortisolLevel >= common.Level(simParams.CortisolHighEffectThreshold) {
		reductionRange := simParams.CortisolMaxLevel - simParams.CortisolHighEffectThreshold
		cortisolSynReduction := simParams.SynaptogenesisReductionFactor
		if reductionRange > 0 {
			t := (float64(env.CortisolLevel) - simParams.CortisolHighEffectThreshold) / reductionRange
			t = math.Max(0, math.Min(1, t))
			effectiveReduction := 1.0 - t*(1.0-cortisolSynReduction)
			synFactor *= effectiveReduction
		} else {
			synFactor *= cortisolSynReduction
		}
	}
	// Efeito da Dopamina na Sinaptogênese (aumento)
	if simParams.DopamineMaxLevel > 0 {
		if env.DopamineLevel > 0 { // Apenas aplicar se houver dopamina
			normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
			dopamineSynIncrease := 1.0 + (simParams.DopamineSynaptogenesisIncreaseFactor-1.0)*normalizedDopamine
			synFactor *= dopamineSynIncrease
		}
	}
	env.SynaptogenesisModulationFactor = common.Factor(math.Max(0.0, synFactor)) // Garantir não negativo
}

// ApplyEffectsToNeurons modifica os limiares de disparo dos neurônios
// com base nos níveis atuais de cortisol e dopamina.
func (env *Environment) ApplyEffectsToNeurons(neurons []*neuron.Neuron, simParams *config.SimulationParameters) {
	for _, n := range neurons {
		baseThreshold := float64(n.BaseFiringThreshold)
		currentEffectiveThreshold := baseThreshold

		// Efeito do Cortisol no Limiar (U-shaped)
		cortisolEffectFactor := 1.0
		if env.CortisolLevel < common.Level(simParams.CortisolMinEffectThreshold) {
			// Sem efeito significativo
		} else if env.CortisolLevel < common.Level(simParams.CortisolOptimalLowThreshold) {
			t := (float64(env.CortisolLevel) - simParams.CortisolMinEffectThreshold) / (simParams.CortisolOptimalLowThreshold - simParams.CortisolMinEffectThreshold)
			cortisolEffectFactor = 1.0 - t*(1.0-simParams.MaxThresholdReductionFactor)
		} else if env.CortisolLevel <= common.Level(simParams.CortisolOptimalHighThreshold) {
			cortisolEffectFactor = simParams.MaxThresholdReductionFactor
		} else if env.CortisolLevel < common.Level(simParams.CortisolHighEffectThreshold) {
			t := (float64(env.CortisolLevel) - simParams.CortisolOptimalHighThreshold) / (simParams.CortisolHighEffectThreshold - simParams.CortisolOptimalHighThreshold)
			cortisolEffectFactor = simParams.MaxThresholdReductionFactor + t*(1.0-simParams.MaxThresholdReductionFactor)
		} else { // CortisolLevel >= CortisolHighEffectThreshold
			if simParams.CortisolMaxLevel > simParams.CortisolHighEffectThreshold {
				t := (float64(env.CortisolLevel) - simParams.CortisolHighEffectThreshold) / (simParams.CortisolMaxLevel - simParams.CortisolHighEffectThreshold)
				t = math.Max(0, math.Min(1, t))
				cortisolEffectFactor = 1.0 + t*(simParams.ThresholdIncreaseFactorHigh-1.0)
			} else { // Cortisol está no ou acima do HighEffectThreshold que é igual ao MaxLevel
				cortisolEffectFactor = simParams.ThresholdIncreaseFactorHigh
			}
		}
		currentEffectiveThreshold *= cortisolEffectFactor

		// Efeito da Dopamina no Limiar (multiplicativo sobre o efeito do cortisol)
		dopamineEffectFactor := 1.0
		if simParams.DopamineMaxLevel > 0 && env.DopamineLevel > 0 {
			normalizedDopamine := math.Min(1.0, float64(env.DopamineLevel)/simParams.DopamineMaxLevel)
			dopamineEffectFactor = 1.0 + (simParams.DopamineThresholdIncreaseFactor-1.0)*normalizedDopamine
		}
		// Dopamina geralmente aumenta o limiar, então o fator é >= 1.0
		currentEffectiveThreshold *= math.Max(1.0, dopamineEffectFactor)

		n.CurrentFiringThreshold = common.Threshold(math.Max(0.01, currentEffectiveThreshold)) // Limiar mínimo absoluto
	}
}
```
