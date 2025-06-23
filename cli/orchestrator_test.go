package cli_test

import (
	"bytes"
	"crownet/cli"
	"crownet/common"
	"crownet/config"
	"crownet/datagen"
	"crownet/network"
	"crownet/neuron"
	"crownet/storage"
	"crownet/synaptic"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestOrchestrator creates a new Orchestrator with a basic AppConfig for testing.
// It allows for overriding CLIConfig and SimParams.
func setupTestOrchestrator(t *testing.T, cliCfgOverride *config.CLIConfig, simParamsOverride *config.SimulationParameters) (*cli.Orchestrator, *config.AppConfig) {
	cliCfg := config.CLIConfig{ // Basic defaults, can be overridden
		Mode:         config.ModeSim,
		TotalNeurons: 10,
		Seed:         1, // Fixed seed for reproducibility
	}
	if cliCfgOverride != nil {
		cliCfg = *cliCfgOverride
	}

	simParams := config.DefaultSimulationParameters()
	if simParamsOverride != nil {
		simParams = *simParamsOverride
	}

	appCfg := &config.AppConfig{
		Cli:       cliCfg,
		SimParams: simParams,
	}

	// Validate the constructed AppConfig
	if err := appCfg.Validate(); err != nil {
		t.Fatalf("Failed to create valid AppConfig for test: %v", err)
	}

	return cli.NewOrchestrator(appCfg), appCfg
}

// captureOutput executes a function and captures its stdout and stderr.
// It also captures output sent to the standard log package.
func captureOutput(action func() error) (output string, err error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldLogOutput := log.Writer()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr
	log.SetOutput(wErr) // Redirect log to stderr pipe to capture it along with other errors

	actionErr := action()

	wOut.Close()
	wErr.Close()

	var bufOut bytes.Buffer
	var bufErr bytes.Buffer

	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	os.Stdout = oldStdout
	os.Stderr = oldStderr
	log.SetOutput(oldLogOutput)

	fullOutput := "STDOUT:\n" + bufOut.String() + "\nSTDERR / LOG:\n" + bufErr.String()
	return fullOutput, actionErr
}

func TestSetupContinuousInputStimulus_ErrorCases(t *testing.T) {
	simParamsDefault := config.DefaultSimulationParameters()
	baseCliCfg := config.CLIConfig{
		StimInputFreqHz: 10.0,
		Mode:            config.ModeSim,
		TotalNeurons:    20, // Ensure enough neurons for various test cases
		Seed:            123,
	}

	testCases := []struct {
		name            string
		modifyCliCfg    func(cfg *config.CLIConfig)
		networkInputIDs []common.NeuronID
		expectedError   bool
		errorContains   string
	}{
		{
			name:            "Valid ID",
			modifyCliCfg:    func(cfg *config.CLIConfig) { cfg.StimInputID = 10 },
			networkInputIDs: []common.NeuronID{10, 11},
			expectedError:   false,
		},
		{
			name:            "ID -1 (first available)",
			modifyCliCfg:    func(cfg *config.CLIConfig) { cfg.StimInputID = -1 },
			networkInputIDs: []common.NeuronID{10, 11},
			expectedError:   false,
		},
		{
			name:            "Invalid ID (not in network input IDs)",
			modifyCliCfg:    func(cfg *config.CLIConfig) { cfg.StimInputID = 99 },
			networkInputIDs: []common.NeuronID{10, 11},
			expectedError:   true,
			errorContains:   "não encontrado ou inválido",
		},
		{
			name:            "ID -1 with no input neurons in network",
			modifyCliCfg:    func(cfg *config.CLIConfig) { cfg.StimInputID = -1 },
			networkInputIDs: []common.NeuronID{},
			expectedError:   false, // Should not error, just won't configure anything
		},
		// Note: Testing ConfigureFrequencyInput error directly is harder without network interface mocking.
		// This test primarily covers orchestrator's logic for ID validation.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentCliCfg := baseCliCfg
			if tc.modifyCliCfg != nil {
				tc.modifyCliCfg(&currentCliCfg)
			}

			appCfg := &config.AppConfig{Cli: currentCliCfg, SimParams: simParamsDefault}
			// We need a real network to test ConfigureFrequencyInput's ID validation path
			net := network.NewCrowNet(currentCliCfg.TotalNeurons, common.Rate(currentCliCfg.BaseLearningRate), &simParamsDefault, currentCliCfg.Seed)
			net.InputNeuronIDs = tc.networkInputIDs // Override for the test case

			orchestrator := cli.NewOrchestrator(appCfg)
			orchestrator.Net = net // Assign the specifically configured network

			err := orchestrator.SetupContinuousInputStimulusForTest()

			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}

func TestExposeMode_RunAndSave(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	// Simplify pattern params for easier mocking if datagen is not the focus
	simParams.PatternHeight = 1
	simParams.PatternWidth = 1
	simParams.PatternSize = 1
	simParams.MinInputNeurons = 1 // Critical for PresentPattern
	simParams.MinOutputNeurons = 1

	tempDir := t.TempDir()
	weightsFilePath := filepath.Join(tempDir, "expose_test_weights.json")

	cliCfg := config.CLIConfig{
		Mode:             config.ModeExpose,
		WeightsFile:      weightsFilePath,
		Epochs:           1,
		CyclesPerPattern: 1,
		TotalNeurons:     2, // Min 1 input, 1 output
		BaseLearningRate: 0.01,
		Seed:             777,
	}
	orchestrator, appCfg := setupTestOrchestrator(t, &cliCfg, &simParams)

	// Mock persistence and data generation
	saveWeightsCalled := false
	var savedFilePath string
	orchestrator.SetSaveWeightsFn(func(weights synaptic.NetworkWeights, filepathStr string) error {
		saveWeightsCalled = true
		savedFilePath = filepathStr
		return nil // Simulate successful save
	})
	orchestrator.SetLoadWeightsFn(func(filepathStr string) (synaptic.NetworkWeights, error) {
		return nil, fmt.Errorf("mock load: weights file not found (expected for new training)")
	})
	originalGetDigitPatternFn := datagen.GetDigitPatternFn
	datagen.GetDigitPatternFn = func(digit int, sp *config.SimulationParameters) ([]float64, error) {
		if sp.PatternSize != 1 { // Ensure params are passed down
			return nil, fmt.Errorf("mock GetDigitPatternFn: PatternSize expected 1, got %d", sp.PatternSize)
		}
		return []float64{1.0}, nil // Simple 1x1 pattern
	}
	defer func() { datagen.GetDigitPatternFn = originalGetDigitPatternFn }()

	// Orchestrator.Run() calls createNetwork(), so o.Net will be initialized.
	// We need to ensure the created network has correct Input/Output neuron IDs for PresentPattern.
	// This is tricky as NewCrowNet internally sets these up.
	// For more control, tests might need to directly create and assign a pre-configured network.
	// However, RunExposeModeForTest calls createNetwork.

	// To ensure PresentPattern works, we let createNetwork run, then adjust Input/Output IDs if necessary
	// This is a bit of a workaround for not mocking the network interface.
	orchestrator.CreateNetworkForTest() // Manually call to allow modification before RunExposeMode
	if len(orchestrator.Net.Neurons) > 0 && simParams.MinInputNeurons > 0 {
		orchestrator.Net.InputNeuronIDs = []common.NeuronID{orchestrator.Net.Neurons[0].ID}
		orchestrator.Net.Neurons[0].Type = neuron.Input
	}
	if len(orchestrator.Net.Neurons) > 1 && simParams.MinOutputNeurons > 0 {
		orchestrator.Net.OutputNeuronIDs = []common.NeuronID{orchestrator.Net.Neurons[1].ID}
		orchestrator.Net.Neurons[1].Type = neuron.Output
	}


	_, err := captureOutput(orchestrator.RunExposeModeForTest)
	if err != nil {
		t.Fatalf("RunExposeModeForTest returned an unexpected error: %v", err)
	}

	if !saveWeightsCalled {
		t.Errorf("Expected saveWeightsFn to be called, but it wasn't")
	}
	if savedFilePath != weightsFilePath {
		t.Errorf("saveWeightsFn called with incorrect filepath: expected '%s', got '%s'", weightsFilePath, savedFilePath)
	}
}

func TestSimMode_DBCreation(t *testing.T) {
	tempDir := t.TempDir()
	dbFilePath := filepath.Join(tempDir, "sim_mode_test.db")
	// Ensure DB doesn't exist from a previous failed run
	os.Remove(dbFilePath)

	cliCfg := config.CLIConfig{
		Mode:         config.ModeSim,
		Cycles:       1, // Minimal cycles
		TotalNeurons: 10,
		DbPath:       dbFilePath,
		SaveInterval: 1, // Save on every cycle for this test
		Seed:         123,
	}
	orchestrator, _ := setupTestOrchestrator(t, &cliCfg, nil)

	// Run the sim mode which should trigger logger initialization and logging.
	// The test wrappers are used to isolate parts of Run() if needed,
	// but here we test the integrated path leading to DB creation.

	// Initialize components as done in Orchestrator.Run() before mode execution
	err := orchestrator.InitializeLoggerForTest()
	if err != nil {
		t.Fatalf("InitializeLoggerForTest() failed: %v", err)
	}
	defer orchestrator.CloseLoggerForTest()
	orchestrator.CreateNetworkForTest()


	_, err = captureOutput(orchestrator.RunSimModeForTest)
	if err != nil {
		t.Fatalf("RunSimModeForTest returned an unexpected error: %v", err)
	}

	// Verify DB file creation
	if _, statErr := os.Stat(dbFilePath); os.IsNotExist(statErr) {
		t.Errorf("SQLite database file %s was not created", dbFilePath)
	} else if statErr != nil {
		t.Errorf("Error stating DB file %s: %v", dbFilePath, statErr)
	}
	// Further checks could involve querying the DB for expected tables/data.
}


func TestObserveMode_OutputVerification(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.MinOutputNeurons = 2
	simParams.MinInputNeurons = 3
	simParams.PatternSize = 3 // Matches pattern [1,1,0]
	simParams.PatternHeight = 1
	simParams.PatternWidth = 3
	simParams.AccumulatedPulseDecayRate = 0.0 // No decay for predictable potential sum
	simParams.BaseFiringThreshold = 100.0 // High threshold to prevent firing during observation

	tempDir := t.TempDir()
	weightsFilePath := filepath.Join(tempDir, "observe_test_weights.json")

	// Create mock weights: Input 0 -> Output 0 (ID 3), Input 1 -> Output 1 (ID 4)
	mockWeights := synaptic.NewNetworkWeights()
	// Neuron IDs: Input 0,1,2; Output 3,4 (assuming sequential assignment by NewCrowNet)
	mockWeights.SetWeight(common.NeuronID(0), common.NeuronID(3), 1.0, &simParams) // I0 -> O0
	mockWeights.SetWeight(common.NeuronID(0), common.NeuronID(4), 0.5, &simParams) // I0 -> O1
	mockWeights.SetWeight(common.NeuronID(1), common.NeuronID(3), 0.2, &simParams) // I1 -> O0
	mockWeights.SetWeight(common.NeuronID(1), common.NeuronID(4), 1.0, &simParams) // I1 -> O1
	// Input 2 is not connected to Output 0 or 1 in this mock setup for simplicity.

	if err := storage.SaveNetworkWeightsToJSON(mockWeights, weightsFilePath); err != nil {
		t.Fatalf("Failed to save mock weights file: %v", err)
	}

	cliCfg := config.CLIConfig{
		Mode:           config.ModeObserve,
		WeightsFile:    weightsFilePath,
		Digit:          0, // Digit to observe (mocked pattern will be [1,1,0])
		CyclesToSettle: 1, // Minimal settling
		TotalNeurons:   simParams.MinInputNeurons + simParams.MinOutputNeurons, // 3 input + 2 output
		Seed:           12345,
	}
	orchestrator, appCfg := setupTestOrchestrator(t, &cliCfg, &simParams)

	// Mock GetDigitPatternFn to return a specific pattern
	originalGetDigitPatternFn := datagen.GetDigitPatternFn
	datagen.GetDigitPatternFn = func(digit int, sp *config.SimulationParameters) ([]float64, error) {
		if sp.PatternSize != 3 {
			return nil, fmt.Errorf("mock GetDigitPatternFn: PatternSize expected 3, got %d", sp.PatternSize)
		}
		return []float64{1.0, 1.0, 0.0}, nil // Input neurons 0 and 1 active
	}
	defer func() { datagen.GetDigitPatternFn = originalGetDigitPatternFn }()

	// Orchestrator.Run() would call createNetwork and loadWeights.
	// For RunObserveModeForTest, we need to ensure these are done.
	orchestrator.CreateNetworkForTest() // Populates o.Net

	// Manually assign neuron IDs if NewCrowNet's assignment is not perfectly predictable for test
	// This ensures that neuron IDs 0,1,2 are inputs and 3,4 are outputs.
	for i := 0; i < appCfg.SimParams.MinInputNeurons; i++ {
		orchestrator.Net.Neurons[i].ID = common.NeuronID(i)
		orchestrator.Net.Neurons[i].Type = neuron.Input
	}
	orchestrator.Net.InputNeuronIDs = []common.NeuronID{0,1,2}
	for i := 0; i < appCfg.SimParams.MinOutputNeurons; i++ {
		offsetIdx := appCfg.SimParams.MinInputNeurons + i
		orchestrator.Net.Neurons[offsetIdx].ID = common.NeuronID(offsetIdx)
		orchestrator.Net.Neurons[offsetIdx].Type = neuron.Output
	}
	orchestrator.Net.OutputNeuronIDs = []common.NeuronID{3,4}


	if err := orchestrator.LoadWeightsForTest(weightsFilePath); err != nil {
		t.Fatalf("LoadWeightsForTest failed: %v", err)
	}

	output, err := captureOutput(orchestrator.RunObserveModeForTest)
	if err != nil {
		t.Fatalf("RunObserveModeForTest failed: %v\nOutput: %s", err, output)
	}

	// Expected activations based on pattern [1,1,0] and weights:
	// OutputNeuron[0] (ID 3): (Input 0 active * weight I0->O0) + (Input 1 active * weight I1->O0)
	//                      = (1.0 * 1.0) + (1.0 * 0.2) = 1.2
	// OutputNeuron[1] (ID 4): (Input 0 active * weight I0->O1) + (Input 1 active * weight I1->O1)
	//                      = (1.0 * 0.5) + (1.0 * 1.0) = 1.5
	expectedOutputID0 := orchestrator.Net.OutputNeuronIDs[0] // Should be ID 3
	expectedOutputID1 := orchestrator.Net.OutputNeuronIDs[1] // Should be ID 4

	expectedPatternStrings := []string{
		fmt.Sprintf("OutNeurônio[0] (ID %d): %.4f", expectedOutputID0, 1.2000),
		fmt.Sprintf("OutNeurônio[1] (ID %d): %.4f", expectedOutputID1, 1.5000),
		"Dígito Apresentado: 0",
	}
	for _, substr := range expectedPatternStrings {
		if !strings.Contains(output, substr) {
			t.Errorf("Output missing expected string '%s'. Full output:\n%s", substr, output)
		}
	}
}

func TestObserveMode_LoadWeightsError(t *testing.T) {
	cliCfg := config.CLIConfig{
		Mode:        config.ModeObserve,
		WeightsFile: "non_existent_weights.json", // This file won't exist
		Digit:       0,
		Seed:        1,
	}
	orchestrator, appCfg := setupTestOrchestrator(t, &cliCfg, nil)

	// Orchestrator.Run() calls createNetwork.
	// For RunObserveModeForTest, ensure network exists.
	orchestrator.Net = network.NewCrowNet(appCfg.Cli.TotalNeurons, 0.01, &appCfg.SimParams, appCfg.Cli.Seed)


	err := orchestrator.RunObserveModeForTest()
	if err == nil {
		t.Errorf("Expected RunObserveModeForTest to fail due to missing weights file, but it succeeded.")
	} else {
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "Exponha a rede primeiro") {
			t.Errorf("Expected error message to indicate missing weights, got: %v", err)
		}
	}
}

func TestExposeMode_SaveWeightsError(t *testing.T) {
	tempDir := t.TempDir()
	// Attempt to save to a read-only path or similar to trigger error (hard to do reliably cross-platform)
	// Instead, we'll mock saveWeightsFn to return an error.
	weightsFilePath := filepath.Join(tempDir, "save_error_test.json")

	cliCfg := config.CLIConfig{
		Mode:        config.ModeExpose,
		WeightsFile: weightsFilePath,
		Epochs:      1, // Minimal epochs
		TotalNeurons: 2,  // Minimal neurons
		Seed:        1,
	}
	orchestrator, appCfg := setupTestOrchestrator(t, &cliCfg, nil)

	mockSaveError := fmt.Errorf("simulated saveWeights error")
	orchestrator.SetSaveWeightsFn(func(weights synaptic.NetworkWeights, filepathStr string) error {
		return mockSaveError
	})
	// Mock load to succeed or not be called, as focus is on save error
	orchestrator.SetLoadWeightsFn(func(filepathStr string) (synaptic.NetworkWeights, error) {
		return synaptic.NewNetworkWeights(), nil
	})
	// Mock datagen to avoid issues with pattern generation if it's not robust
	originalGetDigitPatternFn := datagen.GetDigitPatternFn
	datagen.GetDigitPatternFn = func(digit int, sp *config.SimulationParameters) ([]float64, error) {
		return []float64{1.0}, nil // Assuming MinInputNeurons=1, PatternSize=1 for simplicity
	}
	defer func() { datagen.GetDigitPatternFn = originalGetDigitPatternFn }()

	// Ensure network is created with minimal setup for expose mode to run
	appCfg.SimParams.MinInputNeurons = 1
	appCfg.SimParams.MinOutputNeurons = 1
	appCfg.SimParams.PatternSize = 1
	orchestrator.AppCfg = appCfg // Update orchestrator's appCfg if modified
	orchestrator.CreateNetworkForTest()
	if len(orchestrator.Net.Neurons) > 0 { // Ensure input neuron exists for PresentPattern
		orchestrator.Net.InputNeuronIDs = []common.NeuronID{orchestrator.Net.Neurons[0].ID}
		orchestrator.Net.Neurons[0].Type = neuron.Input
	}


	err := orchestrator.RunExposeModeForTest()
	if err == nil {
		t.Errorf("Expected RunExposeModeForTest to fail due to saveWeights error, but it succeeded.")
	} else {
		if !strings.Contains(err.Error(), "simulated saveWeights error") {
			t.Errorf("Expected error message to indicate save failure, got: %v", err)
		}
	}
}
