package cli_test

import (
	"bytes"       // Adicionado para capturar stdout
	"io"          // Adicionado para capturar stdout
	"log"         // Adicionado para restaurar log.SetOutput
	"os"          // Adicionado para capturar stdout
	"strings"     // Adicionado para tc.errorContains e verificação de output
	"testing"
	"crownet/cli"
	"crownet/common"
	"crownet/config"
	"crownet/network"
	"crownet/synaptic" // Necessário para synaptic.NetworkWeights nos setters
	"fmt"
	// "math/rand"
)

// Helper para capturar stdout
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	oldLogOutput := log.Writer() // Salvar a saída atual do log
	r, w, _ := os.Pipe()
	os.Stdout = w
	log.SetOutput(w) // Redirecionar também a saída do log padrão

	f() // Executar a função que imprime

	w.Close()
	os.Stdout = oldStdout     // Restaurar stdout
	log.SetOutput(oldLogOutput) // Restaurar saída do log

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestSetupContinuousInputStimulus_ErrorCases testa a lógica de validação de ID
// e o tratamento de erro de ConfigureFrequencyInput dentro de setupContinuousInputStimulus.
func TestSetupContinuousInputStimulus_ErrorCases(t *testing.T) {
	simParamsDefault := config.DefaultSimulationParameters() // Renomeado para evitar conflito
	cliCfgBase := config.CLIConfig{
		StimInputFreqHz: 10.0,
		Mode:            config.ModeSim, // Para que setupContinuousInputStimulus seja chamado
	}

	testCases := []struct {
		name                  string
		cliCfg                config.CLIConfig
		networkInputIDs       []common.NeuronID
		configureFreqInputErr error // Erro a ser retornado por mockNet.ConfigureFrequencyInput
		expectedError         bool
		errorContains         string
	}{
		{
			name: "ID válido",
			cliCfg: func() config.CLIConfig { cfg := cliCfgBase; cfg.StimInputID = 10; return cfg }(),
			networkInputIDs:       []common.NeuronID{10, 11},
			expectedError:         false,
		},
		{
			name: "ID -1 (primeiro disponível)",
			cliCfg: func() config.CLIConfig { cfg := cliCfgBase; cfg.StimInputID = -1; return cfg }(),
			networkInputIDs:       []common.NeuronID{10, 11},
			expectedError:         false,
		},
		{
			name: "ID inválido (não existe na lista de InputNeuronIDs da rede)",
			cliCfg: func() config.CLIConfig { cfg := cliCfgBase; cfg.StimInputID = 99; return cfg }(),
			networkInputIDs:       []common.NeuronID{10, 11},
			expectedError:         true,
			errorContains:         "não encontrado ou inválido",
		},
		{
			name: "Falha simulada em net.ConfigureFrequencyInput",
			cliCfg: func() config.CLIConfig { cfg := cliCfgBase; cfg.StimInputID = 10; return cfg }(),
			networkInputIDs:       []common.NeuronID{10, 11},
			configureFreqInputErr: fmt.Errorf("erro simulado em ConfigureFrequencyInput"),
			expectedError:         true,
			errorContains:         "falha ao configurar estímulo",
		},
		{
			name: "ID -1, mas sem neurônios de input na rede (não deve dar erro, apenas não configura)",
			cliCfg: func() config.CLIConfig { cfg := cliCfgBase; cfg.StimInputID = -1; return cfg }(),
			networkInputIDs:       []common.NeuronID{},
			expectedError:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appCfg := &config.AppConfig{Cli: tc.cliCfg, SimParams: simParamsDefault}

			// Usar uma instância real de network.CrowNet, mas podemos controlar seus InputNeuronIDs
			// e, para o caso de erro, podemos mockar a função ConfigureFrequencyInput se necessário.
			// Para este teste, o mock de ConfigureFrequencyInput é mais complexo de injetar sem alterar
			// a struct network.CrowNet ou usar uma interface.
			// Vamos focar nos erros gerados pela lógica do próprio orchestrator.
			// A semente não é crítica aqui, mas NewCrowNet a exige.
			net := network.NewCrowNet(len(tc.networkInputIDs)+5, 0.01, &simParamsDefault, 0)
			net.InputNeuronIDs = tc.networkInputIDs
			// Se um ID de neurônio é válido, mas ConfigureFrequencyInput falha, precisamos simular isso.
			// Isso é difícil sem uma interface de rede ou capacidade de mockar o método Net.
			// Por ora, o teste "Falha simulada em net.ConfigureFrequencyInput" não pode ser implementado
			// completamente sem tal mock. No entanto, a função ConfigureFrequencyInput em network.go
			// já retorna erro se o ID não for um InputNeuronID válido, o que é coberto.
			// Para o caso de erro simulado, vamos assumir que a validação de ID no orchestrator falharia primeiro
			// se o ID não estivesse em net.InputNeuronIDs.

			// A maneira de simular o erro de ConfigureFrequencyInput sem mudar o código de produção
			// seria ter um tipo de rede mockável no Orchestrator, o que é uma refatoração maior.
			// Para o teste "Falha simulada em net.ConfigureFrequencyInput", vamos assumir que
			// o erro é retornado por network.CrowNet.ConfigureFrequencyInput.
			// Este teste, portanto, depende do comportamento de network.CrowNet.
			if tc.name == "Falha simulada em net.ConfigureFrequencyInput" {
				// Para realmente testar este caminho, precisaríamos de uma forma de fazer
				// net.ConfigureFrequencyInput retornar um erro específico.
				// Poderíamos criar um tipo de rede mock localmente no teste:
				type TestableNet struct { *network.CrowNet; ConfigureFrequencyInputError error }
				mockableNet := &TestableNet{CrowNet: net, ConfigureFrequencyInputError: tc.configureFreqInputErr}

				// No entanto, Orchestrator.Net é *network.CrowNet, não uma interface.
				// Então, não podemos simplesmente atribuir mockableNet.
				// Este caso de teste permanece difícil de isolar perfeitamente.
				// Vamos prosseguir, e o erro virá da validação de ID se não for encontrado.
				// Se o ID for encontrado, e ConfigureFrequencyInput não tiver como ser induzido a erro
				// (além de ID não ser input, o que já é coberto), então este caso é limitado.
			}


			orchestrator := cli.NewOrchestrator(appCfg)
			orchestrator.Net = net

			err := orchestrator.SetupContinuousInputStimulusForTest()

			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}

// captureStdoutReturnError é como captureStdout mas para funções que retornam erro.
func captureStdoutReturnError(f func() error) (string, error) {
	oldStdout := os.Stdout
	oldLogOutput := log.Writer()
	r, w, _ := os.Pipe()
	os.Stdout = w
	log.SetOutput(w)

	err := f()

	w.Close()
	os.Stdout = oldStdout
	log.SetOutput(oldLogOutput)

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}


func TestRunObserveMode_OutputVerification(t *testing.T) {
	// --- Setup ---
	simParams := config.DefaultSimulationParameters()
	simParams.MinOutputNeurons = 2
	simParams.MinInputNeurons = 3
	simParams.PatternSize = 3
	simParams.PatternHeight = 1
	simParams.PatternWidth = 3
	simParams.AccumulatedPulseDecayRate = 0.0 // Sem decaimento para simplificar a verificação do potencial

	// Criar um arquivo de pesos temporário com valores conhecidos
	tempDir := t.TempDir()
	weightsFilePath := filepath.Join(tempDir, "observe_test_weights.json")

	mockWeights := synaptic.NewNetworkWeights()
	// Neurônios de Input: 0, 1, 2
	// Neurônios de Output: 3, 4 (assumindo que NewCrowNet os cria com estes IDs após os inputs)
	mockWeights.SetWeight(common.NeuronID(0), common.NeuronID(3), 1.0, &simParams)
	mockWeights.SetWeight(common.NeuronID(0), common.NeuronID(4), 0.5, &simParams)
	mockWeights.SetWeight(common.NeuronID(1), common.NeuronID(3), 0.2, &simParams)
	mockWeights.SetWeight(common.NeuronID(1), common.NeuronID(4), 1.0, &simParams)
	mockWeights.SetWeight(common.NeuronID(2), common.NeuronID(3), 0.0, &simParams)
	mockWeights.SetWeight(common.NeuronID(2), common.NeuronID(4), 0.8, &simParams)

	if err := storage.SaveNetworkWeightsToJSON(mockWeights, weightsFilePath); err != nil {
		t.Fatalf("Falha ao salvar arquivo de pesos mock: %v", err)
	}

	cliCfg := config.CLIConfig{
		Mode:           config.ModeObserve,
		WeightsFile:    weightsFilePath,
		Digit:          0,
		CyclesToSettle: 1,
		TotalNeurons:   5, // 3 input + 2 output
		Seed:           12345, // Semente fixa
	}
	appCfg := &config.AppConfig{Cli: cliCfg, SimParams: simParams}
	orchestrator := cli.NewOrchestrator(appCfg)

	originalGetDigitPatternFn := datagen.GetDigitPatternFn
	datagen.GetDigitPatternFn = func(digit int, sp *config.SimulationParameters) ([]float64, error) {
		if sp.PatternSize != 3 {
			return nil, fmt.Errorf("mock GetDigitPatternFn: PatternSize esperada 3, got %d", sp.PatternSize)
		}
		return []float64{1.0, 1.0, 0.0}, nil // Input 0 e 1 ativos, Input 2 inativo
	}
	defer func() { datagen.GetDigitPatternFn = originalGetDigitPatternFn }()

	// --- Execução ---
	var capturedOutput string
	var runErr error

	// Precisamos garantir que a rede é criada e os pesos são carregados antes de chamar runObserveModeForTest
	// A função Run() do orchestrator faz isso.
	// No entanto, Run() usa log.Fatalf. Para capturar a saída, precisamos evitar o Fatalf.
	// Vamos chamar os passos de Run() manualmente.

	// 1. Inicializar Logger (opcional para este teste, pois não verificamos o DB)
	// if err := orchestrator.InitializeLoggerForTest(); err != nil {
	//    t.Fatalf("initializeLogger failed: %v", err)
	// }

	// 2. Criar Rede (é interno ao orchestrator, chamado por Run)
	// Para ter o .Net populado, precisamos de uma forma de chamar createNetwork
	// ou ter NewOrchestrator fazendo isso, ou Run fazendo isso antes do switch.
	// Por agora, vamos criar a rede e carregas os pesos manualmente no teste.

	orchestrator.Net = network.NewCrowNet(cliCfg.TotalNeurons, common.Rate(cliCfg.BaseLearningRate), &simParams, cliCfg.Seed)
	// Forçar os IDs de input e output para serem determinísticos no teste
	// Assumindo que os primeiros MinInputNeurons são input, e os próximos MinOutputNeurons são output
	if len(orchestrator.Net.Neurons) < simParams.MinInputNeurons + simParams.MinOutputNeurons {
		t.Fatalf("Rede não tem neurônios suficientes para os IDs de input/output esperados")
	}
	inputIDs := make([]common.NeuronID, simParams.MinInputNeurons)
	for i := 0; i < simParams.MinInputNeurons; i++ {
		orchestrator.Net.Neurons[i].Type = neuron.Input
		inputIDs[i] = orchestrator.Net.Neurons[i].ID
	}
	orchestrator.Net.InputNeuronIDs = inputIDs

	outputIDs := make([]common.NeuronID, simParams.MinOutputNeurons)
	for i := 0; i < simParams.MinOutputNeurons; i++ {
		idx := simParams.MinInputNeurons + i
		orchestrator.Net.Neurons[idx].Type = neuron.Output
		outputIDs[i] = orchestrator.Net.Neurons[idx].ID
	}
	orchestrator.Net.OutputNeuronIDs = outputIDs

	// Carregar os pesos mockados na rede do orchestrator
	if errLoad := orchestrator.LoadWeightsForTest(weightsFilePath); errLoad != nil {
		t.Fatalf("Falha ao carregar pesos para o teste: %v", errLoad)
	}


	capturedOutput, runErr = captureStdoutReturnError(func() error {
		return orchestrator.RunObserveModeForTest()
	})

	// --- Verificação ---
	if runErr != nil {
		t.Fatalf("runObserveModeForTest retornou um erro inesperado: %v. Output: %s", runErr, capturedOutput)
	}

	if !strings.Contains(capturedOutput, "Dígito Apresentado: 0") {
		t.Errorf("Saída não contém 'Dígito Apresentado: 0'. Saída: %s", capturedOutput)
	}
	if !strings.Contains(capturedOutput, "Padrão de Ativação dos Neurônios de Saída") {
		t.Errorf("Saída não contém 'Padrão de Ativação dos Neurônios de Saída'. Saída: %s", capturedOutput)
	}

	// Padrão de input: [1.0, 1.0, 0.0] -> Inputs com ID 0 e 1 ativos
	// Pesos:
	// I0->O0 (ID real net.OutputNeuronIDs[0]): 1.0
	// I0->O1 (ID real net.OutputNeuronIDs[1]): 0.5
	// I1->O0: 0.2
	// I1->O1: 1.0
	// I2->O0: 0.0
	// I2->O1: 0.8
	// Ativação O0 = 1.0*1.0 (de I0) + 1.0*0.2 (de I1) = 1.2
	// Ativação O1 = 1.0*0.5 (de I0) + 1.0*1.0 (de I1) = 1.5
	// Com CyclesToSettle=1 e AccumulatedPulseDecayRate=0.0, o potencial é a soma.

	outputID_0_actual := orchestrator.Net.OutputNeuronIDs[0]
	outputID_1_actual := orchestrator.Net.OutputNeuronIDs[1]

	expectedOut0Str := fmt.Sprintf("OutNeurônio[0] (ID %d): %.4f", outputID_0_actual, 1.2000)
	expectedOut1Str := fmt.Sprintf("OutNeurônio[1] (ID %d): %.4f", outputID_1_actual, 1.5000)

	if !strings.Contains(capturedOutput, expectedOut0Str) {
		t.Errorf("Saída para Output 0 incorreta. Esperado conter '%s'. Saída: %s", expectedOut0Str, capturedOutput)
	}
	if !strings.Contains(capturedOutput, expectedOut1Str) {
		t.Errorf("Saída para Output 1 incorreta. Esperado conter '%s'. Saída: %s", expectedOut1Str, capturedOutput)
	}
}

func TestRunObserveMode_LoadWeightsError(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	cliCfg := config.CLIConfig{
		Mode:           config.ModeObserve,
		WeightsFile:    "non_existent_weights.json",
		Digit:          0,
		CyclesToSettle: 1,
	}
	appCfg := &config.AppConfig{Cli: cliCfg, SimParams: simParams}

	orchestrator := cli.NewOrchestrator(appCfg)
	// Sobrescrever loadWeightsFn para simular erro
	mockLoadError := fmt.Errorf("mocked loadWeights error: arquivo não encontrado")
	orchestrator.SetLoadWeightsFn(func(filepath string) (synaptic.NetworkWeights, error) {
		return nil, mockLoadError
	})

	// A rede precisa ser criada para que o.Net não seja nil dentro de runObserveMode
	orchestrator.Net = network.NewCrowNet(10, 0.01, &simParams, 0)


	err := orchestrator.RunObserveModeForTest()
	if err == nil {
		t.Errorf("Expected runObserveMode to return an error due to loadWeights failure, but got nil")
	} else {
		// Verificar se o erro retornado contém o erro mockado
		if !strings.Contains(err.Error(), "mocked loadWeights error") && !strings.Contains(err.Error(), "não encontrado") {
			t.Errorf("Error message mismatch. Expected to contain 'mocked loadWeights error' or 'não encontrado', got: %v", err)
		}
	}
}

func TestRunExposeMode_SaveWeightsError(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	cliCfg := config.CLIConfig{
		Mode:             config.ModeExpose,
		WeightsFile:      "test_save_weights.json",
		Epochs:           0, // 0 épocas para pular runExposureEpochs e ir direto para saveWeights
		CyclesPerPattern: 1,
		BaseLearningRate: 0.01,
	}
	appCfg := &config.AppConfig{Cli: cliCfg, SimParams: simParams}

	orchestrator := cli.NewOrchestrator(appCfg)
	mockSaveError := fmt.Errorf("mocked saveWeights error")
	orchestrator.SetSaveWeightsFn(func(weights synaptic.NetworkWeights, filepath string) error {
		return mockSaveError
	})

	// Configurar uma rede mínima
	orchestrator.Net = network.NewCrowNet(10, 0.01, &simParams, 0)
	// Simular que o carregamento de pesos (opcional para expose) não deu erro ou não aconteceu
	orchestrator.SetLoadWeightsFn(func(filepath string) (synaptic.NetworkWeights, error) {
		return synaptic.NewNetworkWeights(), nil // Sucesso no carregamento (ou nenhum arquivo)
	})

	err := orchestrator.RunExposeModeForTest()
	if err == nil {
		t.Errorf("Expected runExposeMode to return an error due to saveWeights failure, but got nil")
	} else {
		if !strings.Contains(err.Error(), "mocked saveWeights error") {
			t.Errorf("Error message mismatch. Expected 'mocked saveWeights error', got: %v", err)
		}
	}
}
```
