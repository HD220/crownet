package config

import (
	"flag"
	"os"
	"strings"
	"testing"
	"time"
	"crownet/common"
	"fmt"
)

func TestDefaultSimulationParameters(t *testing.T) {
	params := DefaultSimulationParameters()

	if params.SpaceMaxDimension != 10.0 {
		t.Errorf("Expected SpaceMaxDimension 10.0, got %f", params.SpaceMaxDimension)
	}
	if params.BaseFiringThreshold != 1.0 {
		t.Errorf("Expected BaseFiringThreshold 1.0, got %f", params.BaseFiringThreshold)
	}
	if params.MinInputNeurons != 35 {
		t.Errorf("Expected MinInputNeurons 35, got %d", params.MinInputNeurons)
	}
	// Add more checks for other important default values if necessary
	// For brevity, only a few are checked here. A more thorough test would check all critical defaults.
	if params.PatternHeight != 7 || params.PatternWidth != 5 || params.PatternSize != 35 {
		t.Errorf("Pattern dimensions/size default incorrect")
	}
	if params.CyclesPerSecond != 100.0 {
		t.Errorf("Expected CyclesPerSecond 100.0, got %f", params.CyclesPerSecond)
	}
}

func TestLoadCLIConfig_DefaultValues(t *testing.T) {
	fSet := flag.NewFlagSet("testDefaults", flag.ContinueOnError)
	args := []string{}
	cfg, err := LoadCLIConfig(fSet, args)
	if err != nil {
		t.Fatalf("LoadCLIConfig failed with default args: %v", err)
	}

	if cfg.Mode != ModeSim {
		t.Errorf("Expected default Mode %s, got %s", ModeSim, cfg.Mode)
	}
	if cfg.TotalNeurons != 200 {
		t.Errorf("Expected default TotalNeurons 200, got %d", cfg.TotalNeurons)
	}
	if cfg.WeightsFile != "crownet_weights.json" {
		t.Errorf("Expected default WeightsFile crownet_weights.json, got %s", cfg.WeightsFile)
	}
	if cfg.BaseLearningRate != 0.01 {
		t.Errorf("Expected default BaseLearningRate 0.01, got %f", cfg.BaseLearningRate)
	}
	if cfg.Seed == 0 { // Seed should be non-zero after defaulting to time
		t.Error("Expected default Seed to be initialized from time, but was 0")
	}
	// Test a few mode-specific defaults
	if cfg.Cycles != 1000 { // sim mode default
		t.Errorf("Expected default Cycles 1000, got %d", cfg.Cycles)
	}
	if cfg.Epochs != 50 { // expose mode default (though mode is sim)
		t.Errorf("Expected default Epochs 50, got %d", cfg.Epochs)
	}
}

func TestLoadCLIConfig_CustomValues(t *testing.T) {
	fSet := flag.NewFlagSet("testCustom", flag.ContinueOnError)
	args := []string{
		"-mode", "expose",
		"-neurons", "150",
		"-seed", "12345",
		"-weightsFile", "custom_weights.json",
		"-lrBase", "0.005",
		"-epochs", "100",
		"-cyclesPerPattern", "25",
	}
	cfg, err := LoadCLIConfig(fSet, args)
	if err != nil {
		t.Fatalf("LoadCLIConfig failed with custom args: %v", err)
	}

	if cfg.Mode != ModeExpose {
		t.Errorf("Expected Mode expose, got %s", cfg.Mode)
	}
	if cfg.TotalNeurons != 150 {
		t.Errorf("Expected TotalNeurons 150, got %d", cfg.TotalNeurons)
	}
	if cfg.Seed != 12345 {
		t.Errorf("Expected Seed 12345, got %d", cfg.Seed)
	}
	if cfg.WeightsFile != "custom_weights.json" {
		t.Errorf("Expected WeightsFile custom_weights.json, got %s", cfg.WeightsFile)
	}
	if cfg.BaseLearningRate != 0.005 {
		t.Errorf("Expected BaseLearningRate 0.005, got %f", cfg.BaseLearningRate)
	}
	if cfg.Epochs != 100 {
		t.Errorf("Expected Epochs 100, got %d", cfg.Epochs)
	}
	if cfg.CyclesPerPattern != 25 {
		t.Errorf("Expected CyclesPerPattern 25, got %d", cfg.CyclesPerPattern)
	}
}

func TestLoadCLIConfig_ErrorOnUnknownFlag(t *testing.T) {
	fSet := flag.NewFlagSet("testError", flag.ContinueOnError) // ContinueOnError is important for testing errors
	args := []string{"-unknownFlag", "value"}
	_, err := LoadCLIConfig(fSet, args)
	if err == nil {
		t.Error("Expected error for unknown flag, got nil")
	}
}

func TestNewAppConfig_Valid(t *testing.T) {
	// Temporarily set os.Args for this test if NewAppConfig uses os.Args directly.
	// However, our NewAppConfig now takes args []string.
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }() // Restore original os.Args

	testArgs := []string{ // Simulating os.Args[1:]
		"-mode", ModeSim,
		"-neurons", "50", // Valid number
	}

	appCfg, err := NewAppConfig(testArgs)
	if err != nil {
		t.Fatalf("NewAppConfig failed with valid args: %v", err)
	}
	if appCfg.Cli.Mode != ModeSim {
		t.Errorf("Expected Mode %s, got %s", ModeSim, appCfg.Cli.Mode)
	}
	if appCfg.Cli.TotalNeurons != 50 {
		t.Errorf("Expected TotalNeurons 50, got %d", appCfg.Cli.TotalNeurons)
	}
	// Check that SimParams are the default ones
	if appCfg.SimParams.SpaceMaxDimension != DefaultSimulationParameters().SpaceMaxDimension {
		t.Error("SimParams were not the default ones")
	}
}

func TestNewAppConfig_Invalid(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	testArgs := []string{
		"-mode", "invalid_mode",
	}
	_, err := NewAppConfig(testArgs)
	if err == nil {
		t.Fatal("NewAppConfig should have failed with invalid mode, but succeeded")
	}
	if !strings.Contains(err.Error(), "invalid mode 'invalid_mode'") {
		t.Errorf("Expected error message to contain 'invalid mode', got: %v", err)
	}
}


func TestAppConfig_Validate_ValidCases(t *testing.T) {
	defaultSimParams := DefaultSimulationParameters()

	tests := []struct {
		name    string
		cliCfg  CLIConfig
		wantErr bool
	}{
		{
			name: "valid sim mode",
			cliCfg: CLIConfig{Mode: ModeSim, TotalNeurons: 100, BaseLearningRate: 0.01, Cycles: 100, SaveInterval: 10},
			wantErr: false,
		},
		{
			name: "valid expose mode",
			cliCfg: CLIConfig{Mode: ModeExpose, TotalNeurons: 100, BaseLearningRate: 0.01, WeightsFile: "w.json", Epochs: 10, CyclesPerPattern: 10},
			wantErr: false,
		},
		{
			name: "valid observe mode",
			cliCfg: CLIConfig{Mode: ModeObserve, TotalNeurons: 100, BaseLearningRate: 0.01, WeightsFile: "w.json", Digit: 5, CyclesToSettle: 10},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCfg := &AppConfig{SimParams: defaultSimParams, Cli: tt.cliCfg}
			err := appCfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func TestAppConfig_Validate_InvalidCases(t *testing.T) {
	defaultSimParams := DefaultSimulationParameters()

	// Helper to quickly create a valid CLIConfig and then modify one field
	makeValidCliCfg := func() CLIConfig {
		return CLIConfig{
			Mode:             ModeSim,
			TotalNeurons:     100,
			Seed:             1,
			WeightsFile:      "weights.json",
			BaseLearningRate: 0.01,
			Cycles:           1000,
			DbPath:           "test.db",
			SaveInterval:     100,
			Epochs:           10, // Valid for expose
			CyclesPerPattern: 20, // Valid for expose
			Digit:            5,  // Valid for observe
			CyclesToSettle:   50, // Valid for observe
		}
	}

	tests := []struct {
		name        string
		modifier    func(cfg *CLIConfig) // Modifies a valid config to make it invalid
		expectedErr string
	}{
		{"invalid mode", func(cfg *CLIConfig) { cfg.Mode = "unknown_mode" }, "invalid mode 'unknown_mode'"},
		{"negative total neurons", func(cfg *CLIConfig) { cfg.TotalNeurons = -10 }, "total neurons must be positive"},
		{"negative learning rate", func(cfg *CLIConfig) { cfg.BaseLearningRate = -0.1 }, "baseLearningRate must be non-negative"},
		// Sim mode
		{"negative cycles sim", func(cfg *CLIConfig) { cfg.Mode = ModeSim; cfg.Cycles = -1 }, "cycles for sim mode must be non-negative"},
		{"negative save_interval sim", func(cfg *CLIConfig) { cfg.Mode = ModeSim; cfg.SaveInterval = -1 }, "saveInterval for sim mode must be non-negative"},
		// Expose mode
		{"missing weightsfile expose", func(cfg *CLIConfig) { cfg.Mode = ModeExpose; cfg.WeightsFile = "" }, "weightsFile must be specified for mode 'expose'"},
		{"zero epochs expose", func(cfg *CLIConfig) { cfg.Mode = ModeExpose; cfg.Epochs = 0 }, "epochs must be positive for mode 'expose'"},
		{"zero cycles_per_pattern expose", func(cfg *CLIConfig) { cfg.Mode = ModeExpose; cfg.CyclesPerPattern = 0 }, "cyclesPerPattern must be positive for mode 'expose'"},
		// Observe mode
		{"missing weightsfile observe", func(cfg *CLIConfig) { cfg.Mode = ModeObserve; cfg.WeightsFile = "" }, "weightsFile must be specified for mode 'observe'"},
		{"negative digit observe", func(cfg *CLIConfig) { cfg.Mode = ModeObserve; cfg.Digit = -1 }, "digit must be between 0-9 for mode 'observe'"},
		{"too large digit observe", func(cfg *CLIConfig) { cfg.Mode = ModeObserve; cfg.Digit = 10 }, "digit must be between 0-9 for mode 'observe'"},
		{"zero cycles_to_settle observe", func(cfg *CLIConfig) { cfg.Mode = ModeObserve; cfg.CyclesToSettle = 0 }, "cyclesToSettle must be positive for mode 'observe'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliCfg := makeValidCliCfg()
			tt.modifier(&cliCfg)
			appCfg := &AppConfig{SimParams: defaultSimParams, Cli: cliCfg}
			err := appCfg.Validate()
			if err == nil {
				t.Fatalf("Validate() expected error for %s, but got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Validate() error = %q, expected to contain %q", err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestAppConfig_Validate_SimParams(t *testing.T) {
	makeValidCliCfg := func() CLIConfig { // Re-define or pass if not in same file scope
		return CLIConfig{Mode: ModeSim, TotalNeurons: 100, BaseLearningRate: 0.01}
	}

	tests := []struct {
		name        string
		modifier    func(p *SimulationParameters)
		expectedErr string
	}{
		{"negative SpaceMaxDimension", func(p *SimulationParameters) { p.SpaceMaxDimension = -1.0 }, "SpaceMaxDimension must be positive"},
		{"zero BaseFiringThreshold", func(p *SimulationParameters) { p.BaseFiringThreshold = 0 }, "BaseFiringThreshold must be positive"},
		{"negative AccumulatedPulseDecayRate", func(p *SimulationParameters) { p.AccumulatedPulseDecayRate = -0.1 }, "AccumulatedPulseDecayRate must be non-negative"},
		{"negative AbsoluteRefractoryCycles", func(p *SimulationParameters) { p.AbsoluteRefractoryCycles = -1 }, "AbsoluteRefractoryCycles must be non-negative"},
		{"zero PulsePropagationSpeed", func(p *SimulationParameters) { p.PulsePropagationSpeed = 0 }, "PulsePropagationSpeed must be positive"},
		{"negative DopaminergicPercent", func(p *SimulationParameters) { p.DopaminergicPercent = -0.1 }, "DopaminergicPercent must be between 0.0 and 1.0"},
		{"too large InhibitoryPercent", func(p *SimulationParameters) { p.InhibitoryPercent = 1.1 }, "InhibitoryPercent must be between 0.0 and 1.0"},
		{"sum of percentages > 1.0", func(p *SimulationParameters) { p.DopaminergicPercent = 0.6; p.InhibitoryPercent = 0.5; }, "sum of DopaminergicPercent"},
		{"negative InitialSynapticWeightMin", func(p *SimulationParameters) { p.InitialSynapticWeightMin = -0.1 }, "InitialSynapticWeightMin must be non-negative"},
		{"InitialSynapticWeightMax < Min", func(p *SimulationParameters) { p.InitialSynapticWeightMin = 0.2; p.InitialSynapticWeightMax = 0.1 }, "InitialSynapticWeightMax"},
        {"MaxSynapticWeight < InitialMax", func(p *SimulationParameters) { p.InitialSynapticWeightMax = 0.6; p.MaxSynapticWeight = 0.5 }, "MaxSynapticWeight"},
		{"negative SynapticWeightDecayRate", func(p *SimulationParameters) { p.SynapticWeightDecayRate = -0.001 }, "SynapticWeightDecayRate must be non-negative"},
		{"zero HebbianCoincidenceWindow", func(p *SimulationParameters) { p.HebbianCoincidenceWindow = 0 }, "HebbianCoincidenceWindow must be positive"},
		{"negative HebbPositiveReinforceFactor", func(p *SimulationParameters) { p.HebbPositiveReinforceFactor = -0.1 }, "HebbPositiveReinforceFactor must be non-negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliCfg := makeValidCliCfg()
			simParams := DefaultSimulationParameters() // Start with valid defaults
			tt.modifier(&simParams)                   // Apply invalidating modification

			appCfg := &AppConfig{SimParams: simParams, Cli: cliCfg}
			err := appCfg.Validate()
			if err == nil {
				t.Fatalf("Validate() expected error for SimParams check %s, but got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Validate() error = %q, expected to contain %q for SimParams check %s", err.Error(), tt.expectedErr, tt.name)
			}
		})
	}
}

// Example of how to test LoadCLIConfig with specific arguments
func TestLoadCLIConfig_SimModeSpecificFlags(t *testing.T) {
	fSet := flag.NewFlagSet("testSimFlags", flag.ContinueOnError)
	args := []string{
		"-mode", "sim",
		"-cycles", "500",
		"-dbPath", "test.sqlite",
		"-saveInterval", "50",
		"-stimInputID", "10",
		"-stimInputFreqHz", "5.0",
		"-monitorOutputID", "20",
		"-debugChem",
	}
	cfg, err := LoadCLIConfig(fSet, args)
	if err != nil {
		t.Fatalf("LoadCLIConfig failed for sim specific flags: %v", err)
	}

	if cfg.Mode != ModeSim {
		t.Errorf("Expected Mode sim, got %s", cfg.Mode)
	}
	if cfg.Cycles != 500 {
		t.Errorf("Expected Cycles 500, got %d", cfg.Cycles)
	}
	if cfg.DbPath != "test.sqlite" {
		t.Errorf("Expected DbPath test.sqlite, got %s", cfg.DbPath)
	}
	if cfg.SaveInterval != 50 {
		t.Errorf("Expected SaveInterval 50, got %d", cfg.SaveInterval)
	}
	if cfg.StimInputID != 10 {
		t.Errorf("Expected StimInputID 10, got %d", cfg.StimInputID)
	}
	if cfg.StimInputFreqHz != 5.0 {
		t.Errorf("Expected StimInputFreqHz 5.0, got %f", cfg.StimInputFreqHz)
	}
	if cfg.MonitorOutputID != 20 {
		t.Errorf("Expected MonitorOutputID 20, got %d", cfg.MonitorOutputID)
	}
	if !cfg.DebugChem {
		t.Error("Expected DebugChem to be true")
	}
}
