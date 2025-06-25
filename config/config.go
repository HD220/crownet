// Package config provides types and functions for managing application
// configuration, including simulation parameters and command-line interface (CLI)
// settings. It handles loading defaults, parsing CLI flags, and validating
// the overall configuration.
package config

import (
	"flag"
	"fmt"
	"path/filepath" // Added for path cleaning
	"strings"
	"time"

	"crownet/common"
)

const (
	// ModeSim instructs the application to run a general simulation.
	ModeSim = "sim"
	// ModeExpose instructs the application to run in pattern exposure mode for training.
	ModeExpose = "expose"
	// ModeObserve instructs the application to observe network response to a specific input.
	ModeObserve = "observe"
	// ModeLogUtil instructs the application to run the SQLite log utility.
	ModeLogUtil = "logutil" // FEATURE-004
)

// SupportedModes lists all valid operation modes for the application.
// It is used for validating the mode provided via CLI or configuration file.
var SupportedModes = []string{ModeSim, ModeExpose, ModeObserve, ModeLogUtil} // FEATURE-004: Added ModeLogUtil

// SimulationParameters defines the detailed parameters that govern the behavior
// of the neural network simulation. These parameters control aspects from spatial
// properties and neuron behavior to learning rules and neurochemical influences.
type SimulationParameters struct {
	// Spatial and General Network Parameters

	// SpaceMaxDimension defines the boundary of the N-dimensional simulation space.
	// For example, if SpaceMaxDimension is 10.0, coordinates typically range from -5.0 to +5.0.
	SpaceMaxDimension float64
	CyclesPerSecond   float64 // Simulation cycles that represent one second of real time.

	// Neuron Definition and Behavior
	BaseFiringThreshold      common.Threshold // Base threshold for a neuron to fire.
	AccumulatedPulseDecayRate common.Rate    // Rate at which accumulated pulse potential decays per cycle.
	AbsoluteRefractoryCycles common.CycleCount // Cycles a neuron cannot fire after firing.
	RelativeRefractoryCycles common.CycleCount // Cycles a neuron has increased threshold after firing.
	PulsePropagationSpeed    common.Rate    // Speed at which pulses travel in the space.

	// Neuron Type Distribution and Influence Radii
	DopaminergicPercent      common.Percentage // Percentage of internal neurons that are dopaminergic.
	InhibitoryPercent        common.Percentage // Percentage of internal neurons that are inhibitory.
	ExcitatoryRadiusFactor   common.Factor     // Factor for excitatory neuron influence radius relative to a base.
	DopaminergicRadiusFactor common.Factor     // Factor for dopaminergic neuron influence radius.
	InhibitoryRadiusFactor   common.Factor     // Factor for inhibitory neuron influence radius.

	// Input/Output and Pattern Definition
	MinInputNeurons  int // Minimum number of input neurons required.
	MinOutputNeurons int // Minimum number of output neurons required.
	PatternHeight    int // Height of the input patterns (e.g., for digits).
	PatternWidth     int // Width of the input patterns.
	PatternSize      int // Total size of the input patterns (Height * Width).
	OutputFrequencyWindowCycles float64 // Number of cycles to average output neuron firing frequency.

	// Synaptic Weights and Learning Parameters
	InitialSynapticWeightMin    common.SynapticWeight // Minimum initial synaptic weight.
	InitialSynapticWeightMax    common.SynapticWeight // Maximum initial synaptic weight.
	MaxSynapticWeight           common.SynapticWeight // Absolute maximum for any synaptic weight.
	HebbianWeightMin            common.SynapticWeight // Minimum weight for Hebbian learning (can be negative).
	HebbianWeightMax            common.SynapticWeight // Maximum weight for Hebbian learning.
	SynapticWeightDecayRate     common.Rate           // Rate at which synaptic weights decay per cycle.
	HebbianCoincidenceWindow    common.CycleCount     // Time window (cycles) for Hebbian learning co-activation.
	HebbPositiveReinforceFactor common.Factor         // Factor for strengthening synaptic weights in Hebbian learning.
	HebbNegativeReinforceFactor common.Factor         // Factor for weakening synaptic weights (if applicable, or for LTD).
	MinLearningRateFactor       common.Factor         // Minimum modulation factor for learning rate.

	// Synaptogenesis (Neuronal Movement) Parameters
	SynaptogenesisInfluenceRadius common.Coordinate // Radius within which neurons influence each other for movement.
	AttractionForceFactor         common.Factor     // Factor for attractive forces between neurons.
	RepulsionForceFactor          common.Factor     // Factor for repulsive forces between neurons.
	DampeningFactor               common.Factor     // Dampening factor for neuron movement.
	MaxMovementPerCycle           common.Coordinate // Maximum distance a neuron can move in one cycle.

	// Neurochemical System Parameters
	CortisolProductionRate        common.Rate    // Base rate of cortisol production.
	CortisolDecayRate             common.Rate    // Rate at which cortisol decays.
	CortisolProductionPerHit    common.Level   // Amount of cortisol produced per 'stress' event.
	CortisolMaxLevel              common.Level   // Maximum possible cortisol level.
	DopamineProductionRate        common.Rate    // Base rate of dopamine production.
	DopamineDecayRate             common.Rate    // Rate at which dopamine decays.
	DopamineProductionPerEvent  common.Level   // Amount of dopamine produced per 'reward' event.
	DopamineMaxLevel              common.Level   // Maximum possible dopamine level.

	// Neurochemical Influence Factors
	CortisolInfluenceOnLR         common.Factor // How cortisol influences the learning rate.
	DopamineInfluenceOnLR         common.Factor // How dopamine influences the learning rate.
	CortisolInfluenceOnSynapto    common.Factor // How cortisol influences synaptogenesis.
	DopamineInfluenceOnSynapto    common.Factor // How dopamine influences synaptogenesis.
	FiringThresholdIncreaseOnDopa common.Factor // How dopamine influences neuron firing thresholds.
	FiringThresholdIncreaseOnCort common.Factor // How cortisol influences neuron firing thresholds.

	// CortisolGlandPosition defines the fixed N-dimensional coordinates of the cortisol gland
	// within the simulation space.
	CortisolGlandPosition common.Point
}

// CLIConfig holds configuration parameters that are typically set or overridden
// via command-line flags. It includes general settings as well as mode-specific options.
type CLIConfig struct {
	// General Configuration

	// Mode specifies the operation mode for the application (e.g., "sim", "expose", "observe").
	Mode             string      `json:"mode"`
	TotalNeurons     int         `json:"total_neurons"`     // Total number of neurons in the network.
	Seed             int64       `json:"seed"`              // Seed for random number generator (0 for time-based).
	WeightsFile      string      `json:"weights_file"`      // File to save/load synaptic weights.
	BaseLearningRate common.Rate `json:"base_learning_rate"` // Base learning rate for Hebbian plasticity.

	// Mode 'sim' Specific Configuration
	Cycles          int     `json:"cycles"`            // Total simulation cycles for 'sim' mode.
	DbPath          string  `json:"db_path"`           // Path for the SQLite database file for logging in 'sim' mode.
	SaveInterval    int     `json:"save_interval"`     // Cycle interval for saving to DB in 'sim' mode (0 to disable periodic).
	StimInputID     int     `json:"stim_input_id"`     // ID of input neuron for continuous stimulus in 'sim' mode (-1 for first, -2 to disable).
	StimInputFreqHz float64 `json:"stim_input_freq_hz"` // Frequency (Hz) for stimulus in 'sim' mode (0.0 to disable).
	MonitorOutputID int     `json:"monitor_output_id"` // ID of output neuron to monitor frequency in 'sim' mode (-1 for first, -2 to disable).
	DebugChem       bool    `json:"debug_chem"`        // Enable debug prints for chemical production in 'sim' mode.

	// Mode 'expose' Specific Configuration
	Epochs           int `json:"epochs"`             // Number of exposure epochs for 'expose' mode.
	CyclesPerPattern int `json:"cycles_per_pattern"` // Number of cycles to run per pattern presentation in 'expose' mode.

	// Mode 'observe' Specific Configuration
	Digit          int `json:"digit"`            // Digit (0-9) to present for 'observe' mode.
	CyclesToSettle int `json:"cycles_to_settle"` // Number of cycles for network settling in 'observe' mode.

	// Mode 'logutil' Specific Configuration (FEATURE-004)
	LogUtilSubcommand string `json:"logutil_subcommand"` // e.g., "export"
	LogUtilDbPath     string `json:"logutil_dbpath"`     // Path to the SQLite DB file
	LogUtilTable      string `json:"logutil_table"`      // Table to export (e.g., "NetworkSnapshots", "NeuronStates")
	LogUtilFormat     string `json:"logutil_format"`     // Output format (e.g., "csv")
	LogUtilOutput     string `json:"logutil_output"`     // Output file path (stdout if empty)
}

// AppConfig is the top-level configuration structure, aggregating both
// SimulationParameters and CLIConfig.
type AppConfig struct {
	SimParams SimulationParameters // Detailed parameters for the simulation behavior.
	Cli       CLIConfig            // Parameters typically set via command-line flags.
}

// DefaultSimulationParameters returns a SimulationParameters struct populated with
// sensible default values for all simulation settings.
func DefaultSimulationParameters() SimulationParameters {
	return SimulationParameters{
		SpaceMaxDimension:             10.0,
		CortisolGlandPosition:         common.Point{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Default to origin
		BaseFiringThreshold:           1.0,
		PulsePropagationSpeed:         1.0,
		HebbianCoincidenceWindow:      2,
		DopaminergicPercent:           0.1,
		InhibitoryPercent:             0.2,
		ExcitatoryRadiusFactor:        1.0,
		DopaminergicRadiusFactor:      0.8,
		InhibitoryRadiusFactor:        0.9,
		MinInputNeurons:               35,
		MinOutputNeurons:              10,
		PatternHeight:                 7,
		PatternWidth:                  5,
		PatternSize:                   7 * 5,
		AccumulatedPulseDecayRate:     common.Rate(0.1),
		AbsoluteRefractoryCycles:      common.CycleCount(2),
		RelativeRefractoryCycles:      common.CycleCount(3),
		SynaptogenesisInfluenceRadius: common.Coordinate(2.0),
		AttractionForceFactor:         common.Factor(0.01),
		RepulsionForceFactor:          common.Factor(0.005),
		DampeningFactor:               common.Factor(0.5),
		MaxMovementPerCycle:           common.Coordinate(0.1),
		CyclesPerSecond:               100.0,
		OutputFrequencyWindowCycles:   50.0,
		InitialSynapticWeightMin:      common.SynapticWeight(0.1),
		InitialSynapticWeightMax:      common.SynapticWeight(0.5),
		MaxSynapticWeight:             common.SynapticWeight(1.0),
		HebbianWeightMin:              common.SynapticWeight(-0.1), // Allow weakening below zero
		HebbianWeightMax:              common.SynapticWeight(1.0),  // Can reach absolute max
		HebbPositiveReinforceFactor:   common.Factor(0.1),
		HebbNegativeReinforceFactor:   common.Factor(0.05),
		CortisolProductionRate:        common.Rate(0.01),
		CortisolDecayRate:             common.Rate(0.005),
		DopamineProductionRate:        common.Rate(0.02),
		DopamineDecayRate:             common.Rate(0.01),
		CortisolInfluenceOnLR:         common.Factor(-0.5),
		DopamineInfluenceOnLR:         common.Factor(0.8),
		CortisolInfluenceOnSynapto:    common.Factor(-0.3),
		DopamineInfluenceOnSynapto:    common.Factor(0.5),
		FiringThresholdIncreaseOnDopa: common.Factor(-0.2),
		FiringThresholdIncreaseOnCort: common.Factor(0.3),
		SynapticWeightDecayRate:       common.Rate(0.0001),
		CortisolProductionPerHit:    common.Level(0.05),
		CortisolMaxLevel:            common.Level(1.0),
		DopamineProductionPerEvent:  common.Level(0.1),
		DopamineMaxLevel:            common.Level(1.0),
		MinLearningRateFactor:         common.Factor(0.1),
	}
}

// LoadCLIConfig populates a CLIConfig struct by parsing flags from the given
// arguments string slice using the provided FlagSet.
//
// NOTE: With the introduction of Cobra for CLI handling (REFACTOR-CLI-001),
// this function is no longer the primary mechanism for parsing application-wide CLI flags
// during normal execution (main.go now uses cmd.Execute()). Cobra commands define
// and parse their own flags.
//
// This function remains useful for:
// 1. Testing: To parse a controlled set of arguments into a CLIConfig struct.
// 2. Programmatic Configuration: If there's a need to build CLIConfig from a
//    string slice outside the Cobra execution flow.
//
// The `args` slice should not include the program name. If `args` is nil,
// and fSet is a named FlagSet (not flag.CommandLine), no arguments will be parsed,
// which is typically the desired behavior for isolated tests. Ginkgo/test-specific
// flags are filtered out from the provided `args`.
//
// Parameters:
//   fSet: The flag.FlagSet instance to define and parse flags.
//   args: A slice of strings representing the command-line arguments (excluding the program name).
//
// Returns:
//   A CLIConfig struct populated from the parsed flags and an error if parsing fails.
//   If the "seed" flag is 0 after parsing, it's updated to the current time's nanoseconds.
func LoadCLIConfig(fSet *flag.FlagSet, args []string) (CLIConfig, error) {
	cfg := CLIConfig{}

	// General Configuration Flags
	fSet.StringVar(&cfg.Mode, "mode", ModeSim, fmt.Sprintf("Operation mode: '%s', '%s', or '%s'.", ModeSim, ModeExpose, ModeObserve))
	fSet.IntVar(&cfg.TotalNeurons, "neurons", 200, "Total number of neurons in the network.")
	fSet.Int64Var(&cfg.Seed, "seed", 0, "Seed for random number generator (0 uses current time, other values are used directly).")
	fSet.StringVar(&cfg.WeightsFile, "weightsFile", "crownet_weights.json", "File to save/load synaptic weights.")
	fSet.Float64Var((*float64)(&cfg.BaseLearningRate), "lrBase", 0.01, "Base learning rate for Hebbian plasticity.")

	// Mode 'sim' Specific Flags
	fSet.IntVar(&cfg.Cycles, "cycles", 1000, "Total simulation cycles for 'sim' mode.")
	fSet.StringVar(&cfg.DbPath, "dbPath", "crownet_sim_run.db", "Path for the SQLite database file for logging.")
	fSet.IntVar(&cfg.SaveInterval, "saveInterval", 100, "Cycle interval for saving to DB (0 to disable periodic saves, only final if any).")
	fSet.IntVar(&cfg.StimInputID, "stimInputID", -1, "ID of an input neuron for general continuous stimulus in 'sim' mode (-1 for first available, -2 to disable).")
	fSet.Float64Var(&cfg.StimInputFreqHz, "stimInputFreqHz", 0.0, "Frequency (Hz) for general stimulus in 'sim' mode (0.0 to disable).")
	fSet.IntVar(&cfg.MonitorOutputID, "monitorOutputID", -1, "ID of an output neuron to monitor for frequency reporting in 'sim' mode (-1 for first available, -2 to disable).")
	fSet.BoolVar(&cfg.DebugChem, "debugChem", false, "Enable debug prints for chemical production.")

	// Mode 'expose' Specific Flags
	fSet.IntVar(&cfg.Epochs, "epochs", 50, "Number of exposure epochs (for 'expose' mode).")
	fSet.IntVar(&cfg.CyclesPerPattern, "cyclesPerPattern", 20, "Number of cycles to run per pattern presentation during 'expose' mode.")

	// Mode 'observe' Specific Flags
	fSet.IntVar(&cfg.Digit, "digit", 0, "Digit (0-9) to present (for 'observe' mode).")
	fSet.IntVar(&cfg.CyclesToSettle, "cyclesToSettle", 50, "Number of cycles to run for network settling during 'observe' mode.")

	// Mode 'logutil' Specific Flags (FEATURE-004)
	// Note: These flags will only be relevant if -mode=logutil is set.
	// Consider using subcommands for better CLI ergonomics if more logutil features are added.
	// For now, simple flags are fine.
	fSet.StringVar(&cfg.LogUtilSubcommand, "logutil.subcommand", "export", "Log utility subcommand (e.g., 'export').")
	fSet.StringVar(&cfg.LogUtilDbPath, "logutil.dbPath", "", "Path to SQLite DB for logutil mode.")
	fSet.StringVar(&cfg.LogUtilTable, "logutil.table", "", "Table to process in logutil mode (e.g., 'NetworkSnapshots', 'NeuronStates').")
	fSet.StringVar(&cfg.LogUtilFormat, "logutil.format", "csv", "Output format for logutil export (e.g., 'csv').")
	fSet.StringVar(&cfg.LogUtilOutput, "logutil.output", "", "Output file for logutil export (stdout if empty).")

	// Filter out Ginkgo-specific flags before parsing if they exist
	// to prevent "flag provided but not defined" errors when running tests
	// with `go test ./...` if Ginkgo is indirectly included or its flags are passed.
	var nonGinkgoArgs []string
	if args != nil { // args could be nil if called directly without specific test arguments
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-ginkgo.") && !strings.HasPrefix(arg, "-test.") {
				nonGinkgoArgs = append(nonGinkgoArgs, arg)
			}
		}
	}


	// Only parse if not already parsed. In tests, fSet might be parsed multiple times
	// if not careful, but Parse is idempotent for defined flags.
	// However, calling Parse on an already parsed FlagSet with new arguments can lead to issues.
	// For production, os.Args[1:] would be passed. For tests, specific arg slices.
	if err := fSet.Parse(nonGinkgoArgs); err != nil {
		return cfg, fmt.Errorf("error parsing flags: %w", err)
	}

	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	// Clean file paths
	if cfg.WeightsFile != "" {
		cfg.WeightsFile = filepath.Clean(cfg.WeightsFile)
	}
	if cfg.DbPath != "" {
		cfg.DbPath = filepath.Clean(cfg.DbPath)
	}

	return cfg, nil
}

// NewAppConfig creates a new AppConfig primarily for testing or programmatic use,
// by loading default simulation parameters, parsing a given slice of command-line
// arguments (via LoadCLIConfig) to populate CLIConfig, and then validating the
// combined configuration.
//
// NOTE: With the introduction of Cobra for CLI handling (REFACTOR-CLI-001),
// this function is **no longer the primary entry point for application configuration**
// called by `main.go`. Cobra commands now construct `AppConfig` instances by
// combining default parameters, TOML file loading (if specified via `--configFile`),
// and parsed CLI flags.
//
// This function can still be useful for:
// 1. Testing: To create an AppConfig from a controlled set of string arguments
//    as if they were passed on the command line (without Cobra's involvement).
// 2. Programmatic Setup: If an AppConfig needs to be built from arguments
//    outside the main Cobra CLI flow.
//
// It does NOT handle TOML file loading; that logic is now within the Cobra command handlers.
//
// Parameters:
//   args: A slice of strings representing the command-line arguments (excluding the program name).
//
// Returns:
//  An *AppConfig struct containing the fully resolved and validated configuration,
//  or an error if loading or validation fails.
func NewAppConfig(args []string) (*AppConfig, error) {
	// Create a new FlagSet for command-line parsing.
	// os.Args[0] is the program name, actual flags start from os.Args[1:].
	// Here, 'args' is expected to be os.Args[1:].
	cliCfg, err := LoadCLIConfig(flag.NewFlagSet("crownet", flag.ContinueOnError), args)
	if err != nil {
		return nil, fmt.Errorf("failed to load CLI config: %w", err)
	}

	appCfg := &AppConfig{
		SimParams: DefaultSimulationParameters(), // Start with defaults
		Cli:       cliCfg,                        // Apply CLI overrides
	}

	// Future step: Insert loading from config file here if -configFile is specified in cliCfg.
	// The order would be: Defaults -> Config File -> CLI Flags.

	if err := appCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return appCfg, nil
}

// Validate checks the AppConfig for consistency and valid values across
// CLIConfig and SimulationParameters.
// It ensures that parameters meet their required constraints (e.g., positivity,
// ranges, interdependencies).
//
// Returns:
//  An error if any validation rule is violated, nil otherwise.
func (ac *AppConfig) Validate() error {
	// General CLI parameter validation
	if ac.Cli.TotalNeurons <= 0 { // Basic positivity check
		return fmt.Errorf("total neurons must be positive, got %d", ac.Cli.TotalNeurons)
	}
	// More specific check against minimums will be done after SimParams are validated for MinInput/OutputNeurons
	if ac.Cli.BaseLearningRate < 0 {
		return fmt.Errorf("baseLearningRate must be non-negative, got %f", ac.Cli.BaseLearningRate)
	}

	// Mode validation
	modeValid := false
	for _, m := range SupportedModes {
		if ac.Cli.Mode == m {
			modeValid = true
			break
		}
	}
	if !modeValid {
		return fmt.Errorf("invalid mode '%s', supported modes are: %s", ac.Cli.Mode, strings.Join(SupportedModes, ", "))
	}

	// Mode-specific CLI parameter validation
	switch ac.Cli.Mode {
	case ModeSim:
		if ac.Cli.Cycles < 0 { // Allow 0 for a "setup only" type run, though typically positive.
			return fmt.Errorf("cycles for sim mode must be non-negative, got %d", ac.Cli.Cycles)
		}
		if ac.Cli.SaveInterval < 0 {
			return fmt.Errorf("saveInterval for sim mode must be non-negative, got %d", ac.Cli.SaveInterval)
		}
	case ModeExpose:
		if ac.Cli.WeightsFile == "" {
			return fmt.Errorf("weightsFile must be specified for mode '%s'", ac.Cli.Mode)
		}
		if ac.Cli.Epochs <= 0 {
			return fmt.Errorf("epochs must be positive for mode '%s', got %d", ac.Cli.Mode, ac.Cli.Epochs)
		}
		if ac.Cli.CyclesPerPattern <= 0 {
			return fmt.Errorf("cyclesPerPattern must be positive for mode '%s', got %d", ac.Cli.Mode, ac.Cli.CyclesPerPattern)
		}
	case ModeObserve:
		if ac.Cli.WeightsFile == "" {
			return fmt.Errorf("weightsFile must be specified for mode '%s'", ac.Cli.Mode)
		}
		if ac.Cli.Digit < 0 || ac.Cli.Digit > 9 {
			return fmt.Errorf("digit must be between 0-9 for mode '%s', got %d", ac.Cli.Mode, ac.Cli.Digit)
		}
		if ac.Cli.CyclesToSettle <= 0 {
			return fmt.Errorf("cyclesToSettle must be positive for mode '%s', got %d", ac.Cli.Mode, ac.Cli.CyclesToSettle)
		}
	case ModeLogUtil: // FEATURE-004
		if ac.Cli.LogUtilSubcommand != "export" { // Initially only "export" is supported
			return fmt.Errorf("invalid logutil.subcommand '%s', currently only 'export' is supported", ac.Cli.LogUtilSubcommand)
		}
		if strings.TrimSpace(ac.Cli.LogUtilDbPath) == "" {
			return fmt.Errorf("logutil.dbPath must be specified for mode '%s'", ac.Cli.Mode)
		}
		if strings.TrimSpace(ac.Cli.LogUtilTable) == "" {
			return fmt.Errorf("logutil.table must be specified for mode '%s'", ac.Cli.Mode)
		}
		if ac.Cli.LogUtilTable != "NetworkSnapshots" && ac.Cli.LogUtilTable != "NeuronStates" {
			return fmt.Errorf("invalid logutil.table '%s', must be 'NetworkSnapshots' or 'NeuronStates'", ac.Cli.LogUtilTable)
		}
		if ac.Cli.LogUtilFormat != "csv" { // Initially only "csv" is supported
			return fmt.Errorf("invalid logutil.format '%s', currently only 'csv' is supported", ac.Cli.LogUtilFormat)
		}
		// LogUtilOutput can be empty (stdout). Path validation for it will be done by the logutil itself if provided.
		// No SimulationParameters validation needed for LogUtil mode.
		return nil // Early exit for LogUtil mode after its specific checks
	}

	// SimulationParameters validation (Order can be important for dependent checks)
	// This section is skipped if Mode is LogUtil due to the early exit above.
	if ac.SimParams.MinInputNeurons <= 0 {
		return fmt.Errorf("MinInputNeurons must be positive, got %d", ac.SimParams.MinInputNeurons)
	}
	if ac.SimParams.MinOutputNeurons <= 0 {
		return fmt.Errorf("MinOutputNeurons must be positive, got %d", ac.SimParams.MinOutputNeurons)
	}

	// Check TotalNeurons from CLI against the (now validated positive) MinInput/Output from SimParams
	if ac.Cli.TotalNeurons < (ac.SimParams.MinInputNeurons + ac.SimParams.MinOutputNeurons) {
		return fmt.Errorf("total neurons from CLI (%d) is less than the sum of required MinInputNeurons (%d) and MinOutputNeurons (%d)",
			ac.Cli.TotalNeurons, ac.SimParams.MinInputNeurons, ac.SimParams.MinOutputNeurons)
	}

	if ac.SimParams.PatternSize <= 0 {
		return fmt.Errorf("PatternSize must be positive, got %d", ac.SimParams.PatternSize)
	}
	if ac.SimParams.PatternHeight <= 0 {
		return fmt.Errorf("PatternHeight must be positive, got %d", ac.SimParams.PatternHeight)
	}
	if ac.SimParams.PatternWidth <= 0 {
		return fmt.Errorf("PatternWidth must be positive, got %d", ac.SimParams.PatternWidth)
	}
	if ac.SimParams.PatternSize != (ac.SimParams.PatternHeight * ac.SimParams.PatternWidth) {
		return fmt.Errorf("PatternSize (%d) must equal PatternHeight (%d) * PatternWidth (%d) = %d",
			ac.SimParams.PatternSize, ac.SimParams.PatternHeight, ac.SimParams.PatternWidth,
			ac.SimParams.PatternHeight*ac.SimParams.PatternWidth)
	}
	if ac.SimParams.CyclesPerSecond <= 0 {
		return fmt.Errorf("CyclesPerSecond must be positive, got %f", ac.SimParams.CyclesPerSecond)
	}
	if ac.SimParams.DopaminergicPercent < 0 || ac.SimParams.DopaminergicPercent > 1.0 {
		return fmt.Errorf("DopaminergicPercent must be between 0.0 and 1.0, got %f", ac.SimParams.DopaminergicPercent)
	}
	if ac.SimParams.InhibitoryPercent < 0 || ac.SimParams.InhibitoryPercent > 1.0 {
		return fmt.Errorf("InhibitoryPercent must be between 0.0 and 1.0, got %f", ac.SimParams.InhibitoryPercent)
	}
	if ac.SimParams.DopaminergicPercent+ac.SimParams.InhibitoryPercent > 1.0 {
		return fmt.Errorf("sum of DopaminergicPercent (%f) and InhibitoryPercent (%f) cannot exceed 1.0", ac.SimParams.DopaminergicPercent, ac.SimParams.InhibitoryPercent)
	}

	// Validate non-negative or positive constraints for SimParams
	if ac.SimParams.SpaceMaxDimension <= 0 {
		return fmt.Errorf("SpaceMaxDimension must be positive, got %f", ac.SimParams.SpaceMaxDimension)
	}
	if ac.SimParams.BaseFiringThreshold <= 0 { // Assuming threshold should be positive
		return fmt.Errorf("BaseFiringThreshold must be positive, got %f", ac.SimParams.BaseFiringThreshold)
	}
	if ac.SimParams.AccumulatedPulseDecayRate < 0 {
		return fmt.Errorf("AccumulatedPulseDecayRate must be non-negative, got %f", ac.SimParams.AccumulatedPulseDecayRate)
	}
	if ac.SimParams.AbsoluteRefractoryCycles < 0 { // Should likely be >= 0
		return fmt.Errorf("AbsoluteRefractoryCycles must be non-negative, got %d", ac.SimParams.AbsoluteRefractoryCycles)
	}
	if ac.SimParams.RelativeRefractoryCycles < 0 { // Should likely be >= 0
		return fmt.Errorf("RelativeRefractoryCycles must be non-negative, got %d", ac.SimParams.RelativeRefractoryCycles)
	}
	if ac.SimParams.PulsePropagationSpeed <= 0 {
		return fmt.Errorf("PulsePropagationSpeed must be positive, got %f", ac.SimParams.PulsePropagationSpeed)
	}
	if ac.SimParams.OutputFrequencyWindowCycles <= 0 {
		return fmt.Errorf("OutputFrequencyWindowCycles must be positive, got %f", ac.SimParams.OutputFrequencyWindowCycles)
	}
	if ac.SimParams.InitialSynapticWeightMin < 0 { // Assuming weights can be 0 but not negative
		return fmt.Errorf("InitialSynapticWeightMin must be non-negative, got %f", ac.SimParams.InitialSynapticWeightMin)
	}
	if ac.SimParams.InitialSynapticWeightMax < ac.SimParams.InitialSynapticWeightMin {
		return fmt.Errorf("InitialSynapticWeightMax (%f) must be >= InitialSynapticWeightMin (%f)", ac.SimParams.InitialSynapticWeightMax, ac.SimParams.InitialSynapticWeightMin)
	}
    if ac.SimParams.MaxSynapticWeight < ac.SimParams.InitialSynapticWeightMax {
        return fmt.Errorf("MaxSynapticWeight (%f) must be >= InitialSynapticWeightMax (%f)", ac.SimParams.MaxSynapticWeight, ac.SimParams.InitialSynapticWeightMax)
    }
	if ac.SimParams.SynapticWeightDecayRate < 0 {
		return fmt.Errorf("SynapticWeightDecayRate must be non-negative, got %f", ac.SimParams.SynapticWeightDecayRate)
	}
	if ac.SimParams.HebbianCoincidenceWindow <= 0 { // Should be positive for a window to exist
		return fmt.Errorf("HebbianCoincidenceWindow must be positive, got %d", ac.SimParams.HebbianCoincidenceWindow)
	}
	// Factors can be positive or negative depending on their meaning, e.g. CortisolInfluenceOnLR is negative.
	// For now, not adding generic positive checks for all factors, but specific ones like ReinforceFactor.
	if ac.SimParams.HebbPositiveReinforceFactor < 0 {
		return fmt.Errorf("HebbPositiveReinforceFactor must be non-negative, got %f", ac.SimParams.HebbPositiveReinforceFactor)
	}
	if ac.SimParams.MinLearningRateFactor < 0 {
		return fmt.Errorf("MinLearningRateFactor must be non-negative, got %f", ac.SimParams.MinLearningRateFactor)
	}
	if ac.SimParams.SynaptogenesisInfluenceRadius <= 0 {
		return fmt.Errorf("SynaptogenesisInfluenceRadius must be positive, got %f", ac.SimParams.SynaptogenesisInfluenceRadius)
	}
	if ac.SimParams.MaxMovementPerCycle < 0 {
		return fmt.Errorf("MaxMovementPerCycle must be non-negative, got %f", ac.SimParams.MaxMovementPerCycle)
	}
	if ac.SimParams.CortisolProductionRate < 0 || ac.SimParams.CortisolDecayRate < 0 || ac.SimParams.CortisolProductionPerHit < 0 || ac.SimParams.CortisolMaxLevel < 0 {
		return fmt.Errorf("Cortisol parameters (ProductionRate, DecayRate, ProductionPerHit, MaxLevel) must be non-negative")
	}
	if ac.SimParams.DopamineProductionRate < 0 || ac.SimParams.DopamineDecayRate < 0 || ac.SimParams.DopamineProductionPerEvent < 0 || ac.SimParams.DopamineMaxLevel < 0 {
		return fmt.Errorf("Dopamine parameters (ProductionRate, DecayRate, ProductionPerEvent, MaxLevel) must be non-negative")
	}
	if ac.SimParams.CortisolMaxLevel < ac.SimParams.CortisolProductionPerHit {
		// This is a logical check, a single hit shouldn't exceed max. Could be more complex if rate also contributes significantly before decay.
		// For simplicity, basic check.
	}
	if ac.SimParams.DopamineMaxLevel < ac.SimParams.DopamineProductionPerEvent {
		// Similar logical check for dopamine
	}

	return nil
}
