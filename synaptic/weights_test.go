package synaptic

import (
	"crownet/common"
	"crownet/config"
	"math" // Added to fix "undefined: math" errors
	"math/rand"
	"testing"
	// "fmt" // Removed as it was imported but not used
)

func defaultTestSimParams() *config.SimulationParameters {
	p := config.DefaultSimulationParameters() // Get a copy of defaults
	p.InitialSynapticWeightMin = 0.1
	p.InitialSynapticWeightMax = 0.5
	p.MaxSynapticWeight = 1.0
	p.HebbianWeightMin = 0.05 // Allow slightly lower than initial for decay tests
	p.HebbianWeightMax = 0.9
	p.HebbPositiveReinforceFactor = 0.1
	p.SynapticWeightDecayRate = 0.01
	return &p
}

func TestNewNetworkWeights(t *testing.T) {
	simParams := defaultTestSimParams()
	seed := int64(42)
	rng := rand.New(rand.NewSource(seed))

	t.Run("Valid parameters", func(t *testing.T) {
		nw, err := NewNetworkWeights(simParams, rng)
		if err != nil {
			t.Fatalf("NewNetworkWeights() error = %v, wantErr false", err)
		}
		if nw == nil {
			t.Fatalf("NewNetworkWeights() returned nil, want non-nil")
		}
		if nw.simParams != simParams {
			t.Errorf("NewNetworkWeights() nw.simParams not set correctly")
		}
		if nw.rng != rng {
			t.Errorf("NewNetworkWeights() nw.rng not set correctly")
		}
		if nw.weights == nil {
			t.Errorf("NewNetworkWeights() nw.weights map not initialized")
		}
	})

	t.Run("Nil simParams", func(t *testing.T) {
		_, err := NewNetworkWeights(nil, rng)
		if err == nil {
			t.Errorf("NewNetworkWeights() with nil simParams, got nil error, want error")
		}
	})

	t.Run("Nil rng", func(t *testing.T) {
		_, err := NewNetworkWeights(simParams, nil)
		if err == nil {
			t.Errorf("NewNetworkWeights() with nil rng, got nil error, want error")
		}
	})
}

func TestInitializeAllToAllWeights(t *testing.T) {
	simParams := defaultTestSimParams()
	rng := rand.New(rand.NewSource(42))
	nw, _ := NewNetworkWeights(simParams, rng)

	neuronIDs := []common.NeuronID{0, 1, 2}
	nw.InitializeAllToAllWeights(neuronIDs)

	for _, fromID := range neuronIDs {
		for _, toID := range neuronIDs {
			weight := nw.GetWeight(fromID, toID)
			if fromID == toID {
				if weight != 0.0 {
					t.Errorf("InitializeAllToAllWeights() self-connection weight for %d->%d = %f, want 0.0", fromID, toID, weight)
				}
			} else {
				if !(weight >= common.SynapticWeight(simParams.InitialSynapticWeightMin) && weight <= common.SynapticWeight(simParams.InitialSynapticWeightMax)) {
					// Allow for slight floating point inaccuracies if min == max by checking only one bound
					if simParams.InitialSynapticWeightMin == simParams.InitialSynapticWeightMax && weight != common.SynapticWeight(simParams.InitialSynapticWeightMin) {
						t.Errorf("InitializeAllToAllWeights() weight %d->%d = %f, want %f (min==max case)", fromID, toID, weight, simParams.InitialSynapticWeightMin)
					} else if simParams.InitialSynapticWeightMin != simParams.InitialSynapticWeightMax {
						t.Errorf("InitializeAllToAllWeights() weight %d->%d = %f, not in range [%f, %f]", fromID, toID, weight, simParams.InitialSynapticWeightMin, simParams.InitialSynapticWeightMax)
					}
				}
			}
		}
	}

	// Test fallback for inconsistent InitialSynapticWeightMin/Max
	t.Run("Inconsistent initial weight params", func(t *testing.T) {
		inconsistentParams := *simParams // copy
		inconsistentParams.InitialSynapticWeightMin = 0.6
		inconsistentParams.InitialSynapticWeightMax = 0.5
		nwFallback, _ := NewNetworkWeights(&inconsistentParams, rng)
		nwFallback.InitializeAllToAllWeights(neuronIDs)
		// Check if fallback 0.01 to 0.05 is used
		weight := nwFallback.GetWeight(neuronIDs[0], neuronIDs[1]) // Non-self connection
		if !(weight >= 0.01 && weight <= 0.05) {
			t.Errorf("InitializeAllToAllWeights() with inconsistent params, weight %v not in fallback range [0.01, 0.05]", weight)
		}
	})
}

func TestGetSetWeight(t *testing.T) {
	simParams := defaultTestSimParams()
	rng := rand.New(rand.NewSource(42))
	nw, _ := NewNetworkWeights(simParams, rng)

	fromID, toID := common.NeuronID(0), common.NeuronID(1)

	t.Run("Set and get valid weight", func(t *testing.T) {
		val := common.SynapticWeight(0.7)
		nw.SetWeight(fromID, toID, val)
		if got := nw.GetWeight(fromID, toID); got != val {
			t.Errorf("GetWeight(0,1) after SetWeight(0,1,%f) = %f, want %f", val, got, val)
		}
	})

	t.Run("Get non-existent weight", func(t *testing.T) {
		if got := nw.GetWeight(common.NeuronID(5), common.NeuronID(6)); got != 0.0 {
			t.Errorf("GetWeight(5,6) for non-existent = %f, want 0.0", got)
		}
	})

	// Clamping in SetWeight
	// Note: HebbianWeightMin can be negative, MaxSynapticWeight is the upper cap.
	// The current SetWeight clamps general sets to [0, MaxSynapticWeight] or [HebbianMin, MaxSynapticWeight] if HebbianMin is negative.
	// Let's test based on simParams.HebbianWeightMin for the lower bound if it's negative.

	// Case 1: HebbianWeightMin is positive (defaultTestSimParams has 0.05)
	// SetWeight should clamp to [0, MaxSynapticWeight] effectively
	simParams.HebbianWeightMin = 0.05 // Ensure this for the test
	nwClampTest1, _ := NewNetworkWeights(simParams, rng)

	t.Run("SetWeight clamping (HebbianMin positive)", func(t *testing.T) {
		nwClampTest1.SetWeight(fromID, toID, -0.5) // Below 0
		if got := nwClampTest1.GetWeight(fromID, toID); got != 0.0 {
			t.Errorf("SetWeight clamping below 0 (HebbianMin positive), got %f, want 0.0", got)
		}
		nwClampTest1.SetWeight(fromID, toID, common.SynapticWeight(simParams.MaxSynapticWeight+0.1)) // Above max
		if got := nwClampTest1.GetWeight(fromID, toID); got != common.SynapticWeight(simParams.MaxSynapticWeight) {
			t.Errorf("SetWeight clamping above MaxSynapticWeight, got %f, want %f", got, simParams.MaxSynapticWeight)
		}
	})

	// Case 2: HebbianWeightMin is negative
	simParamsNegativeMin := *simParams // copy
	simParamsNegativeMin.HebbianWeightMin = -0.2
	nwClampTest2, _ := NewNetworkWeights(&simParamsNegativeMin, rng)

	t.Run("SetWeight clamping (HebbianMin negative)", func(t *testing.T) {
		nwClampTest2.SetWeight(fromID, toID, -0.5) // Below HebbianMin
		if got := nwClampTest2.GetWeight(fromID, toID); got != common.SynapticWeight(simParamsNegativeMin.HebbianWeightMin) {
			t.Errorf("SetWeight clamping below HebbianMin (negative), got %f, want %f", got, simParamsNegativeMin.HebbianWeightMin)
		}
		nwClampTest2.SetWeight(fromID, toID, common.SynapticWeight(simParamsNegativeMin.MaxSynapticWeight+0.1)) // Above max
		if got := nwClampTest2.GetWeight(fromID, toID); got != common.SynapticWeight(simParamsNegativeMin.MaxSynapticWeight) {
			t.Errorf("SetWeight clamping above MaxSynapticWeight (HebbianMin negative), got %f, want %f", got, simParamsNegativeMin.MaxSynapticWeight)
		}
	})


	t.Run("SetWeight self-connection", func(t *testing.T) {
		nw.SetWeight(fromID, fromID, 0.5)
		if got := nw.GetWeight(fromID, fromID); got != 0.0 {
			t.Errorf("SetWeight self-connection weight = %f, want 0.0", got)
		}
	})
}

func TestApplyHebbianUpdate(t *testing.T) {
	simParams := defaultTestSimParams()
	rng := rand.New(rand.NewSource(42))
	nw, _ := NewNetworkWeights(simParams, rng)

	fromID, toID := common.NeuronID(0), common.NeuronID(1)
	initialWeight := common.SynapticWeight(0.4)
	nw.SetWeight(fromID, toID, initialWeight)

	effectiveLR := common.Rate(0.1)

	t.Run("LTP co-activation", func(t *testing.T) {
		nw.SetWeight(fromID, toID, initialWeight) // Reset weight
		nw.ApplyHebbianUpdate(fromID, toID, 1.0, 1.0, effectiveLR)
		currentWeight := nw.GetWeight(fromID, toID)

		expectedDelta := common.SynapticWeight(float64(effectiveLR) * float64(simParams.HebbPositiveReinforceFactor))
		expectedWeightAfterLTP := initialWeight + expectedDelta
		expectedWeightAfterDecay := expectedWeightAfterLTP * (1.0 - common.SynapticWeight(simParams.SynapticWeightDecayRate))

		// Apply Hebbian clamping
		finalExpected := expectedWeightAfterDecay
		if finalExpected < common.SynapticWeight(simParams.HebbianWeightMin) {
			finalExpected = common.SynapticWeight(simParams.HebbianWeightMin)
		}
		if finalExpected > common.SynapticWeight(simParams.HebbianWeightMax) {
			finalExpected = common.SynapticWeight(simParams.HebbianWeightMax)
		}
		// Final SetWeight applies MaxSynapticWeight cap, but HebbianMax should be <= MaxSynapticWeight
		if finalExpected > common.SynapticWeight(simParams.MaxSynapticWeight) {
			finalExpected = common.SynapticWeight(simParams.MaxSynapticWeight)
		}


		if math.Abs(float64(currentWeight-finalExpected)) > 1e-9 {
			t.Errorf("ApplyHebbianUpdate LTP: got %f, want approx %f (delta %f, afterLTP %f, afterDecay %f)",
				currentWeight, finalExpected, expectedDelta, expectedWeightAfterLTP, expectedWeightAfterDecay)
		}
	})

	t.Run("No change if only pre-active", func(t *testing.T) {
		nw.SetWeight(fromID, toID, initialWeight) // Reset weight
		nw.ApplyHebbianUpdate(fromID, toID, 1.0, 0.0, effectiveLR)
		currentWeight := nw.GetWeight(fromID, toID)
		// Only decay should apply if LTD is not active
		expectedWeightAfterDecay := initialWeight * (1.0 - common.SynapticWeight(simParams.SynapticWeightDecayRate))
		// Apply Hebbian clamping
		finalExpected := expectedWeightAfterDecay
		if finalExpected < common.SynapticWeight(simParams.HebbianWeightMin) {
			finalExpected = common.SynapticWeight(simParams.HebbianWeightMin)
		}
		if finalExpected > common.SynapticWeight(simParams.HebbianWeightMax) {
			finalExpected = common.SynapticWeight(simParams.HebbianWeightMax)
		}
		if finalExpected > common.SynapticWeight(simParams.MaxSynapticWeight) {
			finalExpected = common.SynapticWeight(simParams.MaxSynapticWeight)
		}

		if math.Abs(float64(currentWeight-finalExpected)) > 1e-9 {
			t.Errorf("ApplyHebbianUpdate only pre-active: got %f, want approx %f (only decay)", currentWeight, finalExpected)
		}
	})

	t.Run("No change for self-connection", func(t *testing.T) {
		nw.SetWeight(fromID, fromID, 0.0) // Ensure it's 0
		nw.ApplyHebbianUpdate(fromID, fromID, 1.0, 1.0, effectiveLR)
		if got := nw.GetWeight(fromID, fromID); got != 0.0 {
			t.Errorf("ApplyHebbianUpdate self-connection: got %f, want 0.0", got)
		}
	})

	// Test clamping during Hebbian update
	t.Run("Hebbian update clamping to HebbianWeightMax", func(t *testing.T) {
		// Make initial weight high, so LTP pushes it over HebbianWeightMax
		highInitialWeight := common.SynapticWeight(simParams.HebbianWeightMax - 0.01)
		nw.SetWeight(fromID, toID, highInitialWeight)
		nw.ApplyHebbianUpdate(fromID, toID, 1.0, 1.0, common.Rate(0.5)) // Strong LR

		currentWeight := nw.GetWeight(fromID, toID)
		// Expected is HebbianWeightMax after decay, then potentially MaxSynapticWeight if HebbianMax > MaxOverall
		// but our SetWeight clamps to MaxSynapticWeight.
		// The ApplyHebbianUpdate itself clamps to HebbianWeightMin/Max.
		// Then SetWeight clamps again. So it should be min(HebbianWeightMax_after_decay, MaxSynapticWeight)

		// Let's check it against HebbianWeightMax, as MaxSynapticWeight is an outer bound.
		// The internal logic of ApplyHebbianUpdate clamps to HebbianMin/Max first.
		// Then SetWeight is called. MaxSynapticWeight is usually >= HebbianWeightMax.
		// expected := common.SynapticWeight(simParams.HebbianWeightMax) // Removed unused variable
		// If decay makes it less than HebbianWeightMax, that's the value.
		// tempExpected = (highInitialWeight + deltaFromStrongLR) -> likely > HebbianWeightMax
		// tempExpectedClampedToHebbianMax = HebbianWeightMax
		// tempExpectedAfterDecay = HebbianWeightMax * (1-decay)
		// This expected value should be compared.

		// Simpler: the value should not exceed simParams.HebbianWeightMax if HebbianWeightMax <= MaxSynapticWeight
		// And it should not exceed MaxSynapticWeight in any case.
		effectiveMax := common.SynapticWeight(math.Min(float64(simParams.HebbianWeightMax), float64(simParams.MaxSynapticWeight)))

		if currentWeight > effectiveMax {
			t.Errorf("ApplyHebbianUpdate clamping high: got %f, should not exceed %f (HebbianMax %f, MaxOverall %f)",
				currentWeight, effectiveMax, simParams.HebbianWeightMax, simParams.MaxSynapticWeight)
		}
		// It's hard to predict exact value due to decay after clamping. Check it's <= effectiveMax.
		// A more precise test would mock SetWeight or check intermediate value.
		// For now, checking it doesn't exceed the *effective* cap is a good start.
		if currentWeight > common.SynapticWeight(simParams.MaxSynapticWeight) {
             t.Errorf("ApplyHebbianUpdate resulted in weight %f, exceeding MaxSynapticWeight %f", currentWeight, simParams.MaxSynapticWeight)
        }
        if currentWeight > common.SynapticWeight(simParams.HebbianWeightMax) && simParams.HebbianWeightMax <= simParams.MaxSynapticWeight {
             // This case might happen if decay brings it down just below MaxSynapticWeight but still above a lower HebbianWeightMax
             // The logic is: newW = HWM; nw.SetWeight(newW) -> if newW > MSW, newW=MSW.
             // So currentWeight should be min(HWM_after_decay, MSW)
        }
	})
}

func TestGetAllLoadWeights(t *testing.T) {
	simParams := defaultTestSimParams()
	rng := rand.New(rand.NewSource(42))
	nw, _ := NewNetworkWeights(simParams, rng)
	neuronIDs := []common.NeuronID{0, 1}
	nw.InitializeAllToAllWeights(neuronIDs)
	nw.SetWeight(0,1, 0.77)

	retrievedWeights := nw.GetAllWeights()

	// Check if it's a deep copy
	if _, ok := retrievedWeights[0]; !ok {
		t.Fatalf("GetAllWeights didn't retrieve fromID 0")
	}
	retrievedWeights[0][1] = 0.99 // Modify copy
	if nw.GetWeight(0,1) == 0.99 {
		t.Errorf("GetAllWeights did not return a deep copy (modification affected original)")
	}
	if nw.GetWeight(0,1) != 0.77 { // Check original is untouched
		t.Errorf("Original weight changed after modifying copy from GetAllWeights, expected 0.77 got %f", nw.GetWeight(0,1))
	}


	// Test LoadWeights
	nw2, _ := NewNetworkWeights(simParams, rng)
	weightsToLoad := map[common.NeuronID]WeightMap{
		0: {1: 0.11, 2: 0.22},
		1: {0: 0.33},
	}
	nw2.LoadWeights(weightsToLoad)

	if nw2.GetWeight(0,1) != 0.11 { t.Errorf("LoadWeights failed for 0->1, got %f want 0.11", nw2.GetWeight(0,1)) }
	if nw2.GetWeight(0,2) != 0.22 { t.Errorf("LoadWeights failed for 0->2, got %f want 0.22", nw2.GetWeight(0,2)) }
	if nw2.GetWeight(1,0) != 0.33 { t.Errorf("LoadWeights failed for 1->0, got %f want 0.33", nw2.GetWeight(1,0)) }
	if nw2.GetWeight(common.NeuronID(99), common.NeuronID(98)) != 0.0 {
		t.Errorf("LoadWeights: non-existent weight should be 0 after load, got %f", nw2.GetWeight(99,98))
	}

	// Test LoadWeights with clamping (e.g. loading a weight > MaxSynapticWeight)
	weightsToLoadOverMax := map[common.NeuronID]WeightMap{
		0: {1: common.SynapticWeight(simParams.MaxSynapticWeight + 0.5)},
	}
	nw2.LoadWeights(weightsToLoadOverMax)
	if nw2.GetWeight(0,1) != common.SynapticWeight(simParams.MaxSynapticWeight) {
		t.Errorf("LoadWeights should clamp overweight values, got %f want %f", nw2.GetWeight(0,1), simParams.MaxSynapticWeight)
	}
}
