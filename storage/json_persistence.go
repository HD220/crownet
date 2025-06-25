// Package storage provides utilities for data persistence, including saving
// and loading network state (like synaptic weights) to/from files, and logging
// simulation data to databases.
package storage

import (
	"crownet/common"
	"crownet/synaptic"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
func SaveNetworkWeightsToJSON(weights synaptic.NetworkWeights, filePath string) error {
	// Prepare a structure with string keys for JSON serialization.
	serializableWeights := make(map[string]map[string]float64)
	for fromID, toMap := range weights {
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
//   - synaptic.NetworkWeights: The loaded synaptic weights.
//   - error: An error if file reading, JSON unmarshalling, or NeuronID parsing fails.
//            Specific error for os.IsNotExist if the file is not found.
func LoadNetworkWeightsFromJSON(filePath string) (synaptic.NetworkWeights, error) {
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

	// TODO: The following instantiation of synaptic.NetworkWeights is problematic
	// as NewNetworkWeights requires simParams and rng, and returns an error.
	// This function should ideally return map[common.NeuronID]synaptic.WeightMap
	// or handle NetworkWeights creation differently. For now, proceeding with original logic.
	loadedWeights := synaptic.NewNetworkWeights()
	for strFromID, toMap := range serializableWeights {
		fromIDVal, err := strconv.ParseInt(strFromID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid source neuron ID in JSON '%s': %w", strFromID, err)
		}
		fromID := common.NeuronID(fromIDVal)

		loadedWeights[fromID] = make(synaptic.WeightMap)
		for strToID, weightVal := range toMap {
			toIDVal, err := strconv.ParseInt(strToID, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid target neuron ID in JSON '%s' for source '%s': %w", strToID, strFromID, err)
			}
			toID := common.NeuronID(toIDVal)
			loadedWeights[fromID][toID] = common.SynapticWeight(weightVal)
		}
	}
	return loadedWeights, nil
}
