package cmd

import (
	"fmt"
	"log"

	"crownet/cli"
	"crownet/config"
	"github.com/spf13/cobra"
)

var (
	// Flags para o comando observe
	observeDigit          int
	observeCyclesToSettle int
	// Flags de simulação também usadas por observe
	observeTotalNeurons int
	observeWeightsFile  string
	observeDebugChem    bool
)

var observeCmd = &cobra.Command{
	Use:   "observe",
	Short: "Observa a resposta da rede a um dígito específico.",
	Long: `Carrega pesos sinápticos previamente treinados e apresenta um padrão
de dígito específico à rede, observando o padrão de ativação resultante
nos neurônios de saída.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Executando modo observe via Cobra...")

		cliCfg := config.CLIConfig{
			Mode:           config.ModeObserve,
			TotalNeurons:   observeTotalNeurons,
			Seed:           seed, // Flag global/persistente de rootCmd
			WeightsFile:    observeWeightsFile,
			Digit:          observeDigit,
			CyclesToSettle: observeCyclesToSettle,
			DebugChem:      observeDebugChem, // Embora menos usado, mantido por consistência
		}

		appCfg := &config.AppConfig{
			SimParams: config.DefaultSimulationParameters(),
			Cli:       cliCfg,
		}

		// TODO: Lógica de configFile
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
