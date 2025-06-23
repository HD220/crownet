package cli_test

import (
	"bytes"
	"crownet/cli"
	"crownet/common"
	"crownet/config"
	"crownet/datagen"
	"crownet/network"
	"crownet/neuron"
	"crownet/synaptic"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// helperCaptureAndFormatOutput executes an action and captures its console output (stdout & stderr/log)
// and any returned error, formatting it for inclusion in a Markdown report.
func helperCaptureAndFormatOutput(actionName string, actionFunc func() error, reportBuilder *strings.Builder) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldLogOutput := log.Writer()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr
	log.SetOutput(wErr) // Capture standard log output via stderr pipe

	err := actionFunc()

	wOut.Close()
	wErr.Close()

	var bufOut bytes.Buffer
	var bufErr bytes.Buffer

	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	os.Stdout = oldStdout
	os.Stderr = oldStderr
	log.SetOutput(oldLogOutput)

	reportBuilder.WriteString(fmt.Sprintf("### Console Output for: %s\n\n```text\n", actionName))
	if bufOut.Len() > 0 {
		reportBuilder.WriteString("--- STDOUT ---\n")
		reportBuilder.WriteString(strings.TrimSpace(bufOut.String()) + "\n")
	}
	if bufErr.Len() > 0 {
		reportBuilder.WriteString("--- STDERR / LOG ---\n")
		reportBuilder.WriteString(strings.TrimSpace(bufErr.String()) + "\n")
	}
	reportBuilder.WriteString("```\n")

	if err != nil {
		reportBuilder.WriteString(fmt.Sprintf("\n**Error returned:** `%v`\n", err))
	}
}

// helperStructToJSONString converts a struct to a formatted JSON string.
// Useful for embedding configuration in reports.
func helperStructToJSONString(data interface{}) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error serializing to JSON: %v", err)
	}
	return string(jsonData)
}

// setupOrchestratorForReportTest configures an orchestrator and its network for report generation tests.
// It handles common setup for network, input/output neurons, and mocks.
func setupOrchestratorForReportTest(t *testing.T, appCfg *config.AppConfig, reportBuilder *strings.Builder) *cli.Orchestrator {
	orchestrator := cli.NewOrchestrator(appCfg)

	// Create and configure the network instance manually for predictable NeuronIDs
	orchestrator.Net = network.NewCrowNet(
		appCfg.Cli.TotalNeurons,
		common.Rate(appCfg.Cli.BaseLearningRate),
		&appCfg.SimParams,
		appCfg.Cli.Seed,
	)

	// Ensure deterministic Neuron IDs and types for the report
	if len(orchestrator.Net.Neurons) < appCfg.SimParams.MinInputNeurons+appCfg.SimParams.MinOutputNeurons {
		t.Fatalf("Network does not have enough neurons for MinInput/Output as per SimParams.")
	}

	var inputIDs []common.NeuronID
	for i := 0; i < appCfg.SimParams.MinInputNeurons; i++ {
		orchestrator.Net.Neurons[i].ID = common.NeuronID(i) // Force ID
		orchestrator.Net.Neurons[i].Type = neuron.Input
		inputIDs = append(inputIDs, orchestrator.Net.Neurons[i].ID)
	}
	orchestrator.Net.InputNeuronIDs = inputIDs

	var outputIDs []common.NeuronID
	for i := 0; i < appCfg.SimParams.MinOutputNeurons; i++ {
		neuronIndex := appCfg.SimParams.MinInputNeurons + i
		orchestrator.Net.Neurons[neuronIndex].ID = common.NeuronID(neuronIndex) // Force ID
		orchestrator.Net.Neurons[neuronIndex].Type = neuron.Output
		outputIDs = append(outputIDs, orchestrator.Net.Neurons[neuronIndex].ID)
	}
	orchestrator.Net.OutputNeuronIDs = outputIDs

	reportBuilder.WriteString(fmt.Sprintf(
		"**Note:** Network manually configured with %d inputs (IDs: %v) and %d outputs (IDs: %v) for this report.\n",
		len(inputIDs), inputIDs, len(outputIDs), outputIDs,
	))

	return orchestrator
}

// TestGenerateDigitRecognitionReport generates a Markdown report demonstrating a simulated
// training and observation flow for recognizing a digit (e.g., "1").
func TestGenerateDigitRecognitionReport(t *testing.T) {
	reportBuilder := &strings.Builder{}
	reportFilePath := filepath.Join("..", "docs", "execution_reports", "report_digit_1_recognition_simulation.md")

	// --- Report Header ---
	reportBuilder.WriteString(fmt.Sprintf("# Simulated Execution Report: Digit \"1\" Recognition\n\n"))
	reportBuilder.WriteString(fmt.Sprintf("Generated programmatically on: %s\n\n", time.Now().Format(time.RFC1123)))
	reportBuilder.WriteString("## Objective:\nDemonstrate a training and observation workflow for the CrowNet neural network, focusing on the digit \"1\".\n\n")

	// --- Base Simulation Configuration ---
	reportBuilder.WriteString("## Base Simulation Parameters:\n\n")
	simParams := config.DefaultSimulationParameters()
	// Customize parameters for a simple, illustrative report
	simParams.MinInputNeurons = 3  // e.g., for a 1x3 pattern
	simParams.MinOutputNeurons = 1 // Single output neuron to represent "recognition" of digit "1"
	simParams.PatternHeight = 1
	simParams.PatternWidth = 3
	simParams.PatternSize = simParams.PatternHeight * simParams.PatternWidth
	simParams.AccumulatedPulseDecayRate = 0.0 // No decay for simpler potential tracking in example
	simParams.SynapticWeightDecayRate = 0.0   // No weight decay for this example
	reportBuilder.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", helperStructToJSONString(simParams)))

	// --- Step 1: Network Training (Expose Mode) ---
	reportBuilder.WriteString("## Step 1: Network Training (Expose Mode)\n\n")
	exposeCliCfg := config.CLIConfig{
		Mode:             config.ModeExpose,
		WeightsFile:      "report_temp_weights.json", // Temporary, conceptual path for report
		Epochs:           1,                          // Minimal epoch for demonstration
		CyclesPerPattern: 2,
		TotalNeurons:     simParams.MinInputNeurons + simParams.MinOutputNeurons + 1, // e.g., 3 input, 1 output, 1 internal
		BaseLearningRate: 0.1,
		Seed:             12345, // Fixed seed for reproducibility
	}
	reportBuilder.WriteString(fmt.Sprintf("### Expose Mode CLI Configuration:\n\n```json\n%s\n```\n\n", helperStructToJSONString(exposeCliCfg)))

	exposeAppCfg := &config.AppConfig{Cli: exposeCliCfg, SimParams: simParams}
	exposeOrchestrator := setupOrchestratorForReportTest(t, exposeAppCfg, reportBuilder)

	var trainedWeights synaptic.NetworkWeights
	saveWeightsCalled := false
	exposeOrchestrator.SetSaveWeightsFn(func(weights synaptic.NetworkWeights, filepathStr string) error {
		trainedWeights = weights // Capture weights in memory for the next step
		saveWeightsCalled = true
		reportBuilder.WriteString(fmt.Sprintf("**Mock `saveWeightsFn`:** Weights captured from '%s' (simulated save).\n", filepathStr))
		return nil
	})
	exposeOrchestrator.SetLoadWeightsFn(func(filepathStr string) (synaptic.NetworkWeights, error) {
		reportBuilder.WriteString(fmt.Sprintf("**Mock `loadWeightsFn`:** Called for '%s'; returning 'not found' (simulating new training session).\n", filepathStr))
		return nil, fmt.Errorf("weights file not found (mock)")
	})

	originalGetDigitPatternFn := datagen.GetDigitPatternFn
	datagen.GetDigitPatternFn = func(digit int, sp *config.SimulationParameters) ([]float64, error) {
		// For this report, all digits will present the same pattern [1,1,1] to train the single output neuron
		reportBuilder.WriteString(fmt.Sprintf("**Mock `GetDigitPatternFn`:** Called for digit %d; returning fixed pattern [1,1,1] for simplicity.\n", digit))
		return []float64{1.0, 1.0, 1.0}, nil // Pattern to strongly activate the 3 input neurons
	}
	defer func() { datagen.GetDigitPatternFn = originalGetDigitPatternFn }()

	helperCaptureAndFormatOutput("Expose Mode Execution", exposeOrchestrator.RunExposeModeForTest, reportBuilder)
	if !saveWeightsCalled || trainedWeights == nil {
		reportBuilder.WriteString("\n**Training Outcome:** FAILED (Simulated: `saveWeightsFn` not called or no weights captured).\n")
		t.Fatalf("exposeOrchestrator.RunExposeModeForTest did not result in weights being 'saved'.")
	} else {
		reportBuilder.WriteString("\n**Training Outcome:** Completed (Simulated).\n")
		reportBuilder.WriteString(fmt.Sprintf("Simulated 'trained' weights (captured in memory):\n```json\n%s\n```\n\n", helperStructToJSONString(trainedWeights)))
	}

	// --- Step 2: Network Observation (Observe Mode) ---
	reportBuilder.WriteString("## Step 2: Network Observation (Observe Mode for Digit \"1\")\n\n")
	observeCliCfg := config.CLIConfig{
		Mode:           config.ModeObserve,
		WeightsFile:    exposeCliCfg.WeightsFile, // Use the same conceptual path
		Digit:          1,                        // The digit we are "observing"
		CyclesToSettle: 1,
		TotalNeurons:   exposeCliCfg.TotalNeurons, // Consistent network size
		Seed:           exposeCliCfg.Seed,         // Consistent seed
	}
	reportBuilder.WriteString(fmt.Sprintf("### Observe Mode CLI Configuration:\n\n```json\n%s\n```\n\n", helperStructToJSONString(observeCliCfg)))

	observeAppCfg := &config.AppConfig{Cli: observeCliCfg, SimParams: simParams}
	observeOrchestrator := setupOrchestratorForReportTest(t, observeAppCfg, reportBuilder)

	observeOrchestrator.SetLoadWeightsFn(func(filepathStr string) (synaptic.NetworkWeights, error) {
		reportBuilder.WriteString(fmt.Sprintf("**Mock `loadWeightsFn`:** Called for '%s'; returning 'trained' weights captured from expose phase.\n", filepathStr))
		if trainedWeights == nil {
			return nil, fmt.Errorf("mock error: trained weights from previous step are unexpectedly nil")
		}
		return trainedWeights, nil
	})
	// datagen.GetDigitPatternFn is still mocked to return [1,1,1] for the observed digit "1"

	// Assign the "trained" weights to the network instance for observe mode
	observeOrchestrator.Net.SynapticWeights = trainedWeights

	helperCaptureAndFormatOutput("Observe Mode Execution (Digit 1)", observeOrchestrator.RunObserveModeForTest, reportBuilder)
	reportBuilder.WriteString("\n**Observation Outcome:** Completed. Please review console output for output neuron activation.\n")

	// Theoretical activation calculation for the report
	if len(observeOrchestrator.Net.OutputNeuronIDs) > 0 && trainedWeights != nil {
		outputNeuronID := observeOrchestrator.Net.OutputNeuronIDs[0]
		var expectedActivation float64 = 0
		// Pattern [1,1,1] was presented to InputNeuronIDs [0,1,2]
		for _, inputID := range observeOrchestrator.Net.InputNeuronIDs {
			weight := trainedWeights.GetWeight(inputID, outputNeuronID)
			expectedActivation += float64(weight) // Assumes pattern value 1.0 for active inputs
		}
		reportBuilder.WriteString(fmt.Sprintf("\n**Theoretical Output Neuron Activation (ID %d):** %.4f (based on pattern [1,1,1] and captured weights, no decay).\n",
			outputNeuronID, expectedActivation))
	}

	reportBuilder.WriteString("\n---\nEnd of Report.\n")

	// --- Save Report File ---
	reportDir := filepath.Dir(reportFilePath)
	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(reportDir, 0755); mkdirErr != nil {
			t.Fatalf("Failed to create report directory %s: %v", reportDir, mkdirErr)
		}
	}

	err := os.WriteFile(reportFilePath, []byte(reportBuilder.String()), 0644)
	if err != nil {
		t.Fatalf("Failed to write report file %s: %v", reportFilePath, err)
	}
	t.Logf("Digit recognition simulation report generated at: %s", reportFilePath)
}
