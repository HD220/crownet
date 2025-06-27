package cmd

import (
	"fmt"
	"log"

	"crownet/cli"
	"crownet/common"
	"crownet/config"
	"github.com/spf13/cobra"
	"github.com/BurntSushi/toml" // Added for TOML decoding
	"os"                         // For pprof file creation
	"runtime/pprof"              // For CPU and memory profiling
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

	// Profiling flags
	simCpuProfileFile string
	simMemProfileFile string
)

// simCmd represents the sim command
var simCmd = &cobra.Command{
	Use:   "sim",
	Short: "Executa uma simulação geral da rede CrowNet.",
	Long: `Executa uma simulação geral com todas as dinâmicas da rede (aprendizado,
sinaptogênese, neuromodulação) ativas. Útil para observação de comportamento
ou logging detalhado para análise posterior.`,
	// "github.com/BurntSushi/toml" // This was a misplaced import comment

	RunE: func(cmd *cobra.Command, args []string) error {
		// CPU Profiling
		if simCpuProfileFile != "" {
			f, err := os.Create(simCpuProfileFile)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			defer pprof.StopCPUProfile()
			fmt.Printf("CPU profiling enabled, saving to %s\n", simCpuProfileFile)
		}

		fmt.Println("Executando modo sim via Cobra...")

		// 1. Inicializar AppConfig com valores padrão das flags Cobra e SimParams defaults
		appCfg := &config.AppConfig{
			SimParams: config.DefaultSimulationParameters(),
			Cli: config.CLIConfig{ // Populate com os valores padrão das flags (que já estão nas vars)
				Mode:             config.ModeSim,
				TotalNeurons:     simTotalNeurons,
				Seed:             seed, // da flag global
				WeightsFile:      simWeightsFile,
				BaseLearningRate: common.Rate(simBaseLearningRate),
				Cycles:           simCycles,
				DbPath:           simDbPath,
				SaveInterval:     simSaveInterval,
				StimInputID:      simStimInputID,
				StimInputFreqHz:  simStimInputFreqHz,
				MonitorOutputID:  simMonitorOutputID,
				DebugChem:        simDebugChem,
			},
		}

		// 2. Carregar de arquivo TOML se especificado (sobrescreve os padrões acima)
		if configFile != "" {
			fmt.Printf("Carregando configuração do arquivo TOML: %s\n", configFile)
			// Salvar uma cópia da CLIConfig antes de DecodeFile, para aplicar flags CLI depois
			cliCfgBeforeToml := appCfg.Cli // Make sure toml is imported for DecodeFile
			if _, err := toml.DecodeFile(configFile, appCfg); err != nil { // toml will be undefined if not imported
				log.Printf("Aviso: erro ao decodificar arquivo TOML '%s': %v. Continuando com padrões/flags CLI.", configFile, err)
				// Restaurar CLIConfig se TOML falhou, para que flags CLI ainda possam funcionar sobre defaults
				appCfg.Cli = cliCfgBeforeToml
			}
		}

		// 3. Aplicar flags CLI que foram *explicitamente setadas* pelo usuário,
		//    sobrescrevendo valores do TOML ou dos padrões das flags.
		//    A flag global 'seed' já foi aplicada na inicialização de appCfg.Cli.Seed.
		//    Se 'configFile' setou 'seed', a flag global '--seed' (se usada) irá sobrescrevê-la.
		if cmd.Flags().Changed("seed") { appCfg.Cli.Seed = seed }


		if cmd.Flags().Changed("neurons") { appCfg.Cli.TotalNeurons = simTotalNeurons }
		if cmd.Flags().Changed("weightsFile") { appCfg.Cli.WeightsFile = simWeightsFile }
		if cmd.Flags().Changed("lrBase") { appCfg.Cli.BaseLearningRate = common.Rate(simBaseLearningRate) }
		if cmd.Flags().Changed("cycles") { appCfg.Cli.Cycles = simCycles }
		if cmd.Flags().Changed("dbPath") { appCfg.Cli.DbPath = simDbPath }
		if cmd.Flags().Changed("saveInterval") { appCfg.Cli.SaveInterval = simSaveInterval }
		if cmd.Flags().Changed("stimInputID") { appCfg.Cli.StimInputID = simStimInputID }
		if cmd.Flags().Changed("stimInputFreqHz") { appCfg.Cli.StimInputFreqHz = simStimInputFreqHz }
		if cmd.Flags().Changed("monitorOutputID") { appCfg.Cli.MonitorOutputID = simMonitorOutputID }
		if cmd.Flags().Changed("debugChem") { appCfg.Cli.DebugChem = simDebugChem }

		// Nota: Se flags CLI puderem modificar SimParams diretamente, essa lógica de merge
		// precisaria ser estendida para SimParams também. Por ora, SimParams só vem de
		// DefaultSimulationParameters() e do arquivo TOML.

		if err := appCfg.Validate(); err != nil {
			return fmt.Errorf("configuração inválida para o modo sim: %w", err)
		}

		// A flag configFile é global, mas sua lógica de carregamento (FEATURE-CONFIG-001)
		// ainda não está implementada. Quando estiver, precisará ser integrada aqui
		// para potencialmente sobrescrever SimParams ou Cli antes da validação.

		orchestrator := cli.NewOrchestrator(appCfg)
		runErr := orchestrator.Run() // Run agora vai internamente chamar runSimMode

		// Memory Profiling (Heap)
		if simMemProfileFile != "" && runErr == nil { // Only write mem profile if run was successful
			f, err := os.Create(simMemProfileFile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			defer f.Close()
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			fmt.Printf("Memory heap profile saved to %s\n", simMemProfileFile)
		}

		if runErr != nil {
			// log.Fatalf já lida com a saída e exit em Orchestrator.Run se houver erro crítico.
			// Se Run retornar erro, é um erro de execução do modo.
			return fmt.Errorf("erro durante a execução do modo sim: %w", runErr)
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

	// Profiling flags
	simCmd.Flags().StringVar(&simCpuProfileFile, "cpuprofile", "", "Escreve perfil de CPU para este arquivo.")
	simCmd.Flags().StringVar(&simMemProfileFile, "memprofile", "", "Escreve perfil de memória para este arquivo.")
}
