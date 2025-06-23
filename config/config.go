package config

import (
	"flag"
	"fmt"
	"strings"
	"time"
	"crownet/common"
)

const (
	ModeSim     = "sim"
	ModeExpose  = "expose"
	ModeObserve = "observe"
)

var SupportedModes = []string{ModeSim, ModeExpose, ModeObserve}

type SimulationParameters struct {
	// Spatial and General Network Parameters
	SpaceMaxDimension float64 // Defines the boundary of the 16D space.
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
}

type CLIConfig struct {
	// General Configuration
	Mode             string // Operation mode (sim, expose, observe).
	TotalNeurons     int    // Total number of neurons in the network.
	Seed             int64  // Seed for random number generator.
	WeightsFile      string // File to save/load synaptic weights.
	BaseLearningRate common.Rate // Base learning rate for Hebbian plasticity.

	// Mode 'sim' Specific Configuration
	Cycles          int    // Total simulation cycles.
	DbPath          string // Path for the SQLite database file for logging.
	SaveInterval    int    // Cycle interval for saving to DB.
	StimInputID     int    // ID of an input neuron for continuous stimulus (-1 for first, -2 to disable).
	StimInputFreqHz float64 // Frequency (Hz) for stimulus (0.0 to disable).
	MonitorOutputID int    // ID of an output neuron to monitor frequency (-1 for first, -2 to disable).
	DebugChem       bool   // Enable debug prints for chemical production.

	// Mode 'expose' Specific Configuration
	Epochs           int // Number of exposure epochs.
	CyclesPerPattern int // Number of cycles to run per pattern presentation.

	// Mode 'observe' Specific Configuration
	Digit          int // Digit (0-9) to present.
	CyclesToSettle int // Number of cycles for network settling.
}

type AppConfig struct {
	SimParams SimulationParameters
	Cli       CLIConfig
}

func DefaultSimulationParameters() SimulationParameters {
	return SimulationParameters{
		SpaceMaxDimension:             10.0,
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

func LoadCLIConfig() CLIConfig {
	cfg := CLIConfig{}

	// General Configuration Flags
	flag.StringVar(&cfg.Mode, "mode", ModeSim, fmt.Sprintf("Operation mode: '%s', '%s', or '%s'.", ModeSim, ModeExpose, ModeObserve))
	flag.IntVar(&cfg.TotalNeurons, "neurons", 200, "Total number of neurons in the network.")
	flag.Int64Var(&cfg.Seed, "seed", 0, "Seed for random number generator (0 uses current time, other values are used directly).")
	flag.StringVar(&cfg.WeightsFile, "weightsFile", "crownet_weights.json", "File to save/load synaptic weights.")
	flag.Float64Var((*float64)(&cfg.BaseLearningRate), "lrBase", 0.01, "Base learning rate for Hebbian plasticity.") // Cast to *float64 as common.Rate is float64

	// Mode 'sim' Specific Flags
	flag.IntVar(&cfg.Cycles, "cycles", 1000, "Total simulation cycles for 'sim' mode.")
	flag.StringVar(&cfg.DbPath, "dbPath", "crownet_sim_run.db", "Path for the SQLite database file for logging.")
	flag.IntVar(&cfg.SaveInterval, "saveInterval", 100, "Cycle interval for saving to DB (0 to disable periodic saves, only final if any).")
	flag.IntVar(&cfg.StimInputID, "stimInputID", -1, "ID of an input neuron for general continuous stimulus in 'sim' mode (-1 for first available, -2 to disable).")
	flag.Float64Var(&cfg.StimInputFreqHz, "stimInputFreqHz", 0.0, "Frequency (Hz) for general stimulus in 'sim' mode (0.0 to disable).")
	flag.IntVar(&cfg.MonitorOutputID, "monitorOutputID", -1, "ID of an output neuron to monitor for frequency reporting in 'sim' mode (-1 for first available, -2 to disable).")
	flag.BoolVar(&cfg.DebugChem, "debugChem", false, "Enable debug prints for chemical production.")

	// Mode 'expose' Specific Flags
	flag.IntVar(&cfg.Epochs, "epochs", 50, "Number of exposure epochs (for 'expose' mode).")
	flag.IntVar(&cfg.CyclesPerPattern, "cyclesPerPattern", 20, "Number of cycles to run per pattern presentation during 'expose' mode.")

	// Mode 'observe' Specific Flags
	flag.IntVar(&cfg.Digit, "digit", 0, "Digit (0-9) to present (for 'observe' mode).")
	flag.IntVar(&cfg.CyclesToSettle, "cyclesToSettle", 50, "Number of cycles to run for network settling during 'observe' mode.")

	if !flag.Parsed() {
		flag.Parse()
	}

	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	return cfg
}

func NewAppConfig() (*AppConfig, error) {
	cliCfg := LoadCLIConfig()
	appCfg := &AppConfig{
		SimParams: DefaultSimulationParameters(),
		Cli:       cliCfg,
	}
	if err := appCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return appCfg, nil
}

func (ac *AppConfig) Validate() error {
	// General CLI parameter validation
	if ac.Cli.TotalNeurons <= 0 {
		return fmt.Errorf("total neurons must be positive, got %d", ac.Cli.TotalNeurons)
	}
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
	}

	// SimulationParameters validation
	if ac.SimParams.MinInputNeurons <= 0 || ac.SimParams.MinOutputNeurons <= 0 {
		return fmt.Errorf("MinInputNeurons (%d) and MinOutputNeurons (%d) must be positive", ac.SimParams.MinInputNeurons, ac.SimParams.MinOutputNeurons)
	}
	if ac.SimParams.PatternSize <= 0 {
		return fmt.Errorf("PatternSize must be positive, got %d", ac.SimParams.PatternSize)
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
	// Add more SimParams checks as needed, e.g., for rates, factors, thresholds to be within logical ranges.
	// For example, decay rates should likely be non-negative.
	// MaxLevels should be >= 0.
	// Coincidence windows should be positive.

	return nil
}
