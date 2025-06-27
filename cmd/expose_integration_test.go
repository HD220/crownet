package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time" // For unique temp dir names, though t.TempDir() handles this

	"encoding/json" // For JSON validation

	"crownet/cli"
	"crownet/common"
	"crownet/config"
)

// Helper function to create a minimal AppConfig for expose tests
func newTestExposeAppConfig(tempDir string, weightsFileName string) *config.AppConfig {
	tempWeightsFilePath := filepath.Join(tempDir, weightsFileName)
	return &config.AppConfig{
		SimParams: config.DefaultSimulationParameters(),
		Cli: config.CLIConfig{
			Mode:             config.ModeExpose,
			TotalNeurons:     50, // Satisfies MinInput (35) + MinOutput (10)
			Seed:             time.Now().UnixNano(),
			WeightsFile:      tempWeightsFilePath,
			BaseLearningRate: common.Rate(0.01),
			Epochs:           1,
			CyclesPerPattern: 1,
		},
	}
}

func TestExposeCommand_BasicRun(t *testing.T) {
	// 1. Create a temporary directory for test artifacts
	tempDir := t.TempDir()

	// 2. Define path for a temporary weights file
	weightsFileName := "test_expose_weights.json"
	tempWeightsFilePath := filepath.Join(tempDir, weightsFileName)

	// 3. Construct an AppConfig for a minimal expose run
	appCfg := &config.AppConfig{
		SimParams: config.DefaultSimulationParameters(), // Use default simulation parameters
		Cli: config.CLIConfig{
			Mode:             config.ModeExpose,
			TotalNeurons:     50,                    // Increased from 10 to be >= MinInput (35) + MinOutput (10)
			Seed:             time.Now().UnixNano(), // Use a unique seed
			WeightsFile:      tempWeightsFilePath,
			BaseLearningRate: common.Rate(0.01),
			Epochs:           1, // Minimal epochs
			CyclesPerPattern: 1, // Minimal cycles per pattern
			// DbPath, SaveInterval, DebugChem can be left as default (empty/0/false)
			// as they are not critical for this basic execution test.
		},
	}

	// Validate the constructed AppConfig (optional here, but good practice)
	if err := appCfg.Validate(); err != nil {
		t.Fatalf("Constructed AppConfig is invalid: %v", err)
	}

	// 4. Create an orchestrator
	// Note: NewOrchestrator defaults to using actual storage functions.
	// For this basic execution test, that's acceptable.
	// For more advanced tests (like TEST-003.1.2), mocking might be needed.
	orchestrator := cli.NewOrchestrator(appCfg)

	// 5. Run the orchestrator
	// We are primarily checking if the expose mode runs to completion without panic/error.
	// Console output capture is optional as per task TSK-TEST-003.1.1 and can be complex.
	err := orchestrator.Run()

	// 6. Assert that no error is returned
	if err != nil {
		t.Fatalf("Orchestrator.Run() for expose mode failed: %v", err)
	}

	// 7. Check if the weights file was created (basic check for this task)
	// More detailed validation of the weights file is for TSK-TEST-003.1.2
	if _, errStat := os.Stat(tempWeightsFilePath); os.IsNotExist(errStat) {
		t.Errorf("Expected weights file to be created at %s, but it was not found.", tempWeightsFilePath)
	}
	// BasicRun already covers creation, more detailed checks in specific tests below.
}

func TestExposeCommand_NewWeightsFileCreation(t *testing.T) {
	tempDir := t.TempDir()
	weightsFileName := "new_weights.json"
	tempWeightsFilePath := filepath.Join(tempDir, weightsFileName)

	appCfg := newTestExposeAppConfig(tempDir, weightsFileName)

	orchestrator := cli.NewOrchestrator(appCfg)
	err := orchestrator.Run()
	if err != nil {
		t.Fatalf("Orchestrator.Run() for new weights file creation failed: %v", err)
	}

	// 1. Check existence
	fileInfo, errStat := os.Stat(tempWeightsFilePath)
	if os.IsNotExist(errStat) {
		t.Fatalf("Expected weights file '%s' to be created, but it was not found.", tempWeightsFilePath)
	}
	if errStat != nil {
		t.Fatalf("Error stating weights file '%s': %v", tempWeightsFilePath, errStat)
	}

	// 2. Check not empty
	if fileInfo.Size() == 0 {
		t.Errorf("Expected weights file '%s' to be non-empty, but it was empty.", tempWeightsFilePath)
	}

	// 3. Check valid JSON
	content, errRead := os.ReadFile(tempWeightsFilePath)
	if errRead != nil {
		t.Fatalf("Failed to read created weights file '%s': %v", tempWeightsFilePath, errRead)
	}
	var jsonData interface{} // Use interface{} for generic JSON validation
	if errJSON := json.Unmarshal(content, &jsonData); errJSON != nil {
		t.Errorf("Expected weights file '%s' to contain valid JSON, but unmarshal failed: %v", tempWeightsFilePath, errJSON)
	}
}

func TestExposeCommand_ModifyWeightsFile(t *testing.T) {
	tempDir := t.TempDir()
	weightsFileName := "existing_weights.json"
	tempWeightsFilePath := filepath.Join(tempDir, weightsFileName)

	// 1. Create a dummy initial weights file
	initialContent := []byte(`{"0":{"1":0.5}}`) // Simple valid JSON
	errWrite := os.WriteFile(tempWeightsFilePath, initialContent, 0644)
	if errWrite != nil {
		t.Fatalf("Failed to create initial dummy weights file '%s': %v", tempWeightsFilePath, errWrite)
	}
	initialFileInfo, statErr := os.Stat(tempWeightsFilePath)
	if statErr != nil {
		t.Fatalf("Failed to stat initial dummy weights file '%s': %v", tempWeightsFilePath, statErr)
	}
	initialModTime := initialFileInfo.ModTime()

	// Ensure there's a slight delay for ModTime to change reliably
	time.Sleep(10 * time.Millisecond)

	appCfg := newTestExposeAppConfig(tempDir, weightsFileName)
	// Potentially reduce neuron count further if possible, or ensure SimParams match dummy file
	// For this test, we primarily care that it *overwrites* or *modifies*.
	// The default SimParams might generate a different structure if neuron counts don't align.
	// Let's use a small number of neurons consistent with the dummy file structure if needed.
	// For simplicity, we'll assume the new run with 50 neurons will overwrite it completely.

	orchestrator := cli.NewOrchestrator(appCfg)
	err := orchestrator.Run()
	if err != nil {
		t.Fatalf("Orchestrator.Run() for modifying weights file failed: %v", err)
	}

	// 2. Check modification
	finalFileInfo, errStat := os.Stat(tempWeightsFilePath)
	if os.IsNotExist(errStat) {
		t.Fatalf("Expected weights file '%s' to still exist, but it was not found.", tempWeightsFilePath)
	}
	if errStat != nil {
		t.Fatalf("Error stating weights file '%s' after run: %v", tempWeightsFilePath, errStat)
	}

	if !finalFileInfo.ModTime().After(initialModTime) {
		t.Errorf("Expected weights file '%s' to be modified (modTime after initial), but it was not. Initial: %v, Final: %v",
			tempWeightsFilePath, initialModTime, finalFileInfo.ModTime())
	}

	// 3. Check not empty (it should be overwritten with new weights)
	if finalFileInfo.Size() == 0 {
		t.Errorf("Expected modified weights file '%s' to be non-empty.", tempWeightsFilePath)
	}

	// 4. Check valid JSON
	finalContent, errRead := os.ReadFile(tempWeightsFilePath)
	if errRead != nil {
		t.Fatalf("Failed to read modified weights file '%s': %v", tempWeightsFilePath, errRead)
	}
	var jsonData interface{}
	if errJSON := json.Unmarshal(finalContent, &jsonData); errJSON != nil {
		t.Errorf("Expected modified weights file '%s' to contain valid JSON, but unmarshal failed: %v",
			tempWeightsFilePath, errJSON)
	}

	// Optional: Check content different from initial (could be tricky if format changes but values are similar)
	if string(finalContent) == string(initialContent) && appCfg.Cli.Epochs > 0 { // Only expect change if epochs ran
		// This check might be flaky if the "training" with 1 epoch/1 cycle doesn't change weights much
		// or if the number of neurons (50) results in a vastly different structure than the dummy.
		// A more robust check would be to load both JSONs and compare structures/values if necessary.
		// For now, ModTime is a strong indicator.
		t.Logf("Warning: Content of weights file '%s' did not change after running expose. "+
			"This might be okay for minimal run, but check if intended.", tempWeightsFilePath)
	}
}
