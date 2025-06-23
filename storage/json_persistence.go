package storage

import (
	"crownet/common"
	"crownet/synaptic"
	"encoding/json"
	"fmt"
	"os"
	"strconv" // Consolidado e "os" duplicado removido
)

// SaveNetworkWeightsToJSON serializa a estrutura NetworkWeights para um arquivo JSON.
// A estrutura do JSON será: map[string]map[string]float64
// onde as chaves string são os NeuronIDs.
func SaveNetworkWeightsToJSON(weights synaptic.NetworkWeights, filePath string) error {
	// Converter NeuronIDs (int) para string para chaves JSON
	serializableWeights := make(map[string]map[string]float64)
	for fromID, toMap := range weights {
		strFromID := strconv.FormatInt(int64(fromID), 10) // Usa strconv
		serializableWeights[strFromID] = make(map[string]float64)
		for toID, weightVal := range toMap {
			strToID := strconv.FormatInt(int64(toID), 10) // Usa strconv
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
		fromIDVal, err := strconv.ParseInt(strFromID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ID de neurônio de origem inválido no JSON '%s': %w", strFromID, err)
		}
		fromID := common.NeuronID(fromIDVal)

		loadedWeights[fromID] = make(synaptic.WeightMap)
		for strToID, weightVal := range toMap {
			toIDVal, err := strconv.ParseInt(strToID, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("ID de neurônio de destino inválido no JSON '%s' para origem '%s': %w", strToID, strFromID, err)
			}
			toID := common.NeuronID(toIDVal)
			loadedWeights[fromID][toID] = common.SynapticWeight(weightVal)
		}
	}
	return loadedWeights, nil
}
```
