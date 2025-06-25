package cmd

import (
	"fmt"
	"log"

	"crownet/cli"
	"crownet/common"
	"crownet/config"
	"github.com/spf13/cobra"
)

var (
	// Flags para o comando expose
	exposeEpochs           int
	exposeCyclesPerPattern int
	// Flags de simulação também usadas por expose
	exposeTotalNeurons     int
	exposeWeightsFile      string
	exposeBaseLearningRate float64
	exposeDbPath           string // Opcional para expose
	exposeSaveInterval    int    // Opcional para expose
	exposeDebugChem        bool
)

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Treina a rede expondo-a a padrões de dígitos.",
	Long: `Executa a fase de treinamento da rede. Durante esta fase, a rede é
apresentada repetidamente a padrões de dígitos (0-9), permitindo que o
aprendizado Hebbiano ocorra e os pesos sinápticos sejam ajustados.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Executando modo expose via Cobra...")

		cliCfg := config.CLIConfig{
			Mode:             config.ModeExpose,
			TotalNeurons:     exposeTotalNeurons,
			Seed:             seed, // Flag global/persistente de rootCmd
			WeightsFile:      exposeWeightsFile,
			BaseLearningRate: common.Rate(exposeBaseLearningRate),
			Epochs:           exposeEpochs,
			CyclesPerPattern: exposeCyclesPerPattern,
			DbPath:           exposeDbPath,       // Usado se saveInterval > 0
			SaveInterval:     exposeSaveInterval, // Para logging opcional durante expose
			DebugChem:        exposeDebugChem,
		}

		appCfg := &config.AppConfig{
			SimParams: config.DefaultSimulationParameters(),
			Cli:       cliCfg,
		}

		// TODO: Lógica de configFile
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
