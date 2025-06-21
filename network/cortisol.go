package network

import (
	"crownet/utils"
	"fmt" // Re-add for debug printing
	"math"
)

// updateCortisolLevel adjusts the network's cortisol level based on production and decay.
func (cn *CrowNet) updateCortisolLevel() {
	// Production: Check for excitatory pulses hitting the cortisol gland
	pulsesHittingGland := 0
	for _, p := range cn.ActivePulses {
		// Only consider excitatory pulses for cortisol production
		// Assuming positive value means excitatory. This should be based on emitting neuron type ideally.
		// For now, p.Value > 0 from ExcitatoryNeuron, InputNeuron, OutputNeuron is 0.3.
		if p.Value > 0 { // Simplistic check for excitatory
			// Check if the pulse's current effective area overlaps with the gland
			// A pulse "hits" if the gland is within its current propagation shell
			pulseEffectStartDist, pulseEffectEndDist := p.GetEffectRangeForCycle()
			distToGland := utils.EuclideanDistance(p.OriginPosition, CortisolGlandPosition)

			if distToGland >= pulseEffectStartDist && distToGland < pulseEffectEndDist {
				pulsesHittingGland++
			}
		}
	}

	if pulsesHittingGland > 0 {
		production := float64(pulsesHittingGland) * CortisolProductionPerHit
		cn.CortisolLevel += production
		if DebugCortisolHit { // Check the flag
			fmt.Printf("[DEBUG] Cycle %d: Cortisol gland hit by %d excitatory pulse(s). Produced: %.3f. New Level (before decay): %.3f\n", cn.CycleCount, pulsesHittingGland, production, cn.CortisolLevel)
		}
	}

	// Decay
	cn.CortisolLevel -= cn.CortisolLevel * CortisolDecayRate
	if cn.CortisolLevel < 0 {
		cn.CortisolLevel = 0
	}
	// Clamp cortisol to a maximum level
	if cn.CortisolLevel > CortisolMaxLevel {
		cn.CortisolLevel = CortisolMaxLevel
	}
	// fmt.Printf("Cycle %d: Cortisol level after decay: %.3f\n", cn.CycleCount, cn.CortisolLevel)
}

// applyCortisolEffects updates neuron firing thresholds and synaptogenesis based on cortisol levels.
func (cn *CrowNet) applyCortisolEffects() {
	var currentSynaptogenesisFactor float64 = 1.0 // Default: no change

	// Synaptogenesis reduction at high cortisol
	if cn.CortisolLevel >= CortisolHighEffectThreshold {
		// Scale factor: 1.0 at CortisolHighEffectThreshold, SynaptogenesisReductionFactor at CortisolMaxLevel
		// Linear interpolation for simplicity
		if CortisolMaxLevel > CortisolHighEffectThreshold { // Avoid division by zero
			t := (cn.CortisolLevel - CortisolHighEffectThreshold) / (CortisolMaxLevel - CortisolHighEffectThreshold)
			currentSynaptogenesisFactor = 1.0 - t*(1.0-SynaptogenesisReductionFactor) // Lerp between 1.0 and SynaptogenesisReductionFactor
			currentSynaptogenesisFactor = math.Max(SynaptogenesisReductionFactor, math.Min(1.0, currentSynaptogenesisFactor))
		} else if cn.CortisolLevel >= CortisolMaxLevel { // handles CortisolMaxLevel == CortisolHighEffectThreshold
			currentSynaptogenesisFactor = SynaptogenesisReductionFactor
		}
		// fmt.Printf("Cycle %d: High cortisol (%.2f). Synaptogenesis factor: %.2f\n", cn.CycleCount, cn.CortisolLevel, currentSynaptogenesisFactor)
	}
	// Store this factor to be used in applySynaptogenesis calculations
	cn.currentSynaptogenesisModulationFactor = currentSynaptogenesisFactor

	// Update firing thresholds for all neurons
	for _, n := range cn.Neurons {
		thresholdFactor := 1.0 // Default: no change from base threshold

		if cn.CortisolLevel < CortisolMinEffectThreshold {
			// No significant effect or slight increase if we want to model it. For now, no change.
			thresholdFactor = 1.0
		} else if cn.CortisolLevel < CortisolOptimalLowThreshold {
			// Linearly decrease threshold from 1.0 down to MaxThresholdReductionFactor
			t := (cn.CortisolLevel - CortisolMinEffectThreshold) / (CortisolOptimalLowThreshold - CortisolMinEffectThreshold)
			thresholdFactor = 1.0 - t*(1.0-MaxThresholdReductionFactor)
		} else if cn.CortisolLevel <= CortisolOptimalHighThreshold {
			// Max threshold reduction
			thresholdFactor = MaxThresholdReductionFactor
		} else if cn.CortisolLevel < CortisolHighEffectThreshold {
			// Linearly increase threshold from MaxThresholdReductionFactor up to 1.0 (baseline)
			t := (cn.CortisolLevel - CortisolOptimalHighThreshold) / (CortisolHighEffectThreshold - CortisolOptimalHighThreshold)
			thresholdFactor = MaxThresholdReductionFactor + t*(1.0-MaxThresholdReductionFactor)
		} else { // CortisolLevel >= CortisolHighEffectThreshold
			// Linearly increase threshold from 1.0 up to ThresholdIncreaseFactorHigh
			if CortisolMaxLevel > CortisolHighEffectThreshold { // Avoid division by zero
				t := (cn.CortisolLevel - CortisolHighEffectThreshold) / (CortisolMaxLevel - CortisolHighEffectThreshold)
				thresholdFactor = 1.0 + t*(ThresholdIncreaseFactorHigh-1.0)
				thresholdFactor = math.Min(thresholdFactor, ThresholdIncreaseFactorHigh) // Cap at max increase
			} else { // handles CortisolMaxLevel == CortisolHighEffectThreshold
				thresholdFactor = ThresholdIncreaseFactorHigh
			}
		}

		n.CurrentFiringThreshold = n.BaseFiringThreshold * thresholdFactor
		if n.CurrentFiringThreshold < 0.01 { // Ensure threshold doesn't become zero or negative
			n.CurrentFiringThreshold = 0.01
		}
		// if n.ID < 5 { // Debug print for a few neurons
		// 	fmt.Printf("  Neuron %d: Cortisol %.2f, BaseThr %.2f, CurrThr %.2f (Factor %.2f)\n", n.ID, cn.CortisolLevel, n.BaseFiringThreshold, n.CurrentFiringThreshold, thresholdFactor)
		// }
	}
}

// Add a field to CrowNet to store this factor temporarily per cycle
// In network.go:
// currentSynaptogenesisModulationFactor float64
// Initialize it to 1.0 in NewCrowNet
// Call updateCortisolLevel and applyCortisolEffects in RunCycle
// Modify applySynaptogenesis to use getSynaptogenesisModulationFactor
func (cn *CrowNet) initializeChemicalModulation() {
	cn.currentSynaptogenesisModulationFactor = 1.0
	cn.CortisolLevel = 0.0
	cn.DopamineLevel = 0.0
	// Other chemical initializations if any
}
