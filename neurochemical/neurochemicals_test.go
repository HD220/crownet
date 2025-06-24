package neurochemical

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	// "crownet/pulse" // May not be directly needed for environment tests if events are mocked
	"math"
	"testing"
)

func defaultTestSimParamsForChem() *config.SimulationParameters {
	p := config.DefaultSimulationParameters()
	p.CortisolProductionRate = 0.01
	p.CortisolDecayRate = 0.005
	p.CortisolProductionPerHit = 0.05
	p.CortisolMaxLevel = 1.0
	p.DopamineProductionRate = 0.02 // Renamed from DopamineProductionPerEvent in some earlier thoughts
	p.DopamineDecayRate = 0.01
	p.DopamineProductionPerEvent = 0.1 // This is for specific event-based production
	p.DopamineMaxLevel = 1.0

	p.FiringThresholdIncreaseOnCort = 0.3  // Positive means cortisol increases threshold
	p.FiringThresholdIncreaseOnDopa = -0.2 // Negative means dopamine decreases threshold
	p.CortisolInfluenceOnLR = -0.5
	p.DopamineInfluenceOnLR = 0.8
	p.MinLearningRateFactor = 0.1 // Ensure LR factor doesn't go below this
	return &p
}

func TestNewNeurochemical(t *testing.T) {
	name := "TestChem"
	decayRate := common.Rate(0.1)
	maxLevel := common.Level(1.0)
	nc := NewNeurochemical(name, decayRate, maxLevel)

	if nc.Name != name { t.Errorf("Name got %s, want %s", nc.Name, name) }
	if nc.DecayRate != decayRate { t.Errorf("DecayRate got %f, want %f", nc.DecayRate, decayRate) }
	if nc.MaxLevel != maxLevel { t.Errorf("MaxLevel got %f, want %f", nc.MaxLevel, maxLevel) }
	if nc.CurrentLevel != 0.0 { t.Errorf("CurrentLevel got %f, want 0.0", nc.CurrentLevel) }
}

func TestNeurochemical_UpdateLevel(t *testing.T) {
	nc := NewNeurochemical("TestChem", 0.1, 1.0) // Decay 10%, Max 1.0

	t.Run("Production", func(t *testing.T) {
		nc.CurrentLevel = 0.0
		nc.UpdateLevel(0.5, nil) // Produce 0.5, no simParams needed for this direct update
		if nc.CurrentLevel != 0.5 {
			t.Errorf("Production: Level got %f, want 0.5", nc.CurrentLevel)
		}
	})

	t.Run("Production exceeding max level", func(t *testing.T) {
		nc.CurrentLevel = 0.8
		nc.UpdateLevel(0.5, nil) // Produce 0.5 (total 1.3, clamped to 1.0)
		if nc.CurrentLevel != 1.0 {
			t.Errorf("Production over max: Level got %f, want 1.0", nc.CurrentLevel)
		}
	})

	t.Run("Decay", func(t *testing.T) {
		nc.CurrentLevel = 0.5
		nc.UpdateLevel(0.0, nil) // Produce 0, only decay (0.5 * (1-0.1) = 0.45)
		if math.Abs(float64(nc.CurrentLevel)-0.45) > 1e-9 {
			t.Errorf("Decay: Level got %f, want 0.45", nc.CurrentLevel)
		}
	})

	t.Run("Production and Decay", func(t *testing.T) {
		nc.CurrentLevel = 0.5
		nc.UpdateLevel(0.2, nil) // Produce 0.2. Level becomes 0.5 + 0.2 = 0.7. Then decay: 0.7 * (1-0.1) = 0.63
		if math.Abs(float64(nc.CurrentLevel)-0.63) > 1e-9 {
			t.Errorf("Production and Decay: Level got %f, want 0.63", nc.CurrentLevel)
		}
	})

	t.Run("Decay to zero", func(t *testing.T) {
		nc.CurrentLevel = 0.001 // Very small level
		nc.DecayRate = 0.5      // High decay
		nc.UpdateLevel(0.0, nil)
		if nc.CurrentLevel != 0.0 { // Should clamp to 0 if it goes effectively negative or very small
			t.Errorf("Decay to zero: Level got %f, want 0.0", nc.CurrentLevel)
		}
	})
}

func TestNewEnvironment(t *testing.T) {
	simParams := defaultTestSimParamsForChem()
	env := NewEnvironment(simParams) // NewEnvironment now takes simParams

	if env.Cortisol == nil || env.Cortisol.Name != "Cortisol" {
		t.Error("Environment Cortisol not initialized correctly")
	}
	if env.Dopamine == nil || env.Dopamine.Name != "Dopamine" {
		t.Error("Environment Dopamine not initialized correctly")
	}
	if env.LearningRateModulationFactor != 1.0 {
		t.Errorf("Initial LearningRateModulationFactor got %f, want 1.0", env.LearningRateModulationFactor)
	}
}

// Mocking neurons and pulses for Environment.UpdateLevels is complex.
// We'll test parts of the logic, assuming sub-functions are correct.
// For a full test, more involved setup or mocks would be needed.
func TestEnvironment_UpdateLevels_Simplified(t *testing.T) {
	simParams := defaultTestSimParamsForChem()
	env := NewEnvironment(simParams)

	// Test Cortisol production (simplified: assume one "hit" occurs)
	// To properly test this, we'd need a mock neuron near CortisolGlandPosition and a mock pulse hitting it.
	// For now, we'll manually adjust the internal 'production' value that UpdateLevels uses.
	// This is not ideal as it tests internal calculation rather than the trigger.
	// A better approach would be to have a helper in Environment for tests, or more complex mocks.

	// For this unit test, let's focus on the effect of ApplyEffectsToNeurons and GetModulationFactor,
	// assuming CurrentLevel of chemicals can be set.
	env.Cortisol.CurrentLevel = 0.5
	env.Dopamine.CurrentLevel = 0.2

	// Test GetModulationFactor
	// simParams: CortisolInfluenceOnLR = -0.5, DopamineInfluenceOnLR = 0.8, MinLearningRateFactor = 0.1
	// Expected: (1.0 + C_level * C_inf) * (1.0 + D_level * D_inf)
	// (1.0 + 0.5 * -0.5) = (1.0 - 0.25) = 0.75
	// (1.0 + 0.2 * 0.8)  = (1.0 + 0.16) = 1.16
	// Combined = 0.75 * 1.16 = 0.87. Clamped by MinLearningRateFactor (0.1). Should be 0.87.
	lrMod := env.GetModulationFactor(simParams.CortisolInfluenceOnLR, simParams.DopamineInfluenceOnLR, simParams.MinLearningRateFactor)
	if math.Abs(float64(lrMod)-0.87) > 1e-9 {
		t.Errorf("GetModulationFactor (LR): got %f, want 0.87", lrMod)
	}

	// Test with MinLearningRateFactor clamping
	simParams.CortisolInfluenceOnLR = -2.0 // Strong negative cortisol effect
	// (1.0 + 0.5 * -2.0) = (1.0 - 1.0) = 0.0
	// (1.0 + 0.2 * 0.8)  = 1.16
	// Combined = 0.0 * 1.16 = 0.0. Should be clamped to MinLearningRateFactor (0.1)
	lrModClamped := env.GetModulationFactor(simParams.CortisolInfluenceOnLR, simParams.DopamineInfluenceOnLR, simParams.MinLearningRateFactor)
	if math.Abs(float64(lrModClamped)-simParams.MinLearningRateFactor) > 1e-9 {
		t.Errorf("GetModulationFactor (LR) clamped: got %f, want %f", lrModClamped, simParams.MinLearningRateFactor)
	}
	simParams.CortisolInfluenceOnLR = -0.5 // Reset for next test

}

func TestEnvironment_ApplyEffectsToNeurons(t *testing.T) {
	simParams := defaultTestSimParamsForChem()
	env := NewEnvironment(simParams)

	n1 := neuron.New(0, neuron.Excitatory, common.Point{}, simParams) // BaseThreshold = 1.0

	neurons := []*neuron.Neuron{n1}

	env.Cortisol.CurrentLevel = 0.0
	env.Dopamine.CurrentLevel = 0.0
	env.ApplyEffectsToNeurons(neurons, simParams)
	// Expected: BaseThreshold * (1 + 0*0.3) * (1 + 0*-0.2) = BaseThreshold * 1 * 1 = 1.0
	if math.Abs(float64(n1.CurrentFiringThreshold)-simParams.BaseFiringThreshold) > 1e-9 {
		t.Errorf("ApplyEffects (no chem): Threshold got %f, want %f", n1.CurrentFiringThreshold, simParams.BaseFiringThreshold)
	}

	env.Cortisol.CurrentLevel = 0.5 // Cortisol = 0.5
	env.Dopamine.CurrentLevel = 0.0
	env.ApplyEffectsToNeurons(neurons, simParams)
	// Expected: Base * (1 + 0.5 * FiringThresholdIncreaseOnCort) = 1.0 * (1 + 0.5 * 0.3) = 1.0 * (1 + 0.15) = 1.15
	expectedThreshCort := simParams.BaseFiringThreshold * (1.0 + common.Threshold(env.Cortisol.CurrentLevel)*common.Threshold(simParams.FiringThresholdIncreaseOnCort))
	if math.Abs(float64(n1.CurrentFiringThreshold)-float64(expectedThreshCort)) > 1e-9 {
		t.Errorf("ApplyEffects (cortisol only): Threshold got %f, want %f", n1.CurrentFiringThreshold, expectedThreshCort)
	}

	env.Cortisol.CurrentLevel = 0.0
	env.Dopamine.CurrentLevel = 0.8 // Dopamine = 0.8
	env.ApplyEffectsToNeurons(neurons, simParams)
	// Expected: Base * (1 + 0.8 * FiringThresholdIncreaseOnDopa) = 1.0 * (1 + 0.8 * -0.2) = 1.0 * (1 - 0.16) = 0.84
	expectedThreshDopa := simParams.BaseFiringThreshold * (1.0 + common.Threshold(env.Dopamine.CurrentLevel)*common.Threshold(simParams.FiringThresholdIncreaseOnDopa))
	if math.Abs(float64(n1.CurrentFiringThreshold)-float64(expectedThreshDopa)) > 1e-9 {
		t.Errorf("ApplyEffects (dopamine only): Threshold got %f, want %f", n1.CurrentFiringThreshold, expectedThreshDopa)
	}

	env.Cortisol.CurrentLevel = 0.5
	env.Dopamine.CurrentLevel = 0.8
	env.ApplyEffectsToNeurons(neurons, simParams)
	// Sequential application:
	// After Cortisol: Base * (1 + 0.5 * 0.3) = 1.15
	// After Dopamine: 1.15 * (1 + 0.8 * -0.2) = 1.15 * (1 - 0.16) = 1.15 * 0.84 = 0.966
	intermediateThresh := simParams.BaseFiringThreshold * (1.0 + common.Threshold(env.Cortisol.CurrentLevel)*common.Threshold(simParams.FiringThresholdIncreaseOnCort))
	finalExpectedThresh := intermediateThresh * (1.0 + common.Threshold(env.Dopamine.CurrentLevel)*common.Threshold(simParams.FiringThresholdIncreaseOnDopa))
	if math.Abs(float64(n1.CurrentFiringThreshold)-float64(finalExpectedThresh)) > 1e-9 {
		t.Errorf("ApplyEffects (both chems): Threshold got %f, want %f", n1.CurrentFiringThreshold, finalExpectedThresh)
	}

	// Test threshold doesn't go below a very small positive value (e.g. 0.01)
	// Let FiringThresholdIncreaseOnDopa be very negative
	simParams.FiringThresholdIncreaseOnDopa = -5.0 // Strong decrease
	env.Cortisol.CurrentLevel = 0.0
	env.Dopamine.CurrentLevel = 1.0 // Max Dopa
	env.ApplyEffectsToNeurons(neurons, simParams)
	// Expected: 1.0 * (1 + 1.0 * -5.0) = 1.0 * -4.0 = -4.0. Should be clamped.
	// The ApplyEffectsToNeurons clamps to MinimumFiringThreshold (0.01).
	if math.Abs(float64(n1.CurrentFiringThreshold)-MinimumFiringThreshold) > 1e-9 {
		 t.Errorf("ApplyEffects (threshold clamping): Threshold got %f, want %f", n1.CurrentFiringThreshold, MinimumFiringThreshold)
	}
}

```
