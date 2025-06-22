package synaptic_test

import (
	"crownet/common"
	"crownet/config"
	"crownet/synaptic"
	"math"
	"math/rand"
	"testing"
)

// Helper para comparar floats com tolerância
func floatEquals(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) < tolerance
}

func TestNewNetworkWeights(t *testing.T) {
	nw := synaptic.NewNetworkWeights()
	if nw == nil {
		t.Fatalf("NewNetworkWeights returned nil, expected a valid map.")
	}
	if len(nw) != 0 {
		t.Errorf("NewNetworkWeights should be empty, got length %d", len(nw))
	}
}

func TestInitializeAllToAllWeights(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.InitialSynapticWeightMin = 0.1
	simParams.InitialSynapticWeightMax = 0.5

	neuronIDs := []common.NeuronID{1, 2, 3}
	nw := synaptic.NewNetworkWeights()
	nw.InitializeAllToAllWeights(neuronIDs, &simParams)

	if len(nw) != len(neuronIDs) {
		t.Errorf("Expected %d entries in NetworkWeights, got %d", len(neuronIDs), len(nw))
	}

	for _, fromID := range neuronIDs {
		if _, ok := nw[fromID]; !ok {
			t.Errorf("Missing WeightMap for fromID %d", fromID)
			continue
		}
		if len(nw[fromID]) != len(neuronIDs) {
			t.Errorf("WeightMap for fromID %d should have %d entries, got %d", fromID, len(neuronIDs), len(nw[fromID]))
		}
		for _, toID := range neuronIDs {
			weight, exists := nw[fromID][toID]
			if !exists {
				t.Errorf("Missing weight from %d to %d", fromID, toID)
				continue
			}
			if fromID == toID {
				if weight != 0.0 {
					t.Errorf("Self-connection weight from %d to %d should be 0.0, got %f", fromID, toID, weight)
				}
			} else {
				if float64(weight) < simParams.InitialSynapticWeightMin || float64(weight) > simParams.InitialSynapticWeightMax {
					t.Errorf("Weight %f from %d to %d is outside range [%f, %f]",
						weight, fromID, toID, simParams.InitialSynapticWeightMin, simParams.InitialSynapticWeightMax)
				}
			}
		}
	}

	// Teste com minW > maxW (deve usar fallback)
	simParamsFallback := config.DefaultSimulationParameters()
	simParamsFallback.InitialSynapticWeightMin = 0.8
	simParamsFallback.InitialSynapticWeightMax = 0.2 // Invertido
	nwFallback := synaptic.NewNetworkWeights()
	nwFallback.InitializeAllToAllWeights(neuronIDs, &simParamsFallback)
	// Fallback é para [0.01, 0.05]
	for _, fromID := range neuronIDs {
		for _, toID := range neuronIDs {
			if fromID != toID {
				weight := nwFallback.GetWeight(fromID, toID)
				if float64(weight) < 0.01 || float64(weight) > 0.05 {
					if !(float64(weight) == 0.01 && 0.01 == 0.05) { // Handle case where min=max
						t.Errorf("Fallback weight %f from %d to %d is outside range [0.01, 0.05]", weight, fromID, toID)
					}
				}
			}
		}
	}
}

func TestGetAndSetWeight(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.MaxSynapticWeight = 1.0
	// InitialSynapticWeightMin é usado como o limite inferior de clamping em SetWeight na implementação atual,
	// mas a refatoração mudou para 0.0.
	// Vamos testar com 0.0 como limite inferior.

	nw := synaptic.NewNetworkWeights()
	id1, id2 := common.NeuronID(1), common.NeuronID(2)

	// 1. Get peso não existente
	if w := nw.GetWeight(id1, id2); w != 0.0 {
		t.Errorf("Expected 0.0 for non-existent weight, got %f", w)
	}

	// 2. Set peso normal
	nw.SetWeight(id1, id2, 0.5, &simParams)
	if w := nw.GetWeight(id1, id2); w != 0.5 {
		t.Errorf("Expected 0.5 after SetWeight, got %f", w)
	}

	// 3. Set peso acima do MaxSynapticWeight (deve ser clampeado)
	nw.SetWeight(id1, id2, 1.5, &simParams)
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), simParams.MaxSynapticWeight, 1e-9) {
		t.Errorf("Expected weight to be clamped to MaxSynapticWeight (%f), got %f", simParams.MaxSynapticWeight, w)
	}

	// 4. Set peso abaixo de 0.0 (deve ser clampeado para 0.0)
	nw.SetWeight(id1, id2, -0.5, &simParams)
	// O limite inferior em SetWeight é 0.0
	expectedMinClamp := 0.0
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedMinClamp, 1e-9) {
		t.Errorf("Expected weight to be clamped to %f, got %f", expectedMinClamp, w)
	}

	// 5. Set peso para um novo par de IDs
	id3, id4 := common.NeuronID(3), common.NeuronID(4)
	nw.SetWeight(id3, id4, 0.7, &simParams)
	if w := nw.GetWeight(id3, id4); w != 0.7 {
		t.Errorf("Expected 0.7 for new ID pair, got %f", w)
	}
	// Verificar se o mapa interno para id3 foi criado
	if _, ok := nw[id3]; !ok {
		t.Errorf("WeightMap for id3 was not created after SetWeight")
	}
}


func TestApplyHebbianUpdate(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.MaxSynapticWeight = 1.0
	simParams.SynapticWeightDecayRate = 0.01 // 1% decay
	simParams.HebbPositiveReinforceFactor = 1.0 // Simplificar: delta = LR * pre * post

	nw := synaptic.NewNetworkWeights()
	id1, id2 := common.NeuronID(1), common.NeuronID(2)
	learningRate := common.Rate(0.1)

	// Caso 1: LTP (Long-Term Potentiation)
	// Peso inicial 0.2. pre=1, post=1. LR=0.1. HebbFactor=1.0
	// delta = 0.1 * 1 * 1 * 1.0 = 0.1
	// newWeightBeforeDecay = 0.2 + 0.1 = 0.3
	// newWeightAfterDecay = 0.3 * (1 - 0.01) = 0.3 * 0.99 = 0.297
	nw.SetWeight(id1, id2, 0.2, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 1.0, 1.0, learningRate, &simParams)
	expectedWeight1 := 0.297
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight1, 1e-9) {
		t.Errorf("LTP: Expected weight %f, got %f", expectedWeight1, w)
	}

	// Caso 2: Sem co-atividade (apenas decaimento)
	// Peso inicial 0.5. pre=0, post=1 (ou pre=1, post=0; ou pre=0, post=0)
	// delta = 0.1 * 0 * 1 * 1.0 = 0.0
	// newWeightBeforeDecay = 0.5 + 0.0 = 0.5
	// newWeightAfterDecay = 0.5 * (1 - 0.01) = 0.5 * 0.99 = 0.495
	nw.SetWeight(id1, id2, 0.5, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 0.0, 1.0, learningRate, &simParams)
	expectedWeight2 := 0.495
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight2, 1e-9) {
		t.Errorf("No co-activity (decay): Expected weight %f, got %f", expectedWeight2, w)
	}

	// Caso 3: LTP leva ao clamping no máximo
	// Peso inicial 0.95. pre=1, post=1. LR=0.1. HebbFactor=1.0
	// delta = 0.1 * 1 * 1 * 1.0 = 0.1
	// newWeightBeforeDecay = 0.95 + 0.1 = 1.05
	// newWeightAfterDecay = 1.05 * (1 - 0.01) = 1.05 * 0.99 = 1.0395
	// Clampeado para MaxSynapticWeight = 1.0
	nw.SetWeight(id1, id2, 0.95, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 1.0, 1.0, learningRate, &simParams)
	expectedWeight3 := simParams.MaxSynapticWeight // 1.0
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight3, 1e-9) {
		t.Errorf("LTP with clamping: Expected weight %f, got %f", expectedWeight3, w)
	}

	// Caso 4: Decaimento leva ao clamping em zero
	// Peso inicial muito pequeno, ex: 0.00001. pre=0, post=0. LR=0.1
	// delta = 0
	// newWeightBeforeDecay = 0.00001
	// newWeightAfterDecay = 0.00001 * (1-0.01) = 0.0000099
	// Não deve ser clampeado para zero A MENOS que a lógica de SetWeight o faça para valores muito pequenos.
	// SetWeight atualmente clampa em 0.0 se o valor for < 0.0.
	// A lógica de decaimento é newWeight * (1 - decayRate). Se newWeight é positivo, permanece positivo.
	// Se o peso se torna muito pequeno (ex: < 1e-9), ele poderia ser arredondado/considerado zero, mas a lógica não faz isso.
	// Vamos testar o decaimento normal de um peso pequeno.
	smallPositiveWeight := common.SynapticWeight(0.001)
	nw.SetWeight(id1, id2, smallPositiveWeight, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 0.0, 0.0, learningRate, &simParams) // No Hebbian change
	expectedWeight4 := float64(smallPositiveWeight) * (1.0 - simParams.SynapticWeightDecayRate) // 0.001 * 0.99 = 0.00099
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight4, 1e-9) {
		t.Errorf("Decay of small positive weight: Expected %f, got %f", expectedWeight4, w)
	}


	// Caso 5: Auto-conexão não deve ser afetada
	nw.SetWeight(id1, id1, 0.5, &simParams) // Definir um peso de auto-conexão (embora Initialize não o faça)
	nw.ApplyHebbianUpdate(id1, id1, 1.0, 1.0, learningRate, &simParams)
	if w := nw.GetWeight(id1, id1); w != 0.5 { // ApplyHebbianUpdate tem um `if fromID == toID { return }`
		t.Errorf("Self-connection weight should not be affected by ApplyHebbianUpdate, expected 0.5, got %f", w)
	}
}

func TestMain(m *testing.M) {
	// Seed rand para testes determinísticos, especialmente para InitializeAllToAllWeights
	rand.Seed(12345)
	m.Run()
}
```
