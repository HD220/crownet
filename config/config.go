package config

import (
	"flag"
	"fmt"
	"strings"
	"time" // Adicionado para time.Now().UnixNano()
	"crownet/common" // Adicionado para common.CycleCount
)

// Constantes para modos de operação
const (
	ModeSim     = "sim"
	ModeExpose  = "expose"
	ModeObserve = "observe"
)

var SupportedModes = []string{ModeSim, ModeExpose, ModeObserve}

// SimulationParameters holds intrinsic parameters of the simulation dynamics.
type SimulationParameters struct {
	SpaceMaxDimension             float64 // Max coordinate value for neuron positions (radius of hyper-sphere)
	BaseFiringThreshold           float64 // Base threshold for a neuron to fire
	PulsePropagationSpeed         float64 // Speed at which pulses travel
	HebbianCoincidenceWindow      int     // Time window (cycles) for Hebbian learning coincidence
	DopaminergicPercent           float64 // Percentage of internal neurons that are dopaminergic
	InhibitoryPercent             float64 // Percentage of internal neurons that are inhibitory
	ExcitatoryRadiusFactor        float64 // Factor for placement radius of excitatory/input/output neurons
	DopaminergicRadiusFactor      float64 // Factor for placement radius of dopaminergic neurons
	InhibitoryRadiusFactor        float64 // Factor for placement radius of inhibitory neurons
	MinInputNeurons               int     // Minimum number of input neurons
	MinOutputNeurons              int     // Minimum number of output neurons (e.g., for digit classification)
	PatternHeight                 int     // Height of the input pattern (e.g., 7 for a 7x5 digit)
	PatternWidth                  int     // Width of the input pattern (e.g., 5 for a 7x5 digit)
	PatternSize                   int     // Total size of the input pattern (Height * Width)
	AccumulatedPulseDecayRate     float64 // Rate at which accumulated potential in a neuron decays per cycle
	AbsoluteRefractoryCycles      common.CycleCount // Cycles a neuron stays in absolute refractory state
	RelativeRefractoryCycles      common.CycleCount // Cycles a neuron stays in relative refractory state
	SynaptogenesisInfluenceRadius float64 // Max distance for synaptogenesis interaction
	AttractionForceFactor         float64 // Factor for attractive force in synaptogenesis
	RepulsionForceFactor          float64 // Factor for repulsive force in synaptogenesis
	DampeningFactor               float64 // Dampening factor for neuron velocity in synaptogenesis
	MaxMovementPerCycle           float64 // Maximum distance a neuron can move in one cycle
	CyclesPerSecond               float64 // Simulation cycles that represent one second (for Hz calculations)
	OutputFrequencyWindowCycles   float64 // Window (in cycles) for calculating output neuron frequency
	InitialSynapticWeightMin      float64 // Minimum initial synaptic weight
	InitialSynapticWeightMax      float64 // Maximum initial synaptic weight
	MaxSynapticWeight             float64 // Maximum allowed synaptic weight
	HebbPositiveReinforceFactor   float64 // Factor for LTP (Long-Term Potentiation)
	HebbNegativeReinforceFactor   float64 // Factor for LTD (Long-Term Depression)
	CortisolProductionRate        float64
	CortisolDecayRate             float64
	DopamineProductionRate        float64
	DopamineDecayRate             float64
	CortisolInfluenceOnLR         float64 // How much cortisol decreases learning rate
	DopamineInfluenceOnLR         float64 // How much dopamine increases learning rate
	CortisolInfluenceOnSynapto    float64 // How much cortisol influences synaptogenesis
	DopamineInfluenceOnSynapto    float64 // How much dopamine influences synaptogenesis
	FiringThresholdIncreaseOnDopa float64 // How much dopamine decreases firing threshold (e.g., a factor like -0.2 for 20% reduction from base)
	FiringThresholdIncreaseOnCort float64 // How much cortisol increases firing threshold (e.g., a factor like 0.3 for 30% increase from base)
	SynapticWeightDecayRate       float64 // Rate at which synaptic weights decay towards zero if not reinforced

	// Neurochemical production, decay, and levels
	CortisolProductionPerHit    float64 // Amount of cortisol produced per pulse hitting the gland
	CortisolMaxLevel            float64 // Maximum cortisol level
	DopamineProductionPerEvent  float64 // Amount of dopamine produced per firing dopaminergic neuron
	DopamineMaxLevel            float64 // Maximum dopamine level

	// Simplified modulation factors (used by current neurochemical logic)
	MinLearningRateFactor         float64 // Minimum learning rate factor (e.g., 0.05)
	// Os parâmetros mais complexos para modulação de LR e Threshold (MaxDopamineLearningMultiplier,
	// CortisolHighEffectThreshold, CortisolLearningSuppressionFactor, SynaptogenesisReductionFactor,
	// DopamineSynaptogenesisIncreaseFactor, CortisolMinEffectThreshold, CortisolOptimalLowThreshold,
	// CortisolOptimalHighThreshold, MaxThresholdReductionFactor, ThresholdIncreaseFactorHigh)
	// foram removidos pois a lógica em neurochemicals.go foi simplificada para usar
	// os parâmetros de influência direta (ex: CortisolInfluenceOnLR, FiringThresholdIncreaseOnCort).
}

// CLIConfig holds parameters that are typically set via command-line flags.
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
	Seed             int64 // Seed for random number generator
}

// AppConfig bundles all configuration.
type AppConfig struct {
	SimParams SimulationParameters
	Cli       CLIConfig
}

// DefaultSimulationParameters returns a SimulationParameters struct with sensible defaults.
func DefaultSimulationParameters() SimulationParameters {
	return SimulationParameters{
		SpaceMaxDimension:             10.0,  // Increased from 4.0 for more space
		BaseFiringThreshold:           1.0,
		PulsePropagationSpeed:         1.0,   // Increased from 0.6
		HebbianCoincidenceWindow:      2,     // Short window
		DopaminergicPercent:           0.1,   // 10% of internal neurons
		InhibitoryPercent:             0.2,   // 20% of internal neurons
		ExcitatoryRadiusFactor:        1.0,   // Placed throughout the main sphere volume
		DopaminergicRadiusFactor:      0.8,   // Slightly more central
		InhibitoryRadiusFactor:        0.9,   // Generally distributed
		MinInputNeurons:               35,    // e.g., 7x5 grid, should align with PatternSize
		MinOutputNeurons:              10,    // For 0-9 digits
		PatternHeight:                 7,
		PatternWidth:                  5,
		PatternSize:                   7 * 5, // Explicitly Height * Width
		AccumulatedPulseDecayRate:     0.1,   // 10% decay per cycle
		AbsoluteRefractoryCycles:      common.CycleCount(2),
		RelativeRefractoryCycles:      common.CycleCount(3),
		SynaptogenesisInfluenceRadius: 2.0,   // Neurons interact if within this distance
		AttractionForceFactor:         0.01,
		RepulsionForceFactor:          0.005,
		DampeningFactor:               0.5,   // Velocity is halved if no new forces
		MaxMovementPerCycle:           0.1,   // Max distance a neuron can move
		CyclesPerSecond:               100.0, // 100 cycles = 1 second
		OutputFrequencyWindowCycles:   50.0,  // Calculate frequency over last 50 cycles
		InitialSynapticWeightMin:      0.1,
		InitialSynapticWeightMax:      0.5,
		MaxSynapticWeight:             1.0,
		HebbPositiveReinforceFactor:   0.1,
		HebbNegativeReinforceFactor:   0.05,
		CortisolProductionRate:        0.01,
		CortisolDecayRate:             0.005,
		DopamineProductionRate:        0.02,
		DopamineDecayRate:             0.01,
		CortisolInfluenceOnLR:         -0.5, // Reduces LR by up to 50%
		DopamineInfluenceOnLR:         0.8,  // Increases LR by up to 80%
		CortisolInfluenceOnSynapto:    -0.3,
		DopamineInfluenceOnSynapto:    0.5,
	FiringThresholdIncreaseOnDopa: -0.2, // Lowers threshold (e.g. base * (1 + (-0.2)))
	FiringThresholdIncreaseOnCort: 0.3,  // Raises threshold (e.g. base * (1 + 0.3))
	SynapticWeightDecayRate:       0.0001,

	// Neurochemical production, decay, and levels - Defaults
	CortisolProductionPerHit:    0.05,
	CortisolMaxLevel:            1.0,
	DopamineProductionPerEvent:  0.1,
	DopamineMaxLevel:            1.0,

	// Simplified modulation factors - Defaults
	MinLearningRateFactor:         0.1,   // LR factor won't go below 10%
	// Os valores padrão para os parâmetros de modulação complexos removidos não são mais necessários aqui.
	}
}

// LoadCLIConfig parses command-line flags and returns a CLIConfig struct.
func LoadCLIConfig() CLIConfig {
	cfg := CLIConfig{}

	flag.IntVar(&cfg.TotalNeurons, "neurons", 200, "Total number of neurons in the network.") // Increased default
	flag.IntVar(&cfg.Cycles, "cycles", 1000, "Total simulation cycles for 'sim' mode.")      // Increased default
	flag.IntVar(&cfg.StimInputID, "stimInputID", -1, "ID of an input neuron for general continuous stimulus in 'sim' mode (-1 for first available, -2 to disable).")
	flag.Float64Var(&cfg.StimInputFreqHz, "stimInputFreqHz", 0.0, "Frequency (Hz) for general stimulus in 'sim' mode (0.0 to disable).")
	flag.IntVar(&cfg.MonitorOutputID, "monitorOutputID", -1, "ID of an output neuron to monitor for frequency reporting in 'sim' mode (-1 for first available, -2 to disable).")
	flag.StringVar(&cfg.DbPath, "dbPath", "crownet_sim_run.db", "Path for the SQLite database file.") // Changed default name
	flag.IntVar(&cfg.SaveInterval, "saveInterval", 100, "Cycle interval for saving to DB (0 to disable periodic saves, only final if any).") // Default to 100
	flag.BoolVar(&cfg.DebugChem, "debugChem", false, "Enable debug prints for chemical production.")

	flag.StringVar(&cfg.Mode, "mode", ModeSim, fmt.Sprintf("Operation mode: '%s', '%s', or '%s'.", ModeSim, ModeExpose, ModeObserve))
	flag.IntVar(&cfg.Epochs, "epochs", 50, "Number of exposure epochs (for 'expose' mode).")
	flag.StringVar(&cfg.WeightsFile, "weightsFile", "crownet_weights.json", "File to save/load synaptic weights.") // Changed default name
	flag.IntVar(&cfg.Digit, "digit", 0, "Digit (0-9) to present (for 'observe' mode).")
	flag.Float64Var(&cfg.BaseLearningRate, "lrBase", 0.01, "Base learning rate for Hebbian plasticity.") // Adjusted default
	flag.IntVar(&cfg.CyclesPerPattern, "cyclesPerPattern", 20, "Number of cycles to run per pattern presentation during 'expose' mode.")
	flag.IntVar(&cfg.CyclesToSettle, "cyclesToSettle", 50, "Number of cycles to run for network settling during 'observe' mode.")
	flag.Int64Var(&cfg.Seed, "seed", 0, "Seed for random number generator (0 uses current time, other values are used directly).")

	if !flag.Parsed() {
		flag.Parse()
	}

	// Se a semente for 0, usar o tempo atual para aleatoriedade.
	// Qualquer outro valor (positivo ou negativo) será usado diretamente.
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	return cfg
}

// NewAppConfig creates a new AppConfig, loading CLI parameters and using default simulation parameters.
// It also validates the loaded configuration.
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

// Validate checks the configuration for basic correctness.
func (ac *AppConfig) Validate() error {
	// Validate CLIConfig
	if ac.Cli.TotalNeurons <= 0 {
		return fmt.Errorf("total neurons must be positive, got %d", ac.Cli.TotalNeurons)
	}
	if ac.Cli.Cycles < 0 { // 0 cycles might be valid for some modes if only setup is needed
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

	// Validate SimulationParameters (exemplos)
	if ac.SimParams.MinInputNeurons <= 0 || ac.SimParams.MinOutputNeurons <= 0 {
		return fmt.Errorf("MinInputNeurons and MinOutputNeurons must be positive")
	}
	if ac.SimParams.PatternSize <= 0 {
		return fmt.Errorf("PatternSize must be positive")
	}
	if ac.SimParams.CyclesPerSecond <= 0 {
		return fmt.Errorf("CyclesPerSecond must be positive")
	}

	// Poderia adicionar mais validações para SimParams, como percentuais entre 0 e 1, etc.
	// Por exemplo:
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
```
