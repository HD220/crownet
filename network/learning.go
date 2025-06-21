package network

import (
	"crownet/neuron"
	"fmt"
	"math"
	// "log"
)

// ResetNeuronActivations clears accumulated pulses and resets states for relevant neurons
// typically called before presenting a new pattern.
func (cn *CrowNet) ResetNeuronActivations() {
	for _, n := range cn.Neurons {
		n.AccumulatedPulse = 0.0
		// For self-organization, allow states to evolve more naturally,
		// but ensure they are not stuck in Firing/Refractory from a previous unrelated event
		// if a new distinct pattern presentation is beginning.
		// A simpler approach for now: reset if not already resting.
		if n.State != neuron.RestingState {
			n.State = neuron.RestingState
			n.CyclesInCurrentState = 0
		}
	}
}

// ExposeToPatterns runs the simulation, presenting digit patterns repeatedly
// to allow the network to self-organize via neuromodulated Hebbian learning.
func (cn *CrowNet) ExposeToPatterns(
	getPatternFunc func(digit int) ([]float64, error),
	numDigits int, // e.g., 10 for 0-9
	epochs int,
	cyclesPerPatternPresentation int,
	// learningRate float64, // learningRate is now cn.BaseLearningRate, modulated internally
) error {
	if len(cn.OutputNeuronIDs) < numDigits {
		return fmt.Errorf("not enough output neurons (%d) for %d classes", len(cn.OutputNeuronIDs), numDigits)
	}
	if cyclesPerPatternPresentation <= 0 {
		return fmt.Errorf("cyclesPerPatternPresentation must be positive")
	}

	// Ensure all dynamics are enabled for self-learning by default by CrowNet settings
	// cn.EnableSynaptogenesis = true
	// cn.EnableChemicalModulation = true

	for epoch := 0; epoch < epochs; epoch++ {
		fmt.Printf("Epoch %d/%d starting...\n", epoch+1, epochs)
		totalPatternsProcessed := 0
		for digit := 0; digit < numDigits; digit++ {
			pattern, err := getPatternFunc(digit)
			if err != nil {
				return fmt.Errorf("epoch %d: failed to get pattern for digit %d: %w", epoch, digit, err)
			}

			cn.ResetNeuronActivations()
			cn.ActivePulses = make([]*Pulse, 0)

			err = cn.PresentPattern(pattern)
			if err != nil {
				return fmt.Errorf("epoch %d, digit %d: failed to present pattern: %w", epoch, digit, err)
			}

			for i := 0; i < cyclesPerPatternPresentation; i++ {
				cn.RunCycle()
			}
			totalPatternsProcessed++
		}
		fmt.Printf("Epoch %d/%d completed. Processed %d patterns. Cortisol: %.3f, Dopamine: %.3f, Eff. LR Factor (example): %.4f\n",
			epoch+1, epochs, totalPatternsProcessed, cn.CortisolLevel, cn.DopamineLevel, cn.calculateEffectiveLearningRateMultiplier())
	}
	return nil
}

// GetOutputPatternForInput presents an input pattern and returns the activation state
// of the output neurons after a few cycles of network settling.
func (cn *CrowNet) GetOutputPatternForInput(inputPattern []float64, cyclesToSettle int) ([]float64, error) {
	if len(cn.OutputNeuronIDs) < 10 {
		return nil, fmt.Errorf("not enough output neurons for 10 classes (found %d)", len(cn.OutputNeuronIDs))
	}
	if cyclesToSettle <= 0 {
		return nil, fmt.Errorf("cyclesToSettle must be positive")
	}

	// Store original dynamic states
	originalEnableSynaptogenesis := cn.EnableSynaptogenesis
	originalEnableChemicalModulation := cn.EnableChemicalModulation

	// Disable complex dynamics for a cleaner feed-forward observation pass
	cn.EnableSynaptogenesis = false
	cn.EnableChemicalModulation = false // This will also cause Hebbian LR to be just BaseLearningRate

	cn.ResetNeuronActivations()
	cn.ActivePulses = make([]*Pulse, 0)

	err := cn.PresentPattern(inputPattern)
	if err != nil {
		// Restore original dynamic states before returning error
		cn.EnableSynaptogenesis = originalEnableSynaptogenesis
		cn.EnableChemicalModulation = originalEnableChemicalModulation
		return nil, fmt.Errorf("failed to present pattern for classification: %w", err)
	}

	for i := 0; i < cyclesToSettle; i++ {
		cn.RunCycle()
	}

	// Restore original dynamic states
	cn.EnableSynaptogenesis = originalEnableSynaptogenesis
	cn.EnableChemicalModulation = originalEnableChemicalModulation

	outputPattern := make([]float64, 10)
	for i := 0; i < 10; i++ {
		outputNeuronID := cn.OutputNeuronIDs[i]
		var outputNeuron *neuron.Neuron
		for _, n := range cn.Neurons {
			if n.ID == outputNeuronID {
				outputNeuron = n
				break
			}
		}
		if outputNeuron == nil {
			return nil, fmt.Errorf("output neuron ID %d for digit %d not found", outputNeuronID, i)
		}
		outputPattern[i] = outputNeuron.AccumulatedPulse
	}

	return outputPattern, nil
}

// calculateEffectiveLearningRateMultiplier computes the combined multiplier from dopamine and cortisol.
// This is a helper to see the modulation effect. Actual ELR is BaseLR * multiplier.
func (cn *CrowNet) calculateEffectiveLearningRateMultiplier() float64 {
	multiplier := 1.0

	normalizedDopamine := 0.0
	if DopamineMaxLevel > 0 && cn.DopamineLevel > 0 {
		normalizedDopamine = math.Min(1.0, cn.DopamineLevel/DopamineMaxLevel)
	}
	dopamineMultiplier := 1.0 + (MaxDopamineLearningMultiplier-1.0)*normalizedDopamine
	multiplier *= dopamineMultiplier

	if cn.CortisolLevel >= CortisolHighEffectThreshold {
		suppressionRange := CortisolMaxLevel - CortisolHighEffectThreshold
		cortisolFactor := CortisolLearningSuppressionFactor
		if suppressionRange > 0 {
			currentPosInRange := math.Max(0, cn.CortisolLevel-CortisolHighEffectThreshold)
			suppressionScale := math.Min(1.0, currentPosInRange/suppressionRange)
			cortisolFactor = 1.0 - suppressionScale*(1.0-CortisolLearningSuppressionFactor)
		}
		cortisolFactor = math.Max(MinLearningRateFactor, math.Min(1.0, cortisolFactor))
		multiplier *= cortisolFactor
	}
	return multiplier
}

// ApplyHebbianPlasticity updates synaptic weights based on neuron co-activity,
// modulated by dopamine and cortisol levels.
// Called from RunCycle.
func (cn *CrowNet) ApplyHebbianPlasticity() {
	effectiveLearningRate := cn.BaseLearningRate * cn.calculateEffectiveLearningRateMultiplier()

	if effectiveLearningRate < 1e-9 { // If learning is effectively off, skip
		return
	}

	// For Hebbian, we need to know which neurons fired *in the current cycle processing*.
	// Neuron.State == neuron.FiringState indicates it just crossed threshold in *this* cycle's ReceivePulse.
	// Neuron.LastFiredCycle == cn.CycleCount indicates it fired in the *current simulation cycle*.
	// HebbianCoincidenceWindow allows looking back slightly.

	for fromID, toMap := range cn.SynapticWeights {
		var fromNeuron *neuron.Neuron
		for _, n := range cn.Neurons {
			if n.ID == fromID {
				fromNeuron = n
				break
			}
		}
		if fromNeuron == nil {
			continue
		}

		preActivity := 0.0
		// Neuron fired in the window ending with the current cycle count
		if fromNeuron.LastFiredCycle != -1 && (cn.CycleCount-fromNeuron.LastFiredCycle) <= HebbianCoincidenceWindow {
			preActivity = 1.0
		}
		if preActivity == 0.0 {
			continue
		} // No pre-synaptic activity, no update for outgoing weights from this neuron

		for toID, currentWeight := range toMap {
			if fromID == toID {
				continue
			}

			var toNeuron *neuron.Neuron
			for _, n := range cn.Neurons {
				if n.ID == toID {
					toNeuron = n
					break
				}
			}
			if toNeuron == nil {
				continue
			}

			postActivity := 0.0
			if toNeuron.LastFiredCycle != -1 && (cn.CycleCount-toNeuron.LastFiredCycle) <= HebbianCoincidenceWindow {
				postActivity = 1.0
			}

			newWeight := currentWeight

			if preActivity > 0 && postActivity > 0 { // Both pre and post active
				deltaWeight := effectiveLearningRate * preActivity * postActivity
				newWeight += deltaWeight
			}
			// Note: Standard Hebbian rule only strengthens. Other variants include weakening.
			// For MVP, this simple strengthening on co-activity is a start.

			newWeight -= newWeight * HebbianWeightDecay

			if newWeight > HebbianWeightMax {
				newWeight = HebbianWeightMax
			} else if newWeight < HebbianWeightMin {
				newWeight = HebbianWeightMin
			}

			if math.Abs(newWeight-currentWeight) > 1e-7 { // Only update if changed significantly
				cn.SetWeight(fromID, toID, newWeight)
			}
		}
	}
}
