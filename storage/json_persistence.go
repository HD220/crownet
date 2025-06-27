// Package storage provides utilities for data persistence, including saving
// and loading network state (like synaptic weights) to/from files, and logging
// simulation data to databases.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"crownet/common"
	"crownet/synaptic"
)

// SaveNetworkWeightsToJSON serializes the given network synaptic weights into a JSON file
// at the specified filePath.
// The synaptic.NetworkWeights type (map[common.NeuronID]map[common.NeuronID]common.SynapticWeight)
// is converted to map[string]map[string]float64 for JSON compatibility, as JSON object keys
// must be strings. Neuron IDs are converted to their string representations.
// The JSON output is indented for human readability.
// File permissions are set to 0644.
//
// Parameters:
//   - weights: The synaptic.NetworkWeights data structure to save.
//   - filePath: The path to the file where the JSON data will be written.
//
// Returns:
//   - error: An error if serialization or file writing fails, nil otherwise.
func SaveNetworkWeightsToJSON(networkWeights *synaptic.NetworkWeights, filePath string) error {
	if networkWeights == nil {
		return fmt.Errorf("cannot save nil NetworkWeights")
	}
	// BUG-STORAGE-001: Use networkWeights.GetAllWeights() to get the map for serialization
	weightsToSerialize := networkWeights.GetAllWeights()

	// Prepare a structure with string keys for JSON serialization.
	serializableWeights := make(map[string]map[string]float64)
	for fromID, toMap := range weightsToSerialize {
		strFromID := strconv.FormatInt(int64(fromID), 10)
		serializableWeights[strFromID] = make(map[string]float64)
		for toID, weightVal := range toMap {
			strToID := strconv.FormatInt(int64(toID), 10)
			serializableWeights[strFromID][strToID] = float64(weightVal)
		}
	}

	data, err := json.MarshalIndent(serializableWeights, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize weights to JSON: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON weights file %s: %w", filePath, err)
	}
	return nil
}

// LoadNetworkWeightsFromJSON deserializes network synaptic weights from a JSON file
// located at the specified filePath.
// It expects the JSON structure to be map[string]map[string]float64, where keys
// are string representations of NeuronIDs. These are converted back to their numeric types.
//
// Parameters:
//   - filePath: The path to the JSON file containing the weights.
//
// Returns:
//   - map[common.NeuronID]synaptic.WeightMap: A map representing the loaded synaptic weights.
//     The outer map key is the presynaptic neuron ID, and the inner map (synaptic.WeightMap)
//     maps postsynaptic neuron IDs to their synaptic weights.
//   - error: An error if file reading, JSON unmarshalling, or NeuronID parsing fails.
//     Returns a specific error wrapping os.ErrNotExist if the file is not found.
func LoadNetworkWeightsFromJSON(filePath string) (map[common.NeuronID]synaptic.WeightMap, error) {
	// Read the JSON file content.
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON weights file %s not found: %w", filePath, err)
		}
		return nil, fmt.Errorf("failed to read JSON weights file %s: %w", filePath, err)
	}

	serializableWeights := make(map[string]map[string]float64)
	err = json.Unmarshal(data, &serializableWeights)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal weights from JSON from %s: %w", filePath, err)
	}

	// BUG-STORAGE-001: Changed to return the map directly instead of attempting to create NetworkWeights struct.
	deserializedMap := make(map[common.NeuronID]synaptic.WeightMap)
	for strFromID, toMap := range serializableWeights {
		fromIDVal, errConv := strconv.ParseInt(strFromID, 10, 64)
		if errConv != nil {
			return nil, fmt.Errorf("invalid source neuron ID in JSON '%s': %w", strFromID, errConv)
		}
		fromID := common.NeuronID(fromIDVal)

		deserializedMap[fromID] = make(synaptic.WeightMap)
		for strToID, weightVal := range toMap {
			toIDVal, errConvTo := strconv.ParseInt(strToID, 10, 64)
			if errConvTo != nil {
				return nil, fmt.Errorf("invalid target neuron ID in JSON '%s' for source '%s': %w", strToID, strFromID, errConvTo)
			}
			toID := common.NeuronID(toIDVal)
			deserializedMap[fromID][toID] = common.SynapticWeight(weightVal)
		}
	}
	return deserializedMap, nil
}
