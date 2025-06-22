package neurochemical_test

import (
	"crownet/common"
	"crownet/config"
	"crownet/neurochemical"
	"crownet/neuron"
	"crownet/pulse"
	"math"
	"testing"
)

// Helper para comparar floats com tolerância
func floatEquals(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) < tolerance
}

func TestNewEnvironment(t *testing.T) {
	env := neurochemical.NewEnvironment()
	if env == nil {
		t.Fatalf("NewEnvironment returned nil")
	}
	if !floatEquals(float64(env.CortisolLevel), 0.0, 1e-9) {
		t.Errorf("Initial CortisolLevel: expected 0.0, got %f", env.CortisolLevel)
	}
	if !floatEquals(float64(env.DopamineLevel), 0.0, 1e-9) {
		t.Errorf("Initial DopamineLevel: expected 0.0, got %f", env.DopamineLevel)
	}
	if !floatEquals(float64(env.LearningRateModulationFactor), 1.0, 1e-9) {
		t.Errorf("Initial LearningRateModulationFactor: expected 1.0, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 1.0, 1e-9) {
		t.Errorf("Initial SynaptogenesisModulationFactor: expected 1.0, got %f", env.SynaptogenesisModulationFactor)
	}
}

func TestUpdateLevels_Cortisol(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.CortisolProductionPerHit = 0.1
	simParams.CortisolDecayRate = 0.05
	simParams.CortisolMaxLevel = 1.0
	simParams.PulsePropagationSpeed = 1.0 // Para simplificar cálculo de shell

	env := neurochemical.NewEnvironment()
	cortisolGlandPos := common.Point{0, 0, 0}

	// --- Caso 1: Sem pulsos atingindo a glândula ---
	activePulses := []*pulse.Pulse{}
	neurons := []*neuron.Neuron{} // Não usado para cortisol diretamente
	env.UpdateLevels(neurons, activePulses, cortisolGlandPos, &simParams)
	if !floatEquals(float64(env.CortisolLevel), 0.0, 1e-9) {
		t.Errorf("Cortisol C1: Expected 0.0 with no pulses, got %f", env.CortisolLevel)
	}

	// --- Caso 2: Um pulso excitatório atinge a glândula ---
	// Pulso se origina perto, propaga e sua casca de efeito cobre a glândula.
	// Origem do pulso: {0,0,0}, GlandPos: {0,0,0}
	// Pulso criado no ciclo anterior, CurrentDistance = 0, PropSpeed = 1.0
	// Após Propagate(): CurrentDistance = 1.0. Shell = [0, 1.0)
	// Gland (dist 0) está na casca.
	p1 := pulse.New(1, common.Point{0,0,0}, 1.0, 0, 10.0)
	p1.CurrentDistance = 0 // Simular que acabou de ser criado ou resetado para teste

	activePulses = []*pulse.Pulse{p1}
	env.CortisolLevel = 0 // Reset
	env.UpdateLevels(neurons, activePulses, cortisolGlandPos, &simParams)
	// Produção = 1 * 0.1 = 0.1. Nível = 0.1.
	// Decaimento = 0.1 * 0.05 = 0.005. Nível = 0.1 - 0.005 = 0.095
	expectedCortisol2 := 0.1 * (1.0 - simParams.CortisolDecayRate)
	if !floatEquals(float64(env.CortisolLevel), expectedCortisol2, 1e-9) {
		t.Errorf("Cortisol C2: Expected %f, got %f", expectedCortisol2, env.CortisolLevel)
	}

	// --- Caso 3: Pulso inibitório (não deve causar produção) ---
	pInhib := pulse.New(2, common.Point{0,0,0}, -1.0, 0, 10.0)
	pInhib.CurrentDistance = 0
	activePulses = []*pulse.Pulse{pInhib}
	env.CortisolLevel = 0 // Reset
	env.UpdateLevels(neurons, activePulses, cortisolGlandPos, &simParams)
	if !floatEquals(float64(env.CortisolLevel), 0.0, 1e-9) { // Apenas decaimento de 0 é 0
		t.Errorf("Cortisol C3: Expected 0.0 with inhibitory pulse, got %f", env.CortisolLevel)
	}

	// --- Caso 4: Produção leva ao clamping no MaxLevel ---
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel - 0.05) // Quase no máximo
	// p1 ainda está configurado para atingir a glândula. Produção = 0.1
	// Nível antes de decaimento = MaxLevel - 0.05 + 0.1 = MaxLevel + 0.05. Deve ser clampeado para MaxLevel.
	// Decaimento de MaxLevel = MaxLevel * decayRate.
	// Nível final = MaxLevel * (1 - decayRate)
	activePulses = []*pulse.Pulse{p1} // p1 já propagou uma vez no teste anterior, resetar para teste limpo
	p1_reset := pulse.New(1, common.Point{0,0,0}, 1.0, 0, 10.0)
	p1_reset.CurrentDistance = 0
	activePulses = []*pulse.Pulse{p1_reset}

	env.UpdateLevels(neurons, activePulses, cortisolGlandPos, &simParams)
	expectedCortisol4 := simParams.CortisolMaxLevel * (1.0 - simParams.CortisolDecayRate)
	if !floatEquals(float64(env.CortisolLevel), expectedCortisol4, 1e-9) {
		t.Errorf("Cortisol C4: Expected clamped level %f, got %f", expectedCortisol4, env.CortisolLevel)
	}

	// --- Caso 5: Pulso não atinge a glândula (longe demais) ---
	pFar := pulse.New(3, common.Point{5,5,5}, 1.0, 0, 10.0) // Origem Longe
	pFar.CurrentDistance = 0
	activePulses = []*pulse.Pulse{pFar}
	env.CortisolLevel = 0.5 // Nível inicial para observar decaimento
	env.UpdateLevels(neurons, activePulses, cortisolGlandPos, &simParams)
	expectedCortisol5 := 0.5 * (1.0 - simParams.CortisolDecayRate)
	if !floatEquals(float64(env.CortisolLevel), expectedCortisol5, 1e-9) {
		t.Errorf("Cortisol C5: Expected decay only %f, got %f", expectedCortisol5, env.CortisolLevel)
	}
}

func TestUpdateLevels_Dopamine(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.DopamineProductionPerEvent = 0.2
	simParams.DopamineDecayRate = 0.1
	simParams.DopamineMaxLevel = 1.0

	env := neurochemical.NewEnvironment()
	activePulses := []*pulse.Pulse{} // Não usado para dopamina

	n1Dopa := neuron.New(1, neuron.Dopaminergic, common.Point{}, &simParams)
	n2Excit := neuron.New(2, neuron.Excitatory, common.Point{}, &simParams)
	n3DopaNotFiring := neuron.New(3, neuron.Dopaminergic, common.Point{}, &simParams)
	n3DopaNotFiring.CurrentState = neuron.Resting


	// --- Caso 1: Sem neurônios dopaminérgicos disparando ---
	neurons1 := []*neuron.Neuron{n2Excit, n3DopaNotFiring}
	env.UpdateLevels(neurons1, activePulses, common.Point{}, &simParams)
	if !floatEquals(float64(env.DopamineLevel), 0.0, 1e-9) {
		t.Errorf("Dopamine C1: Expected 0.0 with no firing dopaminergic neurons, got %f", env.DopamineLevel)
	}

	// --- Caso 2: Um neurônio dopaminérgico dispara ---
	n1Dopa.CurrentState = neuron.Firing // Simular disparo
	neurons2 := []*neuron.Neuron{n1Dopa, n2Excit}
	env.DopamineLevel = 0 // Reset
	env.UpdateLevels(neurons2, activePulses, common.Point{}, &simParams)
	// Produção = 1 * 0.2 = 0.2. Nível = 0.2
	// Decaimento = 0.2 * 0.1 = 0.02. Nível = 0.2 - 0.02 = 0.18
	expectedDopamine2 := 0.2 * (1.0 - simParams.DopamineDecayRate)
	if !floatEquals(float64(env.DopamineLevel), expectedDopamine2, 1e-9) {
		t.Errorf("Dopamine C2: Expected %f, got %f", expectedDopamine2, env.DopamineLevel)
	}
	n1Dopa.CurrentState = neuron.Resting // Reset state para próximos testes se n1Dopa for reutilizado

	// --- Caso 3: Produção leva ao clamping no MaxLevel ---
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel - 0.1) // Quase no máximo: 1.0 - 0.1 = 0.9
	n1Dopa.CurrentState = neuron.Firing
	neurons3 := []*neuron.Neuron{n1Dopa} // Apenas um disparando
	// Produção = 0.2. Nível antes do decaimento = 0.9 + 0.2 = 1.1. Deve ser clampeado para 1.0 (MaxLevel).
	// Decaimento de MaxLevel = 1.0 * 0.1 = 0.1.
	// Nível final = 1.0 - 0.1 = 0.9
	env.UpdateLevels(neurons3, activePulses, common.Point{}, &simParams)
	expectedDopamine3 := simParams.DopamineMaxLevel * (1.0 - simParams.DopamineDecayRate)
	if !floatEquals(float64(env.DopamineLevel), expectedDopamine3, 1e-9) {
		t.Errorf("Dopamine C3: Expected clamped level %f, got %f", expectedDopamine3, env.DopamineLevel)
	}
	n1Dopa.CurrentState = neuron.Resting
}

func TestRecalculateModulationFactors(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	// Configurar influências para facilitar o teste
	simParams.DopamineMaxLevel = 1.0
	simParams.CortisolMaxLevel = 1.0
	simParams.DopamineInfluenceOnLR = 0.5  // Max LR factor contrib from Dopa: 1.5x
	simParams.CortisolInfluenceOnLR = -0.4 // Max LR factor supp from Cort: 0.6x
	simParams.MinLearningRateFactor = 0.1
	simParams.DopamineInfluenceOnSynapto = 0.2 // Max Synapto factor contrib from Dopa: 1.2x
	simParams.CortisolInfluenceOnSynapto = -0.8 // Max Synapto factor supp from Cort: 0.2x

	env := neurochemical.NewEnvironment()

	// Caso 1: Sem químicos, fatores devem ser 1.0
	env.CortisolLevel = 0.0
	env.DopamineLevel = 0.0
	// UpdateLevels chama recalculateModulationFactors internamente
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	if !floatEquals(float64(env.LearningRateModulationFactor), 1.0, 1e-9) {
		t.Errorf("LR Factor C1: Expected 1.0, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 1.0, 1e-9) {
		t.Errorf("Synapto Factor C1: Expected 1.0, got %f", env.SynaptogenesisModulationFactor)
	}

	// Caso 2: Max Dopamina, Sem Cortisol
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel) // 1.0
	env.CortisolLevel = 0.0
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	// LR: 1.0 * (1 + 0.5*1.0) = 1.5
	// Synapto: 1.0 * (1 + 0.2*1.0) = 1.2
	if !floatEquals(float64(env.LearningRateModulationFactor), 1.5, 1e-9) {
		t.Errorf("LR Factor C2: Expected 1.5, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 1.2, 1e-9) {
		t.Errorf("Synapto Factor C2: Expected 1.2, got %f", env.SynaptogenesisModulationFactor)
	}

	// Caso 3: Max Cortisol, Sem Dopamina
	env.DopamineLevel = 0.0
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel) // 1.0
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	// LR: 1.0 * (1 - 0.4*1.0) = 0.6
	// Synapto: 1.0 * (1 - 0.8*1.0) = 0.2
	if !floatEquals(float64(env.LearningRateModulationFactor), 0.6, 1e-9) {
		t.Errorf("LR Factor C3: Expected 0.6, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 0.2, 1e-9) {
		t.Errorf("Synapto Factor C3: Expected 0.2, got %f", env.SynaptogenesisModulationFactor)
	}

	// Caso 4: Max Dopamina, Max Cortisol
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	// LR: 1.0 * (1 + 0.5*1.0) * (1 - 0.4*1.0) = 1.5 * 0.6 = 0.9
	// Synapto: 1.0 * (1 + 0.2*1.0) * (1 - 0.8*1.0) = 1.2 * 0.2 = 0.24
	if !floatEquals(float64(env.LearningRateModulationFactor), 0.9, 1e-9) {
		t.Errorf("LR Factor C4: Expected 0.9, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 0.24, 1e-9) {
		t.Errorf("Synapto Factor C4: Expected 0.24, got %f", env.SynaptogenesisModulationFactor)
	}

	// Caso 5: LR Clampeado pelo MinLearningRateFactor
	simParams.MinLearningRateFactor = 0.7 // Aumentar o mínimo para forçar clamp
	env.DopamineLevel = 0.0
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel) // LR seria 0.6, mas clampeado para 0.7
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	if !floatEquals(float64(env.LearningRateModulationFactor), 0.7, 1e-9) {
		t.Errorf("LR Factor C5: Expected clamped 0.7, got %f", env.LearningRateModulationFactor)
	}
}

func TestApplyEffectsToNeurons(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.DopamineMaxLevel = 1.0
	simParams.CortisolMaxLevel = 1.0
	simParams.FiringThresholdIncreaseOnDopa = -0.2 // Reduz limiar em até 20% (fator 0.8)
	simParams.FiringThresholdIncreaseOnCort = 0.5  // Aumenta limiar em até 50% (fator 1.5)

	baseThresholdVal := 1.0
	simParams.BaseFiringThreshold = baseThresholdVal // Para neuron.New

	n1 := neuron.New(1, neuron.Excitatory, common.Point{}, &simParams)
	neurons := []*neuron.Neuron{n1}
	env := neurochemical.NewEnvironment()

	// Caso 1: Sem químicos
	env.DopamineLevel = 0.0
	env.CortisolLevel = 0.0
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), baseThresholdVal, 1e-9) {
		t.Errorf("Threshold C1: Expected %f, got %f", baseThresholdVal, n1.CurrentFiringThreshold)
	}

	// Caso 2: Max Dopamina, Sem Cortisol
	// Limiar = Base * (1 + FiringThresholdIncreaseOnDopa * NormDopa)
	// Limiar = 1.0 * (1 + (-0.2) * 1.0) = 1.0 * 0.8 = 0.8
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = 0.0
	n1.BaseFiringThreshold = common.Threshold(baseThresholdVal) // Reset base
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), 0.8, 1e-9) {
		t.Errorf("Threshold C2 (Max Dopa): Expected 0.8, got %f", n1.CurrentFiringThreshold)
	}

	// Caso 3: Max Cortisol, Sem Dopamina
	// Limiar = Base * (1 + FiringThresholdIncreaseOnCort * NormCort)
	// Limiar = 1.0 * (1 + 0.5 * 1.0) = 1.0 * 1.5 = 1.5
	env.DopamineLevel = 0.0
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	n1.BaseFiringThreshold = common.Threshold(baseThresholdVal) // Reset base
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), 1.5, 1e-9) {
		t.Errorf("Threshold C3 (Max Cort): Expected 1.5, got %f", n1.CurrentFiringThreshold)
	}

	// Caso 4: Max Dopamina, Max Cortisol
	// Efeito Cortisol primeiro: 1.0 * (1 + 0.5 * 1.0) = 1.5
	// Efeito Dopamina sobre isso: 1.5 * (1 + (-0.2) * 1.0) = 1.5 * 0.8 = 1.2
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	n1.BaseFiringThreshold = common.Threshold(baseThresholdVal) // Reset base
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), 1.2, 1e-9) {
		t.Errorf("Threshold C4 (Max Dopa & Cort): Expected 1.2, got %f", n1.CurrentFiringThreshold)
	}

	// Caso 5: Limiar mínimo (0.01)
	n1.BaseFiringThreshold = 0.005 // Base muito baixa
	simParams.FiringThresholdIncreaseOnDopa = -0.99 // Tentativa de reduzir muito
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = 0.0
	env.ApplyEffectsToNeurons(neurons, &simParams)
	// Base * (1 - 0.99) = 0.005 * 0.01 = 0.00005. Deve ser clampeado para 0.01
	if !floatEquals(float64(n1.CurrentFiringThreshold), 0.01, 1e-9) {
		t.Errorf("Threshold C5 (Min Clamp): Expected 0.01, got %f", n1.CurrentFiringThreshold)
	}
	// Reset FiringThresholdIncreaseOnDopa para não afetar outros testes se simParams for reutilizado (não é o caso aqui)
	simParams.FiringThresholdIncreaseOnDopa = -0.2
}
```
