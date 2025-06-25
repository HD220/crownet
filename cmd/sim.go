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
	// Flags para o comando sim
	simCycles          int
	simDbPath          string
	simSaveInterval    int
	simStimInputID     int
	simStimInputFreqHz float64
	simMonitorOutputID int
	simDebugChem       bool

	// Flags que eram globais, agora específicas para comandos de simulação
	simTotalNeurons     int
	simWeightsFile      string
	simBaseLearningRate float64
)

// simCmd represents the sim command
var simCmd = &cobra.Command{
	Use:   "sim",
	Short: "Executa uma simulação geral da rede CrowNet.",
	Long: `Executa uma simulação geral com todas as dinâmicas da rede (aprendizado,
sinaptogênese, neuromodulação) ativas. Útil para observação de comportamento
ou logging detalhado para análise posterior.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Executando modo sim via Cobra...")

		cliCfg := config.CLIConfig{
			Mode:             config.ModeSim, // Definido pelo comando
			TotalNeurons:     simTotalNeurons,
			Seed:             seed, // Flag global/persistente de rootCmd
			WeightsFile:      simWeightsFile,
			BaseLearningRate: common.Rate(simBaseLearningRate),
			Cycles:           simCycles,
			DbPath:           simDbPath,
			SaveInterval:     simSaveInterval,
			StimInputID:      simStimInputID,
			StimInputFreqHz:  simStimInputFreqHz,
			MonitorOutputID:  simMonitorOutputID,
			DebugChem:        simDebugChem,
			// Outras flags de outros modos não são relevantes aqui
		}

		appCfg := &config.AppConfig{
			SimParams: config.DefaultSimulationParameters(), // Começa com padrões
			Cli:       cliCfg,
		}

		// TODO: Implementar carregamento de configFile se fornecido globalmente,
		// e mesclar com appCfg ANTES da validação.

		if err := appCfg.Validate(); err != nil {
			return fmt.Errorf("configuração inválida para o modo sim: %w", err)
		}

		// A flag configFile é global, mas sua lógica de carregamento (FEATURE-CONFIG-001)
		// ainda não está implementada. Quando estiver, precisará ser integrada aqui
		// para potencialmente sobrescrever SimParams ou Cli antes da validação.

		orchestrator := cli.NewOrchestrator(appCfg)
		if err := orchestrator.Run(); err != nil { // Run agora vai internamente chamar runSimMode
			// log.Fatalf já lida com a saída e exit em Orchestrator.Run se houver erro crítico.
			// Se Run retornar erro, é um erro de execução do modo.
			return fmt.Errorf("erro durante a execução do modo sim: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(simCmd)

	// Flags específicas do comando 'sim'
	simCmd.Flags().IntVarP(&simCycles, "cycles", "c", 1000, "Total de ciclos de simulação para o modo 'sim'.")
	simCmd.Flags().StringVar(&simDbPath, "dbPath", "crownet_sim_run.db", "Caminho para o arquivo SQLite para logging.")
	simCmd.Flags().IntVar(&simSaveInterval, "saveInterval", 100, "Intervalo de ciclos para salvar no BD (0 desabilita saves periódicos).")
	simCmd.Flags().IntVar(&simStimInputID, "stimInputID", -1, "ID do neurônio de entrada para estímulo contínuo (-1: primeiro disponível, -2: desabilitado).")
	simCmd.Flags().Float64Var(&simStimInputFreqHz, "stimInputFreqHz", 0.0, "Frequência (Hz) para estímulo contínuo (0.0 desabilita).")
	simCmd.Flags().IntVar(&simMonitorOutputID, "monitorOutputID", -1, "ID do neurônio de saída para monitorar frequência (-1: primeiro disponível, -2: desabilitado).")
	simCmd.Flags().BoolVar(&simDebugChem, "debugChem", false, "Habilita logs de depuração para produção de neuroquímicos.")

	// Flags que eram "globais" mas são contextuais aos modos de simulação
	simCmd.Flags().IntVarP(&simTotalNeurons, "neurons", "n", 200, "Total de neurônios na rede.")
	simCmd.Flags().StringVarP(&simWeightsFile, "weightsFile", "w", "crownet_weights.json", "Arquivo para salvar/carregar pesos sinápticos.")
	simCmd.Flags().Float64Var(&simBaseLearningRate, "lrBase", 0.01, "Taxa de aprendizado base para plasticidade Hebbiana.")
	// A flag 'seed' é persistente no rootCmd
}
