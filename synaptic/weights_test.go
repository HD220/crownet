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
	rng := rand.New(rand.NewSource(1)) // Create a local rng instance
	nw.InitializeAllToAllWeights(neuronIDs, &simParams, rng)

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

	simParamsFallback := config.DefaultSimulationParameters()
	simParamsFallback.InitialSynapticWeightMin = 0.8
	simParamsFallback.InitialSynapticWeightMax = 0.2
	nwFallback := synaptic.NewNetworkWeights()
	rngFallback := rand.New(rand.NewSource(2)) // Create another local rng instance
	nwFallback.InitializeAllToAllWeights(neuronIDs, &simParamsFallback, rngFallback)
	for _, fromID := range neuronIDs {
		for _, toID := range neuronIDs {
			if fromID != toID {
				weight := nwFallback.GetWeight(fromID, toID)
				if float64(weight) < 0.01 || float64(weight) > 0.05 {
					if !(float64(weight) == 0.01 && 0.01 == 0.05) {
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

	nw := synaptic.NewNetworkWeights()
	id1, id2 := common.NeuronID(1), common.NeuronID(2)

	if w := nw.GetWeight(id1, id2); w != 0.0 {
		t.Errorf("Expected 0.0 for non-existent weight, got %f", w)
	}

	nw.SetWeight(id1, id2, 0.5, &simParams)
	if w := nw.GetWeight(id1, id2); w != 0.5 {
		t.Errorf("Expected 0.5 after SetWeight, got %f", w)
	}

	nw.SetWeight(id1, id2, 1.5, &simParams)
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), simParams.MaxSynapticWeight, 1e-9) {
		t.Errorf("Expected weight to be clamped to MaxSynapticWeight (%f), got %f", simParams.MaxSynapticWeight, w)
	}

	nw.SetWeight(id1, id2, -0.5, &simParams)
	expectedMinClamp := 0.0
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedMinClamp, 1e-9) {
		t.Errorf("Expected weight to be clamped to %f, got %f", expectedMinClamp, w)
	}

	id3, id4 := common.NeuronID(3), common.NeuronID(4)
	nw.SetWeight(id3, id4, 0.7, &simParams)
	if w := nw.GetWeight(id3, id4); w != 0.7 {
		t.Errorf("Expected 0.7 for new ID pair, got %f", w)
	}
	if _, ok := nw[id3]; !ok {
		t.Errorf("WeightMap for id3 was not created after SetWeight")
	}
}

func TestApplyHebbianUpdate(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.MaxSynapticWeight = 1.0
	simParams.SynapticWeightDecayRate = 0.01
	simParams.HebbPositiveReinforceFactor = 1.0

	nw := synaptic.NewNetworkWeights()
	id1, id2 := common.NeuronID(1), common.NeuronID(2)
	learningRate := common.Rate(0.1)

	nw.SetWeight(id1, id2, 0.2, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 1.0, 1.0, learningRate, &simParams)
	expectedWeight1 := 0.297
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight1, 1e-9) {
		t.Errorf("LTP: Expected weight %f, got %f", expectedWeight1, w)
	}

	nw.SetWeight(id1, id2, 0.5, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 0.0, 1.0, learningRate, &simParams)
	expectedWeight2 := 0.495
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight2, 1e-9) {
		t.Errorf("No co-activity (decay): Expected weight %f, got %f", expectedWeight2, w)
	}

	nw.SetWeight(id1, id2, 0.95, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 1.0, 1.0, learningRate, &simParams)
	expectedWeight3 := simParams.MaxSynapticWeight
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight3, 1e-9) {
		t.Errorf("LTP with clamping: Expected weight %f, got %f", expectedWeight3, w)
	}

	smallPositiveWeight := common.SynapticWeight(0.001)
	nw.SetWeight(id1, id2, smallPositiveWeight, &simParams)
	nw.ApplyHebbianUpdate(id1, id2, 0.0, 0.0, learningRate, &simParams)
	expectedWeight4 := float64(smallPositiveWeight) * (1.0 - simParams.SynapticWeightDecayRate)
	if w := nw.GetWeight(id1, id2); !floatEquals(float64(w), expectedWeight4, 1e-9) {
		t.Errorf("Decay of small positive weight: Expected %f, got %f", expectedWeight4, w)
	}

	nw.SetWeight(id1, id1, 0.5, &simParams)
	nw.ApplyHebbianUpdate(id1, id1, 1.0, 1.0, learningRate, &simParams)
	if w := nw.GetWeight(id1, id1); w != 0.5 {
		t.Errorf("Self-connection weight should not be affected by ApplyHebbianUpdate, expected 0.5, got %f", w)
	}
}

func TestMain(m *testing.M) {
	rand.Seed(12345)
	m.Run()
}
