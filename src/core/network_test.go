package core

import (
	"testing"
	"math"
)

// Helper para comparar floats com tolerância, pode ser movido para um test_utils
func floatEquals(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestNewNetwork(t *testing.T) {
	config := DefaultNetworkConfig()
	config.NumNeurons = 50
	config.RandomSeed = 12345

	network := NewNetwork(config)

	if len(network.Neurons) != config.NumNeurons {
		t.Errorf("NewNetwork: expected %d neurons, got %d", config.NumNeurons, len(network.Neurons))
	}
	if network.Config.NumNeurons != config.NumNeurons {
		t.Errorf("NewNetwork: config mismatch on NumNeurons")
	}
	if network.CurrentCycle != 0 {
		t.Errorf("NewNetwork: CurrentCycle should be 0, got %d", network.CurrentCycle)
	}
	if network.CortisolLevel != 0.0 {
		t.Errorf("NewNetwork: CortisolLevel should be 0.0, got %f", network.CortisolLevel)
	}
	if network.DopamineLevel != 0.0 {
		t.Errorf("NewNetwork: DopamineLevel should be 0.0, got %f", network.DopamineLevel)
	}

	expectedGlandPos := [SpaceDimensions]float64{}
	for i := range expectedGlandPos {
		expectedGlandPos[i] = config.SpaceSize / 2.0
	}
	if !AreEqualVector(network.Gland.Position, expectedGlandPos) { // AreEqualVector de space_test.go
		t.Errorf("NewNetwork: Gland position incorrect. Expected %v, got %v", expectedGlandPos, network.Gland.Position)
	}

	if network.rng == nil {
		t.Errorf("NewNetwork: RNG not initialized")
	}
	if network.Config.RandomSeed != config.RandomSeed {
		t.Errorf("NewNetwork: RandomSeed in network.Config is %d, expected %d", network.Config.RandomSeed, config.RandomSeed)
	}
}

func TestNetwork_SimulateCycle_SimpleFiring(t *testing.T) {
	config := DefaultNetworkConfig()
	config.NumNeurons = 2
	config.SpaceSize = 10.0
	config.PulsePropagationSpeed = 1.0
	config.MaxCycles = 10
	config.RandomSeed = 1
	config.NeuronDistribution = map[NeuronType]float64{
		ExcitatoryNeuron: 1.0,
	}
	network := NewNetwork(config)

	if len(network.Neurons) < 2 {
		t.Fatalf("Need at least 2 neurons for this test, got %d", len(network.Neurons))
	}
	emitter := network.Neurons[0]
	receiver := network.Neurons[1]
	emitter.Position = [SpaceDimensions]float64{1,0}
	receiver.Position = [SpaceDimensions]float64{2,0}
	emitter.FiringThreshold = 0.5
	emitter.BaseFiringThreshold = 0.5
	receiver.FiringThreshold = 0.2
	receiver.BaseFiringThreshold = 0.2
	emitter.CurrentPotential = 1.0

	// Ciclo 1: Emitter dispara
	network.SimulateCycle()

	if emitter.State != FiringState {
		t.Errorf("Cycle 1: Emitter state expected FiringState, got %v", emitter.State)
	}
	if len(network.Pulses) != 1 {
		t.Fatalf("Cycle 1: Expected 1 pulse, got %d", len(network.Pulses))
	}
	pulse := network.Pulses[0]
	if pulse.OriginNeuronID != emitter.ID {
		t.Errorf("Cycle 1: Pulse origin ID mismatch")
	}
	if !floatEquals(pulse.Strength, 0.3, 1e-9) {
		t.Errorf("Cycle 1: Pulse strength expected 0.3, got %f", pulse.Strength)
	}

	// Ciclo 2: Pulso atinge receptor. Receptor dispara.
	network.SimulateCycle()

	if emitter.State != RefractoryAbsoluteState {
		t.Errorf("Cycle 2: Emitter state expected RefractoryAbsoluteState, got %v", emitter.State)
	}
	// Verificando se o receptor disparou (seu estado deve ser FiringState)
	if receiver.State != FiringState {
		t.Errorf("Cycle 2: Receiver state expected FiringState, got %v. Potential was %f, Threshold %f",
			receiver.State, receiver.CurrentPotential, receiver.FiringThreshold)
	}
	// Após o receptor disparar, seu potencial também é resetado.
	if receiver.State == FiringState && !floatEquals(receiver.CurrentPotential, 0.0, 1e-9) {
		t.Errorf("Cycle 2: Receiver potential expected ~0.0 after firing, got %f", receiver.CurrentPotential)
	}

	// O pulso original do emitter ainda deve estar ativo (CurrentRadius=2, MaxRadius=8)
	// E um novo pulso do receiver deve ter sido adicionado.
	if len(network.Pulses) != 2 {
		t.Errorf("Cycle 2: Expected 2 pulses (original still active, new from receiver), got %d. Pulses: %+v", len(network.Pulses), network.Pulses)
	}
}


func TestNetwork_SetInput_GetOutput(t *testing.T) {
	config := DefaultNetworkConfig()
	config.NumNeurons = 20
	// Garantir que a distribuição some 1.0 e tenha Input/Output neurons
	delete(config.NeuronDistribution, ExcitatoryNeuron) // Remover para recalcular
	delete(config.NeuronDistribution, InhibitoryNeuron)
	delete(config.NeuronDistribution, DopaminergicNeuron)
	config.NeuronDistribution[InputNeuron] = 0.5  // 10 neurônios de input
	config.NeuronDistribution[OutputNeuron] = 0.5 // 10 neurônios de output
	config.RandomSeed = 2
	network := NewNetwork(config)

	inputNeurons := network.GetNeuronsByType(InputNeuron)
	outputNeurons := network.GetNeuronsByType(OutputNeuron)

	if len(inputNeurons) == 0 {
		t.Fatalf("Test setup error: No InputNeurons. Counts: %v, Total: %d", neuronCountsForTest(network), len(network.Neurons))
	}
	if len(outputNeurons) == 0 {
		t.Fatalf("Test setup error: No OutputNeurons. Counts: %v, Total: %d", neuronCountsForTest(network), len(network.Neurons))
	}

	inputPattern := make([]float64, len(inputNeurons))
	for i := range inputPattern {
		inputPattern[i] = 0.5 * float64(i+1)
	}
	network.SetInput(inputPattern)

	for i, neuron := range inputNeurons {
		if i < len(inputPattern) {
			if !floatEquals(neuron.CurrentPotential, inputPattern[i], 1e-9) {
				t.Errorf("SetInput: Neuron %d (Input) potential expected %f, got %f", i, inputPattern[i], neuron.CurrentPotential)
			}
		}
	}

	expectedOutput := make([]float64, len(outputNeurons))
	for i, neuron := range outputNeurons {
		val := 0.25 * float64(i+1)
		neuron.CurrentPotential = val // Set manual para testar GetOutput isoladamente
		expectedOutput[i] = val
	}

	outputValues := network.GetOutput()
	if len(outputValues) != len(expectedOutput) {
		t.Fatalf("GetOutput: length mismatch. Expected %d, got %d", len(expectedOutput), len(outputValues))
	}
	for i, val := range outputValues {
		if !floatEquals(val, expectedOutput[i], 1e-9) {
			t.Errorf("GetOutput: Neuron %d (Output) value expected %f, got %f", i, expectedOutput[i], val)
		}
	}
}
// Helper para TestNetwork_SetInput_GetOutput
func neuronCountsForTest(net *Network) map[NeuronType]int {
    counts := make(map[NeuronType]int)
    for _, n := range net.Neurons {
        counts[n.Type]++
    }
    return counts
}


func TestNetwork_ResetNetworkState(t *testing.T) {
	config := DefaultNetworkConfig()
	config.NumNeurons = 5
	config.RandomSeed = 3
	network := NewNetwork(config)

	network.CurrentCycle = 10
	network.CortisolLevel = 0.5
	network.DopamineLevel = 0.3
	if len(network.Neurons) > 0 {
		network.Neurons[0].CurrentPotential = 1.0
		network.Neurons[0].State = FiringState
		network.Neurons[0].LastFiringCycle = 9
	}
	network.Pulses = append(network.Pulses, &Pulse{})

	network.ResetNetworkState()

	if network.CurrentCycle != 0 {
		t.Errorf("ResetNetworkState: CurrentCycle not reset. Got %d", network.CurrentCycle)
	}
	if len(network.Pulses) != 0 {
		t.Errorf("ResetNetworkState: Pulses not cleared. Got %d", len(network.Pulses))
	}
	for _, neuron := range network.Neurons {
		if neuron.CurrentPotential != 0.0 {
			t.Errorf("ResetNetworkState: Neuron %d potential not reset. Got %f", neuron.ID, neuron.CurrentPotential)
		}
		if neuron.State != RestingState {
			t.Errorf("ResetNetworkState: Neuron %d state not reset. Got %v", neuron.ID, neuron.State)
		}
		if neuron.LastFiringCycle != -1 {
			t.Errorf("ResetNetworkState: Neuron %d LastFiringCycle not reset. Got %d", neuron.ID, neuron.LastFiringCycle)
		}
		if !floatEquals(neuron.FiringThreshold, neuron.BaseFiringThreshold, 1e-9) {
			t.Errorf("ResetNetworkState: Neuron %d FiringThreshold not reset to base. Got %f, base %f", neuron.ID, neuron.FiringThreshold, neuron.BaseFiringThreshold)
		}
	}
}

func TestNetwork_NeurochemicalEffects(t *testing.T) {
	config := DefaultNetworkConfig()
	config.NumNeurons = 1
	config.RandomSeed = 4
	config.CortisolEffectOnThreshold = 0.1
	config.DopamineEffectOnThreshold = 0.2
	config.DopamineDecayRate = 0.05 // Explicit for test calculation
    config.CortisolDecayRate = 0.01 // Explicit for test calculation

	network := NewNetwork(config)
	if len(network.Neurons) == 0 { t.Fatal("No neurons in network") }
	neuron := network.Neurons[0]
	baseThreshold := neuron.BaseFiringThreshold

	// Teste 1: Apenas Dopamina
	network.DopamineLevel = 1.0
	network.CortisolLevel = 0.0
	// Antes de SimulateCycle, resetar o limiar do neurônio para base,
    // pois SimulateCycle aplica ajustes ao FiringThreshold atual.
    neuron.FiringThreshold = baseThreshold
	network.SimulateCycle()

	expectedDopamineAfterDecay := 1.0 * (1.0 - config.DopamineDecayRate)
	expectedThreshDopa := baseThreshold + (expectedDopamineAfterDecay * config.DopamineEffectOnThreshold)
	if !floatEquals(neuron.FiringThreshold, expectedThreshDopa, 1e-9) {
		t.Errorf("Dopamine effect: FiringThreshold expected %f, got %f. Dopamine level used for effect: %f",
			expectedThreshDopa, neuron.FiringThreshold, expectedDopamineAfterDecay)
	}

	// Teste 2: Cortisol Alto
	network.DopamineLevel = 0.0
	network.CortisolLevel = 1.5
	neuron.FiringThreshold = baseThreshold // Reset
	network.SimulateCycle()

	expectedCortisolHighAfterDecay := 1.5 * (1.0 - config.CortisolDecayRate)
	expectedThreshCortHigh := baseThreshold + (expectedCortisolHighAfterDecay - 1.0) * config.CortisolEffectOnThreshold
	if !floatEquals(neuron.FiringThreshold, expectedThreshCortHigh, 1e-9) {
		t.Errorf("High Cortisol effect: FiringThreshold expected %f, got %f. Cortisol level used for effect: %f",
			expectedThreshCortHigh, neuron.FiringThreshold, expectedCortisolHighAfterDecay)
	}

	// Teste 3: Cortisol Moderado
	network.DopamineLevel = 0.0
	network.CortisolLevel = 0.5
	neuron.FiringThreshold = baseThreshold // Reset
	network.SimulateCycle()

	expectedCortisolModAfterDecay := 0.5 * (1.0 - config.CortisolDecayRate)
	expectedThreshCortMod := baseThreshold - (expectedCortisolModAfterDecay * config.CortisolEffectOnThreshold * 0.5)
	if !floatEquals(neuron.FiringThreshold, expectedThreshCortMod, 1e-9) {
		t.Errorf("Moderate Cortisol effect: FiringThreshold expected %f, got %f. Cortisol level used for effect: %f",
			expectedThreshCortMod, neuron.FiringThreshold, expectedCortisolModAfterDecay)
	}
}
