package pulse_test

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/pulse"
	"crownet/space" // Para EuclideanDistance, se necessário em mocks ou setups
	"crownet/synaptic"
	"testing"
	"math" // Para comparações de float
)

// Helper para comparar floats com tolerância
func floatEquals(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) < tolerance
}

// --- Testes para Pulse ---

func TestNewPulse(t *testing.T) {
	emitterID := common.NeuronID(1)
	origin := common.Point{1, 2, 3}
	signal := common.PulseValue(1.0)
	creationCycle := common.CycleCount(10)
	maxRadius := 100.0

	p := pulse.New(emitterID, origin, signal, creationCycle, maxRadius)

	if p.EmittingNeuronID != emitterID {
		t.Errorf("EmittingNeuronID: expected %d, got %d", emitterID, p.EmittingNeuronID)
	}
	if p.OriginPosition != origin {
		t.Errorf("OriginPosition: expected %v, got %v", origin, p.OriginPosition)
	}
	if p.BaseSignalValue != signal {
		t.Errorf("BaseSignalValue: expected %f, got %f", signal, p.BaseSignalValue)
	}
	if p.CreationCycle != creationCycle {
		t.Errorf("CreationCycle: expected %d, got %d", creationCycle, p.CreationCycle)
	}
	if p.CurrentDistance != 0.0 {
		t.Errorf("CurrentDistance: expected 0.0, got %f", p.CurrentDistance)
	}
	if p.MaxTravelRadius != maxRadius {
		t.Errorf("MaxTravelRadius: expected %f, got %f", maxRadius, p.MaxTravelRadius)
	}
}

func TestPulse_Propagate(t *testing.T) {
	simParams := config.SimulationParameters{PulsePropagationSpeed: 0.5}
	p := pulse.New(0, common.Point{}, 1.0, 0, 1.0) // MaxTravelRadius = 1.0

	// Propagação 1: Ainda ativo
	isActive := p.Propagate(&simParams)
	if !isActive {
		t.Errorf("Expected pulse to be active after first propagation")
	}
	if !floatEquals(p.CurrentDistance, 0.5, 1e-9) {
		t.Errorf("CurrentDistance after first propagation: expected 0.5, got %f", p.CurrentDistance)
	}

	// Propagação 2: Ainda ativo (na borda)
	isActive = p.Propagate(&simParams)
	if !isActive {
		t.Errorf("Expected pulse to be active at the edge of MaxTravelRadius")
	}
	if !floatEquals(p.CurrentDistance, 1.0, 1e-9) {
		t.Errorf("CurrentDistance at edge: expected 1.0, got %f", p.CurrentDistance)
	}

	// Propagação 3: Deve se dissipar
	isActive = p.Propagate(&simParams)
	if isActive {
		t.Errorf("Expected pulse to be inactive after exceeding MaxTravelRadius")
	}
	if !floatEquals(p.CurrentDistance, 1.5, 1e-9) {
		t.Errorf("CurrentDistance after dissipation: expected 1.5, got %f", p.CurrentDistance)
	}
}

func TestPulse_GetEffectShellForCycle(t *testing.T) {
	simParams := config.SimulationParameters{PulsePropagationSpeed: 0.5}
	p := pulse.New(0, common.Point{}, 1.0, 0, 10.0)

	// Primeiro ciclo de propagação
	p.Propagate(&simParams) // CurrentDistance = 0.5
	start, end := p.GetEffectShellForCycle(&simParams)
	if !floatEquals(start, 0.0, 1e-9) || !floatEquals(end, 0.5, 1e-9) {
		t.Errorf("Shell 1: expected {0.0, 0.5}, got {%f, %f}", start, end)
	}

	// Segundo ciclo de propagação
	p.Propagate(&simParams) // CurrentDistance = 1.0
	start, end = p.GetEffectShellForCycle(&simParams)
	if !floatEquals(start, 0.5, 1e-9) || !floatEquals(end, 1.0, 1e-9) {
		t.Errorf("Shell 2: expected {0.5, 1.0}, got {%f, %f}", start, end)
	}
}

// --- Testes para PulseList ---

func TestNewPulseList(t *testing.T) {
	pl := pulse.NewPulseList()
	if pl == nil {
		t.Fatalf("NewPulseList returned nil")
	}
	if pl.Count() != 0 {
		t.Errorf("New PulseList should have count 0, got %d", pl.Count())
	}
	if len(pl.GetAll()) != 0 {
		t.Errorf("New PulseList's GetAll should return empty slice, got %d elements", len(pl.GetAll()))
	}
}

func TestPulseList_Add_AddAll_Clear_Count_GetAll(t *testing.T) {
	pl := pulse.NewPulseList()
	p1 := pulse.New(1, common.Point{}, 1.0, 0, 10.0)
	p2 := pulse.New(2, common.Point{}, 1.0, 0, 10.0)
	p3 := pulse.New(3, common.Point{}, 1.0, 0, 10.0)

	// Add
	pl.Add(p1)
	if pl.Count() != 1 {
		t.Errorf("Count after Add(p1): expected 1, got %d", pl.Count())
	}
	if pl.GetAll()[0] != p1 {
		t.Errorf("GetAll after Add(p1): did not find p1")
	}

	// AddAll
	pl.AddAll([]*pulse.Pulse{p2, p3})
	if pl.Count() != 3 {
		t.Errorf("Count after AddAll(p2,p3): expected 3, got %d", pl.Count())
	}
	allPulses := pl.GetAll()
	if len(allPulses) != 3 || allPulses[0] != p1 || allPulses[1] != p2 || allPulses[2] != p3 {
		t.Errorf("GetAll after AddAll: content mismatch")
	}

	// Clear
	pl.Clear()
	if pl.Count() != 0 {
		t.Errorf("Count after Clear: expected 0, got %d", pl.Count())
	}
	if len(pl.GetAll()) != 0 {
		t.Errorf("GetAll after Clear: expected empty slice, got %d elements", len(pl.GetAll()))
	}
}

// --- Testes para PulseList.ProcessCycle ---

func TestPulseList_ProcessCycle_PropagationAndRemoval(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.PulsePropagationSpeed = 0.5

	pl := pulse.NewPulseList()
	pActive := pulse.New(1, common.Point{0,0}, 1.0, 0, 1.0) // MaxRadius 1.0, Speed 0.5. Ativo após 1 ciclo, dissipado após 3.
	pDissipate := pulse.New(2, common.Point{0,0}, 1.0, 0, 0.2) // MaxRadius 0.2, dissipará na primeira propagação (dist 0.5)

	pl.Add(pActive)
	pl.Add(pDissipate)

	// Teste sem neurônios, focando apenas na propagação e remoção de pulsos
	testNeurons := []*neuron.Neuron{}
	testWeights := synaptic.NewNetworkWeights()

	newPulses := pl.ProcessCycle(testNeurons, testWeights, common.CycleCount(1), &simParams)

	if len(newPulses) != 0 {
		t.Errorf("Expected 0 new pulses when no neurons are present, got %d", len(newPulses))
	}
	if pl.Count() != 1 {
		t.Errorf("Expected 1 active pulse remaining after one cycle, got %d. pActive dist: %f, pDissipate dist: %f",
			pl.Count(), pActive.CurrentDistance, pDissipate.CurrentDistance)
	}
	if pl.Count() > 0 && pl.GetAll()[0].EmittingNeuronID != pActive.EmittingNeuronID {
		t.Errorf("The wrong pulse remained active.")
	}
	if pl.Count() > 0 && !floatEquals(pl.GetAll()[0].CurrentDistance, 0.5, 1e-9) {
		t.Errorf("Active pulse distance: expected 0.5, got %f", pl.GetAll()[0].CurrentDistance)
	}

	// Propagar pActive novamente, deve continuar ativo
	pl.ProcessCycle(testNeurons, testWeights, common.CycleCount(2), &simParams) // pActive distance = 1.0
	if pl.Count() != 1 {
		t.Errorf("Expected pActive to still be active after 2nd propagation, got count %d", pl.Count())
	}
    if pl.Count() > 0 && !floatEquals(pl.GetAll()[0].CurrentDistance, 1.0, 1e-9) {
		t.Errorf("pActive distance after 2nd propagation: expected 1.0, got %f", pl.GetAll()[0].CurrentDistance)
	}

	// Propagar pActive mais uma vez, agora deve dissipar
	pl.ProcessCycle(testNeurons, testWeights, common.CycleCount(3), &simParams) // pActive distance = 1.5
	if pl.Count() != 0 {
		t.Errorf("Expected pActive to dissipate after 3rd propagation, got count %d. Distance: %f", pl.Count(), pActive.CurrentDistance)
	}
}


func TestPulseList_ProcessCycle_EffectAndNewPulseGeneration(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.PulsePropagationSpeed = 0.6
	simParams.BaseFiringThreshold = 0.9
	simParams.SpaceMaxDimension = 10.0
	// Garantir que AbsoluteRefractoryCycles seja realisticamente > 0 para evitar disparos múltiplos no mesmo ciclo se não tratado
	simParams.AbsoluteRefractoryCycles = 2 // common.CycleCount

	emitterNeuron := neuron.New(1, neuron.Excitatory, common.Point{0,0,0}, &simParams)

	// Neurônio Alvo - posicionado para ser afetado e disparar
	// Usar neuron.New diretamente, pois o MockNeuron não intercepta chamadas como esperado sem interfaces.
	targetNeuronAffected := neuron.New(2, neuron.Excitatory, common.Point{0.5, 0,0}, &simParams) // Distância 0.5

	// Neurônio Alvo - posicionado para NÃO ser afetado (longe demais)
	targetNeuronFar := neuron.New(3, neuron.Excitatory, common.Point{5,0,0}, &simParams) // Distância 5.0

	// Neurônio Alvo - posicionado para ser afetado mas não disparar (potencial insuficiente)
	targetNeuronNoFire := neuron.New(4, neuron.Excitatory, common.Point{0.4,0,0}, &simParams) // Distância 0.4

	// Neurônio Alvo - para testar peso zero
	targetNeuronZeroWeight := neuron.New(5, neuron.Excitatory, common.Point{0.3,0,0}, &simParams) // Distância 0.3


	neuronsForCycle := []*neuron.Neuron{
		emitterNeuron,
		targetNeuronAffected,
		targetNeuronFar,
		targetNeuronNoFire,
		targetNeuronZeroWeight,
	}

	weights := synaptic.NewNetworkWeights()
	// Peso forte para targetNeuronAffected para garantir disparo
	weights.SetWeight(emitterNeuron.ID, targetNeuronAffected.ID, 1.0)
	// Peso para targetNeuronFar (não deve ser alcançado)
	weights.SetWeight(emitterNeuron.ID, targetNeuronFar.ID, 1.0)
	// Peso fraco para targetNeuronNoFire
	weights.SetWeight(emitterNeuron.ID, targetNeuronNoFire.ID, 0.5) // Sinal 1.0 * Peso 0.5 = 0.5 (abaixo do limiar 0.9)
	// Peso zero para targetNeuronZeroWeight
	weights.SetWeight(emitterNeuron.ID, targetNeuronZeroWeight.ID, 0.0)


	pl := pulse.NewPulseList()
	// Pulso se origina em emitterNeuron (0,0,0), viaja 0.6 no primeiro ciclo. Casca de efeito [0, 0.6)
	initialPulse := pulse.New(emitterNeuron.ID, emitterNeuron.Position, emitterNeuron.EmittedPulseSign(), 0, 5.0)
	pl.Add(initialPulse)

	currentSimCycle := common.CycleCount(1)
	newlyGenerated := pl.ProcessCycle(neuronsForCycle, weights, currentSimCycle, &simParams)

	// 1. Verificar targetNeuronAffected (deve ter sido afetado e disparado)
	if targetNeuronAffected.CurrentState != neuron.Firing {
		t.Errorf("targetNeuronAffected: expected state Firing, got %s. Potential: %f", targetNeuronAffected.CurrentState, targetNeuronAffected.AccumulatedPotential)
	}
	foundNewPulseFromAffected := false
	for _, p := range newlyGenerated {
		if p.EmittingNeuronID == targetNeuronAffected.ID {
			foundNewPulseFromAffected = true
			break
		}
	}
	if !foundNewPulseFromAffected {
		t.Errorf("Expected a new pulse from targetNeuronAffected, but none found. Generated: %d", len(newlyGenerated))
	}


	// 2. Verificar targetNeuronFar (não deve ter sido afetado)
	if targetNeuronFar.AccumulatedPotential != 0 {
		t.Errorf("targetNeuronFar potential: expected 0, got %f", targetNeuronFar.AccumulatedPotential)
	}
	if targetNeuronFar.CurrentState == neuron.Firing {
		t.Errorf("targetNeuronFar should not have fired.")
	}

	// 3. Verificar targetNeuronNoFire (afetado, mas não disparou)
	if !floatEquals(float64(targetNeuronNoFire.AccumulatedPotential), 0.5, 1e-9) { // Sinal 1.0 * Peso 0.5
		t.Errorf("targetNeuronNoFire potential: expected 0.5, got %f", targetNeuronNoFire.AccumulatedPotential)
	}
	if targetNeuronNoFire.CurrentState == neuron.Firing {
		t.Errorf("targetNeuronNoFire should not have fired.")
	}

	// 4. Verificar targetNeuronZeroWeight (afetado, mas potencial efetivo zero, não disparou)
	if targetNeuronZeroWeight.AccumulatedPotential != 0 {
		t.Errorf("targetNeuronZeroWeight potential: expected 0, got %f", targetNeuronZeroWeight.AccumulatedPotential)
	}
	if targetNeuronZeroWeight.CurrentState == neuron.Firing {
		t.Errorf("targetNeuronZeroWeight should not have fired.")
	}

	// 5. Verificar número total de pulsos gerados
	// Apenas targetNeuronAffected deve ter disparado e gerado um pulso.
	if len(newlyGenerated) != 1 {
		t.Errorf("Expected 1 new pulse to be generated in total, got %d", len(newlyGenerated))
	}
}


// Mais testes podem ser adicionados para cobrir:
// - Múltiplos pulsos afetando o mesmo neurônio
// - Pulsos inibitórios
// - Pesos sinápticos zero ou negativos (além do que já é coberto pela lógica de effectivePotential == 0)
// - Efeitos em neurônios em diferentes estados refratários (o mock simplificado não lida com isso bem,
//   mas o neurônio real sim, então testar com neurônios reais é importante).
```
