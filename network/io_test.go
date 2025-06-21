package network

import (
	"crownet/datagen"
	"crownet/neuron"
	"math"
	"testing"
	// "fmt" // Not strictly needed for these tests
)

// Helper to create a minimal CrowNet for I/O tests
func newTestCrowNetForIO(numInputs int, numOutputs int) *CrowNet {
	totalNeurons := numInputs + numOutputs + 10 // Add some other neurons
	net := &CrowNet{
		Neurons:                  make([]*neuron.Neuron, 0, totalNeurons),
		InputNeuronIDs:           make([]int, 0),
		OutputNeuronIDs:          make([]int, 0),
		inputTargetFrequencies:   make(map[int]float64),
		timeToNextInputFire:      make(map[int]int),
		outputFiringHistory:      make(map[int][]int),
		SynapticWeights:          make(map[int]map[int]float64), // Needed for PresentPattern -> NewPulse
		EnableSynaptogenesis:     false,                         // Off for these tests
		EnableChemicalModulation: false,                         // Off for these tests
		BaseLearningRate:         0.01,
		CycleCount:               0,
		neuronIDCounter:          0,
		PulseMaxTravelDistance:   8.0, // Default from config
	}

	for i := 0; i < numInputs; i++ {
		id := net.getNextID()
		n := neuron.NewNeuron(id, neuron.Point{}, neuron.InputNeuron, 1.0)
		net.Neurons = append(net.Neurons, n)
		net.InputNeuronIDs = append(net.InputNeuronIDs, id)
	}
	for i := 0; i < numOutputs; i++ {
		id := net.getNextID()
		n := neuron.NewNeuron(id, neuron.Point{}, neuron.OutputNeuron, 1.0)
		net.Neurons = append(net.Neurons, n)
		net.OutputNeuronIDs = append(net.OutputNeuronIDs, id)
		net.outputFiringHistory[id] = make([]int, 0) // Initialize history
	}
	// Add a few other neurons
	for i := 0; i < 10; i++ {
		id := net.getNextID()
		n := neuron.NewNeuron(id, neuron.Point{}, neuron.ExcitatoryNeuron, 1.0)
		net.Neurons = append(net.Neurons, n)
	}
	return net
}

func TestSetInputFrequency(t *testing.T) {
	net := newTestCrowNetForIO(5, 5)
	if len(net.InputNeuronIDs) == 0 {
		t.Fatal("Test setup failed: no input neurons created.")
	}
	inputNeuronID := net.InputNeuronIDs[0]

	// Test valid frequency
	err := net.SetInputFrequency(inputNeuronID, 2.0) // 2 Hz
	if err != nil {
		t.Errorf("SetInputFrequency failed for valid input: %v", err)
	}
	expectedCycles := int(math.Round(CyclesPerSecond / 2.0)) // 10/2 = 5
	if net.timeToNextInputFire[inputNeuronID] != expectedCycles {
		t.Errorf("timeToNextInputFire incorrect. Expected %d, got %d", expectedCycles, net.timeToNextInputFire[inputNeuronID])
	}
	if net.inputTargetFrequencies[inputNeuronID] != 2.0 {
		t.Errorf("inputTargetFrequencies incorrect. Expected 2.0, got %.2f", net.inputTargetFrequencies[inputNeuronID])
	}

	// Test high frequency (should cap at 1 fire per cycle)
	err = net.SetInputFrequency(inputNeuronID, CyclesPerSecond*2) // e.g., 20 Hz if CyclesPerSecond is 10
	if err != nil {
		t.Errorf("SetInputFrequency failed for high frequency: %v", err)
	}
	if net.timeToNextInputFire[inputNeuronID] != 1 { // Should be 1 (fire next cycle)
		t.Errorf("timeToNextInputFire for high freq incorrect. Expected 1, got %d", net.timeToNextInputFire[inputNeuronID])
	}

	// Test zero frequency (disable)
	err = net.SetInputFrequency(inputNeuronID, 0.0)
	if err != nil {
		t.Errorf("SetInputFrequency failed for zero frequency: %v", err)
	}
	if _, exists := net.timeToNextInputFire[inputNeuronID]; exists {
		t.Errorf("timeToNextInputFire should be deleted for zero frequency.")
	}
	if net.inputTargetFrequencies[inputNeuronID] != 0.0 {
		t.Errorf("inputTargetFrequencies incorrect for zero freq. Expected 0.0, got %.2f", net.inputTargetFrequencies[inputNeuronID])
	}

	// Test invalid neuron ID
	err = net.SetInputFrequency(999, 1.0) // Assuming 999 is not a valid input neuron ID
	if err == nil {
		t.Errorf("SetInputFrequency should have failed for invalid neuron ID.")
	}
}

func TestProcessInputs(t *testing.T) {
	net := newTestCrowNetForIO(1, 0)
	inputID := net.InputNeuronIDs[0]

	// Set to fire every 2 cycles (CyclesPerSecond / 5Hz = 2, if CyclesPerSecond = 10)
	freq := CyclesPerSecond / 2.0
	net.SetInputFrequency(inputID, freq)

	initialPulseCount := len(net.ActivePulses)

	// Cycle 0: timeLeft = 2 -> 1. No fire.
	net.processInputs()
	if len(net.ActivePulses) != initialPulseCount {
		t.Errorf("Cycle 0: Pulses created prematurely. Expected %d, got %d", initialPulseCount, len(net.ActivePulses))
	}

	// Cycle 1: timeLeft = 1 -> 0. Fire. Timer reset to 2.
	net.CycleCount = 1
	net.processInputs()
	if len(net.ActivePulses) != initialPulseCount+1 {
		t.Errorf("Cycle 1: Pulse not created. Expected %d, got %d", initialPulseCount+1, len(net.ActivePulses))
	}
	if net.timeToNextInputFire[inputID] != 2 {
		t.Errorf("Cycle 1: Timer not reset correctly. Expected 2, got %d", net.timeToNextInputFire[inputID])
	}
	initialPulseCount = len(net.ActivePulses)

	// Cycle 2: timeLeft = 2 -> 1. No fire.
	net.CycleCount = 2
	net.processInputs()
	if len(net.ActivePulses) != initialPulseCount {
		t.Errorf("Cycle 2: Pulses created prematurely. Expected %d, got %d", initialPulseCount, len(net.ActivePulses))
	}

	// Cycle 3: timeLeft = 1 -> 0. Fire. Timer reset to 2.
	net.CycleCount = 3
	net.processInputs()
	if len(net.ActivePulses) != initialPulseCount+1 {
		t.Errorf("Cycle 3: Pulse not created. Expected %d, got %d", initialPulseCount+1, len(net.ActivePulses))
	}
}

func TestPresentPattern(t *testing.T) {
	net := newTestCrowNetForIO(datagen.PatternSize, 0) // Ensure enough input neurons

	pattern, _ := datagen.GetDigitPattern(1) // Get pattern for '1'

	initialPulseCount := len(net.ActivePulses)
	err := net.PresentPattern(pattern)
	if err != nil {
		t.Fatalf("PresentPattern failed: %v", err)
	}

	expectedPulses := 0
	for _, val := range pattern {
		if val > 0.5 {
			expectedPulses++
		}
	}

	if len(net.ActivePulses) != initialPulseCount+expectedPulses {
		t.Errorf("PresentPattern did not create correct number of pulses. Expected %d, got %d", initialPulseCount+expectedPulses, len(net.ActivePulses))
	}

	// Check if the correct neurons are set to FiringState (they are, by PresentPattern logic)
	// And if pulses have the correct base signal
	for i := 0; i < expectedPulses; i++ {
		pulse := net.ActivePulses[initialPulseCount+i]
		if pulse.Value != 1.0 { // Input neurons have base signal 1.0
			t.Errorf("Pulse %d from PresentPattern has incorrect base signal value %.2f", i, pulse.Value)
		}
	}

	// Test error for too large pattern
	largePattern := make([]float64, datagen.PatternSize+1)
	err = net.PresentPattern(largePattern)
	if err == nil {
		t.Errorf("PresentPattern should have failed for pattern larger than input neurons.")
	}
}

func TestOutputFrequency(t *testing.T) {
	net := newTestCrowNetForIO(0, 1)
	outputID := net.OutputNeuronIDs[0]

	// Simulate some firings
	// Window = 20 cycles (CyclesPerSecond * 2)
	net.recordOutputFiring(outputID, 5)
	net.recordOutputFiring(outputID, 10)
	net.recordOutputFiring(outputID, 15) // 3 firings in last 20 cycles if current is ~20-24

	net.CycleCount = 24 // Current cycle for GetOutputFrequency window end

	freq, err := net.GetOutputFrequency(outputID)
	if err != nil {
		t.Fatalf("GetOutputFrequency failed: %v", err)
	}
	// Expected: 3 firings in 2 seconds (20 cycles / 10 CyclesPerSec) = 1.5 Hz
	expectedFreq := 3.0 / (OutputFrequencyWindowCycles / CyclesPerSecond)
	if math.Abs(freq-expectedFreq) > 1e-9 {
		t.Errorf("GetOutputFrequency incorrect. Expected %.2f Hz, got %.2f Hz", expectedFreq, freq)
	}

	// Test pruning: Add a firing outside the window
	net.recordOutputFiring(outputID, 1) // This should be pruned if current cycle is 25+
	net.CycleCount = 30
	// Now, firings at 5,10,15. Cutoff for window ending at 30 is cycle 10 (30-20).
	// So, firings at 10, 15 should be counted (2 firings).
	freq, _ = net.GetOutputFrequency(outputID)
	expectedFreq = 2.0 / (OutputFrequencyWindowCycles / CyclesPerSecond) // 2 firings / 2s = 1 Hz
	if math.Abs(freq-expectedFreq) > 1e-9 {
		t.Errorf("GetOutputFrequency after pruning incorrect. Expected %.2f Hz, got %.2f Hz", expectedFreq, freq)
		t.Logf("History for %d: %v", outputID, net.outputFiringHistory[outputID])
	}

}
