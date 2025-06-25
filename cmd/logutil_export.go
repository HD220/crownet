package cmd

import (
	"fmt"
	"log"

	"crownet/config" // Para validar e usar as flags de logutil
	"crownet/storage"
	"github.com/spf13/cobra"
)

var (
	logutilExportDbPath string
	logutilExportTable  string
	logutilExportFormat string
	logutilExportOutput string
)

// logutilExportCmd represents the logutil export command
var logutilExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exporta dados de uma tabela do log SQLite para um formato especificado (ex: CSV).",
	Long: `Lê um arquivo de banco de dados SQLite gerado pelo CrowNet e exporta
os dados da tabela especificada. Atualmente, suporta exportação para CSV.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Executando logutil export via Cobra...")

		// Usar as flags globais e as locais para popular uma CLIConfig temporária para validação
		// ou passar diretamente para a função de exportação.
		// A validação das flags já é feita em AppConfig.Validate() se usarmos essa via.
		// Para simplificar, vamos chamar a função de exportação diretamente.
		// Primeiro, validamos as flags usando uma AppConfig temporária.
		tempCliCfg := config.CLIConfig{
			Mode:              config.ModeLogUtil, // Necessário para acionar a validação correta
			LogUtilSubcommand: "export",          // Hardcoded pois este é o comando de exportação
			LogUtilDbPath:     logutilExportDbPath,
			LogUtilTable:      logutilExportTable,
			LogUtilFormat:     logutilExportFormat,
			LogUtilOutput:     logutilExportOutput,
		}
		tempAppCfg := &config.AppConfig{Cli: tempCliCfg}
		if err := tempAppCfg.Validate(); err != nil {
			return fmt.Errorf("configuração inválida para logutil export: %w", err)
		}

		// Path para LogUtilDbPath já foi validado em cli.Orchestrator.validatePath
		// quando chamado por runLogUtilMode. Aqui, assumimos que o Orchestrator
		// não será usado para logutil, então precisamos de uma validação de caminho.
		// No entanto, a validação em config.AppConfig.Validate() já checa se LogUtilDbPath não está vazio.
		// A validação de existência do arquivo será feita por storage.ExportLogData.

		fmt.Printf("  Database: %s\n", logutilExportDbPath)
		fmt.Printf("  Table: %s\n", logutilExportTable)
		fmt.Printf("  Format: %s\n", logutilExportFormat)
		if logutilExportOutput != "" {
			fmt.Printf("  Output: %s\n", logutilExportOutput)
		} else {
			fmt.Println("  Output: stdout")
		}

		err := storage.ExportLogData(
			logutilExportDbPath,
			logutilExportTable,
			logutilExportFormat,
			logutilExportOutput,
		)
		if err != nil {
			// Usar log.Printf para erros não fatais que não devem parar o Cobra em si,
			// mas sim reportar o erro da operação.
			log.Printf("Erro durante a exportação do log: %v", err)
			return err // Retornar o erro para que Cobra o reporte
		}
		fmt.Println("Exportação do log concluída com sucesso.")
		return nil
	},
}

func init() {
	logutilCmd.AddCommand(logutilExportCmd)

	logutilExportCmd.Flags().StringVarP(&logutilExportDbPath, "dbPath", "d", "", "Caminho para o arquivo SQLite DB (obrigatório).")
	_ = logutilExportCmd.MarkFlagRequired("dbPath")

	logutilExportCmd.Flags().StringVarP(&logutilExportTable, "table", "t", "", "Tabela a ser exportada (ex: 'NetworkSnapshots', 'NeuronStates') (obrigatório).")
	_ = logutilExportCmd.MarkFlagRequired("table")

	logutilExportCmd.Flags().StringVarP(&logutilExportFormat, "format", "f", "csv", "Formato de saída (atualmente apenas 'csv').")
	logutilExportCmd.Flags().StringVarP(&logutilExportOutput, "output", "o", "", "Arquivo de saída (stdout se não especificado).")
}
