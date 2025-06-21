package main

import (
	"crownet/cli"
	"crownet/config"
	"fmt"
	"os"
	// "math/rand" // Seed é melhor no orchestrator ou aqui, se global
	// "time"
)

func main() {
	// Carregar configurações da CLI e parâmetros de simulação padrão
	appCfg := config.NewAppConfig()

	// Inicializar o orquestrador da CLI com as configurações
	orchestrator := cli.NewOrchestrator(appCfg)

	// Tratar erros de `flag.Parse()` que podem ocorrer em `config.LoadCLIConfig()`
	// `flag.Parse()` chama `os.Exit(2)` em caso de erro, então não precisamos tratar aqui explicitamente
	// a menos que mudemos o `flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)`

	// Executar a lógica principal da aplicação através do orquestrador
	// `orchestrator.Run()` lidará com a lógica específica do modo e erros fatais.
	// Se `Run()` retornar um erro, podemos tratá-lo aqui.
	// Por enquanto, `Run()` usa `log.Fatalf` para erros que impedem a continuação.

	// Exemplo de como poderia ser se Run retornasse erro:
	// if err := orchestrator.Run(); err != nil {
	// 	 fmt.Fprintf(os.Stderr, "Erro na execução: %v\n", err)
	// 	 os.Exit(1)
	// }

	// Como Run usa log.Fatalf, main apenas chama Run.
	// rand.Seed(time.Now().UnixNano()) // Seed global, se desejado. Pode ser melhor dentro do orchestrator se a reprodutibilidade for por simulação.

	// fmt.Println("Iniciando CrowNet a partir do main.go...") // Debug
	orchestrator.Run()

	// Mensagem de finalização agora está dentro de orchestrator.Run()
	// fmt.Println("CrowNet main.go concluído.")
}
```
