package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// Importar config para acessar AppConfig e CLIConfig futuramente, se necessário aqui
	// "crownet/config"
)

var (
	// Usado para flags globais que podem ser vinculadas a uma struct de configuração
	// cfg *config.AppConfig // Exemplo, se quiséssemos carregar AppConfig aqui

	// Flags Globais/Persistentes
	configFile string
	seed       int64
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crownet",
	Short: "CrowNet: Simulador de Rede Neural Bio-inspirada",
	Long: `CrowNet é uma aplicação de linha de comando escrita em Go que simula
um modelo computacional de rede neural bio-inspirada.
Para mais detalhes sobre um comando específico, use: crownet [comando] --help`,
	// Run: func(cmd *cobra.Command, args []string) { }, // Descomente se o comando raiz precisar fazer algo
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig) // Se precisar de inicialização de config via Viper, por exemplo

	// Definir flags globais/persistentes aqui
	rootCmd.PersistentFlags().StringVar(&configFile, "configFile", "", "Caminho para o arquivo de configuração TOML (funcionalidade planejada).")
	rootCmd.PersistentFlags().Int64Var(&seed, "seed", 0, "Semente para o gerador de números aleatórios (0 usa o tempo atual).")

	// Exemplo de como vincular a uma struct de config global (se necessário)
	// cfg = &config.AppConfig{} // Inicializar
	// rootCmd.PersistentFlags().IntVar(&cfg.Cli.TotalNeurons, "neurons", 200, "Total de neurônios na rede.")
	// ... mas muitas flags são específicas de comando, então serão definidas nos subcomandos.
}

// initConfig seria usado se tivéssemos Viper ou similar para carregar config de arquivo.
// func initConfig() {
// 	if configFile != "" {
// 		viper.SetConfigFile(configFile)
// 	} else {
// 		// ... lógica para encontrar config em locais padrão
// 	}
// 	viper.AutomaticEnv()
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Println("Usando arquivo de configuração:", viper.ConfigFileUsed())
// 	}
// }
