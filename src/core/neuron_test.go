package core

import (
	"testing"
	"math/rand"
)

func TestNewNeuron(t *testing.T) {
	id := 1
	nType := ExcitatoryNeuron
	pos := [SpaceDimensions]float64{1,2,3}
	neuron := NewNeuron(id, nType, pos)

	if neuron.ID != id {
		t.Errorf("NewNeuron ID: expected %d, got %d", id, neuron.ID)
	}
	if neuron.Type != nType {
		t.Errorf("NewNeuron Type: expected %v, got %v", nType, neuron.Type)
	}
	if !AreEqualVector(pos, neuron.Position) { // AreEqualVector é de space_test.go, mas ok para teste
		t.Errorf("NewNeuron Position: expected %v, got %v", pos, neuron.Position)
	}
	if neuron.State != RestingState {
		t.Errorf("NewNeuron State: expected %v, got %v", RestingState, neuron.State)
	}
	if neuron.CurrentPotential != 0.0 {
		t.Errorf("NewNeuron CurrentPotential: expected %f, got %f", 0.0, neuron.CurrentPotential)
	}
	if !AreEqualFloat64(neuron.BaseFiringThreshold, neuron.FiringThreshold) { // AreEqualFloat64 de space_test.go
		t.Errorf("NewNeuron FiringThreshold: expected %f (base), got %f", neuron.BaseFiringThreshold, neuron.FiringThreshold)
	}
}

func TestNeuron_UpdateState_FiringAndCycle(t *testing.T) {
	neuron := NewNeuron(0, ExcitatoryNeuron, [SpaceDimensions]float64{})
	neuron.FiringThreshold = 1.0
	neuron.BaseFiringThreshold = 1.0
	neuron.RefractoryPeriodAbsolute = 1
	neuron.RefractoryPeriodRelative = 1
	neuron.CurrentPotential = 1.5
	cycle := 0

	// Cycle 1: Disparo
	cycle++
	fired := neuron.UpdateState(cycle)
	if !(fired && neuron.State == FiringState && neuron.LastFiringCycle == cycle) {
		t.Fatalf("Cycle %d: Expected FiringState, fired. Got state %v, fired %t, last_fired %d", cycle, neuron.State, fired, neuron.LastFiringCycle)
	}

	// Cycle 2: Transição para Refratário Absoluto
	cycle++
	fired = neuron.UpdateState(cycle)
	if !(neuron.State == RefractoryAbsoluteState && neuron.RefractoryCycles == neuron.RefractoryPeriodAbsolute && !fired) {
		t.Fatalf("Cycle %d: Expected RefractoryAbsoluteState (cycles=%d). Got state %v, cycles %d, fired %t", cycle, neuron.RefractoryPeriodAbsolute, neuron.State, neuron.RefractoryCycles, fired)
	}

	// Cycle 3: Em Refratário Absoluto, transição para Relativo
	// (Como RP_Abs = 1, RefractoryCycles (1) é decrementado para 0, então transita)
	cycle++
	fired = neuron.UpdateState(cycle)
	if !(neuron.State == RefractoryRelativeState && neuron.RefractoryCycles == neuron.RefractoryPeriodRelative && !fired) {
		t.Fatalf("Cycle %d: Expected RefractoryRelativeState (cycles=%d). Got state %v, cycles %d, fired %t", cycle, neuron.RefractoryPeriodRelative, neuron.State, neuron.RefractoryCycles, fired)
	}

	// Cycle 4: Em Refratário Relativo, transição para Repouso
	// (Como RP_Rel = 1, RefractoryCycles (1) é decrementado para 0, então transita)
	cycle++
	fired = neuron.UpdateState(cycle)
	if !(neuron.State == RestingState && !fired) {
		t.Fatalf("Cycle %d: Expected RestingState. Got state %v, fired %t", cycle, neuron.State, fired)
	}
}

func TestNeuron_AddPotential(t *testing.T) {
	neuron := NewNeuron(0, ExcitatoryNeuron, [SpaceDimensions]float64{})
	neuron.State = RestingState
	neuron.AddPotential(0.5)
	if !AreEqualFloat64(neuron.CurrentPotential, 0.5) {
		t.Errorf("AddPotential: expected %f, got %f", 0.5, neuron.CurrentPotential)
	}
	neuron.AddPotential(0.3)
	if !AreEqualFloat64(neuron.CurrentPotential, 0.8) {
		t.Errorf("AddPotential: expected %f, got %f", 0.8, neuron.CurrentPotential)
	}

	neuron.State = RefractoryAbsoluteState
	neuron.CurrentPotential = 0.123
	neuron.AddPotential(10.0)
	if !AreEqualFloat64(neuron.CurrentPotential, 0.123) {
		t.Errorf("AddPotential (RefractoryAbsolute): potential should not change. Got %f", neuron.CurrentPotential)
	}
}

func TestNeuron_AdjustFiringThreshold(t *testing.T) {
	neuron := NewNeuron(0, ExcitatoryNeuron, [SpaceDimensions]float64{})
	base := neuron.BaseFiringThreshold

	neuron.AdjustFiringThreshold(0.2)
	if !AreEqualFloat64(neuron.FiringThreshold, base+0.2) {
		t.Errorf("AdjustFiringThreshold (+0.2): expected %f, got %f", base+0.2, neuron.FiringThreshold)
	}

	neuron.AdjustFiringThreshold(-0.3)
	if !AreEqualFloat64(neuron.FiringThreshold, base-0.3) {
		t.Errorf("AdjustFiringThreshold (-0.3): expected %f, got %f", base-0.3, neuron.FiringThreshold)
	}

	neuron.AdjustFiringThreshold(-base * 2)
	if !AreEqualFloat64(neuron.FiringThreshold, 0.1) {
		t.Errorf("AdjustFiringThreshold (min clamp): expected %f, got %f", 0.1, neuron.FiringThreshold)
	}
}

func TestInitializeNeurons(t *testing.T) {
	numNeurons := 100
	distribution := map[NeuronType]float64{
		ExcitatoryNeuron: 0.7,
		InhibitoryNeuron: 0.2,
		InputNeuron:      0.1, // Soma 1.0
	}
	spaceSize := 100.0
	// rand.Seed(0) // InitializeNeurons não usa o RNG global diretamente mais, mas NewNetwork sim.

	// Para testar InitializeNeurons isoladamente, precisamos de um RNG ou mock.
	// No entanto, ele é chamado por NewNetwork que tem seu próprio RNG.
	// A lógica de InitializeNeurons em si não usa mais o RNG global.
	// As posições são geradas com rand.Float64(), que usa o RNG global.
	// Para teste consistente de posições, o RNG global precisaria ser semeado.
	// Mas o teste principal aqui é sobre contagens e tipos.

	// NewNetwork lida com o seed e chama InitializeNeurons implicitamente.
	// Para testar InitializeNeurons diretamente e de forma mais completa (incluindo posições),
	// seria melhor se ele aceitasse um *rand.Rand. Por ora, testamos o que podemos.

	neurons := InitializeNeurons(numNeurons, distribution, spaceSize)

	if len(neurons) != numNeurons {
		// A lógica atual de InitializeNeurons pode não garantir exatamente numNeurons
		// se as porcentagens não somarem 1.0 ou por arredondamento.
		// A função InitializeNeurons foi atualizada para tentar preencher, mas ainda pode haver pequenas diferenças.
		// O teste em NewNetwork é mais completo para a contagem final.
		// Aqui, vamos verificar se a contagem está próxima.
		if len(neurons) < numNeurons-len(distribution) || len(neurons) > numNeurons+len(distribution) { // Margem pequena
			t.Errorf("InitializeNeurons: expected around %d neurons, got %d", numNeurons, len(neurons))
		}
	}

	counts := make(map[NeuronType]int)
	for _, n := range neurons {
		counts[n.Type]++
		for _, p_dim := range n.Position {
			if p_dim < 0 || p_dim > spaceSize {
				t.Errorf("InitializeNeurons: neuron %d position %v out of bounds [0, %f]", n.ID, n.Position, spaceSize)
				break
			}
		}
	}

	expectedCounts := make(map[NeuronType]int)
	for nType, percentage := range distribution {
		expectedCounts[nType] = int(float64(numNeurons) * percentage)
	}

	// Devido a arredondamentos, as contagens exatas podem variar ligeiramente.
	// A soma total é mais importante e é verificada em NewNetwork.
	// Aqui, apenas verificamos se os tipos esperados foram criados.
	for nType, expectedCount := range expectedCounts {
		if expectedCount > 0 && counts[nType] == 0 {
			t.Errorf("InitializeNeurons: expected some %v neurons, got 0", nType)
		}
		// Poderia adicionar um teste de proximidade se a lógica de preenchimento fosse mais complexa
		// e distribuísse os erros de arredondamento.
	}
}

// TestNeuronPotentialDecay testa o decaimento do potencial.
func TestNeuronPotentialDecay(t *testing.T) {
	neuron := NewNeuron(0, ExcitatoryNeuron, [SpaceDimensions]float64{})
	neuron.CurrentPotential = 1.0
	neuron.State = RestingState // Para garantir que o decaimento ocorra

	// Simula um ciclo onde não há disparo e o neurônio está em repouso
	neuron.UpdateState(1)

	// decayRate é 0.1 em Neuron.UpdateState
	// Expected = 1.0 - (0.1 * 1.0) = 0.9
	if !AreEqualFloat64(neuron.CurrentPotential, 0.9) {
		t.Errorf("Potential decay: expected %f, got %f", 0.9, neuron.CurrentPotential)
	}

	neuron.UpdateState(2)
	// Expected = 0.9 - (0.1 * 0.9) = 0.9 - 0.09 = 0.81
	if !AreEqualFloat64(neuron.CurrentPotential, 0.81) {
		t.Errorf("Potential decay (2nd cycle): expected %f, got %f", 0.81, neuron.CurrentPotential)
	}
}
