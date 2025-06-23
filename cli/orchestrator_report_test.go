package cli_test

import (
	"bytes"
	"crownet/cli"
	"crownet/common"
	"crownet/config"
	"crownet/datagen"
	"crownet/network"
	"crownet/storage"
	"crownet/synaptic"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"encoding/json" // Adicionado para StructToJSONString
)

// captureAndFormatOutput executa uma função e formata sua saída de console e erro para o relatório.
func captureAndFormatOutput(actionName string, f func() error) (string, error) {
	oldStdout := os.Stdout
	oldLogOutput := log.Writer()
	r, w, _ := os.Pipe()
	os.Stdout = w
	log.SetOutput(w)

	actionErr := f()

	w.Close()
	os.Stdout = oldStdout
	log.SetOutput(oldLogOutput)

	var buf bytes.Buffer
	io.Copy(&buf, r)

	reportOutput := fmt.Sprintf("### Saída do Console para: %s\n\n```text\n%s\n```\n", actionName, strings.TrimSpace(buf.String()))
	if actionErr != nil {
		reportOutput += fmt.Sprintf("\n**Erro retornado:** `%v`\n", actionErr)
	}
	return reportOutput, actionErr
}

// TestGenerateDigitRecognitionReport é um "teste" que na verdade gera um relatório.
func TestGenerateDigitRecognitionReport(t *testing.T) {
	reportContent := &strings.Builder{}
	reportFilePath := filepath.Join("..", "docs", "execution_reports", "report_digit_1_recognition_simulation.md") // Salvar na pasta docs/execution_reports

	// --- Cabeçalho do Relatório ---
	reportContent.WriteString(fmt.Sprintf("# Relatório de Execução Simulada: Reconhecimento do Dígito \"1\"\n\n"))
	reportContent.WriteString(fmt.Sprintf("Este relatório foi gerado programaticamente em: %s\n\n", time.Now().Format(time.RFC1123)))
	reportContent.WriteString("## Objetivo:\nDemonstrar um fluxo de treinamento e observação da rede CrowNet para o dígito \"1\".\n\n")

	// --- Configuração Base ---
	reportContent.WriteString("## Configuração da Simulação Base:\n\n")
	simParams := config.DefaultSimulationParameters()
	simParams.MinInputNeurons = 3
	simParams.MinOutputNeurons = 1 // Apenas 1 neurônio de output para simplificar o "reconhecimento" do dígito 1
	simParams.PatternHeight = 1
	simParams.PatternWidth = 3
	simParams.PatternSize = simParams.PatternHeight * simParams.PatternWidth
	simParams.AccumulatedPulseDecayRate = 0.0 // Sem decaimento para este exemplo
	simParams.SynapticWeightDecayRate = 0.0 // Sem decaimento de peso para este exemplo

	reportContent.WriteString(fmt.Sprintf("```json\n%s\n```\n\n",StructToJSONString(simParams)))


	// --- Passo 1: Treinamento (Modo Expose) ---
	reportContent.WriteString("## Passo 1: Treinamento da Rede (Modo `expose`)\n\n")
	exposeCliCfg := config.CLIConfig{
		Mode:             config.ModeExpose,
		WeightsFile:      "temp_report_weights.json", // Usaremos em memória/mock para este relatório
		Epochs:           1,
		CyclesPerPattern: 2, // Ciclos suficientes para aprendizado básico
		TotalNeurons:     simParams.MinInputNeurons + simParams.MinOutputNeurons + 1, // 3 input, 1 output, 1 internal
		BaseLearningRate: 0.1,
		Seed:             12345,
	}
	reportContent.WriteString(fmt.Sprintf("### Configuração CLI para `expose`:\n\n```json\n%s\n```\n\n", StructToJSONString(exposeCliCfg)))

	exposeAppCfg := &config.AppConfig{Cli: exposeCliCfg, SimParams: simParams}
	exposeOrchestrator := cli.NewOrchestrator(exposeAppCfg)

	var trainedWeights synaptic.NetworkWeights
	saveWeightsCalled := false
	exposeOrchestrator.SetSaveWeightsFn(func(weights synaptic.NetworkWeights, filepathStr string) error {
		trainedWeights = weights // Capturar pesos em memória
		saveWeightsCalled = true
		reportContent.WriteString(fmt.Sprintf("**Mock `saveWeightsFn`:** Pesos capturados para o arquivo '%s'.\n", filepathStr))
		return nil
	})
	exposeOrchestrator.SetLoadWeightsFn(func(filepathStr string) (synaptic.NetworkWeights, error) {
		reportContent.WriteString(fmt.Sprintf("**Mock `loadWeightsFn`:** Chamado para '%s', retornando 'não encontrado' (comportamento padrão para novo treino).\n", filepathStr))
		return nil, fmt.Errorf("arquivo não encontrado (mock)")
	})

	originalGetDigitPatternFn := datagen.GetDigitPatternFn
	datagen.GetDigitPatternFn = func(digit int, sp *config.SimulationParameters) ([]float64, error) {
		reportContent.WriteString(fmt.Sprintf("**Mock `GetDigitPatternFn`:** Chamado para dígito %d, retornando padrão `[1,1,1]`.\n", digit))
		return []float64{1.0, 1.0, 1.0}, nil // Padrão para o "dígito 1"
	}
	defer func() { datagen.GetDigitPatternFn = originalGetDigitPatternFn }()

	// Simular criação da rede para o orchestrator
	exposeOrchestrator.Net = network.NewCrowNet(exposeCliCfg.TotalNeurons, common.Rate(exposeCliCfg.BaseLearningRate), &exposeAppCfg.SimParams, exposeCliCfg.Seed)
	// Configurar IDs de input/output manualmente para consistência no teste
	// Assumindo que os primeiros MinInputNeurons são input, e os próximos MinOutputNeurons são output
	if len(exposeOrchestrator.Net.Neurons) < simParams.MinInputNeurons + simParams.MinOutputNeurons {
		t.Fatalf("Rede de expose não tem neurônios suficientes para os IDs de input/output esperados")
	}
	inputIDsExpose := make([]common.NeuronID, simParams.MinInputNeurons)
	for i := 0; i < simParams.MinInputNeurons; i++ {
		exposeOrchestrator.Net.Neurons[i].Type = neuron.Input
		inputIDsExpose[i] = exposeOrchestrator.Net.Neurons[i].ID
	}
	exposeOrchestrator.Net.InputNeuronIDs = inputIDsExpose

	outputIDsExpose := make([]common.NeuronID, simParams.MinOutputNeurons)
	for i := 0; i < simParams.MinOutputNeurons; i++ {
		idx := simParams.MinInputNeurons + i
		exposeOrchestrator.Net.Neurons[idx].Type = neuron.Output
		outputIDsExpose[i] = exposeOrchestrator.Net.Neurons[idx].ID
	}
	exposeOrchestrator.Net.OutputNeuronIDs = outputIDsExpose


	exposeOutput, exposeErr := captureAndFormatOutput("Modo Expose", exposeOrchestrator.RunExposeModeForTest)
	reportContent.WriteString(exposeOutput)
	if exposeErr != nil {
		reportContent.WriteString(fmt.Sprintf("\n**Resultado do Treinamento:** FALHOU (%v)\n", exposeErr))
		t.Errorf("Execução do modo expose falhou: %v", exposeErr)
	} else if !saveWeightsCalled {
		reportContent.WriteString("\n**Resultado do Treinamento:** FALHOU (saveWeightsFn não foi chamado)\n")
		t.Errorf("saveWeightsFn não foi chamado no modo expose")
	} else {
		reportContent.WriteString("\n**Resultado do Treinamento:** Concluído com sucesso (simulado).\n")
		reportContent.WriteString(fmt.Sprintf("Pesos \"treinados\" (capturados em memória):\n```json\n%s\n```\n\n", StructToJSONString(trainedWeights)))
	}

	// --- Passo 2: Observação do Dígito "1" (Modo Observe) ---
	reportContent.WriteString("## Passo 2: Observação do Dígito \"1\" (Modo `observe`)\n\n")
	observeCliCfg := config.CLIConfig{
		Mode:           config.ModeObserve,
		WeightsFile:    exposeCliCfg.WeightsFile, // Usar o mesmo nome de arquivo (embora carreguemos de `trainedWeights`)
		Digit:          1,                        // Observar o "dígito 1"
		CyclesToSettle: 1,
		TotalNeurons:   exposeCliCfg.TotalNeurons,
		Seed:           12345,
	}
	reportContent.WriteString(fmt.Sprintf("### Configuração CLI para `observe`:\n\n```json\n%s\n```\n\n", StructToJSONString(observeCliCfg)))

	observeAppCfg := &config.AppConfig{Cli: observeCliCfg, SimParams: simParams}
	observeOrchestrator := cli.NewOrchestrator(observeAppCfg)

	observeOrchestrator.SetLoadWeightsFn(func(filepathStr string) (synaptic.NetworkWeights, error) {
		reportContent.WriteString(fmt.Sprintf("**Mock `loadWeightsFn`:** Chamado para '%s', retornando pesos \"treinados\" capturados.\n", filepathStr))
		if trainedWeights == nil { // Se o treino falhou
			return nil, fmt.Errorf("pesos treinados não disponíveis do passo anterior")
		}
		return trainedWeights, nil
	})
	// datagen.GetDigitPatternFn já está mockado para retornar [1,1,1]

	// Simular criação da rede para o orchestrator
	observeOrchestrator.Net = network.NewCrowNet(observeCliCfg.TotalNeurons, common.Rate(observeCliCfg.BaseLearningRate), &observeAppCfg.SimParams, observeCliCfg.Seed)
	// Reconfigurar IDs de input/output
	if len(observeOrchestrator.Net.Neurons) < simParams.MinInputNeurons + simParams.MinOutputNeurons {
		t.Fatalf("Rede de observe não tem neurônios suficientes para os IDs de input/output esperados")
	}
	inputIDsObserve := make([]common.NeuronID, simParams.MinInputNeurons)
	for i := 0; i < simParams.MinInputNeurons; i++ {
		observeOrchestrator.Net.Neurons[i].Type = neuron.Input
		inputIDsObserve[i] = observeOrchestrator.Net.Neurons[i].ID
	}
	observeOrchestrator.Net.InputNeuronIDs = inputIDsObserve

	outputIDsObserve := make([]common.NeuronID, simParams.MinOutputNeurons)
	for i := 0; i < simParams.MinOutputNeurons; i++ {
		idx := simParams.MinInputNeurons + i
		observeOrchestrator.Net.Neurons[idx].Type = neuron.Output
		outputIDsObserve[i] = observeOrchestrator.Net.Neurons[idx].ID
	}
	observeOrchestrator.Net.OutputNeuronIDs = outputIDsObserve
	// Atribuir os pesos treinados à rede de observação
	observeOrchestrator.Net.SynapticWeights = trainedWeights


	observeOutput, observeErr := captureAndFormatOutput("Modo Observe (Dígito 1)", observeOrchestrator.RunObserveModeForTest)
	reportContent.WriteString(observeOutput)
	if observeErr != nil {
		reportContent.WriteString(fmt.Sprintf("\n**Resultado da Observação:** FALHOU (%v)\n", observeErr))
		t.Errorf("Execução do modo observe falhou: %v", observeErr)
	} else {
		reportContent.WriteString("\n**Resultado da Observação:** Concluído. Verificar saída do console para ativação do neurônio de output.\n")
		// Uma validação programática da saída capturada poderia ser adicionada aqui se necessário,
		// similar ao TestRunObserveMode_OutputVerification.
		// Por exemplo, verificar se o neurônio de output teve alta ativação.
		// Com 1 neurônio de output e padrão [1,1,1], a ativação dependerá dos pesos I->O.
		// Se I0,I1,I2 -> O0 (ID outputIDsObserve[0])
		// Pesos (exemplo, poderiam ser aprendidos): w(I0,O0), w(I1,O0), w(I2,O0)
		// Ativação = 1*w(I0,O0) + 1*w(I1,O0) + 1*w(I2,O0) (pois padrão é [1,1,1])
		// Os pesos reais foram capturados em trainedWeights. Podemos usá-los para prever a saída.
		if len(observeOrchestrator.Net.OutputNeuronIDs) > 0 {
			outputNeuronID := observeOrchestrator.Net.OutputNeuronIDs[0]
			var expectedActivation float64 = 0
			for _, inputID := range observeOrchestrator.Net.InputNeuronIDs {
				// Assumindo que o padrão de input [1,1,1] ativa todos os inputs
				weight := trainedWeights.GetWeight(inputID, outputNeuronID)
				expectedActivation += float64(weight) // Sinal do pulso é 1.0
			}
			reportContent.WriteString(fmt.Sprintf("\n**Ativação Esperada (Teórica) do Neurônio de Output %d:** %.4f (sem decaimento)\n", outputNeuronID, expectedActivation))

			// Extrair ativação real da saída do console
			// (Isso é frágil e depende do formato exato do log)
			// Ex: "OutNeurônio[0] (ID X): VALOR"
			// Esta parte é complexa de fazer de forma robusta aqui, melhor para um teste dedicado.
		}
	}

	reportContent.WriteString("\n---\nFim do Relatório.\n")

	// --- Salvar Relatório ---
	// Garantir que o diretório docs/execution_reports exista
	reportDir := filepath.Dir(reportFilePath)
	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		os.MkdirAll(reportDir, 0755)
	}

	err := os.WriteFile(reportFilePath, []byte(reportContent.String()), 0644)
	if err != nil {
		t.Fatalf("Falha ao escrever arquivo de relatório %s: %v", reportFilePath, err)
	}
	t.Logf("Relatório de execução simulada gerado em: %s", reportFilePath)
}

// StructToJSONString converte uma struct para uma string JSON formatada.
// Usado para incluir configurações no relatório.
func StructToJSONString(data interface{}) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Erro ao serializar para JSON: %v", err)
	}
	return string(jsonData)
}
