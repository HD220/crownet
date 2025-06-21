package network

import (
	"crownet/neuron"
	"fmt" // Re-add for debug printing
	"math"
)

// updateDopamineLevel adjusts the network's dopamine level.
func (cn *CrowNet) updateDopamineLevel() {
	// Production: Dopaminergic neurons that are in FiringState this cycle
	// (meaning they just fired due to received pulses in the *previous* part of this cycle)
	dopamineProducedThisCycle := 0.0
	for _, n := range cn.Neurons {
		if n.Type == neuron.DopaminergicNeuron && n.State == neuron.FiringState {
			// A dopaminergic neuron "fired". Its primary role here is to produce dopamine.
			dopamineProducedThisCycle += DopamineProductionPerEvent
		}
	}

	if dopamineProducedThisCycle > 0 {
		cn.DopamineLevel += dopamineProducedThisCycle
		if DebugCortisolHit { // Using same debug flag for now for any chemical debug prints
			fmt.Printf("[DEBUG] Cycle %d: Dopamine produced: %.3f by firing dopaminergic neurons. New Level (before decay): %.3f\n", cn.CycleCount, dopamineProducedThisCycle, cn.DopamineLevel)
		}
	}

	// Decay (faster than cortisol)
	cn.DopamineLevel -= cn.DopamineLevel * DopamineDecayRate
	if cn.DopamineLevel < 0 {
		cn.DopamineLevel = 0
	}
	if cn.DopamineLevel > DopamineMaxLevel {
		cn.DopamineLevel = DopamineMaxLevel
	}
	// fmt.Printf("Cycle %d: Dopamine level: %.3f\n", cn.CycleCount, cn.DopamineLevel)
}

// applyDopamineEffects updates neuron firing thresholds and synaptogenesis based on dopamine levels.
// This will modify values potentially already changed by cortisol.
func (cn *CrowNet) applyDopamineEffects() {
	// Calculate Dopamine's effect on Synaptogenesis
	// Synaptogenesis increases with dopamine. Factor is 1.0 at DopamineLevel 0,
	// and DopamineSynaptogenesisIncreaseFactor at DopamineMaxLevel. Linear interpolation.
	dopamineSynFactor := 1.0
	if DopamineMaxLevel > 0 {
		if cn.DopamineLevel >= DopamineMaxLevel {
			dopamineSynFactor = DopamineSynaptogenesisIncreaseFactor
		} else if cn.DopamineLevel > 0 {
			t := cn.DopamineLevel / DopamineMaxLevel
			dopamineSynFactor = 1.0 + t*(DopamineSynaptogenesisIncreaseFactor-1.0) // Lerp
		}
	} // else, if DopamineMaxLevel is 0, factor remains 1.0
	dopamineSynFactor = math.Max(1.0, math.Min(dopamineSynFactor, DopamineSynaptogenesisIncreaseFactor)) // Clamp, ensure it's at least 1.0

	// Combine with cortisol's effect on synaptogenesis factor
	// cn.currentSynaptogenesisModulationFactor was set by cortisol. Now multiply by dopamine's effect.
	cn.currentSynaptogenesisModulationFactor *= dopamineSynFactor
	// fmt.Printf("Cycle %d: Dopamine %.2f. Synaptogenesis factor (after dopamine): %.2f\n", cn.CycleCount, cn.DopamineLevel, cn.currentSynaptogenesisModulationFactor)

	// Update firing thresholds for all neurons
	// Dopamine increases firing threshold. Factor is 1.0 at DopamineLevel 0,
	// and DopamineThresholdIncreaseFactor at DopamineMaxLevel. Linear interpolation.
	dopamineThresholdFactor := 1.0
	if DopamineMaxLevel > 0 {
		if cn.DopamineLevel >= DopamineMaxLevel {
			dopamineThresholdFactor = DopamineThresholdIncreaseFactor
		} else if cn.DopamineLevel > 0 {
			t := cn.DopamineLevel / DopamineMaxLevel
			dopamineThresholdFactor = 1.0 + t*(DopamineThresholdIncreaseFactor-1.0) // Lerp
		}
	} // else, if DopamineMaxLevel is 0, factor remains 1.0
	dopamineThresholdFactor = math.Max(1.0, math.Min(dopamineThresholdFactor, DopamineThresholdIncreaseFactor)) // Clamp, ensure it's at least 1.0

	for _, n := range cn.Neurons {
		// CurrentFiringThreshold was already set by cortisol. Now apply dopamine's multiplicative effect.
		n.CurrentFiringThreshold *= dopamineThresholdFactor

		if n.CurrentFiringThreshold < 0.01 { // Ensure threshold doesn't become zero or negative
			n.CurrentFiringThreshold = 0.01
		}
		// if n.ID < 2 { // Debug print for a few neurons
		// 	fmt.Printf("  Neuron %d: Dopamine %.2f, BaseThr %.2f -> PrevThr %.2f -> CurrThr %.2f (DopaFactor %.2f)\n", n.ID, cn.DopamineLevel, n.BaseFiringThreshold, n.CurrentFiringThreshold / dopamineThresholdFactor, n.CurrentFiringThreshold, dopamineThresholdFactor)
		// }
	}
}

// Call updateDopamineLevel and applyDopamineEffects in RunCycle (network.go)
// Ensure initializeChemicalModulation in network.go also sets DopamineLevel to 0.
// In network.go NewCrowNet:
// cn.DopamineLevel = 0.0
// In network.go RunCycle, ensure order:
// 1. updateCortisolLevel, updateDopamineLevel
// 2. applyCortisolEffects (sets threshold from base, sets initial synFactor)
// 3. applyDopamineEffects (modifies threshold further, modifies synFactor further)
// This order ensures combined effects are calculated correctly.
