package cmd

import (
	// "fmt" // Unused import
	"os"

	"github.com/spf13/cobra"
	// Importar config para acessar AppConfig e CLIConfig futuramente, se necessário aqui
	// "crownet/config"
)

// Flags globais/persistentes que podem ser acessadas por subcomandos.
var (
	configFile string // Caminho para o arquivo de configuração TOML.
	seed       int64  // Semente para o gerador de números aleatórios.
)

// rootCmd representa o comando base da aplicação CrowNet quando chamado sem subcomandos.
// Ele configura as flags globais e adiciona todos os subcomandos da aplicação.
var rootCmd = &cobra.Command{
	Use:   "crownet",
	Short: "CrowNet: Simulador de Rede Neural Bio-inspirada",
	Long: `CrowNet é uma aplicação de linha de comando escrita em Go que simula
um modelo computacional de rede neural bio-inspirada.
Inclui funcionalidades para simulação, treinamento (exposição a padrões),
observação de respostas da rede e utilitários de log.

Para mais detalhes sobre um comando específico, use: crownet [comando] --help`,
	// Run: func(cmd *cobra.Command, args []string) { }, // O comando raiz não executa nenhuma ação direta.
}

// Execute é o principal ponto de entrada para a CLI baseada em Cobra.
// Ele executa o comando raiz, que por sua vez lida com o parsing de argumentos
// e a execução do subcomando apropriado. Chamado por main.main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra já imprime o erro no Stderr por padrão.
		// fmt.Fprintln(os.Stderr, err) // Redundante
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
