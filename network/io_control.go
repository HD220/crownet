package network

import (
	"crownet/common"
	"crownet/neuron"
	"crownet/pulse"
	"fmt"
	"math"
)

// SetDynamicState allows toggling of major dynamic processes in the network.
func (cn *CrowNet) SetDynamicState(learning, synaptogenesis, chemicalModulation bool) {
	cn.isLearningEnabled = learning
	cn.isSynaptogenesisEnabled = synaptogenesis
	cn.isChemicalModulationEnabled = chemicalModulation
}

// ResetNetworkStateForNewPattern resets transient states of the network,
// preparing it for the presentation of a new input pattern.
// It clears accumulated potentials in all neurons and removes all active pulses.
func (cn *CrowNet) ResetNetworkStateForNewPattern() {
	for _, n := range cn.Neurons {
		n.AccumulatedPotential = 0.0
		// Note: Neuron's internal state (Resting, Firing, etc.) is not reset here,
		// only the potential that might lead to immediate firing.
		// Firing state is typically managed by AdvanceState or direct activation (e.g., PresentPattern).
	}
	cn.ActivePulses.Clear()
}

// PresentPattern activates specific input neurons based on the provided patternData.
// Active points in the pattern (value > 0.5) cause corresponding input neurons to fire.
// It returns an error if patternData size doesn't match configuration or if input neurons are insufficient/not found.
func (cn *CrowNet) PresentPattern(patternData []float64) error {
	if cn.SimParams == nil {
		return fmt.Errorf("PresentPattern: SimParams not initialized in CrowNet")
	}
	if len(patternData) != cn.SimParams.PatternSize {
		return fmt.Errorf("pattern data size (%d) does not match configured PatternSize (%d)", len(patternData), cn.SimParams.PatternSize)
	}
	if len(cn.InputNeuronIDs) < cn.SimParams.PatternSize {
		return fmt.Errorf("insufficient input neurons (%d) for PatternSize (%d)", len(cn.InputNeuronIDs), cn.SimParams.PatternSize)
	}

	for i := 0; i < cn.SimParams.PatternSize; i++ {
		if patternData[i] > 0.5 { // Consider this pixel/feature active
			if i >= len(cn.InputNeuronIDs) { // Defensive check
				return fmt.Errorf("pattern index %d out of bounds for InputNeuronIDs (len %d)", i, len(cn.InputNeuronIDs))
			}
			inputNeuronID := cn.InputNeuronIDs[i]

			targetNeuron, ok := cn.neuronMap[inputNeuronID]
			if !ok {
				return fmt.Errorf("input neuron ID %d (at pattern index %d) not found in neuronMap", inputNeuronID, i)
			}
			if targetNeuron.Type != neuron.Input {
				return fmt.Errorf("neuron ID %d (at pattern index %d) is not of Type Input", inputNeuronID, i)
			}

			targetNeuron.CurrentState = neuron.Firing // Set state to firing
			emittedSignal := targetNeuron.EmittedPulseSign()
			if emittedSignal != 0 {
				newP := pulse.New(
					targetNeuron.ID,
					targetNeuron.Position,
					emittedSignal,
					cn.CycleCount,
					cn.SimParams.SpaceMaxDimension*2.0, // Max travel distance for these pulses
				)
				cn.ActivePulses.Add(newP)
			}
		}
	}
	return nil
}

// GetOutputActivation retrieves the current accumulated potentials of the output neurons.
// The order of activations corresponds to the sorted order of OutputNeuronIDs.
// Returns an error if the number of output neurons is less than configured MinOutputNeurons or if an output neuron is not found.
func (cn *CrowNet) GetOutputActivation() ([]float64, error) {
	if cn.SimParams == nil {
		return nil, fmt.Errorf("GetOutputActivation: SimParams not initialized in CrowNet")
	}
	if len(cn.OutputNeuronIDs) < cn.SimParams.MinOutputNeurons {
		return nil, fmt.Errorf("actual number of output neurons (%d) is less than configured MinOutputNeurons (%d)",
			len(cn.OutputNeuronIDs), cn.SimParams.MinOutputNeurons)
	}

	numOutputsToReport := cn.SimParams.MinOutputNeurons
	outputActivations := make([]float64, numOutputsToReport)

	for i := 0; i < numOutputsToReport; i++ {
		if i >= len(cn.OutputNeuronIDs) { // Defensive check
			return nil, fmt.Errorf("logic error: index %d out of bounds for OutputNeuronIDs (len %d)", i, len(cn.OutputNeuronIDs))
		}
		outputNeuronID := cn.OutputNeuronIDs[i]

		targetNeuron, ok := cn.neuronMap[outputNeuronID]
		if !ok {
			return nil, fmt.Errorf("output neuron ID %d (expected at sorted index %d) not found in neuronMap", outputNeuronID, i)
		}
		if targetNeuron.Type != neuron.Output {
			return nil, fmt.Errorf("neuron ID %d (expected at sorted index %d) is not of Type Output", outputNeuronID, i)
		}
		outputActivations[i] = float64(targetNeuron.AccumulatedPotential)
	}
	return outputActivations, nil
}

// ConfigureFrequencyInput sets up a continuous stimulus for a specific input neuron at a given frequency.
// If hz is <= 0, any existing stimulus for that neuron is removed.
// Returns an error if the neuronID is not a valid input neuron.
func (cn *CrowNet) ConfigureFrequencyInput(neuronID common.NeuronID, hz float64) error {
	if _, isInput := cn.InputNeuronIDSet[neuronID]; !isInput {
		return fmt.Errorf("neuron ID %d is not a valid input neuron", neuronID)
	}
	if cn.SimParams == nil {
		return fmt.Errorf("ConfigureFrequencyInput: SimParams not initialized in CrowNet")
	}
	if cn.rng == nil {
		return fmt.Errorf("ConfigureFrequencyInput: RNG not initialized in CrowNet")
	}


	if hz <= 0 {
		delete(cn.inputTargetFrequencies, neuronID)
		delete(cn.timeToNextInputFire, neuronID)
	} else {
		cn.inputTargetFrequencies[neuronID] = hz
		if cn.SimParams.CyclesPerSecond <= 0 { // Avoid division by zero or negative
			return fmt.Errorf("CyclesPerSecond must be positive to calculate cyclesPerFiring, got %f", cn.SimParams.CyclesPerSecond)
		}
		cyclesPerFiring := cn.SimParams.CyclesPerSecond / hz
		cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(cn.rng.Float64()*cyclesPerFiring)+1.0))
	}
	return nil
}

// recordOutputFiring updates the firing history for a given output neuron.
// This method assumes neuronID has already been validated as an output neuron by the caller.
// It maintains a sliding window of recent firing times for frequency calculation.
func (cn *CrowNet) recordOutputFiring(neuronID common.NeuronID) {
	// Assumption: neuronID is a valid output neuron ID.
	// outputFiringHistory[neuronID] is guaranteed to exist due to pre-initialization in finalizeInitialization.
	if cn.SimParams == nil {
		// This should not happen in normal execution. Log or handle error if necessary.
		return
	}

	history := cn.outputFiringHistory[neuronID]
	history = append(history, cn.CycleCount)

	cutoffCycle := cn.CycleCount - common.CycleCount(cn.SimParams.OutputFrequencyWindowCycles)

	startIndex := 0
	for i, fireCycle := range history {
		if fireCycle >= cutoffCycle {
			startIndex = i
			break
		}
		if i == len(history)-1 { // All elements are older than cutoff
			startIndex = len(history)
		}
	}

	cn.outputFiringHistory[neuronID] = history[startIndex:]
}

// GetOutputFrequency calculates the firing frequency (in Hz) of a specific output neuron
// based on its recorded firing history within the OutputFrequencyWindowCycles.
// Returns an error if neuronID is not a valid output neuron or if configured cycle durations are zero/invalid.
func (cn *CrowNet) GetOutputFrequency(neuronID common.NeuronID) (float64, error) {
	if _, isOutput := cn.OutputNeuronIDSet[neuronID]; !isOutput {
		return 0, fmt.Errorf("neuron ID %d is not a valid output neuron", neuronID)
	}
	if cn.SimParams == nil {
		return 0, fmt.Errorf("GetOutputFrequency: SimParams not initialized in CrowNet")
	}


	history, _ := cn.outputFiringHistory[neuronID] // Known to exist
	if len(history) == 0 {
		return 0.0, nil // No firings in history means 0 Hz
	}

	firingsInWindow := len(history)

	if cn.SimParams.CyclesPerSecond <= 0 {
		return 0, fmt.Errorf("CyclesPerSecond (%f) must be positive to calculate frequency", cn.SimParams.CyclesPerSecond)
	}
	if cn.SimParams.OutputFrequencyWindowCycles <= 0 {
		return 0, fmt.Errorf("OutputFrequencyWindowCycles (%f) must be positive to calculate frequency", cn.SimParams.OutputFrequencyWindowCycles)
	}

	windowDurationSeconds := cn.SimParams.OutputFrequencyWindowCycles / cn.SimParams.CyclesPerSecond
	// Epsilon check for windowDurationSeconds is implicitly handled by positive checks above for its components.

	frequencyHz := float64(firingsInWindow) / windowDurationSeconds
	return frequencyHz, nil
}

// processFrequencyInputs handles continuous, frequency-based stimulation of input neurons.
// It checks `timeToNextInputFire` for each configured neuron and fires them if their timer is up,
// then schedules the next firing based on the target frequency and some randomness.
func (cn *CrowNet) processFrequencyInputs() {
	if cn.SimParams == nil || cn.rng == nil {
		// Should not happen with proper initialization.
		return
	}
	for neuronID, timeLeft := range cn.timeToNextInputFire {
		newTimeLeft := timeLeft - 1
		cn.timeToNextInputFire[neuronID] = newTimeLeft

		if newTimeLeft <= 0 {
			targetNeuron, ok := cn.neuronMap[neuronID]
			if !ok || targetNeuron.Type != neuron.Input {
				// Configured for a non-existent or non-input neuron. Clean up.
				delete(cn.inputTargetFrequencies, neuronID)
				delete(cn.timeToNextInputFire, neuronID)
				continue
			}

			targetNeuron.CurrentState = neuron.Firing
			emittedSignal := targetNeuron.EmittedPulseSign()
			if emittedSignal != 0 {
				newP := pulse.New(
					targetNeuron.ID,
					targetNeuron.Position,
					emittedSignal,
					cn.CycleCount,
					cn.SimParams.SpaceMaxDimension*2.0,
				)
				cn.ActivePulses.Add(newP)
			}

			targetHz, stillConfigured := cn.inputTargetFrequencies[neuronID]
			if stillConfigured && targetHz > 0 {
				if cn.SimParams.CyclesPerSecond <=0 { // Should be caught by config validation
					// Log error or handle, for now just stop stimulating this neuron
					delete(cn.inputTargetFrequencies, neuronID)
					delete(cn.timeToNextInputFire, neuronID)
					continue
				}
				cyclesPerFiring := cn.SimParams.CyclesPerSecond / targetHz
				cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(cn.rng.Float64()*cyclesPerFiring)+1.0))
			} else {
				// If targetHz became <=0 or was deleted (e.g. configured off)
				delete(cn.inputTargetFrequencies, neuronID)
				delete(cn.timeToNextInputFire, neuronID)
			}
		}
	}
}
