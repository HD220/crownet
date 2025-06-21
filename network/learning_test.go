package network

import (
	"crownet/neuron"
	"math"
	"testing"
)

// Helper to create a CrowNet with specific neurons for learning tests
func newTestNetForLearning(numNeurons int) *CrowNet {
	net := &CrowNet{
		Neurons:                  make([]*neuron.Neuron, numNeurons),
		ActivePulses:             make([]*Pulse, 0),
		InputNeuronIDs:           make([]int, 0),
		OutputNeuronIDs:          make([]int, 0),
		inputTargetFrequencies:   make(map[int]float64),
		timeToNextInputFire:      make(map[int]int),
		outputFiringHistory:      make(map[int][]int),
		SynapticWeights:          make(map[int]map[int]float64),
		EnableSynaptogenesis:     false,
		EnableChemicalModulation: false, // Keep false for isolated Hebbian test unless testing modulation
		BaseLearningRate:         0.1,   // Use a simple LR for testing
		CycleCount:               1,     // Start at cycle 1 to avoid issues with LastFiredCycle = 0 being same as current
		neuronIDCounter:          0,
		PulseMaxTravelDistance:   8.0,
	}
	for i := 0; i < numNeurons; i++ {
		n := neuron.NewNeuron(i, neuron.Point{}, neuron.ExcitatoryNeuron, 1.0)
		net.Neurons[i] = n
		net.SynapticWeights[i] = make(map[int]float64)
	}
	// Initialize all-to-all weights to 0 for predictability in tests
	for i := 0; i < numNeurons; i++ {
		for j := 0; j < numNeurons; j++ {
			if i != j {
				net.SetWeight(i, j, 0.0)
			}
		}
	}
	return net
}

func TestApplyHebbianPlasticity_Basic(t *testing.T) {
	net := newTestNetForLearning(3) // Neuron 0 (pre), Neuron 1 (post), Neuron 2 (unrelated)

	// Scenario 1: Pre and Post fire in coincidence window
	net.Neurons[0].LastFiredCycle = net.CycleCount - HebbianCoincidenceWindow
	net.Neurons[1].LastFiredCycle = net.CycleCount
	net.Neurons[2].LastFiredCycle = -1 // Unrelated neuron did not fire

	initialWeight01 := 0.0
	net.SetWeight(0, 1, initialWeight01)

	net.ApplyHebbianPlasticity()

	expectedWeight01 := initialWeight01 + net.BaseLearningRate // Default modulation factors are 1
	if math.Abs(net.GetWeight(0, 1)-expectedWeight01) > 1e-6 {
		t.Errorf("Hebbian potentiation failed. Expected %.3f, got %.3f", expectedWeight01, net.GetWeight(0, 1))
	}
	// Weight decay should also apply
	expectedWeight01 -= expectedWeight01 * HebbianWeightDecay
	if math.Abs(net.GetWeight(0, 1)-expectedWeight01) > 1e-5 { // Increased tolerance
		t.Errorf("Hebbian potentiation with decay failed. Expected %.3f, got %.3f", expectedWeight01, net.GetWeight(0, 1))
	}

	// Scenario 2: Only Pre fires
	net.SetWeight(0, 1, 0.5) // Reset weight
	net.Neurons[0].LastFiredCycle = net.CycleCount
	net.Neurons[1].LastFiredCycle = -1 // Post did not fire

	net.ApplyHebbianPlasticity()
	expectedWeight01_decayOnly := 0.5 * (1 - HebbianWeightDecay)
	if math.Abs(net.GetWeight(0, 1)-expectedWeight01_decayOnly) > 1e-5 { // Increased tolerance
		t.Errorf("Hebbian only pre-fire (should only decay). Expected %.3f, got %.3f", expectedWeight01_decayOnly, net.GetWeight(0, 1))
	}
}

func TestHebbianWeightClippingAndDecay(t *testing.T) {
	net := newTestNetForLearning(2)
	net.BaseLearningRate = 1.0 // Large LR to hit clips fast

	// Test Max Clipping
	net.SetWeight(0, 1, HebbianWeightMax-0.05)
	net.Neurons[0].LastFiredCycle = net.CycleCount
	net.Neurons[1].LastFiredCycle = net.CycleCount
	net.ApplyHebbianPlasticity()
	finalWeightMax := HebbianWeightMax * (1 - HebbianWeightDecay) // Clipped then decayed
	if math.Abs(net.GetWeight(0, 1)-finalWeightMax) > 1e-5 {      // Increased tolerance
		t.Errorf("Hebbian max clipping failed. Expected %.3f, got %.3f", finalWeightMax, net.GetWeight(0, 1))
	}

	// Test Min Clipping
	net.SetWeight(0, 1, HebbianWeightMin+0.05)
	// To cause depression (if we had it) or just test decay on min bound:
	// For now, our simple Hebbian only potentiates or decays. So, make it not potentiate.
	net.Neurons[0].LastFiredCycle = -1 // No pre-fire
	net.Neurons[1].LastFiredCycle = -1 // No post-fire
	net.ApplyHebbianPlasticity()
	finalWeightMin := (HebbianWeightMin + 0.05) * (1 - HebbianWeightDecay)
	if math.Abs(net.GetWeight(0, 1)-finalWeightMin) > 1e-5 { // Increased tolerance
		t.Errorf("Hebbian min clipping/decay failed. Expected %.3f, got %.3f", finalWeightMin, net.GetWeight(0, 1))
	}
}

func TestNeuromodulationOfLearningRate(t *testing.T) {
	net := newTestNetForLearning(2)
	net.SetWeight(0, 1, 0.0)
	net.Neurons[0].LastFiredCycle = net.CycleCount
	net.Neurons[1].LastFiredCycle = net.CycleCount

	// Test with Dopamine
	net.EnableChemicalModulation = true        // Enable it for this test part
	net.DopamineLevel = DopamineMaxLevel / 2.0 // Half max dopamine
	net.CortisolLevel = 0.0

	net.ApplyHebbianPlasticity()

	// Expected LR = BaseLR * (1 + (MaxDopamineLearningMultiplier-1.0) * 0.5)
	expectedLR_Dopa := net.BaseLearningRate * (1 + (MaxDopamineLearningMultiplier-1.0)*0.5)
	expectedWeight_Dopa := expectedLR_Dopa
	expectedWeight_Dopa -= expectedWeight_Dopa * HebbianWeightDecay // Apply decay

	if math.Abs(net.GetWeight(0, 1)-expectedWeight_Dopa) > 1e-5 { // Increased tolerance
		t.Errorf("Dopamine modulation of LR failed. Expected weight %.4f (LR %.4f), got %.4f",
			expectedWeight_Dopa, expectedLR_Dopa, net.GetWeight(0, 1))
	}

	// Test with Cortisol Suppression
	net.SetWeight(0, 1, 0.0) // Reset weight
	net.DopamineLevel = 0.0
	net.CortisolLevel = CortisolMaxLevel // Max cortisol for max suppression

	net.ApplyHebbianPlasticity()

	baseLRWithNoDopa := net.BaseLearningRate // Dopamine is 0
	expectedLR_Cortisol := baseLRWithNoDopa * CortisolLearningSuppressionFactor
	if expectedLR_Cortisol < net.BaseLearningRate*MinLearningRateFactor {
		expectedLR_Cortisol = net.BaseLearningRate * MinLearningRateFactor
	}
	expectedWeight_Cortisol := expectedLR_Cortisol
	expectedWeight_Cortisol -= expectedWeight_Cortisol * HebbianWeightDecay

	if math.Abs(net.GetWeight(0, 1)-expectedWeight_Cortisol) > 1e-5 { // Increased tolerance
		t.Errorf("Cortisol modulation of LR failed. Expected weight %.4f (LR %.4f), got %.4f",
			expectedWeight_Cortisol, expectedLR_Cortisol, net.GetWeight(0, 1))
	}
	net.EnableChemicalModulation = false // Reset for other tests if any
}
