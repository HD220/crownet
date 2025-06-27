package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time" // For unique temp dir names, though t.TempDir() handles this

	"crownet/cli"
	"crownet/common"
	"crownet/config"
)

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
			TotalNeurons:     50,    // Increased from 10 to be >= MinInput (35) + MinOutput (10)
			Seed:             time.Now().UnixNano(), // Use a unique seed
			WeightsFile:      tempWeightsFilePath,
			BaseLearningRate: common.Rate(0.01),
			Epochs:           1,     // Minimal epochs
			CyclesPerPattern: 1,     // Minimal cycles per pattern
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

	// t.Cleanup() will automatically handle removal of tempDir if t.TempDir() was used.
	// If manual temp dir creation was used, os.RemoveAll(tempDir) would be needed in t.Cleanup().
}
