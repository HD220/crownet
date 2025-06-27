package cmd

import (
	"bytes" // Needed for buffer
	"io"    // Needed for MultiWriter and pipe reading
	"os"    // Needed for stdout capture
	"path/filepath"
	"strings" // Needed for output assertions
	"testing"
	"time"

	"crownet/cli"
	"crownet/config"
	// "crownet/common" // Not directly needed for this basic AppConfig setup
)

// Helper function to create a minimal AppConfig for observe tests
func newTestObserveAppConfig(weightsFilePath string) *config.AppConfig {
	// Ensure SimParams are consistent with a network that could have generated/used the fixture weights.
	// The fixture is very sparse, so exact neuron counts in SimParams for generation aren't critical,
	// but TotalNeurons in CLIConfig should be enough for what the weights file implies (e.g., IDs up to 45+).
	// Using 50 neurons as a default consistent with other tests.
	simParams := config.DefaultSimulationParameters()
	// If fixture_observe_weights.json implies a specific network structure (e.g. number of I/O neurons),
	// those should be reflected in simParams if they affect loading or observation logic.
	// For now, defaults are used. MinInputNeurons=35, MinOutputNeurons=10.

	return &config.AppConfig{
		SimParams: simParams,
		Cli: config.CLIConfig{
			Mode:           config.ModeObserve,
			TotalNeurons:   50, // Should be consistent with the network that would use/generate the fixture.
			Seed:           time.Now().UnixNano(),
			WeightsFile:    weightsFilePath,
			Digit:          1, // Observe digit '1'
			CyclesToSettle: 1, // Minimal settling cycles
			// DebugChem can be default false
		},
	}
}

func TestObserveCommand_BasicRun(t *testing.T) {
	// 1. Define path to the fixture weights file.
	// Assuming 'testdata' is relative to the package being tested (cmd).
	// For `go test ./...` from repo root, this path needs to be relative to repo root.
	// Or, use runtime path resolution if tests are run from different dirs.
	// For simplicity, assuming `go test ./cmd` or `go test ./...` from root.
	// If running `go test ./cmd`, `testdata` should be `../testdata`.
	// If running from repo root, `testdata` is correct.
	// Let's make it relative to the test file's location for robustness if possible,
	// but that's tricky. For now, assume relative to repo root for `make test`.
	// When running `go test ./cmd/...` or `make test` (which does `go test ./...`),
	// the working directory for tests in package `cmd` is usually `<repo_root>/cmd`.
	// So, to reach `<repo_root>/testdata/`, the path should be `../testdata/`.
	fixtureWeightsPath := filepath.Join("..", "testdata", "fixture_observe_weights.json")

	// This test doesn't create temp dirs as observe mode doesn't write files by default.
	// If it were to log to DB, then a tempDbPath would be needed.

	// 2. Construct an AppConfig for a minimal observe run
	appCfg := newTestObserveAppConfig(fixtureWeightsPath)

	// Validate the constructed AppConfig
	if err := appCfg.Validate(); err != nil {
		t.Fatalf("Constructed AppConfig is invalid: %v", err)
	}

	// 3. Create an orchestrator
	orchestrator := cli.NewOrchestrator(appCfg)

	// 4. Run the orchestrator
	// We are primarily checking if the observe mode runs to completion without panic/error
	// using the fixture weights.
	err := orchestrator.Run()

	// 5. Assert that no error is returned
	if err != nil {
		t.Fatalf("Orchestrator.Run() for observe mode failed: %v", err)
	}

	// Optional: Check for basic console output.
	// This is harder to do reliably without redirecting stdout.
	// For TSK-TEST-003.2.1, just ensuring it runs is the primary goal.
	// t.Log("Observe mode ran successfully. Console output validation is for TSK-TEST-003.2.2")

	// --- Start of TSK-TEST-003.2.2 additions ---

	// Capture output for validation.
	// Store original stdout.
	originalStdout := os.Stdout
	rPipe, wPipe, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("Failed to create pipe for stdout capture: %v", pipeErr)
	}
	os.Stdout = wPipe

	captureOut := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, errCopy := io.Copy(&buf, rPipe) // Capture error from io.Copy
		if errCopy != nil {
			// Not t.Fatalf as this is in a goroutine. Log or send error through channel.
			// For simplicity in this test, we'll log and proceed.
			// A more robust solution might involve passing errors via the channel.
			t.Logf("Error copying from pipe to buffer: %v", errCopy)
		}
		captureOut <- buf.String()
	}()

	// Create a new orchestrator instance for the capture run,
	// as the state of the previous 'orchestrator' might have been affected
	// or might not be suitable for a re-run if it holds state across Run() calls.
	orchestratorForCapture := cli.NewOrchestrator(appCfg)
	runErrForCapture := orchestratorForCapture.Run()

	wPipe.Close()
	os.Stdout = originalStdout // Restore stdout
	capturedStr := <-captureOut

	if runErrForCapture != nil {
		t.Fatalf("Orchestrator.Run() for observe mode (capture run) failed: %v. Captured output:\n%s",
			runErrForCapture, capturedStr)
	}

	// Perform assertions on capturedStr
	// a. Check for header
	expectedHeader := "Output Neuron Activation Pattern"
	if !strings.Contains(capturedStr, expectedHeader) {
		t.Errorf("Expected output to contain header '%s', but it was not found.\nCaptured output:\n%s",
			expectedHeader, capturedStr)
	}

	// b. Check for number of output neuron lines
	// Default SimParams: MinOutputNeurons = 10
	// The actual number of output neurons created is MinOutputNeurons.
	expectedOutputLines := appCfg.SimParams.Structure.MinOutputNeurons
	outputNeuronLineCount := 0
	lines := strings.Split(capturedStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "OutputNeuron[") {
			outputNeuronLineCount++
		}
	}
	if outputNeuronLineCount != expectedOutputLines {
		t.Errorf("Expected %d output neuron lines, got %d.\nCaptured output:\n%s",
			expectedOutputLines, outputNeuronLineCount, capturedStr)
	}

	// c. Check for ASCII bar format (basic check for presence of key chars)
	hasBarFormat := false
	for _, line := range lines {
		if strings.Contains(line, "OutputNeuron[") && strings.Contains(line, "[") &&
			strings.Contains(line, "]") && strings.Contains(line, "|") {
			hasBarFormat = true
			break
		}
	}
	if !hasBarFormat && expectedOutputLines > 0 { // Only expect bar format if there are output lines
		t.Errorf("Expected ASCII bar format in output neuron lines, but it was not found.\nCaptured output:\n%s", capturedStr)
	}
	// --- End of TSK-TEST-003.2.2 additions ---
}
