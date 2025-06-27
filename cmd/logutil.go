package cmd

import (
	// "fmt" // Unused import

	"github.com/spf13/cobra"
)

// logutilCmd represents the base logutil command
var logutilCmd = &cobra.Command{
	Use:   "logutil",
	Short: "Utilitários para interagir com logs SQLite gerados pelo CrowNet.",
	Long: `O commando logutil fornece subcomandos para processar e exportar dados
dos arquivos de log SQLite criados durante as simulações de CrowNet.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error { // cmd and args renamed to _
		// Validar aqui se o subcomando é conhecido, se necessário, embora Cobra faça isso.
		// Pode ser usado para carregar configurações globais para todos os subcomandos de logutil.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logutilCmd)
	// Flags persistentes para todos os subcomandos de logutil podem ser definidas aqui,
	// mas para 'export', as flags são mais específicas e definidas em logutil_export.go
}
