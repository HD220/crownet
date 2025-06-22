package main

import (
	"crownet/cli"
	"crownet/config"
	"fmt"
	"log"
	"os"
	// "math/rand"
	// "time"
)

func main() {
	// Carregar configurações da CLI e parâmetros de simulação padrão
	appCfg, err := config.NewAppConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar configuração: %v\n", err)
		os.Exit(1)
	}

	// Inicializar o orquestrador da CLI com as configurações
	orchestrator := cli.NewOrchestrator(appCfg)

	// Executar a lógica principal da aplicação através do orquestrador
	orchestrator.Run()

}
```
