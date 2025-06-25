// Package synaptic manages synaptic weights within the neural network.
// It provides structures and methods for initializing, accessing, modifying,
// and applying learning rules (like Hebbian updates) to these weights.
package synaptic

import (
	"crownet/common"
	"crownet/config"
	"fmt"
	"math/rand"
)

// WeightMap defines a map from a target NeuronID to a synaptic weight.
// The key is the NeuronID of the postsynaptic (target) neuron.
type WeightMap map[common.NeuronID]common.SynapticWeight

// NetworkWeights stores and manages all synaptic weights in the network.
// It encapsulates the weight map and relevant simulation parameters.
type NetworkWeights struct {
	// weights maps a source NeuronID to its outgoing connections (a WeightMap).
	weights   map[common.NeuronID]WeightMap
	simParams *config.SimulationParameters
	rng       *rand.Rand
}

// NewNetworkWeights creates and returns a new instance of NetworkWeights.
// It requires simulation parameters and a random number generator source (rng).
// Returns an error if simParams or rng is nil.
func NewNetworkWeights(simParams *config.SimulationParameters, rng *rand.Rand) (*NetworkWeights, error) {
	if simParams == nil {
		return nil, fmt.Errorf("NewNetworkWeights: simParams cannot be nil")
	}
	if rng == nil {
		return nil, fmt.Errorf("NewNetworkWeights: rng cannot be nil")
	}
	return &NetworkWeights{
		weights:   make(map[common.NeuronID]WeightMap),
		simParams: simParams,
		rng:       rng,
	}, nil
}

// InitializeAllToAllWeights sets up initial weights between all provided neuron IDs.
// Weights are randomly assigned within the bounds defined in simParams (InitialSynapticWeightMin/Max).
// Self-connections (from a neuron to itself) are initialized with a weight of zero.
func (nw *NetworkWeights) InitializeAllToAllWeights(neuronIDs []common.NeuronID) {
	minW := nw.simParams.InitialSynapticWeightMin
	maxW := nw.simParams.InitialSynapticWeightMax

	// Basic validation of initial weight limits (ideally, this is also done in config.Validate)
	if minW >= maxW {
		// Fallback or log a warning if parameters are inconsistent.
		// For now, using a safe fallback, but this should indicate a configuration error.
		minW = 0.01
		maxW = 0.05
		// Consider logging a warning here:
		// log.Printf("Warning: InitialSynapticWeightMin (%f) >= InitialSynapticWeightMax (%f). Using fallback %f-%f.",
		// nw.simParams.InitialSynapticWeightMin, nw.simParams.InitialSynapticWeightMax, minW, maxW)
	}

	for _, fromID := range neuronIDs {
		if _, exists := nw.weights[fromID]; !exists {
			nw.weights[fromID] = make(WeightMap)
		}
		for _, toID := range neuronIDs {
			if fromID == toID {
				nw.weights[fromID][toID] = 0.0 // Self-connections are zero.
			} else {
				randomFactor := nw.rng.Float64() // Use the struct's rng
				// Ensure arithmetic operations use consistent float64 types before converting to SynapticWeight
				base := float64(minW)
				diff := float64(maxW) - float64(minW)
				calculatedWeight := base + randomFactor*diff
				nw.weights[fromID][toID] = common.SynapticWeight(calculatedWeight)
			}
		}
	}
}

// GetWeight returns the synaptic weight from neuron `fromID` to neuron `toID`.
// It returns 0.0 if the connection does not explicitly exist, assuming zero weight
// for non-connections, which is common in sparse networks or for simplifying potential calculations.
func (nw *NetworkWeights) GetWeight(fromID, toID common.NeuronID) common.SynapticWeight {
	if fromMap, ok := nw.weights[fromID]; ok {
		if weight, ok2 := fromMap[toID]; ok2 {
			return weight
		}
	}
	// Return 0.0 for non-existent connections.
	return 0.0
}

// SetWeight sets the synaptic weight from neuron `fromID` to neuron `toID`.
// The weight is clamped by simParams.MaxSynapticWeight.
// A minimum applicable weight is also considered, defaulting to 0.0 but allowing
// negative values if simParams.HebbianWeightMin is negative (to support specific learning rules).
// Self-connections (fromID == toID) are always set to 0.
// Note: For specific learning rules like Hebbian, their specific min/max limits
// (e.g., HebbianWeightMin/Max) should be applied *before* calling this general SetWeight method
// if those rules require different clamping than the global MaxSynapticWeight or the default 0 minimum.
func (nw *NetworkWeights) SetWeight(fromID, toID common.NeuronID, weight common.SynapticWeight) {
	if _, ok := nw.weights[fromID]; !ok {
		nw.weights[fromID] = make(WeightMap)
	}

	limitedWeight := weight

	// Determine the minimum applicable weight.
	// Default to 0.0 for general use, but allow negative if Hebbian rules permit (e.g. for LTD).
	minApplicableWeight := common.SynapticWeight(0.0)
	if nw.simParams.HebbianWeightMin < 0 { // Check if Hebbian rules might allow negative weights
		minApplicableWeight = nw.simParams.HebbianWeightMin // Use Hebbian min if it's more permissive (negative)
	}
	// However, ensure that the weight doesn't go below 0 if not explicitly allowed by a negative HebbianWeightMin.
	// This interpretation might need refinement based on desired interaction between general SetWeight and Hebbian logic.
	// For now: if HebbianWeightMin is positive, minApplicable is 0. If HebbianWeightMin is negative, minApplicable is HebbianWeightMin.

	maxApplicableWeight := nw.simParams.MaxSynapticWeight // Global maximum.

	if fromID != toID { // Clamping does not apply to self-connections (always 0).
		if limitedWeight < minApplicableWeight {
			limitedWeight = minApplicableWeight
		}
		if limitedWeight > maxApplicableWeight {
			limitedWeight = maxApplicableWeight
		}
	}

	// Self-connections must always be zero.
	if fromID == toID {
		nw.weights[fromID][toID] = 0.0
	} else {
		nw.weights[fromID][toID] = limitedWeight
	}
}

// ApplyHebbianUpdate updates the weight of a specific synapse based on presynaptic
// and postsynaptic activity, modulated by an effective learning rate.
// The change in weight (deltaWeight) is calculated based on Hebbian principles (LTP).
// The new weight is then subject to passive decay and clamped within Hebbian-specific
// min/max bounds (simParams.HebbianWeightMin, simParams.HebbianWeightMax).
// Finally, the general SetWeight method is called to apply the update, which also
// enforces the global simParams.MaxSynapticWeight.
// Self-plasticity (fromID == toID) is not applied.
func (nw *NetworkWeights) ApplyHebbianUpdate(
	fromID, toID common.NeuronID,
	preSynapticActivity, postSynapticActivity float64, // Values > 0 indicate activity
	effectiveLearningRate common.Rate,
) {
	if fromID == toID { // Sem auto-plasticidade
		return
	}

	currentWeight := nw.GetWeight(fromID, toID)
	deltaWeight := common.SynapticWeight(0.0)

	// LTP - Long-Term Potentiation
	if preSynapticActivity > 0 && postSynapticActivity > 0 {
		// Aumenta o peso se ambos os neurônios estiverem ativos
		reinforceFactor := float64(nw.simParams.HebbPositiveReinforceFactor)
		deltaWeight = common.SynapticWeight(float64(effectiveLearningRate) * reinforceFactor)
	} else {
		// LTD - Long-Term Depression (opcional, baseado em parâmetros)
		// Exemplo: se HebbNegativeReinforceFactor > 0 e um dos neurônios (mas não ambos) está ativo.
		// Esta parte da lógica pode ser expandida conforme os requisitos.
		// if nw.simParams.HebbNegativeReinforceFactor > 0 && (preSynapticActivity > 0 || postSynapticActivity > 0) {
		//    deltaWeight = -common.SynapticWeight(float64(effectiveLearningRate) * nw.simParams.HebbNegativeReinforceFactor)
		// }
	}

	newWeight := currentWeight + deltaWeight

	// Aplica decaimento passivo do peso
	if nw.simParams.SynapticWeightDecayRate > 0 {
		newWeight *= (1.0 - common.SynapticWeight(nw.simParams.SynapticWeightDecayRate))
	}

	// Clampeia o novo peso usando os limites específicos para aprendizado Hebbiano
	clampedHebbianWeight := newWeight
	if clampedHebbianWeight < common.SynapticWeight(nw.simParams.HebbianWeightMin) {
		clampedHebbianWeight = common.SynapticWeight(nw.simParams.HebbianWeightMin)
	}
	if clampedHebbianWeight > common.SynapticWeight(nw.simParams.HebbianWeightMax) {
		clampedHebbianWeight = common.SynapticWeight(nw.simParams.HebbianWeightMax)
	}

	// Usa o SetWeight geral que aplica o clamp global MaxSynapticWeight.
	// Isso garante que mesmo o peso Hebbiano clampeado não exceda o máximo absoluto da rede.
	nw.SetWeight(fromID, toID, clampedHebbianWeight)
}

// GetAllWeights returns a deep copy of all weights in the network.
// This is useful for saving the network state or for external analysis.
func (nw *NetworkWeights) GetAllWeights() map[common.NeuronID]WeightMap {
	copiedWeights := make(map[common.NeuronID]WeightMap)
	for fromID, toMap := range nw.weights {
		copiedDestMap := make(WeightMap)
		for toID, weight := range toMap {
			copiedDestMap[toID] = weight
		}
		copiedWeights[fromID] = copiedDestMap
	}
	return copiedWeights
}

// LoadWeights loads a map of weights into the NetworkWeights structure,
// replacing any existing weights.
// It uses the SetWeight method to ensure that loaded weights adhere to current
// simulation parameter constraints (e.g., MaxSynapticWeight).
func (nw *NetworkWeights) LoadWeights(weightsToLoad map[common.NeuronID]WeightMap) {
	nw.weights = make(map[common.NeuronID]WeightMap) // Clear existing weights
	for fromID, toMap := range weightsToLoad {
		nw.weights[fromID] = make(WeightMap)
		for toID, weight := range toMap {
			// When loading, weights are set using SetWeight to ensure they
			// are consistent with the current simulation's clamping rules.
			nw.SetWeight(fromID, toID, weight)
		}
	}
}
