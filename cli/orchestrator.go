// Package cli provides the command-line interface (CLI) orchestrator for the
// CrowNet simulation. It interprets CLI arguments, sets up the simulation
// environment, and manages the execution flow for different modes of operation
// (e.g., simulation, training, observation).
package cli

import (
	"crownet/common"
	"crownet/config"
	"crownet/datagen"
	"crownet/network"
	"crownet/storage" // For JSON persistence and SQLite logging
	"crownet/synaptic" // For synaptic.NetworkWeights type in function signatures
	"fmt"
	"log"
	"os"
	"path/filepath" // Added for path validation
	"strings"       // Added for path validation messages
	"time"
)

// Orchestrator manages the simulation execution based on CLI configuration.
type Orchestrator struct {
	AppCfg *config.AppConfig
	Net    *network.CrowNet
	Logger *storage.SQLiteLogger

	// loadWeightsFn and saveWeightsFn allow for mocking persistence operations in tests.
	loadWeightsFn func(filepath string) (synaptic.NetworkWeights, error)
	saveWeightsFn func(weights synaptic.NetworkWeights, filepath string) error
}

// NewOrchestrator creates a new orchestrator with the given application configuration.
// It defaults to using actual file system operations for loading/saving weights.
func NewOrchestrator(appCfg *config.AppConfig) *Orchestrator {
	return &Orchestrator{
		AppCfg:        appCfg,
		loadWeightsFn: storage.LoadNetworkWeightsFromJSON,
		saveWeightsFn: storage.SaveNetworkWeightsToJSON,
	}
}

// Run executes the selected simulation mode. It's the main entry point for the orchestrator.
func (o *Orchestrator) Run() error {
	fmt.Println("CrowNet Initializing...")
	fmt.Printf("Selected Mode: %s\n", o.AppCfg.Cli.Mode)
	fmt.Printf("Base Configuration: Neurons=%d, WeightsFile='%s'\n",
		o.AppCfg.Cli.TotalNeurons, o.AppCfg.Cli.WeightsFile)
	o.printModeSpecificConfig()

	if err := o.initializeLogger(); err != nil {
		return fmt.Errorf("logger initialization failed: %w", err)
	}
	if o.Logger != nil {
		defer func() {
			if errClose := o.Logger.Close(); errClose != nil {
				// Log error but don't override a primary error from Run()
				log.Printf("Error closing SQLite logger: %v", errClose)
			}
		}()
	}

	o.createNetwork()

	startTime := time.Now()
	var errRun error

	switch o.AppCfg.Cli.Mode {
	case config.ModeSim:
		errRun = o.runSimMode()
	case config.ModeExpose:
		errRun = o.runExposeMode()
	case config.ModeObserve:
		errRun = o.runObserveMode()
	case config.ModeLogUtil: // FEATURE-004
		errRun = o.runLogUtilMode()
	default:
		// This case should ideally be caught by AppConfig.Validate()
		return fmt.Errorf("unknown or unsupported mode in Orchestrator.Run: %s", o.AppCfg.Cli.Mode)
	}

	if errRun != nil {
		return fmt.Errorf("error during execution of mode '%s': %w", o.AppCfg.Cli.Mode, errRun)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nCrowNet session finished. Total duration: %s.\n", duration)
	return nil
}

// initializeLogger sets up the SQLite logger if configured.
func (o *Orchestrator) initializeLogger() error {
	cfg := &o.AppCfg.Cli
	// Logging is active for 'sim' mode, or 'expose' mode if periodic saving to DB is enabled.
	if cfg.DbPath != "" && (cfg.Mode == config.ModeSim || (cfg.Mode == config.ModeExpose && cfg.SaveInterval > 0)) {
		validatedDbPath, err := o.validatePath(cfg.DbPath, false) // false: for writing
		if err != nil {
			// Allow DbPath to be empty if not in sim mode or expose with saveInterval
			// This condition might be redundant if cfg.DbPath == "" is checked first,
			// but validatePath itself checks for empty path.
			// If DbPath was provided but invalid, it's an error.
			return fmt.Errorf("invalid DbPath '%s': %w", cfg.DbPath, err)
		}
		cfg.DbPath = validatedDbPath // Update with cleaned, absolute path

		o.Logger, err = storage.NewSQLiteLogger(cfg.DbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize SQLite logger at %s: %w", cfg.DbPath, err)
		}
		fmt.Printf("SQLite logging enabled: %s\n", cfg.DbPath)
	}
	return nil
}

// validatePath cleans, absolutizes, and performs basic checks on a file path.
// - rawPath: the user-provided path string.
// - forRead: true if the path is intended for reading, false for writing.
// Returns the cleaned, absolute path or an error if validation fails.
func (o *Orchestrator) validatePath(rawPath string, forRead bool) (string, error) {
	if strings.TrimSpace(rawPath) == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Clean the path to resolve ".." etc.
	cleanedPath := filepath.Clean(rawPath)

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanedPath)
	if err != nil {
		return "", fmt.Errorf("could not determine absolute path for '%s': %w", cleanedPath, err)
	}

	// At this point, absPath is a cleaned, absolute path.
	// Further checks for path traversal vulnerabilities could involve ensuring
	// it's within a known-good base directory, but for a general CLI tool,
	// this is harder to enforce without more context or a sandbox.
	// The TSK-SEC-001 asks to "Consider if there is a need to restrict... to a subdiretory".
	// For now, we'll focus on existence and basic type checks.

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			if forRead {
				return "", fmt.Errorf("path '%s' (resolved to '%s') does not exist", rawPath, absPath)
			}
			// If for writing, the path itself might not exist, but its parent directory must.
			parentDir := filepath.Dir(absPath)
			parentInfo, parentErr := os.Stat(parentDir)
			if parentErr != nil {
				if os.IsNotExist(parentErr) {
					return "", fmt.Errorf("parent directory for '%s' (resolved to '%s') does not exist", rawPath, parentDir)
				}
				return "", fmt.Errorf("could not stat parent directory '%s' for path '%s': %w", parentDir, rawPath, parentErr)
			}
			if !parentInfo.IsDir() {
				return "", fmt.Errorf("parent path '%s' for '%s' is not a directory", parentDir, rawPath)
			}
			// Note: Checking actual write permissions is complex and OS-dependent.
			// We rely on the OS to prevent writing to unauthorized locations during the actual write operation.
			return absPath, nil // Parent exists and is a dir, path is OK for writing.
		}
		return "", fmt.Errorf("could not stat path '%s' (resolved to '%s'): %w", rawPath, absPath, err)
	}

	// Path exists.
	if forRead {
		if fileInfo.IsDir() {
			return "", fmt.Errorf("path '%s' (resolved to '%s') is a directory, expected a file for reading", rawPath, absPath)
		}
		// Note: Actual read permission check is typically handled by os.Open().
	} else { // For writing (e.g. weights file, potentially overwriting)
		if fileInfo.IsDir() {
			// This case is for dbPath, which *can* be a directory (SQLite creates file in it).
			// However, for weights file, we expect to write a file.
			// For now, let's assume if it exists and is a dir, it's problematic for a file write.
			// This logic might need refinement based on how DbPath is handled if it points to a dir.
			// The current SQLite logger seems to expect a file path.
			return "", fmt.Errorf("path '%s' (resolved to '%s') exists and is a directory, expected a file path for writing", rawPath, absPath)
		}
	}

	return absPath, nil
}


// createNetwork initializes the main CrowNet neural network instance (o.Net)
// using the application configuration. It passes the necessary parameters
// to network.NewCrowNet to construct and set up the network.
// Note: This function assumes network.NewCrowNet handles detailed setup based on AppConfig.
func (o *Orchestrator) createNetwork() {
	cliCfg := &o.AppCfg.Cli
	// TODO: The call to network.NewCrowNet here needs to be updated
	// to match the signature network.NewCrowNet(appCfg *config.AppConfig).
	// For now, documenting intent.
	o.Net = network.NewCrowNet(
		cliCfg.TotalNeurons,
		common.Rate(cliCfg.BaseLearningRate),
		&o.AppCfg.SimParams,
		cliCfg.Seed,
	)

	// Log initial network state.
	// Showing only the first few input/output neuron IDs for brevity.
	maxInputToShow := 5
	maxOutputToShow := 10
	numInputs := len(o.Net.InputNeuronIDs)
	numOutputs := len(o.Net.OutputNeuronIDs)

	fmt.Printf("Network created: %d neurons. Input IDs: %v..., Output IDs: %v...\n",
		len(o.Net.Neurons),
		o.Net.InputNeuronIDs[:min(maxInputToShow, numInputs)],
		o.Net.OutputNeuronIDs[:min(maxOutputToShow, numOutputs)],
	)
	fmt.Printf("Initial State: Cortisol=%.3f, Dopamine=%.3f\n",
		o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel)
}

// loadWeights loads synaptic weights from the specified file.
// Uses the injected loadWeightsFn for testability.
func (o *Orchestrator) loadWeights(rawFilepath string) error {
	validatedFilepath, err := o.validatePath(rawFilepath, true) // true: for reading
	if err != nil {
		// For loadWeights, if the error is specifically IsNotExist, it might be handled differently
		// by the caller (e.g., expose mode might proceed with random weights).
		// However, validatePath already returns a specific error for IsNotExist.
		// Let's make it clear: if path is invalid (not just not existing), it's an error.
		// If it's valid but doesn't exist, that's also an error from validatePath if forRead=true.
		return fmt.Errorf("invalid weights file path '%s': %w", rawFilepath, err)
	}
	// At this point, validatedFilepath is a valid, existing file path.

	loadedWeights, errLoad := o.loadWeightsFn(validatedFilepath)
	if errLoad != nil {
		return fmt.Errorf("failed to load weights from %s: %w", validatedFilepath, errLoad)
	}
	o.Net.SynapticWeights = loadedWeights
	fmt.Printf("Existing weights loaded from %s\n", validatedFilepath)
	return nil
}

// saveWeights saves the current synaptic weights to the specified file.
// Uses the injected saveWeightsFn for testability.
func (o *Orchestrator) saveWeights(rawFilepath string) error {
	validatedFilepath, err := o.validatePath(rawFilepath, false) // false: for writing
	if err != nil {
		return fmt.Errorf("invalid weights file path '%s' for saving: %w", rawFilepath, err)
	}

	if err := o.saveWeightsFn(o.Net.SynapticWeights, validatedFilepath); err != nil {
		return fmt.Errorf("failed to save trained weights to %s: %w", validatedFilepath, err)
	}
	fmt.Printf("Trained weights saved to %s\n", validatedFilepath)
	return nil
}

// printModeSpecificConfig outputs configuration details relevant to the current execution mode.
func (o *Orchestrator) printModeSpecificConfig() {
	cliCfg := o.AppCfg.Cli
	switch cliCfg.Mode {
	case config.ModeExpose:
		fmt.Printf("  ModeExpose: Epochs=%d, BaseLearningRate=%.4f, CyclesPerPattern=%d\n",
			cliCfg.Epochs, cliCfg.BaseLearningRate, cliCfg.CyclesPerPattern)
	case config.ModeObserve:
		fmt.Printf("  ModeObserve: Digit=%d, CyclesToSettle=%d\n",
			cliCfg.Digit, cliCfg.CyclesToSettle)
	case config.ModeSim:
		fmt.Printf("  ModeSim: TotalCycles=%d, DBPath='%s', DBSaveInterval=%d\n",
			cliCfg.Cycles, cliCfg.DbPath, cliCfg.SaveInterval)
		if cliCfg.StimInputFreqHz > 0 && cliCfg.StimInputID != -2 { // -2 means stimulus disabled
			fmt.Printf("  ModeSim: ContinuousStimulus: InputID=%d at %.1f Hz\n",
				cliCfg.StimInputID, cliCfg.StimInputFreqHz)
		}
	}
}

// setupContinuousInputStimulus configures a continuous input stimulus for simulation mode.
func (o *Orchestrator) setupContinuousInputStimulus() error {
	cliCfg := o.AppCfg.Cli
	if cliCfg.StimInputFreqHz <= 0.0 || cliCfg.StimInputID == -2 || len(o.Net.InputNeuronIDs) == 0 {
		return nil // No stimulus configured or no input neurons to stimulate
	}

	stimID := cliCfg.StimInputID
	if stimID == -1 { // -1 means use the first available input neuron
		stimID = int(o.Net.InputNeuronIDs[0])
	}

	// Validate that stimID is a valid input neuron ID
	isValidStimID := false
	for _, id := range o.Net.InputNeuronIDs {
		if int(id) == stimID {
			isValidStimID = true
			break
		}
	}

	if !isValidStimID {
		return fmt.Errorf("stimulus input neuron ID %d not found or invalid", stimID)
	}

	if err := o.Net.ConfigureFrequencyInput(common.NeuronID(stimID), cliCfg.StimInputFreqHz); err != nil {
		return fmt.Errorf("failed to configure frequency input for neuron %d at %.1f Hz: %w", stimID, cliCfg.StimInputFreqHz, err)
	}
	fmt.Printf("Continuous stimulus: Input Neuron %d at %.1f Hz.\n", stimID, cliCfg.StimInputFreqHz)
	return nil
}

// runSimulationLoop executes the main simulation cycles.
func (o *Orchestrator) runSimulationLoop() error {
	cycles := o.AppCfg.Cli.Cycles
	saveInterval := o.AppCfg.Cli.SaveInterval

	for i := 0; i < cycles; i++ {
		o.Net.RunCycle()
		// Log progress periodically
		if i%10 == 0 || i == cycles-1 {
			fmt.Printf("Cycle %d/%d: Cortisol:%.3f Dopamine:%.3f LRMod:%.3f SynMod:%.3f Pulses:%d\n",
				o.Net.CycleCount-1, cycles, // CycleCount is incremented at the end of RunCycle
				o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel,
				o.Net.ChemicalEnv.LearningRateModulationFactor, o.Net.ChemicalEnv.SynaptogenesisModulationFactor,
				len(o.Net.ActivePulses.GetAll())) // Use GetAll() for length
		}

		// Log network state to DB if enabled and interval is met
		if o.Logger != nil && saveInterval > 0 && o.Net.CycleCount > 0 && int(o.Net.CycleCount)%saveInterval == 0 {
			if err := o.Logger.LogNetworkState(o.Net); err != nil {
				return fmt.Errorf("failed to log network state to DB (periodic) at cycle %d: %w", o.Net.CycleCount, err)
			}
		}
	}

	// Final log if DB is enabled and the last cycle wasn't a save interval point
	if o.Logger != nil && cycles > 0 && (saveInterval == 0 || cycles%saveInterval != 0) {
		if err := o.Logger.LogNetworkState(o.Net); err != nil {
			return fmt.Errorf("failed to log final network state to DB: %w", err)
		}
	}
	return nil
}

// reportMonitoredOutputFrequency prints the firing frequency of a monitored output neuron.
func (o *Orchestrator) reportMonitoredOutputFrequency() error {
	cliCfg := o.AppCfg.Cli
	if cliCfg.MonitorOutputID == -2 || len(o.Net.OutputNeuronIDs) == 0 { // -2 means monitoring disabled
		return nil
	}

	monitorID := cliCfg.MonitorOutputID
	if monitorID == -1 { // -1 means use the first available output neuron
		monitorID = int(o.Net.OutputNeuronIDs[0])
	}

	// Validate that monitorID is a valid output neuron ID
	isValidMonitorID := false
	for _, outID := range o.Net.OutputNeuronIDs {
		if int(outID) == monitorID {
			isValidMonitorID = true
			break
		}
	}

	if !isValidMonitorID {
		return fmt.Errorf("output neuron ID for monitoring (%d) not found or invalid", monitorID)
	}

	freq, err := o.Net.GetOutputFrequency(common.NeuronID(monitorID))
	if err != nil {
		return fmt.Errorf("failed to get frequency for output neuron %d: %w", monitorID, err)
	}
	fmt.Printf("Frequency for Output Neuron %d: %.2f Hz (over last %.0f cycles).\n",
		monitorID, freq, o.AppCfg.SimParams.OutputFrequencyWindowCycles)
	return nil
}

// runSimMode handles the 'sim' execution mode.
func (o *Orchestrator) runSimMode() error {
	fmt.Printf("\nStarting General Simulation for %d cycles...\n", o.AppCfg.Cli.Cycles)
	if err := o.setupContinuousInputStimulus(); err != nil {
		return fmt.Errorf("error in stimulus setup: %w", err)
	}

	o.Net.SetDynamicState(true, true, true) // Neurochemicals, learning, synaptogenesis active

	if err := o.runSimulationLoop(); err != nil {
		return fmt.Errorf("error during simulation loop: %w", err)
	}
	if err := o.reportMonitoredOutputFrequency(); err != nil {
		return fmt.Errorf("error reporting monitored output frequency: %w", err)
	}

	fmt.Printf("Final State: Cortisol=%.3f, Dopamine=%.3f\n", o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel)
	return nil
}

// runExposureEpochs handles the core loop for the 'expose' mode.
func (o *Orchestrator) runExposureEpochs() error {
	allPatterns, err := datagen.GetAllDigitPatterns(&o.AppCfg.SimParams)
	if err != nil {
		return fmt.Errorf("failed to load digit patterns: %w", err)
	}

	cliCfg := o.AppCfg.Cli
	for epoch := 0; epoch < cliCfg.Epochs; epoch++ {
		fmt.Printf("Epoch %d/%d starting...\n", epoch+1, cliCfg.Epochs)
		patternsProcessedThisEpoch := 0
		// Consider randomizing pattern order or using a defined sequence if needed.
		// For now, iterating 0-9.
		for digit := 0; digit <= 9; digit++ {
			pattern, ok := allPatterns[digit]
			if !ok {
				return fmt.Errorf("pattern for digit %d not found in loaded set (epoch %d)", digit, epoch+1)
			}

			o.Net.ResetNetworkStateForNewPattern()
			if errPres := o.Net.PresentPattern(pattern); errPres != nil {
				return fmt.Errorf("failed to present pattern for digit %d in epoch %d: %w", digit, epoch+1, errPres)
			}

			for cycleInPattern := 0; cycleInPattern < cliCfg.CyclesPerPattern; cycleInPattern++ {
				o.Net.RunCycle()
				// Log to DB if enabled and interval is met
				if o.Logger != nil && cliCfg.SaveInterval > 0 && o.Net.CycleCount > 0 && int(o.Net.CycleCount)%cliCfg.SaveInterval == 0 {
					if errLog := o.Logger.LogNetworkState(o.Net); errLog != nil {
						return fmt.Errorf("failed to log network state (epoch %d, digit %d, cycle %d): %w", epoch+1, digit, o.Net.CycleCount, errLog)
					}
				}
			}
			patternsProcessedThisEpoch++
		}
		fmt.Printf("Epoch %d/%d completed. Processed %d patterns. Cortisol: %.3f, Dopamine: %.3f, EffectiveLRFactor: %.4f\n",
			epoch+1, cliCfg.Epochs, patternsProcessedThisEpoch,
			o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel, o.Net.ChemicalEnv.LearningRateModulationFactor)
	}
	return nil
}

// runExposeMode handles the 'expose' execution mode for training the network.
func (o *Orchestrator) runExposeMode() error {
	cliCfg := o.AppCfg.Cli
	fmt.Printf("\nStarting Exposure Phase for %d epochs (BaseLearningRate: %.4f, CyclesPerPattern: %d)...\n",
		cliCfg.Epochs, cliCfg.BaseLearningRate, cliCfg.CyclesPerPattern)

	// Attempt to load weights; if not found, network uses random weights (normal for initial training).
	if err := o.loadWeights(cliCfg.WeightsFile); err != nil {
		// Log the error but continue, as starting from random weights is acceptable.
		fmt.Printf("Note: Could not load weights from %s (%v), starting with new/random weights.\n", cliCfg.WeightsFile, err)
	}

	o.Net.SetDynamicState(true, true, true) // Neurochemicals, learning, synaptogenesis active

	if err := o.runExposureEpochs(); err != nil {
		return fmt.Errorf("error during exposure epochs: %w", err)
	}

	fmt.Println("Exposure phase completed.")
	if err := o.saveWeights(cliCfg.WeightsFile); err != nil {
		return err // Error saving weights is critical after training
	}
	return nil
}

// runObservationPattern presents a single pattern and runs the network for settling cycles.
func (o *Orchestrator) runObservationPattern() ([]float64, error) {
	cliCfg := o.AppCfg.Cli
	patternToObserve, err := datagen.GetDigitPatternFn(cliCfg.Digit, &o.AppCfg.SimParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get pattern for digit %d: %w", cliCfg.Digit, err)
	}

	o.Net.ResetNetworkStateForNewPattern()
	if errPres := o.Net.PresentPattern(patternToObserve); errPres != nil {
		return nil, fmt.Errorf("failed to present pattern for observation: %w", errPres)
	}

	for i := 0; i < cliCfg.CyclesToSettle; i++ {
		o.Net.RunCycle()
	}

	outputActivation, errAct := o.Net.GetOutputActivation()
	if errAct != nil {
		return nil, fmt.Errorf("failed to get output activation: %w", errAct)
	}
	return outputActivation, nil
}

// displayOutputActivation prints the activation levels of output neurons with ASCII art bars.
func (o *Orchestrator) displayOutputActivation(outputActivation []float64) {
	fmt.Printf("Digit Presented: %d\n", o.AppCfg.Cli.Digit)
	fmt.Println("Output Neuron Activation Pattern (Accumulated Potential):")

	if len(outputActivation) == 0 {
		fmt.Println("  No output activation data to display.")
		return
	}

	const maxBarLength = 20 // Length of the ASCII bar

	// Find min and max activation for normalization
	minAct := outputActivation[0]
	maxAct := outputActivation[0]
	for _, val := range outputActivation {
		if val < minAct {
			minAct = val
		}
		if val > maxAct {
			maxAct = val
		}
	}

	activationRange := maxAct - minAct

	for i, val := range outputActivation {
		neuronIDStr := "N/A" // Fallback if ID mapping is off
		if i < len(o.Net.OutputNeuronIDs) {
			neuronIDStr = fmt.Sprintf("%d", o.Net.OutputNeuronIDs[i])
		}

		bar := ""
		if activationRange == 0 { // All values are the same
			if maxAct > 0 { // Or some other threshold for "active"
				bar = strings.Repeat("|", maxBarLength)
			} else {
				bar = strings.Repeat(" ", maxBarLength)
			}
		} else {
			normalizedVal := (val - minAct) / activationRange
			numChars := int(normalizedVal*float64(maxBarLength) + 0.5) // +0.5 for rounding
			if numChars < 0 {
				numChars = 0
			}
			if numChars > maxBarLength {
				numChars = maxBarLength
			}
			bar = strings.Repeat("|", numChars) + strings.Repeat(" ", maxBarLength-numChars)
		}
		fmt.Printf("  OutputNeuron[%2d] (ID %4s) | [%s] | %.4f\n", i, neuronIDStr, bar, val)
	}
}

// runObserveMode handles the 'observe' execution mode.
func (o *Orchestrator) runObserveMode() error {
	cliCfg := o.AppCfg.Cli
	fmt.Printf("\nObserving Network Response for digit %d (%d settling cycles)...\n",
		cliCfg.Digit, cliCfg.CyclesToSettle)

	// Loading weights is critical for observe mode.
	if err := o.loadWeights(cliCfg.WeightsFile); err != nil {
		return fmt.Errorf("failed to load weights for observe mode from %s: %w. Expose/train the network first", cliCfg.WeightsFile, err)
	}

	// Disable dynamics that would alter the network state during observation.
	o.Net.SetDynamicState(false, false, false)

	outputActivation, err := o.runObservationPattern()
	if err != nil {
		return fmt.Errorf("failed to run observation pattern: %w", err)
	}

	o.displayOutputActivation(outputActivation)
	o.Net.SetDynamicState(true, true, true) // Restore default dynamic states
	return nil
}

// min is a simple helper to find the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- Test Wrappers (Exported for use in _test package) ---
// These allow tests to call unexported methods or specific parts of the orchestrator's logic.

// SetupContinuousInputStimulusForTest wraps setupContinuousInputStimulus for testing.
func (o *Orchestrator) SetupContinuousInputStimulusForTest() error {
	return o.setupContinuousInputStimulus()
}

// RunObserveModeForTest wraps runObserveMode for testing.
func (o *Orchestrator) RunObserveModeForTest() error {
	return o.runObserveMode()
}

// RunExposeModeForTest wraps runExposeMode for testing.
func (o *Orchestrator) RunExposeModeForTest() error {
	return o.runExposeMode()
}

// SetLoadWeightsFn allows tests to inject a mock loadWeightsFn.
func (o *Orchestrator) SetLoadWeightsFn(fn func(filepath string) (synaptic.NetworkWeights, error)) {
	o.loadWeightsFn = fn
}

// SetSaveWeightsFn allows tests to inject a mock saveWeightsFn.
func (o *Orchestrator) SetSaveWeightsFn(fn func(weights synaptic.NetworkWeights, filepath string) error) {
	o.saveWeightsFn = fn
}

// InitializeLoggerForTest wraps initializeLogger for testing.
func (o *Orchestrator) InitializeLoggerForTest() error {
	return o.initializeLogger()
}

// CreateNetworkForTest wraps createNetwork for testing.
func (o *Orchestrator) CreateNetworkForTest() {
	o.createNetwork()
}

// RunSimModeForTest wraps runSimMode for testing.
func (o *Orchestrator) RunSimModeForTest() error {
	return o.runSimMode()
}

// CloseLoggerForTest wraps closing the logger, for testing.
// It now returns an error from the underlying Logger.Close() method.
func (o *Orchestrator) CloseLoggerForTest() error {
	if o.Logger != nil {
		if err := o.Logger.Close(); err != nil {
			return fmt.Errorf("error closing logger for test: %w", err)
		}
	}
	return nil
}

// LoadWeightsForTest wraps loadWeights for testing.
func (o *Orchestrator) LoadWeightsForTest(filepath string) error {
	return o.loadWeights(filepath)
}

// runLogUtilMode handles the 'logutil' execution mode (FEATURE-004).
func (o *Orchestrator) runLogUtilMode() error {
	fmt.Println("\nCrowNet Log Utility...")
	cliCfg := &o.AppCfg.Cli

	// Path validation for LogUtilDbPath (read-only for export)
	// Using the existing validatePath method.
	// Note: validatePath might try to check parent dir for writing if forRead=false.
	// Forcing forRead=true as logutil only reads.
	// If dbPath is invalid, validatePath will return an error.
	// The config.Validate() already ensures LogUtilDbPath is not empty.
	_, err := o.validatePath(cliCfg.LogUtilDbPath, true)
	if err != nil {
		return fmt.Errorf("invalid --logutil.dbPath '%s': %w", cliCfg.LogUtilDbPath, err)
	}
	// We use cliCfg.LogUtilDbPath directly in the call to exporter,
	// as validatePath was just for the check here. The exporter will use the raw path.

	fmt.Printf("  Subcommand: %s\n", cliCfg.LogUtilSubcommand)
	fmt.Printf("  Database: %s\n", cliCfg.LogUtilDbPath)
	fmt.Printf("  Table: %s\n", cliCfg.LogUtilTable)
	fmt.Printf("  Format: %s\n", cliCfg.LogUtilFormat)
	if cliCfg.LogUtilOutput != "" {
		fmt.Printf("  Output: %s\n", cliCfg.LogUtilOutput)
	} else {
		fmt.Println("  Output: stdout")
	}

	if cliCfg.LogUtilSubcommand == "export" {
		// Call the main export function (to be created in storage package)
		err := storage.ExportLogData(
			cliCfg.LogUtilDbPath,
			cliCfg.LogUtilTable,
			cliCfg.LogUtilFormat,
			cliCfg.LogUtilOutput,
		)
		if err != nil {
			return fmt.Errorf("log export failed: %w", err)
		}
		fmt.Println("Log export completed successfully.")
		return nil
	}
	// Should be caught by validation, but as a safeguard:
	return fmt.Errorf("unknown logutil subcommand: %s", cliCfg.LogUtilSubcommand)
}
