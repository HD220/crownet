package config

import "flag"

// SimulationParameters holds intrinsic parameters of the simulation dynamics.
type SimulationParameters struct { // Mantendo simplificado por enquanto
	SpaceMaxDimension        float64
	BaseFiringThreshold    float64
	PulsePropagationSpeed  float64
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
}

// AppConfig bundles all configuration.
type AppConfig struct {
	SimParams SimulationParameters
	Cli       CLIConfig
}

// DefaultSimulationParameters returns a drastically simplified SimulationParameters struct.
func DefaultSimulationParameters() SimulationParameters {
	return SimulationParameters{
		SpaceMaxDimension:        4.0,
		BaseFiringThreshold:       1.0,
		PulsePropagationSpeed:     0.6,
	}
}

// LoadCLIConfig parses command-line flags and returns a CLIConfig struct.
func LoadCLIConfig() CLIConfig {
	cfg := CLIConfig{}

	flag.IntVar(&cfg.TotalNeurons, "neurons", 100, "Total number of neurons in the network.")
	flag.IntVar(&cfg.Cycles, "cycles", 100, "Total simulation cycles for 'sim' mode.")
	flag.IntVar(&cfg.StimInputID, "stimInputID", -1, "ID of an input neuron for general continuous stimulus in 'sim' mode (-1 for first available, -2 to disable).")
	flag.Float64Var(&cfg.StimInputFreqHz, "stimInputFreqHz", 0.0, "Frequency (Hz) for general stimulus in 'sim' mode (0.0 to disable).")
	flag.IntVar(&cfg.MonitorOutputID, "monitorOutputID", -1, "ID of an output neuron to monitor for frequency reporting in 'sim' mode (-1 for first available).")
	flag.StringVar(&cfg.DbPath, "dbPath", "crownet_data.db", "Path for the SQLite database file.")
	flag.IntVar(&cfg.SaveInterval, "saveInterval", 0, "Cycle interval for saving to DB (0 to disable periodic saves, only final if any).")
	flag.BoolVar(&cfg.DebugChem, "debugChem", false, "Enable debug prints for chemical production.")

	flag.StringVar(&cfg.Mode, "mode", "sim", "Operation mode: 'sim', 'expose', or 'observe'.")
	// flag.IntVar(&cfg.Epochs, "epochs", 50, "Number of exposure epochs (for 'expose' mode).")
	// flag.StringVar(&cfg.WeightsFile, "weightsFile", "crownet_digit_weights.json", "File to save/load synaptic weights.")
	// flag.IntVar(&cfg.Digit, "digit", 0, "Digit (0-9) to present (for 'observe' mode).")
	// flag.Float64Var(&cfg.BaseLearningRate, "lrBase", 0.005, "Base learning rate for Hebbian plasticity.")
	// flag.IntVar(&cfg.CyclesPerPattern, "cyclesPerPattern", 5, "Number of cycles to run per pattern presentation during 'expose' mode.")
	// flag.IntVar(&cfg.CyclesToSettle, "cyclesToSettle", 5, "Number of cycles to run for network settling during 'observe' mode.")

	if !flag.Parsed() {
		flag.Parse()
	}
	return cfg
}

// NewAppConfig creates a new AppConfig, loading CLI parameters and using default simulation parameters.
func NewAppConfig() *AppConfig {
	return &AppConfig{
		SimParams: DefaultSimulationParameters(),
		Cli:       LoadCLIConfig(),
	}
}
```
