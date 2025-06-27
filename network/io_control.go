package network

import (
	"fmt"
	"math"

	"crownet/common"
	"crownet/neuron"
	"crownet/pulse"
)

// SetDynamicState enables or disables major dynamic processes within the CrowNet simulation.
// This allows for runtime control over learning, synaptogenesis (structural plasticity),
// and neurochemical modulation.
//
// Parameters:
//
//	learning: If true, Hebbian learning and other synaptic plasticity rules are active.
//	synaptogenesis: If true, structural plasticity mechanisms (neuron movement, connection pruning/formation) are active.
//	chemicalModulation: If true, neurochemical systems (e.g., cortisol, dopamine) influence network behavior.
func (cn *CrowNet) SetDynamicState(learning, synaptogenesis, chemicalModulation bool) {
	cn.isLearningEnabled = learning
	cn.isSynaptogenesisEnabled = synaptogenesis
	cn.isChemicalModulationEnabled = chemicalModulation
}

// ResetNetworkStateForNewPattern prepares the network for a new input pattern
// presentation by resetting transient neuronal states and clearing active signals.
// Specifically, it:
// - Clears accumulated potentials in all neurons.
// - Removes all active pulses from the PulseList.
// Note: This method does not reset neuron firing states (e.g., a neuron in a refractory
// period will remain so). It primarily focuses on clearing immediate electrical activity.
func (cn *CrowNet) ResetNetworkStateForNewPattern() {
	for _, n := range cn.Neurons {
		n.AccumulatedPotential = 0.0
		// Note: Neuron's internal state (Resting, Firing, etc.) is not reset here,
		// only the potential that might lead to immediate firing.
		// Firing state is typically managed by AdvanceState or direct activation (e.g., PresentPattern).
	}
	cn.ActivePulses.Clear()
}

// PresentPattern presents an input pattern to the network by activating a subset
// of its input neurons.
// Each element in patternData corresponds to an input neuron (ordered by InputNeuronIDs).
// If a patternData element's value is greater than 0.5, the corresponding input neuron
// is set to a Firing state, and a new pulse is emitted from it.
//
// Parameters:
//
//	patternData: A slice of float64 values representing the input pattern. The length
//	             must match SimParams.PatternSize.
//
// Returns:
//
//	An error if SimParams are not initialized, if the patternData size is incorrect,
//	if there are insufficient configured input neurons, or if an expected input
//	neuron is not found or is not of Type Input. Returns nil on success.
func (cn *CrowNet) PresentPattern(patternData []float64) error {
	if cn.SimParams == nil {
		return fmt.Errorf("PresentPattern: SimParams not initialized in CrowNet")
	}
	if len(patternData) != cn.SimParams.Pattern.PatternSize {
		return fmt.Errorf("pattern data size (%d) does not match configured PatternSize (%d)", len(patternData), cn.SimParams.Pattern.PatternSize)
	}
	if len(cn.InputNeuronIDs) < cn.SimParams.Pattern.PatternSize {
		return fmt.Errorf("insufficient input neurons (%d) for PatternSize (%d)", len(cn.InputNeuronIDs), cn.SimParams.Pattern.PatternSize)
	}

	for i := 0; i < cn.SimParams.Pattern.PatternSize; i++ {
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
					cn.SimParams.General.SpaceMaxDimension*2.0, // Max travel distance for these pulses
				)
				cn.ActivePulses.Add(newP)
			}
		}
	}
	return nil
}

// GetOutputActivation retrieves the current accumulated potentials of the
// network's output neurons.
// The number of output activations returned is determined by SimParams.MinOutputNeurons.
// The activations are ordered corresponding to the sorted OutputNeuronIDs.
//
// Returns:
//
//	A slice of float64 representing the accumulated potentials of the output neurons,
//	or nil and an error if SimParams are not initialized, if the actual number of
//	output neurons is less than configured MinOutputNeurons, or if an expected
//	output neuron is not found or is not of Type Output.
func (cn *CrowNet) GetOutputActivation() ([]float64, error) {
	if cn.SimParams == nil {
		return nil, fmt.Errorf("GetOutputActivation: SimParams not initialized in CrowNet")
	}
	if len(cn.OutputNeuronIDs) < cn.SimParams.Structure.MinOutputNeurons {
		return nil, fmt.Errorf("actual number of output neurons (%d) is less than configured MinOutputNeurons (%d)",
			len(cn.OutputNeuronIDs), cn.SimParams.Structure.MinOutputNeurons)
	}

	numOutputsToReport := cn.SimParams.Structure.MinOutputNeurons
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

// ConfigureFrequencyInput sets up or removes a continuous stimulus for a specific
// input neuron, causing it to fire periodically at approximately the given frequency.
// The time to the next firing event is initialized with some randomness to avoid
// synchronized starts if multiple inputs are configured simultaneously.
//
// Parameters:
//
//	neuronID: The ID of the input neuron to configure.
//	hz: The target firing frequency in Hertz. If hz is <= 0, any existing
//	    continuous stimulus for this neuronID is removed.
//
// Returns:
//
//	An error if the specified neuronID is not a valid input neuron, if SimParams
//	or RNG are not initialized, or if SimParams.CyclesPerSecond is not positive
//	(which is required for frequency calculation). Returns nil on success.
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
		if cn.SimParams.General.CyclesPerSecond <= 0 { // Avoid division by zero or negative
			return fmt.Errorf("CyclesPerSecond must be positive to calculate cyclesPerFiring, got %f", cn.SimParams.General.CyclesPerSecond)
		}
		cyclesPerFiring := cn.SimParams.General.CyclesPerSecond / hz
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

	cutoffCycle := cn.CycleCount - common.CycleCount(cn.SimParams.Structure.OutputFrequencyWindowCycles)

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

// GetOutputFrequency calculates the firing frequency in Hertz (Hz) of a specific
// output neuron. The calculation is based on its recorded firing events within
// the time window defined by SimParams.OutputFrequencyWindowCycles.
//
// Parameters:
//
//	neuronID: The ID of the output neuron whose firing frequency is to be calculated.
//
// Returns:
//
//	The calculated firing frequency in Hz, or 0.0 if there are no firings in the
//	history or if essential SimParams (like CyclesPerSecond or OutputFrequencyWindowCycles)
//	are invalid (e.g., zero or negative).
//	An error if the specified neuronID is not a valid output neuron or if critical
//	SimParams required for frequency calculation are not valid.
func (cn *CrowNet) GetOutputFrequency(neuronID common.NeuronID) (float64, error) {
	if _, isOutput := cn.OutputNeuronIDSet[neuronID]; !isOutput {
		return 0, fmt.Errorf("neuron ID %d is not a valid output neuron", neuronID)
	}
	if cn.SimParams == nil {
		return 0, fmt.Errorf("GetOutputFrequency: SimParams not initialized in CrowNet")
	}

	history := cn.outputFiringHistory[neuronID] // Known to exist
	if len(history) == 0 {
		return 0.0, nil // No firings in history means 0 Hz
	}

	firingsInWindow := len(history)

	if cn.SimParams.General.CyclesPerSecond <= 0 {
		return 0, fmt.Errorf("CyclesPerSecond (%f) must be positive to calculate frequency", cn.SimParams.General.CyclesPerSecond)
	}
	if cn.SimParams.Structure.OutputFrequencyWindowCycles <= 0 {
		return 0, fmt.Errorf("OutputFrequencyWindowCycles (%f) must be positive to calculate frequency", cn.SimParams.Structure.OutputFrequencyWindowCycles)
	}

	windowDurationSeconds := cn.SimParams.Structure.OutputFrequencyWindowCycles / cn.SimParams.General.CyclesPerSecond
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
					cn.SimParams.General.SpaceMaxDimension*2.0,
				)
				cn.ActivePulses.Add(newP)
			}

			targetHz, stillConfigured := cn.inputTargetFrequencies[neuronID]
			if stillConfigured && targetHz > 0 {
				if cn.SimParams.General.CyclesPerSecond <= 0 { // Should be caught by config validation
					// Log error or handle, for now just stop stimulating this neuron
					delete(cn.inputTargetFrequencies, neuronID)
					delete(cn.timeToNextInputFire, neuronID)
					continue
				}
				cyclesPerFiring := cn.SimParams.General.CyclesPerSecond / targetHz
				cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(cn.rng.Float64()*cyclesPerFiring)+1.0))
			} else {
				// If targetHz became <=0 or was deleted (e.g. configured off)
				delete(cn.inputTargetFrequencies, neuronID)
				delete(cn.timeToNextInputFire, neuronID)
			}
		}
	}
}
