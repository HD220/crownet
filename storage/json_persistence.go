package storage

import (
	"crownet/common"
	"crownet/synaptic"
	"encoding/json"
	"fmt"
	"os"
)

// SaveNetworkWeightsToJSON serializa a estrutura NetworkWeights para um arquivo JSON.
// A estrutura do JSON será: map[string]map[string]float64
// onde as chaves string são os NeuronIDs.
func SaveNetworkWeightsToJSON(weights synaptic.NetworkWeights, filePath string) error {
	// Converter NeuronIDs (int) para string para chaves JSON
	serializableWeights := make(map[string]map[string]float64)
	for fromID, toMap := range weights {
		strFromID := fmt.Sprintf("%d", fromID)
		serializableWeights[strFromID] = make(map[string]float64)
		for toID, weightVal := range toMap {
			strToID := fmt.Sprintf("%d", toID)
			serializableWeights[strFromID][strToID] = float64(weightVal)
		}
	}

	data, err := json.MarshalIndent(serializableWeights, "", "  ")
	if err != nil {
		return fmt.Errorf("falha ao serializar pesos para JSON: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("falha ao escrever arquivo de pesos JSON %s: %w", filePath, err)
	}
	return nil
}

// LoadNetworkWeightsFromJSON deserializa NetworkWeights de um arquivo JSON.
func LoadNetworkWeightsFromJSON(filePath string) (synaptic.NetworkWeights, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("arquivo de pesos JSON %s não encontrado: %w", filePath, err)
		}
		return nil, fmt.Errorf("falha ao ler arquivo de pesos JSON %s: %w", filePath, err)
	}

	serializableWeights := make(map[string]map[string]float64)
	err = json.Unmarshal(data, &serializableWeights)
	if err != nil {
		return nil, fmt.Errorf("falha ao deserializar pesos de JSON de %s: %w", filePath, err)
	}

	// Converter chaves string de volta para NeuronID (int)
	loadedWeights := synaptic.NewNetworkWeights()
	for strFromID, toMap := range serializableWeights {
		var fromID common.NeuronID
		_, err := fmt.Sscan(strFromID, &fromID)
		if err != nil {
			return nil, fmt.Errorf("ID de neurônio de origem inválido no JSON '%s': %w", strFromID, err)
		}

		loadedWeights[fromID] = make(synaptic.WeightMap)
		for strToID, weightVal := range toMap {
			var toID common.NeuronID
			_, err := fmt.Sscan(strToID, &toID)
			if err != nil {
				return nil, fmt.Errorf("ID de neurônio de destino inválido no JSON '%s' para origem '%s': %w", strToID, strFromID, err)
			}
			loadedWeights[fromID][toID] = common.SynapticWeight(weightVal)
		}
	}
	return loadedWeights, nil
}
```
