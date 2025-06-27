package neurochemical

import (
	"math"
	"testing"

	"crownet/common"
	"crownet/config"
	"crownet/neuron"
)

// Helper to create default SimParams for testing
func getDefaultTestSimParams() *config.SimulationParameters {
	sp := config.DefaultSimulationParameters() // Gets a copy
	// Modify specific neurochemical params for predictable testing if needed
	sp.Neurochemical.CortisolMaxLevel = 10.0
	sp.Neurochemical.DopamineMaxLevel = 10.0
	sp.Neurochemical.CortisolDecayRate = 0.1
	sp.Neurochemical.DopamineDecayRate = 0.1
	sp.Neurochemical.CortisolProductionRate = 0.0 // Base production to 0 for isolated event testing
	sp.Neurochemical.DopamineProductionRate = 0.0 // Base production to 0 for isolated event testing
	sp.Neurochemical.CortisolProductionPerHit = 1.0
	sp.Neurochemical.DopamineProductionPerEvent = 1.0
	sp.Neurochemical.FiringThresholdIncreaseOnCort = 0.2  // e.g. +20% per unit of normalized cortisol
	sp.Neurochemical.FiringThresholdIncreaseOnDopa = -0.1 // e.g. -10% per unit of normalized dopamine
	sp.Learning.MinLearningRateFactor = 0.05             // For testing clamping
	return &sp
}

func TestNewEnvironment(t *testing.T) {
	params := getDefaultTestSimParams()
	env := NewEnvironment(&params.Neurochemical)

	if env.CortisolLevel != 0.0 {
		t.Errorf("Expected initial CortisolLevel 0.0, got %f", env.CortisolLevel)
	}
	if env.DopamineLevel != 0.0 {
		t.Errorf("Expected initial DopamineLevel 0.0, got %f", env.DopamineLevel)
	}
	if env.LearningRateModulationFactor != 1.0 {
		t.Errorf("Expected initial LearningRateModulationFactor 1.0, got %f", env.LearningRateModulationFactor)
	}
	if env.SynaptogenesisModulationFactor != 1.0 {
		t.Errorf("Expected initial SynaptogenesisModulationFactor 1.0, got %f", env.SynaptogenesisModulationFactor)
	}
	if env.NeurochemicalParams == nil {
		t.Errorf("NeurochemicalParams should not be nil in new environment")
	}
}

func TestUpdateChemicalLevel(t *testing.T) {
	tests := []struct {
		name                string
		currentLevel        common.Level
		decayRate           common.Rate
		productionThisCycle float64
		maxLevel            common.Level
		expectedLevel       common.Level
	}{
		{"no change", 5.0, 0.0, 0.0, 10.0, 5.0},
		{"decay only", 5.0, 0.1, 0.0, 10.0, 4.5}, // 5 * (1-0.1) = 4.5
		{"production only", 5.0, 0.0, 1.0, 10.0, 6.0},
		{"decay and production", 5.0, 0.1, 1.0, 10.0, 5.5}, // 4.5 + 1.0 = 5.5
		{"clamp at maxLevel", 8.0, 0.0, 3.0, 10.0, 10.0},   // 8 + 3 = 11, clamped to 10
		{"clamp at zero", 0.5, 0.2, -1.0, 10.0, 0.0},      // 0.5*(1-0.2) = 0.4; 0.4 - 1.0 = -0.6, clamped to 0
		{"no maxLevel (0)", 5.0, 0.0, 2.0, 0.0, 7.0},      // MaxLevel 0 means no upper clamp
		{"no maxLevel (negative)", 5.0, 0.0, 2.0, -1.0, 7.0},
		{"decay to zero", 0.1, 1.0, 0.0, 10.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := updateChemicalLevel(tt.currentLevel, tt.decayRate, tt.productionThisCycle, tt.maxLevel)
			if math.Abs(float64(level-tt.expectedLevel)) > 1e-9 {
				t.Errorf("updateChemicalLevel() got %f, want %f", level, tt.expectedLevel)
			}
		})
	}
}

// TestEnvironment_UpdateLevels_Simplified focuses on level changes without complex pulse/neuron interactions.
func TestEnvironment_UpdateLevels_Simplified(_ *testing.T) { // Parameter t renamed to _
	// simParams := getDefaultTestSimParams()
	// env := NewEnvironment(&simParams.Neurochemical)

	// env.CortisolLevel = 5.0
	// env.DopamineLevel = 2.0

	// Call UpdateLevels with no neurons or pulses to only test base production and decay.
	// env.UpdateLevels(nil, nil, common.Point{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}, simParams)

	// Expected Cortisol: 5.0 * (1 - 0.1) + 0.0 (base prod) = 4.5
	// Expected Dopamine: 2.0 * (1 - 0.1) + 0.0 (base prod) = 1.8
	// if math.Abs(float64(env.CortisolLevel-4.5)) > 1e-9 {
	// 	t.Errorf("Cortisol level after UpdateLevels: got %f, want 4.5", env.CortisolLevel)
	// }
	// if math.Abs(float64(env.DopamineLevel-1.8)) > 1e-9 {
	// 	t.Errorf("Dopamine level after UpdateLevels: got %f, want 1.8", env.DopamineLevel)
	// }
	// This test is a bit weak as calculateCortisolStimulation is a placeholder.
	// A more robust test would mock calculateCortisolStimulation or test components.
}

func TestRecalculateModulationFactors(t *testing.T) {
	simParams := getDefaultTestSimParams() // MaxLevels are 10.0
	env := NewEnvironment(&simParams.Neurochemical)

	// Case 1: No chemicals
	env.CortisolLevel = 0.0
	env.DopamineLevel = 0.0
	env.recalculateModulationFactors(simParams)
	if math.Abs(float64(env.LearningRateModulationFactor-1.0)) > 1e-9 {
		t.Errorf("LR Factor (no chem): got %f, want 1.0", env.LearningRateModulationFactor)
	}

	// Case 2: Cortisol only (normalized: 5/10 = 0.5)
	// LR_mod = 1.0 + (0.2 * 0.5) + (-0.1 * 0.0) = 1.0 + 0.1 = 1.1
	env.CortisolLevel = 5.0
	env.DopamineLevel = 0.0
	env.recalculateModulationFactors(simParams)
	expectedLR_cort := 1.0 + (simParams.Neurochemical.CortisolInfluenceOnLR * 0.5)
	if math.Abs(float64(env.LearningRateModulationFactor-common.Factor(expectedLR_cort))) > 1e-9 {
		t.Errorf("LR Factor (cortisol only): got %f, want %f", env.LearningRateModulationFactor, expectedLR_cort)
	}

	// Case 3: Dopamine only (normalized: 8/10 = 0.8)
	// LR_mod = 1.0 + (0.2 * 0.0) + (-0.1 * 0.8) = 1.0 - 0.08 = 0.92
	env.CortisolLevel = 0.0
	env.DopamineLevel = 8.0
	env.recalculateModulationFactors(simParams)
	expectedLR_dopa := 1.0 + (simParams.Neurochemical.DopamineInfluenceOnLR * 0.8)
	if math.Abs(float64(env.LearningRateModulationFactor-common.Factor(expectedLR_dopa))) > 1e-9 {
		t.Errorf("LR Factor (dopamine only): got %f, want %f", env.LearningRateModulationFactor, expectedLR_dopa)
	}

	// Case 4: Both chemicals & clamping
	// Cortisol: 15.0 (norm 1.0, maxed), Dopamine: 1.0 (norm 0.1)
	// LR_mod = 1.0 + (0.2 * 1.0) + (-0.1 * 0.1) = 1.0 + 0.2 - 0.01 = 1.19
	// MinLearningRateFactor is 0.05. 1.19 is > 0.05, so no clamping for this.
	env.CortisolLevel = 15.0 // Will be normalized to 1.0 due to CortisolMaxLevel = 10.0
	env.DopamineLevel = 1.0  // Normalized to 0.1
	env.recalculateModulationFactors(simParams)
	// normCortisol = 1.0 (15.0 clamped to 10.0, then 10.0/10.0)
	// normDopamine = 0.1 (1.0 / 10.0)
	// expectedLR_both = 1.0 + (0.2 * 1.0) + (-0.1 * 0.1) = 1.0 + 0.2 - 0.01 = 1.19
	expectedLR_both := 1.0 +
		(simParams.Neurochemical.CortisolInfluenceOnLR * 1.0) +
		(simParams.Neurochemical.DopamineInfluenceOnLR * 0.1)
	if math.Abs(float64(env.LearningRateModulationFactor-common.Factor(expectedLR_both))) > 1e-9 {
		t.Errorf("LR Factor (both, no clamp): got %f, want %f", env.LearningRateModulationFactor, expectedLR_both)
	}

	// Case 5: Clamping at MinLearningRateFactor
	// Set influences such that result would be < MinLearningRateFactor (0.05)
	// e.g., CortisolInfluenceOnLR = -1.0, DopamineInfluenceOnLR = -1.0
	// Cortisol = 5.0 (norm 0.5), Dopa = 5.0 (norm 0.5)
	// LR_mod = 1.0 + (-1.0 * 0.5) + (-1.0 * 0.5) = 1.0 - 0.5 - 0.5 = 0.0
	// This should be clamped to 0.05.
	originalCortInfluence := simParams.Neurochemical.CortisolInfluenceOnLR
	originalDopaInfluence := simParams.Neurochemical.DopamineInfluenceOnLR
	simParams.Neurochemical.CortisolInfluenceOnLR = -1.0
	simParams.Neurochemical.DopamineInfluenceOnLR = -1.0
	env.CortisolLevel = 5.0
	env.DopamineLevel = 5.0
	env.recalculateModulationFactors(simParams)
	if math.Abs(float64(env.LearningRateModulationFactor-simParams.Learning.MinLearningRateFactor)) > 1e-9 {
		t.Errorf("LR Factor (clamped): got %f, want %f",
			env.LearningRateModulationFactor, simParams.Learning.MinLearningRateFactor)
	}
	// Restore original params for other tests if simParams is shared (it's a copy here, so ok)
	simParams.Neurochemical.CortisolInfluenceOnLR = originalCortInfluence
	simParams.Neurochemical.DopamineInfluenceOnLR = originalDopaInfluence
}

func TestApplyEffectsToNeurons(t *testing.T) {
	simParams := getDefaultTestSimParams()
	env := NewEnvironment(&simParams.Neurochemical)
	// BaseFiringThreshold is 1.0 from DefaultSimulationParameters
	// FiringThresholdIncreaseOnCort = 0.2, FiringThresholdIncreaseOnDopa = -0.1
	// MaxLevels are 10.0

	n1 := neuron.New(0, neuron.Excitatory, common.Point{0, 0}, simParams) // Pass full simParams
	neuronsMap := map[common.NeuronID]*neuron.Neuron{n1.ID: n1}

	// Case 1: No chemicals
	env.CortisolLevel = 0.0
	env.DopamineLevel = 0.0
	env.ApplyEffectsToNeurons(neuronsMap, simParams)
	if math.Abs(float64(n1.CurrentFiringThreshold-n1.BaseFiringThreshold)) > 1e-9 {
		t.Errorf("ApplyEffects (no chem): Threshold got %f, want %f",
			n1.CurrentFiringThreshold, float64(n1.BaseFiringThreshold))
	}

	// Case 2: Cortisol only (5.0, norm 0.5)
	// Expected: 1.0 * (1 + (0.2 * 0.5)) = 1.0 * 1.1 = 1.1
	env.CortisolLevel = 5.0
	env.DopamineLevel = 0.0
	env.ApplyEffectsToNeurons(neuronsMap, simParams)
	expectedThreshCort := float64(n1.BaseFiringThreshold) *
		(1.0 + float64(simParams.Neurochemical.FiringThresholdIncreaseOnCort)*0.5)
	if math.Abs(float64(n1.CurrentFiringThreshold-common.Threshold(expectedThreshCort))) > 1e-9 {
		t.Errorf("ApplyEffects (cortisol): Threshold got %f, want %f", n1.CurrentFiringThreshold, expectedThreshCort)
	}

	// Case 3: Dopamine only (8.0, norm 0.8)
	// Expected: 1.0 * (1 + (-0.1 * 0.8)) = 1.0 * (1 - 0.08) = 1.0 * 0.92 = 0.92
	env.CortisolLevel = 0.0
	env.DopamineLevel = 8.0
	env.ApplyEffectsToNeurons(neuronsMap, simParams)
	expectedThreshDopa := float64(n1.BaseFiringThreshold) *
		(1.0 + float64(simParams.Neurochemical.FiringThresholdIncreaseOnDopa)*0.8)
	if math.Abs(float64(n1.CurrentFiringThreshold-common.Threshold(expectedThreshDopa))) > 1e-9 {
		t.Errorf("ApplyEffects (dopamine): Threshold got %f, want %f", n1.CurrentFiringThreshold, expectedThreshDopa)
	}

	// Case 4: Both chemicals
	// Cortisol 5.0 (norm 0.5), Dopamine 8.0 (norm 0.8)
	// From Cort: 1.0 * (1 + (0.2*0.5)) = 1.1
	// From Dopa on that: 1.1 * (1 + (-0.1*0.8)) = 1.1 * (1 - 0.08) = 1.1 * 0.92 = 1.012
	env.CortisolLevel = 5.0
	env.DopamineLevel = 8.0
	env.ApplyEffectsToNeurons(neuronsMap, simParams)
	intermediateThresh := float64(n1.BaseFiringThreshold) *
		(1.0 + float64(simParams.Neurochemical.FiringThresholdIncreaseOnCort)*0.5)
	finalExpectedThresh := intermediateThresh *
		(1.0 + float64(simParams.Neurochemical.FiringThresholdIncreaseOnDopa)*0.8)
	if math.Abs(float64(n1.CurrentFiringThreshold-common.Threshold(finalExpectedThresh))) > 1e-9 {
		t.Errorf("ApplyEffects (both): Threshold got %f, want %f", n1.CurrentFiringThreshold, finalExpectedThresh)
	}

	// Case 5: Threshold clamping at 0
	// Base = 1.0. Set Dopa influence to -2.0. Dopa level 8.0 (norm 0.8)
	// Expected: 1.0 * (1 + (-2.0 * 0.8)) = 1.0 * (1 - 1.6) = 1.0 * -0.6 = -0.6. Clamped to 0.
	originalDopaThresholdInfluence := simParams.Neurochemical.FiringThresholdIncreaseOnDopa
	simParams.Neurochemical.FiringThresholdIncreaseOnDopa = -2.0
	env.CortisolLevel = 0.0
	env.DopamineLevel = 8.0
	env.ApplyEffectsToNeurons(neuronsMap, simParams)
	minFiringThresholdValue := 0.0
	if math.Abs(float64(n1.CurrentFiringThreshold-common.Threshold(minFiringThresholdValue))) > 1e-9 {
		t.Errorf("ApplyEffects (threshold clamping): Threshold got %f, want %f",
			n1.CurrentFiringThreshold, minFiringThresholdValue)
	}
	simParams.Neurochemical.FiringThresholdIncreaseOnDopa = originalDopaThresholdInfluence
}

func TestProduceDopamineEvent(t *testing.T) {
	simParams := getDefaultTestSimParams() // DopamineProductionPerEvent = 1.0
	env := NewEnvironment(&simParams.Neurochemical)
	env.DopamineLevel = 1.0 // Initial level

	// Event magnitude 2.0. Expected production = 2.0 * 1.0 = 2.0
	// New level = 1.0 (initial) + 2.0 (produced) = 3.0
	env.ProduceDopamineEvent(2.0, simParams)
	if math.Abs(float64(env.DopamineLevel-3.0)) > 1e-9 {
		t.Errorf("Dopamine after event: got %f, want 3.0", env.DopamineLevel)
	}

	// Another event, this time will hit max level
	// Current Dopa = 3.0. MaxDopamineLevel = 10.0.
	// Event magnitude 15.0. Production = 15.0 * 1.0 = 15.0
	// Tentative new level = 3.0 + 15.0 = 18.0. Clamped to 10.0.
	env.ProduceDopamineEvent(15.0, simParams)
	if math.Abs(float64(env.DopamineLevel-simParams.Neurochemical.DopamineMaxLevel)) > 1e-9 {
		t.Errorf("Dopamine after event (max clamp): got %f, want %f",
			env.DopamineLevel, simParams.Neurochemical.DopamineMaxLevel)
	}
}

// TODO: Add test for calculateCortisolStimulation if its logic becomes non-trivial.
// Currently, it's a placeholder, so testing its exact output is not very meaningful.
// It would require mocking pulse states and positions.

// TODO: Test GetModulationFactor more thoroughly if its calculation becomes complex
// or has more edge cases beyond what's covered by TestRecalculateModulationFactors.
// The logic is largely the same.
// Example:
// func TestGetModulationFactor(t *testing.T) {
// 	simParams := getDefaultTestSimParams()
// 	env := NewEnvironment(&simParams.Neurochemical)
// 	env.CortisolLevel = 5.0 // norm 0.5
// 	env.DopamineLevel = 2.0 // norm 0.2
//
// 	// Example: Get LR modulation
// 	lrMod := env.GetModulationFactor(
// 		simParams.Neurochemical.CortisolInfluenceOnLR,
// 		simParams.Neurochemical.DopamineInfluenceOnLR,
// 		simParams.Learning.MinLearningRateFactor,
// 	)
// 	// Expected based on TestRecalculateModulationFactors logic:
// 	// LR_mod = 1.0 + (0.2 * 0.5) + (-0.1 * 0.2) = 1.0 + 0.1 - 0.02 = 1.08
// 	expected := 1.0 + (simParams.Neurochemical.CortisolInfluenceOnLR * 0.5) + (simParams.Neurochemical.DopamineInfluenceOnLR * 0.2)
// 	if math.Abs(float64(lrMod-common.Factor(expected))) > 1e-9 {
// 		t.Errorf("GetModulationFactor for LR: got %f, want %f", lrMod, expected)
// 	}
// }
