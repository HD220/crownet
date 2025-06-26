package cmd

import (
	"fmt"
	"log"

	"crownet/cli"
	"crownet/common"
	"crownet/config"
	"github.com/spf13/cobra"
	"github.com/BurntSushi/toml" // FEATURE-CONFIG-001
)

var (
	// Flags para o comando expose
	exposeEpochs           int
	exposeCyclesPerPattern int
	exposeTotalNeurons     int    // Duplicates global 'totalNeurons' but specific to expose if needed, or use global
	exposeWeightsFile      string // Duplicates global 'weightsFile'
	exposeBaseLearningRate float64
	exposeDbPath           string // Duplicates global 'dbPath'
	exposeSaveInterval     int    // Duplicates global 'saveInterval'
	exposeDebugChem        bool   // Duplicates global 'debugChem'
)

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Executa o modo de exposição/treinamento da rede.",
	Long: `O modo expose é usado para treinar a rede neural apresentando
sequências de padrões de entrada (e.g. dígitos) e ajustando os pesos sinápticos
através de aprendizado Hebbiano modulado por neuroquímicos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		if cmd.Flags().Changed("seed") { appCfg.Cli.Seed = seed }
		if cmd.Flags().Changed("neurons") { appCfg.Cli.TotalNeurons = exposeTotalNeurons }
		if cmd.Flags().Changed("weightsFile") { appCfg.Cli.WeightsFile = exposeWeightsFile }
		if cmd.Flags().Changed("lrBase") { appCfg.Cli.BaseLearningRate = common.Rate(exposeBaseLearningRate) }
		if cmd.Flags().Changed("epochs") { appCfg.Cli.Epochs = exposeEpochs }
		if cmd.Flags().Changed("cyclesPerPattern") { appCfg.Cli.CyclesPerPattern = exposeCyclesPerPattern }
		if cmd.Flags().Changed("dbPath") { appCfg.Cli.DbPath = exposeDbPath }
		if cmd.Flags().Changed("saveInterval") { appCfg.Cli.SaveInterval = exposeSaveInterval }
		if cmd.Flags().Changed("debugChem") { appCfg.Cli.DebugChem = exposeDebugChem }

		if err := appCfg.Validate(); err != nil {
			return fmt.Errorf("configuração inválida para o modo expose: %w", err)
		}

		orchestrator := cli.NewOrchestrator(appCfg)
		if err := orchestrator.Run(); err != nil {
			return fmt.Errorf("erro durante a execução do modo expose: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)

	exposeCmd.Flags().IntVarP(&exposeEpochs, "epochs", "e", 50, "Número de épocas de exposição aos padrões.")
	exposeCmd.Flags().IntVar(&exposeCyclesPerPattern, "cyclesPerPattern", 20, "Número de ciclos de simulação por apresentação de padrão.")

	// Flags de simulação relevantes para expose
	exposeCmd.Flags().IntVarP(&exposeTotalNeurons, "neurons", "n", 200, "Total de neurônios na rede.")
	exposeCmd.Flags().StringVarP(&exposeWeightsFile, "weightsFile", "w", "crownet_weights.json", "Arquivo para salvar/carregar pesos sinápticos.")
	_ = exposeCmd.MarkFlagRequired("weightsFile") // Salvar pesos é essencial após expose
	exposeCmd.Flags().Float64Var(&exposeBaseLearningRate, "lrBase", 0.01, "Taxa de aprendizado base.")
	exposeCmd.Flags().StringVar(&exposeDbPath, "dbPath", "", "Caminho opcional para o arquivo SQLite para logging durante o expose.")
	exposeCmd.Flags().IntVar(&exposeSaveInterval, "saveInterval", 0, "Intervalo de ciclos para salvar no BD durante expose (0 desabilita).")
	exposeCmd.Flags().BoolVar(&exposeDebugChem, "debugChem", false, "Habilita logs de depuração para neuroquímicos.")
}
