package storage_test

import (
	"crownet/common"
	"crownet/config" // Necessário para simParams em synaptic.SetWeight
	"crownet/storage"
	"crownet/synaptic"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings" // Adicionado
	"errors"  // Adicionado
	"testing"
)

func TestSaveAndLoadNetworkWeights(t *testing.T) {
	tempDir := t.TempDir() // Cria um diretório temporário que é limpo após o teste
	filePath := filepath.Join(tempDir, "test_weights.json")

	originalWeights := synaptic.NewNetworkWeights()
	simParams := config.DefaultSimulationParameters() // Necessário para SetWeight

	// Popular com alguns dados de exemplo
	id1, id2, id3 := common.NeuronID(1), common.NeuronID(2), common.NeuronID(3)
	originalWeights.SetWeight(id1, id2, 0.5, &simParams)
	originalWeights.SetWeight(id1, id3, 0.75, &simParams)
	originalWeights.SetWeight(id2, id3, -0.25, &simParams) // SetWeight clampará para 0.0

	// Recuperar o valor clampeado para comparação precisa
	clampedValId2Id3 := originalWeights.GetWeight(id2, id3)


	err := storage.SaveNetworkWeightsToJSON(originalWeights, filePath)
	if err != nil {
		t.Fatalf("SaveNetworkWeightsToJSON failed: %v", err)
	}

	loadedWeights, err := storage.LoadNetworkWeightsFromJSON(filePath)
	if err != nil {
		t.Fatalf("LoadNetworkWeightsFromJSON failed: %v", err)
	}

	if loadedWeights == nil {
		t.Fatalf("LoadNetworkWeightsFromJSON returned nil weights")
	}

	// Verificar se os pesos carregados são iguais aos originais
	// Precisamos criar um mapa esperado com os valores clampeados se SetWeight os modifica.
	expectedWeights := synaptic.NewNetworkWeights()
	expectedWeights.SetWeight(id1, id2, 0.5, &simParams)
	expectedWeights.SetWeight(id1, id3, 0.75, &simParams)
	// Usar o valor que GetWeight retornaria após o clamp em SetWeight
	expectedWeights[id2][id3] = clampedValId2Id3


	if !reflect.DeepEqual(loadedWeights, expectedWeights) {
		t.Errorf("Loaded weights do not match original weights.\nOriginal: %v\nLoaded:   %v", originalWeights, loadedWeights)
	}
}

func TestLoadNetworkWeightsFromJSON_FileNotExist(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "non_existent_weights.json")

	_, err := storage.LoadNetworkWeightsFromJSON(filePath)
	if err == nil {
		t.Fatalf("LoadNetworkWeightsFromJSON should have failed for non-existent file, but got nil error")
	}
	if !os.IsNotExist(err) { // Verifica se o erro é do tipo "não existe"
		// A função LoadNetworkWeightsFromJSON encapsula o erro, então precisamos verificar a string.
		// Este teste é um pouco frágil por causa disso.
		// Idealmente, LoadNetworkWeightsFromJSON retornaria um erro específico ou permitiria unwrapping.
		// Contudo, a mensagem de erro atual inclui "não encontrado".
		expectedSubString := "não encontrado"
		if !strings.Contains(err.Error(), expectedSubString) {
			t.Errorf("Expected error to be os.IsNotExist or contain '%s', got: %v", expectedSubString, err)
		}
	}
}

func TestLoadNetworkWeightsFromJSON_MalformedJSON(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "malformed.json")

	malformedData := []byte(`{"1": {"2": 0.5, "3": "not-a-float"}}`) // "not-a-float" causará erro
	err := os.WriteFile(filePath, malformedData, 0644)
	if err != nil {
		t.Fatalf("Failed to write malformed JSON file: %v", err)
	}

	_, err = storage.LoadNetworkWeightsFromJSON(filePath)
	if err == nil {
		t.Fatalf("LoadNetworkWeightsFromJSON should have failed for malformed JSON, but got nil error")
	}
	// Verificar se o erro é do tipo json.UnmarshalTypeError ou similar, ou contém "deserializar"
	var unmarshalTypeError *json.UnmarshalTypeError
	if !errors.As(err, &unmarshalTypeError) {
		expectedSubString := "falha ao deserializar pesos de JSON"
		if !strings.Contains(err.Error(), expectedSubString) {
			t.Errorf("Expected error to be a JSON unmarshal error or contain '%s', got: %v", expectedSubString, err)
		}
	}
}

func TestLoadNetworkWeightsFromJSON_InvalidNeuronID(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid_id.json")

	// JSON com uma chave de ID de neurônio que não é um número
	invalidIDData := []byte(`{"1": {"not-an-id": 0.5}}`)
	err := os.WriteFile(filePath, invalidIDData, 0644)
	if err != nil {
		t.Fatalf("Failed to write JSON file with invalid ID: %v", err)
	}

	_, err = storage.LoadNetworkWeightsFromJSON(filePath)
	if err == nil {
		t.Fatalf("LoadNetworkWeightsFromJSON should have failed for invalid neuron ID, but got nil error")
	}
	expectedSubString := "ID de neurônio de destino inválido no JSON"
	if !strings.Contains(err.Error(), expectedSubString) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedSubString, err)
	}
}

// Adicionar import "strings" e "errors"
```
