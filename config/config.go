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
	SpaceMaxDimension             float64
	BaseFiringThreshold           float64
	PulsePropagationSpeed         float64
	HebbianCoincidenceWindow      int
	DopaminergicPercent           float64
	InhibitoryPercent             float64
	ExcitatoryRadiusFactor        float64
	DopaminergicRadiusFactor      float64
	InhibitoryRadiusFactor        float64
	MinInputNeurons               int
	MinOutputNeurons              int
	PatternHeight                 int
	PatternWidth                  int
	PatternSize                   int
	AccumulatedPulseDecayRate     float64
	AbsoluteRefractoryCycles      common.CycleCount
	RelativeRefractoryCycles      common.CycleCount
	SynaptogenesisInfluenceRadius float64
	AttractionForceFactor         float64
	RepulsionForceFactor          float64
	DampeningFactor               float64
	MaxMovementPerCycle           float64
	CyclesPerSecond               float64
	OutputFrequencyWindowCycles   float64
	InitialSynapticWeightMin      float64
	InitialSynapticWeightMax      float64
	MaxSynapticWeight             float64
	HebbPositiveReinforceFactor   float64
	HebbNegativeReinforceFactor   float64
	CortisolProductionRate        float64
	CortisolDecayRate             float64
	DopamineProductionRate        float64
	DopamineDecayRate             float64
	CortisolInfluenceOnLR         float64
	DopamineInfluenceOnLR         float64
	CortisolInfluenceOnSynapto    float64
	DopamineInfluenceOnSynapto    float64
	FiringThresholdIncreaseOnDopa float64
	FiringThresholdIncreaseOnCort float64
	SynapticWeightDecayRate       float64

	CortisolProductionPerHit    float64
	CortisolMaxLevel            float64
	DopamineProductionPerEvent  float64
	DopamineMaxLevel            float64

	MinLearningRateFactor         float64
}

type CLIConfig struct {
	TotalNeurons     int
	Cycles           int
	DbPath           string
	SaveInterval     int
	StimInputID      int
	StimInputFreqHz  float64
	MonitorOutputID  int
	DebugChem        bool
	Mode             string
	Epochs           int
	WeightsFile      string
	Digit            int
	BaseLearningRate float64
	CyclesPerPattern int
	CyclesToSettle   int
	Seed             int64
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
		AccumulatedPulseDecayRate:     0.1,
		AbsoluteRefractoryCycles:      common.CycleCount(2),
		RelativeRefractoryCycles:      common.CycleCount(3),
		SynaptogenesisInfluenceRadius: 2.0,
		AttractionForceFactor:         0.01,
		RepulsionForceFactor:          0.005,
		DampeningFactor:               0.5,
		MaxMovementPerCycle:           0.1,
		CyclesPerSecond:               100.0,
		OutputFrequencyWindowCycles:   50.0,
		InitialSynapticWeightMin:      0.1,
		InitialSynapticWeightMax:      0.5,
		MaxSynapticWeight:             1.0,
		HebbPositiveReinforceFactor:   0.1,
		HebbNegativeReinforceFactor:   0.05,
		CortisolProductionRate:        0.01,
		CortisolDecayRate:             0.005,
		DopamineProductionRate:        0.02,
		DopamineDecayRate:             0.01,
		CortisolInfluenceOnLR:         -0.5,
		DopamineInfluenceOnLR:         0.8,
		CortisolInfluenceOnSynapto:    -0.3,
		DopamineInfluenceOnSynapto:    0.5,
		FiringThresholdIncreaseOnDopa: -0.2,
		FiringThresholdIncreaseOnCort: 0.3,
		SynapticWeightDecayRate:       0.0001,
		CortisolProductionPerHit:    0.05,
		CortisolMaxLevel:            1.0,
		DopamineProductionPerEvent:  0.1,
		DopamineMaxLevel:            1.0,
		MinLearningRateFactor:         0.1,
	}
}

func LoadCLIConfig() CLIConfig {
	cfg := CLIConfig{}

	flag.IntVar(&cfg.TotalNeurons, "neurons", 200, "Total number of neurons in the network.")
	flag.IntVar(&cfg.Cycles, "cycles", 1000, "Total simulation cycles for 'sim' mode.")
	flag.IntVar(&cfg.StimInputID, "stimInputID", -1, "ID of an input neuron for general continuous stimulus in 'sim' mode (-1 for first available, -2 to disable).")
	flag.Float64Var(&cfg.StimInputFreqHz, "stimInputFreqHz", 0.0, "Frequency (Hz) for general stimulus in 'sim' mode (0.0 to disable).")
	flag.IntVar(&cfg.MonitorOutputID, "monitorOutputID", -1, "ID of an output neuron to monitor for frequency reporting in 'sim' mode (-1 for first available, -2 to disable).")
	flag.StringVar(&cfg.DbPath, "dbPath", "crownet_sim_run.db", "Path for the SQLite database file.")
	flag.IntVar(&cfg.SaveInterval, "saveInterval", 100, "Cycle interval for saving to DB (0 to disable periodic saves, only final if any).")
	flag.BoolVar(&cfg.DebugChem, "debugChem", false, "Enable debug prints for chemical production.")
	flag.StringVar(&cfg.Mode, "mode", ModeSim, fmt.Sprintf("Operation mode: '%s', '%s', or '%s'.", ModeSim, ModeExpose, ModeObserve))
	flag.IntVar(&cfg.Epochs, "epochs", 50, "Number of exposure epochs (for 'expose' mode).")
	flag.StringVar(&cfg.WeightsFile, "weightsFile", "crownet_weights.json", "File to save/load synaptic weights.")
	flag.IntVar(&cfg.Digit, "digit", 0, "Digit (0-9) to present (for 'observe' mode).")
	flag.Float64Var(&cfg.BaseLearningRate, "lrBase", 0.01, "Base learning rate for Hebbian plasticity.")
	flag.IntVar(&cfg.CyclesPerPattern, "cyclesPerPattern", 20, "Number of cycles to run per pattern presentation during 'expose' mode.")
	flag.IntVar(&cfg.CyclesToSettle, "cyclesToSettle", 50, "Number of cycles to run for network settling during 'observe' mode.")
	flag.Int64Var(&cfg.Seed, "seed", 0, "Seed for random number generator (0 uses current time, other values are used directly).")

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
	if ac.Cli.TotalNeurons <= 0 {
		return fmt.Errorf("total neurons must be positive, got %d", ac.Cli.TotalNeurons)
	}
	if ac.Cli.Cycles < 0 {
		return fmt.Errorf("cycles must be non-negative, got %d", ac.Cli.Cycles)
	}
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
	if ac.Cli.Mode == ModeExpose || ac.Cli.Mode == ModeObserve {
		if ac.Cli.WeightsFile == "" {
			return fmt.Errorf("weightsFile must be specified for mode '%s'", ac.Cli.Mode)
		}
	}
	if ac.Cli.Mode == ModeExpose {
		if ac.Cli.Epochs <= 0 {
			return fmt.Errorf("epochs must be positive for mode '%s', got %d", ac.Cli.Mode, ac.Cli.Epochs)
		}
		if ac.Cli.CyclesPerPattern <= 0 {
			return fmt.Errorf("cyclesPerPattern must be positive for mode '%s', got %d", ac.Cli.Mode, ac.Cli.CyclesPerPattern)
		}
	}
	if ac.Cli.Mode == ModeObserve {
		if ac.Cli.Digit < 0 || ac.Cli.Digit > 9 {
			return fmt.Errorf("digit must be between 0-9 for mode '%s', got %d", ac.Cli.Mode, ac.Cli.Digit)
		}
		if ac.Cli.CyclesToSettle <= 0 {
			return fmt.Errorf("cyclesToSettle must be positive for mode '%s', got %d", ac.Cli.Mode, ac.Cli.CyclesToSettle)
		}
	}
	if ac.Cli.BaseLearningRate < 0 {
		return fmt.Errorf("baseLearningRate must be non-negative, got %f", ac.Cli.BaseLearningRate)
	}
	if ac.Cli.SaveInterval < 0 {
		return fmt.Errorf("saveInterval must be non-negative, got %d", ac.Cli.SaveInterval)
	}
	if ac.SimParams.MinInputNeurons <= 0 || ac.SimParams.MinOutputNeurons <= 0 {
		return fmt.Errorf("MinInputNeurons and MinOutputNeurons must be positive")
	}
	if ac.SimParams.PatternSize <= 0 {
		return fmt.Errorf("PatternSize must be positive")
	}
	if ac.SimParams.CyclesPerSecond <= 0 {
		return fmt.Errorf("CyclesPerSecond must be positive")
	}
	if ac.SimParams.DopaminergicPercent < 0 || ac.SimParams.DopaminergicPercent > 1.0 {
		return fmt.Errorf("DopaminergicPercent must be between 0.0 and 1.0, got %f", ac.SimParams.DopaminergicPercent)
	}
	if ac.SimParams.InhibitoryPercent < 0 || ac.SimParams.InhibitoryPercent > 1.0 {
		return fmt.Errorf("InhibitoryPercent must be between 0.0 and 1.0, got %f", ac.SimParams.InhibitoryPercent)
	}
	if ac.SimParams.DopaminergicPercent+ac.SimParams.InhibitoryPercent > 1.0 {
		return fmt.Errorf("sum of DopaminergicPercent and InhibitoryPercent cannot exceed 1.0, got %f", ac.SimParams.DopaminergicPercent+ac.SimParams.InhibitoryPercent)
	}
	return nil
}
