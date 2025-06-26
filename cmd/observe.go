package cmd

import (
	"fmt"
	"log"

	"crownet/cli"
	"crownet/config"
	"github.com/spf13/cobra"
	"github.com/BurntSushi/toml" // FEATURE-CONFIG-001
)

var (
	// Flags para o comando observe
	observeDigit            int
	observeCyclesToSettle   int
	observeTotalNeurons     int    // Duplicates global 'totalNeurons'
	observeWeightsFile      string // Duplicates global 'weightsFile'
	observeDebugChem        bool   // Duplicates global 'debugChem'
)

var observeCmd = &cobra.Command{
	Use:   "observe",
	Short: "Executa o modo de observação da rede.",
	Long: `O modo observe é usado para apresentar um padrão específico (e.g. um dígito)
à rede (com pesos previamente treinados) e observar o padrão de ativação dos neurônios de saída.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Executando modo observe via Cobra...")

		// 1. Inicializar AppConfig com valores padrão das flags Cobra e SimParams defaults
		appCfg := &config.AppConfig{
			SimParams: config.DefaultSimulationParameters(),
			Cli: config.CLIConfig{
				Mode:           config.ModeObserve,
				TotalNeurons:   observeTotalNeurons,
				Seed:           seed,
				WeightsFile:    observeWeightsFile,
				Digit:          observeDigit,
				CyclesToSettle: observeCyclesToSettle,
				DebugChem:      observeDebugChem,
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
		if cmd.Flags().Changed("neurons") { appCfg.Cli.TotalNeurons = observeTotalNeurons }
		if cmd.Flags().Changed("weightsFile") { appCfg.Cli.WeightsFile = observeWeightsFile }
		if cmd.Flags().Changed("digit") { appCfg.Cli.Digit = observeDigit }
		if cmd.Flags().Changed("cyclesToSettle") { appCfg.Cli.CyclesToSettle = observeCyclesToSettle }
		if cmd.Flags().Changed("debugChem") { appCfg.Cli.DebugChem = observeDebugChem }

		if err := appCfg.Validate(); err != nil {
			return fmt.Errorf("configuração inválida para o modo observe: %w", err)
		}

		orchestrator := cli.NewOrchestrator(appCfg)
		if err := orchestrator.Run(); err != nil {
			return fmt.Errorf("erro durante a execução do modo observe: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(observeCmd)

	observeCmd.Flags().IntVarP(&observeDigit, "digit", "d", 0, "Dígito (0-9) a ser apresentado.")
	observeCmd.Flags().IntVar(&observeCyclesToSettle, "cyclesToSettle", 50, "Número de ciclos para acomodação da rede.")

	// Flags de simulação relevantes para observe
	observeCmd.Flags().IntVarP(&observeTotalNeurons, "neurons", "n", 200, "Total de neurônios na rede (deve corresponder à rede dos pesos carregados).")
	observeCmd.Flags().StringVarP(&observeWeightsFile, "weightsFile", "w", "crownet_weights.json", "Arquivo para carregar os pesos sinápticos.")
	_ = observeCmd.MarkFlagRequired("weightsFile")
	observeCmd.Flags().BoolVar(&observeDebugChem, "debugChem", false, "Habilita logs de depuração para neuroquímicos.")
}
