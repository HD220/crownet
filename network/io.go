package network

import (
	"crownet/neuron"
	"fmt"
	"math"
	"sort"
)

// SetInputFrequency sets the target firing frequency for a specific input neuron.
func (cn *CrowNet) SetInputFrequency(neuronID int, hz float64) error {
	isInputNeuron := false
	var selectedNeuron *neuron.Neuron
	for _, id := range cn.InputNeuronIDs {
		if id == neuronID {
			isInputNeuron = true
			break
		}
	}
	if !isInputNeuron {
		return fmt.Errorf("neuron ID %d is not an input neuron", neuronID)
	}
	for _, n := range cn.Neurons { // Get the neuron pointer
		if n.ID == neuronID {
			selectedNeuron = n
			break
		}
	}
	if selectedNeuron == nil {
		return fmt.Errorf("input neuron ID %d not found in neuron list", neuronID)
	}

	cn.inputTargetFrequencies[neuronID] = hz
	if hz <= 0 {
		// If frequency is zero or negative, effectively disable it.
		// No need to set timeToNextInputFire, it won't be checked.
		delete(cn.timeToNextInputFire, neuronID) // Remove if previously set
	} else {
		cyclesPerFiring := CyclesPerSecond / hz
		if cyclesPerFiring < 1 { // Max 1 fire per cycle if hz > CyclesPerSecond
			cyclesPerFiring = 1
		}
		// Start firing at the next cycle (or spread initial firings)
		cn.timeToNextInputFire[neuronID] = int(math.Round(cyclesPerFiring))
		if cn.timeToNextInputFire[neuronID] == 0 && cyclesPerFiring > 0 { // Ensure at least 1 if rounded to 0
			cn.timeToNextInputFire[neuronID] = 1
		}
	}
	// fmt.Printf("Set input freq for Neuron %d to %.2f Hz. Cycles per fire: %.2f, Time to next: %d\n", neuronID, hz, CyclesPerSecond/hz, cn.timeToNextInputFire[neuronID])
	return nil
}

// processInputs checks and fires input neurons based on their set frequencies.
// This should be called early in the RunCycle.
func (cn *CrowNet) processInputs() {
	for neuronID, timeLeft := range cn.timeToNextInputFire {
		timeLeft--
		cn.timeToNextInputFire[neuronID] = timeLeft

		if timeLeft <= 0 {
			var inputNeuron *neuron.Neuron
			for _, n := range cn.Neurons {
				if n.ID == neuronID {
					inputNeuron = n
					break
				}
			}

			if inputNeuron != nil && inputNeuron.Type == neuron.InputNeuron {
				// fmt.Printf("Cycle %d: Firing input neuron %d due to frequency schedule.\n", cn.CycleCount, neuronID)

				// For InputNeurons, frequency-based firing should override refractory states
				// as it represents an external, paced stimulus.
				// The neuron's state machine will still progress (e.g. it will enter refractory after this forced fire),
				// but that won't prevent the *next* scheduled input spike.
				inputNeuron.State = neuron.FiringState // Will transition to refractory in UpdateState by neuron.UpdateState()
				// inputNeuron.LastFiredCycle = cn.CycleCount // UpdateState handles this

				baseSignal := inputNeuron.GetBasePulseSign() // Get +1.0 or -1.0
				if baseSignal != 0 {                         // Only create pulses that have a potential effect type
					p := NewPulse(inputNeuron.ID, inputNeuron.Position, baseSignal, cn.CycleCount, cn.PulseMaxTravelDistance)
					// Add to a temporary list, to be merged with ActivePulses after input processing,
					// or directly if safe (ActivePulses is processed later in cycle).
					// For simplicity, adding directly.
					cn.ActivePulses = append(cn.ActivePulses, p)
					// fmt.Printf("  Pulse from freq-fired input neuron %d added. Value: %.2f\n", inputNeuron.ID, pulseVal)
				}
				// The extra brace was here. Now removed.
			} // This brace closes "if inputNeuron != nil && inputNeuron.Type == neuron.InputNeuron"
			// Reset timer for next firing
			targetHz := cn.inputTargetFrequencies[neuronID]
			if targetHz > 0 {
				cyclesPerFiring := CyclesPerSecond / targetHz
				if cyclesPerFiring < 1 {
					cyclesPerFiring = 1
				}
				cn.timeToNextInputFire[neuronID] = int(math.Max(1.0, math.Round(cyclesPerFiring)))
			} else {
				delete(cn.timeToNextInputFire, neuronID) // Should have been removed by SetInputFrequency if hz <=0
			}
		}
	}
}

// recordOutputFiring is called when an output neuron fires.
func (cn *CrowNet) recordOutputFiring(neuronID int, cycle int) {
	isOutputNeuron := false
	for _, id := range cn.OutputNeuronIDs {
		if id == neuronID {
			isOutputNeuron = true
			break
		}
	}
	if !isOutputNeuron {
		return // Not an output neuron
	}

	history, _ := cn.outputFiringHistory[neuronID]
	history = append(history, cycle)

	// Prune history older than OutputFrequencyWindowCycles
	cutoffCycle := cycle - int(OutputFrequencyWindowCycles)
	prunedHistory := make([]int, 0, len(history))
	for _, fireCycle := range history {
		if fireCycle >= cutoffCycle {
			prunedHistory = append(prunedHistory, fireCycle)
		}
	}
	cn.outputFiringHistory[neuronID] = prunedHistory
	// fmt.Printf("Cycle %d: Output neuron %d fired. History size: %d\n", cycle, neuronID, len(prunedHistory))
}

// GetOutputFrequency calculates the firing frequency of an output neuron in Hz.
func (cn *CrowNet) GetOutputFrequency(neuronID int) (float64, error) {
	isOutputNeuron := false
	for _, id := range cn.OutputNeuronIDs {
		if id == neuronID {
			isOutputNeuron = true
			break
		}
	}
	if !isOutputNeuron {
		return 0, fmt.Errorf("neuron ID %d is not an output neuron", neuronID)
	}

	history, ok := cn.outputFiringHistory[neuronID]
	if !ok || len(history) == 0 {
		return 0.0, nil // No firings recorded or not an output neuron with history
	}

	// Count firings within the window ending at the current cycle
	// (or last recorded firing if network not running)
	currentSimCycle := cn.CycleCount
	// If called post-simulation, use the last firing time in history as reference,
	// or assume cn.CycleCount is the end of simulation.

	cutoffCycle := currentSimCycle - int(OutputFrequencyWindowCycles)
	firingsInWindow := 0
	for _, fireCycle := range history {
		if fireCycle >= cutoffCycle && fireCycle < currentSimCycle { // Count fires up to (but not including) current cycle if mid-run
			firingsInWindow++
		}
	}

	// If history is very short (less than a full window), this might be misleading.
	// The window duration is OutputFrequencyWindowCycles.
	windowDurationSeconds := OutputFrequencyWindowCycles / CyclesPerSecond
	if windowDurationSeconds == 0 {
		return 0, fmt.Errorf("OutputFrequencyWindowCycles or CyclesPerSecond is zero, cannot calculate frequency")
	}

	frequencyHz := float64(firingsInWindow) / windowDurationSeconds
	return frequencyHz, nil
}

// Helper in network.go, in RunCycle, after a neuron fires (ReceivePulse returns true):
//
//	if firedNeuron.Type == neuron.OutputNeuron {
//	   cn.recordOutputFiring(firedNeuron.ID, cn.CycleCount)
//	}
//
// Also, the temporary input firing in RunCycle should be removed.
// Inputs are now handled by processInputs().
func (cn *CrowNet) initializeIO() {
	// Initialize maps if they were nil (they are initialized in NewCrowNet now)
	// cn.inputTargetFrequencies = make(map[int]float64)
	// cn.timeToNextInputFire = make(map[int]int)
	// cn.outputFiringHistory = make(map[int][]int)

	// Ensure output neuron histories are initialized
	for _, outID := range cn.OutputNeuronIDs {
		if _, exists := cn.outputFiringHistory[outID]; !exists {
			cn.outputFiringHistory[outID] = make([]int, 0)
		}
	}
	// For input neurons, SetInputFrequency will populate necessary maps.
}

// Sort OutputNeuronIDs for deterministic GetOutputFrequency iteration if needed by external caller
func (cn *CrowNet) finalizeInitialization() {
	sort.Ints(cn.InputNeuronIDs)
	sort.Ints(cn.OutputNeuronIDs)
	cn.initializeIO() // Make sure IO maps are ready
}

// PresentPattern activates a set of input neurons based on the provided pattern.
// The pattern is a flattened list of activations (0.0 or 1.0).
// The first N input neurons (where N is len(pattern)) will be stimulated.
// Stimulation means forcing them to fire once in the current cycle.
func (cn *CrowNet) PresentPattern(pattern []float64) error {
	if len(pattern) > len(cn.InputNeuronIDs) {
		return fmt.Errorf("pattern size (%d) is larger than the number of available input neurons (%d)", len(pattern), len(cn.InputNeuronIDs))
	}
	if len(pattern) == 0 {
		return fmt.Errorf("pattern is empty")
	}

	// fmt.Printf("[NETWORK] Presenting pattern of size %d to first %d input neurons.\n", len(pattern), len(pattern))

	for i, activation := range pattern {
		if activation > 0.5 { // Consider it "on"
			inputNeuronID := cn.InputNeuronIDs[i] // Get the i-th input neuron ID
			var inputNeuron *neuron.Neuron
			for _, n := range cn.Neurons {
				if n.ID == inputNeuronID {
					inputNeuron = n
					break
				}
			}

			if inputNeuron != nil && inputNeuron.Type == neuron.InputNeuron {
				// Force fire: set state and create pulse
				// This bypasses refractory checks for direct pattern presentation.
				// fmt.Printf("[NETWORK] Forcing Input Neuron %d to fire for pattern.\n", inputNeuron.ID)
				inputNeuron.State = neuron.FiringState

				baseSignal := inputNeuron.GetBasePulseSign() // Should be +1.0 for InputNeuron
				if baseSignal != 0 {
					// Note: Pulses created here will be processed in the *next* call to processPulses,
					// or if processPulses is called after this in the same cycle.
					// For typical training/classification, one might call RunCycle(s) after PresentPattern.
					// For now, add to ActivePulses. The next RunCycle will propagate them.
					p := NewPulse(inputNeuron.ID, inputNeuron.Position, baseSignal, cn.CycleCount, cn.PulseMaxTravelDistance)
					cn.ActivePulses = append(cn.ActivePulses, p)
				} else {
					// This should not happen for an InputNeuron if GetBasePulseSign is set up correctly
					fmt.Printf("[WARN] Input neuron %d generated a base signal of zero.\n", inputNeuron.ID)
				}
			} else {
				return fmt.Errorf("could not find input neuron with ID %d from InputNeuronIDs index %d, or it's not an InputNeuron type", inputNeuronID, i)
			}
		}
	}
	return nil
}
