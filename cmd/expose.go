package cmd

import (
	"fmt"
	"log"

	"os"            // For pprof file creation
	"runtime/pprof" // For CPU and memory profiling

	"github.com/BurntSushi/toml" // FEATURE-CONFIG-001
	"github.com/spf13/cobra"

	"crownet/cli"
	"crownet/common"
	"crownet/config"
)

var (
	// Flags para o commando expose
	exposeEpochs           int
	exposeCyclesPerPattern int
	exposeTotalNeurons     int    // Duplicates global 'totalNeurons' but specific to expose if needed, or use global
	exposeWeightsFile      string // Duplicates global 'weightsFile'
	exposeBaseLearningRate float64
	exposeDbPath           string // Duplicates global 'dbPath'
	exposeSaveInterval     int    // Duplicates global 'saveInterval'
	exposeDebugChem        bool   // Duplicates global 'debugChem'
	// Profiling flags
	exposeCPUProfileFile string // Renamed from exposeCpuProfileFile
	exposeMemProfileFile string
)

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Executa o modo de exposição/treinamento da rede.",
	Long: `O modo expose é usado para treinar a rede neural apresentando
sequências de padrões de entrada (e.g. dígitos) e ajustando os pesos sinápticos
através de aprendizado Hebbiano modulado por neuroquímicos.`,
	RunE: func(cmd *cobra.Command, _ []string) error { // args renamed to _
		// CPU Profiling
		if exposeCPUProfileFile != "" {
			f, err := os.Create(exposeCPUProfileFile)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			defer pprof.StopCPUProfile()
			fmt.Printf("CPU profiling enabled, saving to %s\n", exposeCPUProfileFile)
		}

		fmt.Println("Executando modo expose via Cobra...")

		// 1. Inicializar AppConfig com valores padrão das flags Cobra e SimParams defaults
		appCfg := &config.AppConfig{
			SimParams: config.DefaultSimulationParameters(),
			Cli: config.CLIConfig{
				Mode:             config.ModeExpose,
				TotalNeurons:     exposeTotalNeurons,
				Seed:             seed,
				WeightsFile:      exposeWeightsFile,
				BaseLearningRate: common.Rate(exposeBaseLearningRate),
				Epochs:           exposeEpochs,
				CyclesPerPattern: exposeCyclesPerPattern,
				DbPath:           exposeDbPath,
				SaveInterval:     exposeSaveInterval,
				DebugChem:        exposeDebugChem,
			},
		}

		// 2. Carregar de arquivo TOML se especificado
		if configFile != "" {
			fmt.Printf("Carregando configuração do arquivo TOML: %s\n", configFile)
			cliCfgBeforeToml := appCfg.Cli
			if _, err := toml.DecodeFile(configFile, appCfg); err != nil {
				log.Printf("Aviso: erro ao decodificar arquivo TOML '%s': %v. Continuando.", configFile, err)
				appCfg.Cli = cliCfgBeforeToml
			}
		}

		// 3. Aplicar flags CLI explicitamente setadas
		if cmd.Flags().Changed("seed") {
			appCfg.Cli.Seed = seed
		}
		if cmd.Flags().Changed("neurons") {
			appCfg.Cli.TotalNeurons = exposeTotalNeurons
		}
		if cmd.Flags().Changed("weightsFile") {
			appCfg.Cli.WeightsFile = exposeWeightsFile
		}
		if cmd.Flags().Changed("lrBase") {
			appCfg.Cli.BaseLearningRate = common.Rate(exposeBaseLearningRate)
		}
		if cmd.Flags().Changed("epochs") {
			appCfg.Cli.Epochs = exposeEpochs
		}
		if cmd.Flags().Changed("cyclesPerPattern") {
			appCfg.Cli.CyclesPerPattern = exposeCyclesPerPattern
		}
		if cmd.Flags().Changed("dbPath") {
			appCfg.Cli.DbPath = exposeDbPath
		}
		if cmd.Flags().Changed("saveInterval") {
			appCfg.Cli.SaveInterval = exposeSaveInterval
		}
		if cmd.Flags().Changed("debugChem") {
			appCfg.Cli.DebugChem = exposeDebugChem
		}

		if err := appCfg.Validate(); err != nil {
			return fmt.Errorf("configuração inválida para o modo expose: %w", err)
		}

		orchestrator := cli.NewOrchestrator(appCfg)
		runErr := orchestrator.Run() // Store error from Run

		// Memory Profiling (Heap)
		if exposeMemProfileFile != "" && runErr == nil { // Only write mem profile if run was successful
			f, err := os.Create(exposeMemProfileFile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			defer f.Close()
			// runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			fmt.Printf("Memory heap profile saved to %s\n", exposeMemProfileFile)
		}

		if runErr != nil {
			return fmt.Errorf("erro durante a execução do modo expose: %w", runErr)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)

	exposeCmd.Flags().IntVarP(&exposeEpochs, "epochs", "e", 50, "Número de épocas de exposição aos padrões.")
	exposeCmd.Flags().IntVar(&exposeCyclesPerPattern, "cyclesPerPattern", 20,
		"Número de ciclos de simulação por apresentação de padrão.")

	// Flags de simulação relevantes para expose
	exposeCmd.Flags().IntVarP(&exposeTotalNeurons, "neurons", "n", 200, "Total de neurônios na rede.")
	exposeCmd.Flags().StringVarP(&exposeWeightsFile, "weightsFile", "w", "crownet_weights.json",
		"Arquivo para salvar/carregar pesos sinápticos.")
	if err := exposeCmd.MarkFlagRequired("weightsFile"); err != nil { // Salvar pesos é essential após expose
		log.Printf("Warning: could not mark 'weightsFile' as required for exposeCmd: %v", err)
	}
	exposeCmd.Flags().Float64Var(&exposeBaseLearningRate, "lrBase", 0.01, "Taxa de aprendizado base.")
	exposeCmd.Flags().StringVar(&exposeDbPath, "dbPath", "",
		"Caminho opcional para o arquivo SQLite para logging durante o expose.")
	exposeCmd.Flags().IntVar(&exposeSaveInterval, "saveInterval", 0,
		"Intervalo de ciclos para salvar no BD durante expose (0 desabilita).")
	exposeCmd.Flags().BoolVar(&exposeDebugChem, "debugChem", false, "Habilita logs de depuração para neuroquímicos.")

	// Profiling flags
	exposeCmd.Flags().StringVar(&exposeCPUProfileFile, "cpuprofile", "", "Escreve perfil de CPU para este arquivo.")
	exposeCmd.Flags().StringVar(&exposeMemProfileFile, "memprofile", "", "Escreve perfil de memória para este arquivo.")
}
