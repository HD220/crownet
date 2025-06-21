package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"crownet/src/core"
	"crownet/src/database"
	"crownet/src/io"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds) // Adiciona microssegundos ao log
	log.Println("Iniciando Simulação CrowNet...")

	// Parâmetros da linha de comando
	numCycles := flag.Int("cycles", 0, "Número de ciclos de simulação a executar automaticamente. 0 para modo interativo.")
	logInterval := flag.Int("loginterval", 10, "Intervalo de ciclos para salvar o estado da rede no banco de dados.")
	loadCycle := flag.Uint64("loadfrom", 0, "Carregar estado da rede a partir deste ciclo do banco de dados.")
	configFile := flag.String("config", "", "Caminho para o arquivo de configuração da rede (JSON, não implementado no MVP).")

	flag.Parse()

	// Inicializar banco de dados
	if err := database.InitDB(); err != nil {
		log.Fatalf("Falha ao inicializar banco de dados: %v", err)
	}
	defer database.CloseDB()

	var nn *core.NeuralNetwork
	var initialConfig core.Config

	if *loadCycle > 0 {
		log.Printf("Tentando carregar estado da rede do ciclo %d...", *loadCycle)
		var err error
		nn, err = database.LoadNetworkState(*loadCycle)
		if err != nil {
			log.Fatalf("Falha ao carregar estado da rede: %v. Verifique se o ciclo existe e o DB está correto.", err)
		}
		// A configuração original usada para criar a rede não é salva/carregada explicitamente no LoadNetworkState.
		// Se precisarmos dela (ex: para PulsePropagationSpeed), teríamos que armazená-la ou usar defaults.
		// Por ora, o LoadNetworkState recria os neurônios e seus estados, e o default config é usado implicitamente para alguns params.
		// Vamos carregar uma config default e alguns parâmetros podem ser sobrescritos pelo estado carregado.
		initialConfig = core.GetDefaultConfig() // Carrega defaults
		nn.PulsePropagationSpeed = initialConfig.PulsePropagationSpeed // Garante que está setado
		nn.MaxSpaceDistance = 8.0 // Conforme README, garantir que está setado

		log.Printf("Estado da rede carregado com sucesso do ciclo %d. Próximo ciclo a ser simulado: %d.", *loadCycle, nn.CurrentCycle+1)

	} else {
		// Usar configuração padrão se nenhum arquivo de config for fornecido (ou se o parsing falhar)
		// TODO: Implementar carregamento de config de arquivo JSON
		if *configFile != "" {
			log.Printf("Carregamento de arquivo de configuração (%s) não implementado no MVP. Usando defaults.", *configFile)
		}
		initialConfig = core.GetDefaultConfig()
		nn = core.InitializeNetwork(initialConfig)

		// Salvar estado inicial dos neurônios (IDs, tipos, posições iniciais)
		if err := database.SaveInitialNeurons(nn.Neurons); err != nil {
			log.Printf("Aviso: Falha ao salvar estado inicial dos neurônios: %v", err)
			// Continuar mesmo assim, mas o DB não terá os dados base dos neurônios se for a primeira execução.
		}
		log.Println("Rede neural inicializada com configuração padrão.")
		// Logar o estado inicial da rede (ciclo 0)
		if err := database.LogNetworkState(0, nn); err != nil {
			log.Printf("Falha ao logar estado inicial da rede (ciclo 0): %v", err)
		} else {
			log.Println("Estado inicial da rede (ciclo 0) logado no banco de dados.")
		}
	}

	consoleManager := io.NewConsoleManager(nn)

	if *numCycles > 0 { // Modo automático
		log.Printf("Executando %d ciclos em modo automático...", *numCycles)
		startTime := time.Now()
		for i := 0; i < *numCycles; i++ {
			nn.SimulateCycle()
			if nn.CurrentCycle%uint64(*logInterval) == 0 {
				if err := database.LogNetworkState(nn.CurrentCycle, nn); err != nil {
					log.Printf("Falha ao logar estado da rede no ciclo %d: %v", nn.CurrentCycle, err)
				} else {
					log.Printf("Estado da rede no ciclo %d logado no banco de dados.", nn.CurrentCycle)
				}
			}
		}
		duration := time.Since(startTime)
		log.Printf("Simulação automática de %d ciclos concluída em %s.", *numCycles, duration)
		avgCycleTime := float64(duration.Nanoseconds()) / float64(*numCycles) / 1e6 // em milissegundos
		log.Printf("Tempo médio por ciclo: %.3f ms", avgCycleTime)

		// Log final após o loop automático
		if nn.CurrentCycle%uint64(*logInterval) != 0 { // Logar se o último ciclo não foi um intervalo de log
			if err := database.LogNetworkState(nn.CurrentCycle, nn); err != nil {
				log.Printf("Falha ao logar estado final da rede no ciclo %d: %v", nn.CurrentCycle, err)
			} else {
				log.Printf("Estado final da rede (ciclo %d) logado.", nn.CurrentCycle)
			}
		}
		consoleManager.DisplayNetworkStatus()
		consoleManager.DisplayOutput()

	} else { // Modo interativo
		log.Println("Iniciando modo interativo. Digite 'help' para comandos.")
		// Listar neurônios de input/output disponíveis
		consoleManager.GetInputNeuronIDs()
		consoleManager.GetOutputNeuronIDs()

		running := true
		for running {
			// No modo interativo, o comando 'run <ciclos>' dentro do CheckForInput
			// já lida com a simulação e o logging.
			running = consoleManager.CheckForInput()

			// Se o comando 'run' foi executado dentro de CheckForInput, o CurrentCycle já foi avançado.
			// O logging é feito dentro do loop de 'run' em ConsoleManager.
			// No entanto, precisamos garantir que, se o usuário não rodar ciclos explicitamente
			// mas fizer outras alterações (ex: 'input'), o estado seja salvo ao sair ou periodicamente.
			// Por simplicidade, o logging no modo interativo é acoplado ao comando 'run'.
			// Um 'save' explícito poderia ser um comando.
		}
	}

	log.Println("Simulação CrowNet encerrada.")
}

// Getwd mostra o diretório de trabalho atual (para debug de caminhos de arquivo)
func Getwd() {
	wd, err := os.Getwd()
	if err != nil {
		log.Println("Erro ao obter diretório de trabalho:", err)
	} else {
		log.Println("Diretório de trabalho atual:", wd)
	}
}
