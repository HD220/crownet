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
	p1 := pulse.New(1, common.Point{0,0,0}, 1.0, 0, 10.0)
	p1.CurrentDistance = 0

	activePulses = []*pulse.Pulse{p1}
	env.CortisolLevel = 0 // Reset
	env.UpdateLevels(neurons, activePulses, cortisolGlandPos, &simParams)
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
	if !floatEquals(float64(env.CortisolLevel), 0.0, 1e-9) {
		t.Errorf("Cortisol C3: Expected 0.0 with inhibitory pulse, got %f", env.CortisolLevel)
	}

	// --- Caso 4: Produção leva ao clamping no MaxLevel ---
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel - 0.05)
	p1_reset := pulse.New(1, common.Point{0,0,0}, 1.0, 0, 10.0)
	p1_reset.CurrentDistance = 0
	activePulses = []*pulse.Pulse{p1_reset}

	env.UpdateLevels(neurons, activePulses, cortisolGlandPos, &simParams)
	expectedCortisol4 := simParams.CortisolMaxLevel * (1.0 - simParams.CortisolDecayRate)
	if !floatEquals(float64(env.CortisolLevel), expectedCortisol4, 1e-9) {
		t.Errorf("Cortisol C4: Expected clamped level %f, got %f", expectedCortisol4, env.CortisolLevel)
	}

	// --- Caso 5: Pulso não atinge a glândula (longe demais) ---
	pFar := pulse.New(3, common.Point{5,5,5}, 1.0, 0, 10.0)
	pFar.CurrentDistance = 0
	activePulses = []*pulse.Pulse{pFar}
	env.CortisolLevel = 0.5
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
	activePulses := []*pulse.Pulse{}

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
	n1Dopa.CurrentState = neuron.Firing
	neurons2 := []*neuron.Neuron{n1Dopa, n2Excit}
	env.DopamineLevel = 0 // Reset
	env.UpdateLevels(neurons2, activePulses, common.Point{}, &simParams)
	expectedDopamine2 := 0.2 * (1.0 - simParams.DopamineDecayRate)
	if !floatEquals(float64(env.DopamineLevel), expectedDopamine2, 1e-9) {
		t.Errorf("Dopamine C2: Expected %f, got %f", expectedDopamine2, env.DopamineLevel)
	}
	n1Dopa.CurrentState = neuron.Resting

	// --- Caso 3: Produção leva ao clamping no MaxLevel ---
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel - 0.1)
	n1Dopa.CurrentState = neuron.Firing
	neurons3 := []*neuron.Neuron{n1Dopa}
	env.UpdateLevels(neurons3, activePulses, common.Point{}, &simParams)
	expectedDopamine3 := simParams.DopamineMaxLevel * (1.0 - simParams.DopamineDecayRate)
	if !floatEquals(float64(env.DopamineLevel), expectedDopamine3, 1e-9) {
		t.Errorf("Dopamine C3: Expected clamped level %f, got %f", expectedDopamine3, env.DopamineLevel)
	}
	n1Dopa.CurrentState = neuron.Resting
}

func TestRecalculateModulationFactors(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.DopamineMaxLevel = 1.0
	simParams.CortisolMaxLevel = 1.0
	simParams.DopamineInfluenceOnLR = 0.5
	simParams.CortisolInfluenceOnLR = -0.4
	simParams.MinLearningRateFactor = 0.1
	simParams.DopamineInfluenceOnSynapto = 0.2
	simParams.CortisolInfluenceOnSynapto = -0.8

	env := neurochemical.NewEnvironment()

	// Caso 1: Sem químicos, fatores devem ser 1.0
	env.CortisolLevel = 0.0
	env.DopamineLevel = 0.0
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	if !floatEquals(float64(env.LearningRateModulationFactor), 1.0, 1e-9) {
		t.Errorf("LR Factor C1: Expected 1.0, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 1.0, 1e-9) {
		t.Errorf("Synapto Factor C1: Expected 1.0, got %f", env.SynaptogenesisModulationFactor)
	}

	// Caso 2: Max Dopamina, Sem Cortisol
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = 0.0
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	if !floatEquals(float64(env.LearningRateModulationFactor), 1.5, 1e-9) {
		t.Errorf("LR Factor C2: Expected 1.5, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 1.2, 1e-9) {
		t.Errorf("Synapto Factor C2: Expected 1.2, got %f", env.SynaptogenesisModulationFactor)
	}

	// Caso 3: Max Cortisol, Sem Dopamina
	env.DopamineLevel = 0.0
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
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
	if !floatEquals(float64(env.LearningRateModulationFactor), 0.9, 1e-9) {
		t.Errorf("LR Factor C4: Expected 0.9, got %f", env.LearningRateModulationFactor)
	}
	if !floatEquals(float64(env.SynaptogenesisModulationFactor), 0.24, 1e-9) {
		t.Errorf("Synapto Factor C4: Expected 0.24, got %f", env.SynaptogenesisModulationFactor)
	}

	// Caso 5: LR Clampeado pelo MinLearningRateFactor
	simParams.MinLearningRateFactor = 0.7
	env.DopamineLevel = 0.0
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	env.UpdateLevels([]*neuron.Neuron{}, []*pulse.Pulse{}, common.Point{}, &simParams)
	if !floatEquals(float64(env.LearningRateModulationFactor), 0.7, 1e-9) {
		t.Errorf("LR Factor C5: Expected clamped 0.7, got %f", env.LearningRateModulationFactor)
	}
}

func TestApplyEffectsToNeurons(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.DopamineMaxLevel = 1.0
	simParams.CortisolMaxLevel = 1.0
	simParams.FiringThresholdIncreaseOnDopa = -0.2
	simParams.FiringThresholdIncreaseOnCort = 0.5

	baseThresholdVal := 1.0
	simParams.BaseFiringThreshold = baseThresholdVal

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
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = 0.0
	n1.BaseFiringThreshold = common.Threshold(baseThresholdVal)
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), 0.8, 1e-9) {
		t.Errorf("Threshold C2 (Max Dopa): Expected 0.8, got %f", n1.CurrentFiringThreshold)
	}

	// Caso 3: Max Cortisol, Sem Dopamina
	env.DopamineLevel = 0.0
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	n1.BaseFiringThreshold = common.Threshold(baseThresholdVal)
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), 1.5, 1e-9) {
		t.Errorf("Threshold C3 (Max Cort): Expected 1.5, got %f", n1.CurrentFiringThreshold)
	}

	// Caso 4: Max Dopamina, Max Cortisol
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = common.Level(simParams.CortisolMaxLevel)
	n1.BaseFiringThreshold = common.Threshold(baseThresholdVal)
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), 1.2, 1e-9) {
		t.Errorf("Threshold C4 (Max Dopa & Cort): Expected 1.2, got %f", n1.CurrentFiringThreshold)
	}

	// Caso 5: Limiar mínimo (0.01)
	n1.BaseFiringThreshold = 0.005
	simParams.FiringThresholdIncreaseOnDopa = -0.99
	env.DopamineLevel = common.Level(simParams.DopamineMaxLevel)
	env.CortisolLevel = 0.0
	env.ApplyEffectsToNeurons(neurons, &simParams)
	if !floatEquals(float64(n1.CurrentFiringThreshold), 0.01, 1e-9) {
		t.Errorf("Threshold C5 (Min Clamp): Expected 0.01, got %f", n1.CurrentFiringThreshold)
	}
	simParams.FiringThresholdIncreaseOnDopa = -0.2
}
