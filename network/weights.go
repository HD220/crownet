package network

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

const (
	MinInitialWeight = -0.1
	MaxInitialWeight = 0.1
)

// initializeSynapticWeights sets up initial synaptic weights between neurons.
// For MVP, using random weights. All-to-all potential connectivity.
func (cn *CrowNet) initializeSynapticWeights(numNeurons int) {
	cn.SynapticWeights = make(map[int]map[int]float64)

	neuronIDs := make([]int, len(cn.Neurons))
	for i, n := range cn.Neurons {
		neuronIDs[i] = n.ID
	}

	for _, fromID := range neuronIDs {
		cn.SynapticWeights[fromID] = make(map[int]float64)
		for _, toID := range neuronIDs {
			if fromID == toID {
				cn.SynapticWeights[fromID][toID] = 0 // No self-connections through this weight matrix
			} else {
				// Initialize with small random weights
				weight := MinInitialWeight + rand.Float64()*(MaxInitialWeight-MinInitialWeight)
				cn.SynapticWeights[fromID][toID] = weight
			}
		}
	}
	// fmt.Printf("[WEIGHTS] Initialized synaptic weights for %d neurons.\n", numNeurons)
}

// GetWeight returns the synaptic weight from a source neuron to a target neuron.
func (cn *CrowNet) GetWeight(fromNeuronID, toNeuronID int) float64 {
	if fromMap, ok := cn.SynapticWeights[fromNeuronID]; ok {
		if weight, ok2 := fromMap[toNeuronID]; ok2 {
			return weight
		}
	}
	// Default to 0 if no specific weight is found (implies no connection or error)
	// Consider if this default is appropriate or if missing weights should be an error.
	// For a fully connected initial matrix (even with zeros), this path shouldn't be hit often
	// unless new neurons are added without updating weights.
	return 0.0
}

// SetWeight sets the synaptic weight from a source neuron to a target neuron.
func (cn *CrowNet) SetWeight(fromNeuronID, toNeuronID int, weight float64) {
	if _, ok := cn.SynapticWeights[fromNeuronID]; !ok {
		cn.SynapticWeights[fromNeuronID] = make(map[int]float64)
	}
	cn.SynapticWeights[fromNeuronID][toNeuronID] = weight
}

// SaveWeights saves the synaptic weights to a JSON file.
func (cn *CrowNet) SaveWeights(filePath string) error {
	data, err := json.MarshalIndent(cn.SynapticWeights, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal synaptic weights: %w", err)
	}
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write weights file %s: %w", filePath, err)
	}
	fmt.Printf("Synaptic weights saved to %s\n", filePath)
	return nil
}

// LoadWeights loads synaptic weights from a JSON file.
func (cn *CrowNet) LoadWeights(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("weights file %s not found", filePath) // Specific error for not found
		}
		return fmt.Errorf("failed to read weights file %s: %w", filePath, err)
	}
	err = json.Unmarshal(data, &cn.SynapticWeights)
	if err != nil {
		return fmt.Errorf("failed to unmarshal synaptic weights from %s: %w", filePath, err)
	}
	fmt.Printf("Synaptic weights loaded from %s\n", filePath)
	return nil
}
