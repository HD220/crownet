# Guia para Agentes LLM Trabalhando no Projeto CrowNet

Olá, colega Agente! Este documento serve como um guia para ajudá-lo a entender e trabalhar efetivamente no codebase do projeto CrowNet.

## 1. Visão Geral do Projeto

**CrowNet** é uma simulação de rede neural inspirada em aspectos biológicos, com foco em aprendizado, plasticidade sináptica e os efeitos de neuromoduladores como cortisol e dopamina. O sistema pode operar em diferentes modos:
*   **`sim`**: Simulação geral da dinâmica da rede ao longo do tempo, com possíveis estímulos externos.
*   **`expose`**: Treinamento da rede através da exposição a padrões específicos (atualmente, dígitos de 0 a 9), onde a rede ajusta seus pesos sinápticos via aprendizado Hebbiano.
*   **`observe`**: Apresentação de um padrão específico a uma rede pré-treinada para observar a ativação de seus neurônios de saída.

O objetivo é modelar um sistema que possa aprender e cujo comportamento seja influenciado por um ambiente químico interno.

## 2. Estrutura de Pacotes e Responsabilidades

O projeto é organizado nos seguintes pacotes principais:

*   **`main.go`**: Ponto de entrada da aplicação. Lida com a inicialização da configuração e do orquestrador da CLI.
*   **`cli` (`cli/orchestrator.go`)**: Responsável por interpretar os argumentos da linha de comando e orquestrar a execução do modo de simulação selecionado. Contém a lógica de alto nível para cada modo.
*   **`config` (`config/config.go`)**: Define as estruturas de configuração (`AppConfig`, `SimulationParameters`, `CLIConfig`), carrega parâmetros da CLI (usando o pacote `flag`), fornece valores padrão para a simulação e valida a configuração.
*   **`common` (`common/types.go`)**: Define tipos de dados básicos e fundamentais usados em todo o projeto (ex: `NeuronID`, `CycleCount`, `Point`, `Vector`, `PulseValue`, `Rate`, `Level`).
*   **`network` (`network/network.go`)**: O coração da simulação. Contém a struct `CrowNet` que gerencia todos os componentes da rede (neurônios, pulsos, pesos sinápticos, ambiente químico) e executa o ciclo de simulação principal (`RunCycle`).
*   **`neuron` (`neuron/neuron.go`, `neuron/enums.go`)**: Define a struct `Neuron`, seus estados (`Resting`, `Firing`, etc.), tipos (`Excitatory`, `Input`, etc.), e sua lógica de funcionamento individual (integração de potencial, disparo, períodos refratários, decaimento de potencial).
*   **`pulse` (`pulse/pulse.go`)**: Define a struct `Pulse` e `PulseList`. Gerencia a propagação de pulsos pela rede, seu tempo de vida e a aplicação de seus efeitos nos neurônios.
*   **`space` (`space/geometry.go`)**: Contém funções geométricas para lidar com o espaço N-dimensional (atualmente 16D) onde os neurônios residem, como cálculo de distância e posicionamento aleatório.
*   **`synaptic` (`synaptic/weights.go`)**: Define a estrutura `NetworkWeights` para armazenar os pesos das conexões sinápticas. Inclui lógica para inicialização de pesos, obtenção/configuração de pesos e aplicação da regra de aprendizado Hebbiano.
*   **`neurochemical` (`neurochemical/neurochemicals.go`)**: Modela o ambiente neuroquímico (Cortisol, Dopamina). Gerencia a produção, decaimento e os efeitos desses neuroquímicos na taxa de aprendizado, sinaptogênese e limiares de disparo dos neurônios.
*   **`datagen` (`datagen/digits.go`)**: Responsável por fornecer dados de padrões para a rede (atualmente, padrões visuais para dígitos 0-9).
*   **`storage` (`storage/json_persistence.go`, `storage/sqlite_logger.go`)**: Lida com a persistência de dados. `json_persistence.go` salva e carrega os pesos da rede em formato JSON. `sqlite_logger.go` registra o estado detalhado da rede (neurônios, químicos, etc.) em um banco de dados SQLite durante a simulação.

## 3. Principais Decisões de Design e Refatorações Recentes

O projeto passou por uma refatoração sistemática significativa. As principais mudanças e decisões incluem:

*   **Configuração Centralizada e Validada:** Todos os parâmetros da simulação e da CLI são agora definidos e validados no pacote `config`. `SimulationParameters` é abrangente.
*   **Reprodutibilidade (RNG Semeado):** A simulação agora usa uma fonte de números aleatórios (`*rand.Rand`) local para a instância `CrowNet`, semeada através de um parâmetro da CLI (`-seed`). Isso permite execuções reproduzíveis. O `rand` global não é mais usado diretamente pela lógica principal da simulação.
*   **Tratamento de Erros Aprimorado:** Funções nos módulos de lógica e no orquestrador da CLI foram refatoradas para retornar erros em vez de usar `log.Fatalf` indiscriminadamente, permitindo melhor tratamento e propagação de erros.
*   **Modularidade e Coesão:** Funções longas foram quebradas em unidades menores e mais focadas. Pacotes têm responsabilidades mais claras.
*   **Testes Unitários:** Foram adicionados testes unitários para a maioria dos pacotes de lógica (`neuron`, `pulse`, `space`, `synaptic`, `neurochemical`, `datagen`, `network` parcialmente, `cli` parcialmente), cobrindo funcionalidades chave e casos de borda.
*   **Lógica Neuroquímica Simplificada:** A modulação dos parâmetros da rede (taxa de aprendizado, sinaptogênese, limiar de disparo) por cortisol e dopamina foi simplificada para usar fatores de influência direta (configuráveis em `SimulationParameters`) em vez de modelos anteriores mais complexos baseados em múltiplos limiares e curvas em U. Os parâmetros de configuração para a lógica complexa foram removidos para alinhar com a implementação atual. Se for necessário maior fidelidade, essa lógica complexa pode ser reintroduzida.
*   **Injeção de Dependência para Testes:** Em `cli.Orchestrator`, as funções de acesso ao `storage` foram transformadas em campos de função para permitir mocking durante os testes.

## 4. Como Executar e Testar

*   **Compilação:** `go build` na raiz do projeto.
*   **Execução:** O executável aceita várias flags. Use `-h` ou `--help` para ver todas as opções.
    *   Exemplo Simulação: `./crownet -mode sim -cycles 200 -neurons 100 -dbPath sim_results.db -saveInterval 50`
    *   Exemplo Treino: `./crownet -mode expose -epochs 10 -weightsFile my_weights.json`
    *   Exemplo Observação: `./crownet -mode observe -digit 5 -weightsFile my_weights.json`
*   **Testes Unitários:** Execute `go test ./...` na raiz do projeto para rodar todos os testes unitários.
*   **Testes de Casos de Uso (Sistema):** Consulte o arquivo `docs/TESTING_SCENARIOS.md` para uma lista de cenários de teste de ponta a ponta e como executá-los manualmente. Estes testes validam o comportamento da CLI e a interação entre os módulos.

## 5. Pontos de Atenção e Futuras Melhorias

*   **Lógica Neuroquímica:** A lógica atual é simplificada. Se for necessário um modelo mais complexo (ex: curva em U para o efeito do cortisol no limiar de disparo, ou múltiplos estágios de efeito na taxa de aprendizado), isso precisará ser reimplementado em `neurochemical/neurochemicals.go` e os parâmetros correspondentes em `config/config.go` reativados/ajustados.
*   **Testes de Sistema Automatizados:** Os cenários em `TESTING_SCENARIOS.md` são manuais. Seria valioso automatizá-los usando scripts ou um framework de teste em Go que execute o binário.
*   **Performance:** Para simulações muito grandes, a performance de `PulseList.ProcessCycle` (complexidade `O(num_pulsos * num_neurônios)`) e da sinaptogênese pode precisar de otimização (ex: usando estruturas de dados espaciais).
*   **Logging:** O logging atual usa `fmt.Printf` para informações e `log.Printf` / `log.Fatalf` para avisos/erros. Um sistema de logging mais estruturado e configurável (ex: Logrus, Zap) seria uma melhoria.
*   **Interface de Rede para Testes:** Para testar `cli.Orchestrator` de forma mais isolada, a dependência `Orchestrator.Net (*network.CrowNet)` poderia ser substituída por uma interface, permitindo mocks completos da rede.

## 6. Convenções de Código

*   Siga as convenções padrão de Go (`gofmt`, `golint` implicitamente).
*   Mantenha os comentários atualizados, especialmente para funções e tipos públicos.
*   Priorize a clareza e a legibilidade.
*   Adicione testes unitários para novas funcionalidades ou ao corrigir bugs.

## 7. Abordagem de Testes e Uso de Mocks

O projeto CrowNet emprega uma estratégia de testes em várias camadas:

*   **Testes Unitários:** Cada pacote (`neuron`, `pulse`, `synaptic`, etc.) possui seus próprios testes (`_test.go`) que verificam a lógica interna de suas funções e tipos de forma isolada.
*   **Testes de "Quase Integração" para a CLI:** O pacote `cli_test` contém testes que verificam o `Orchestrator`. Para isolar a lógica do orquestrador de dependências externas como o sistema de arquivos (`storage`) ou a geração de dados (`datagen`), são usadas algumas técnicas:
    *   **Injeção de Dependência de Função:** O `Orchestrator` possui campos como `loadWeightsFn` e `saveWeightsFn`. No código de produção, eles usam as funções reais do pacote `storage`. Nos testes, eles podem ser sobrescritos por funções mock para simular sucesso, falha ou capturar dados.
    *   **Variáveis de Função de Pacote Mockáveis:** Funções como `datagen.GetDigitPattern` são expostas como variáveis (`datagen.GetDigitPatternFn`) que podem ser temporariamente substituídas nos testes por implementações mock.
    *   **Importante:** Esses "mocks" são apenas para fins de teste e não indicam funcionalidade ausente no código de produção. As implementações reais são usadas por padrão.
*   **Testes de Sistema (Manuais/Planejados):** Os documentos `TESTING_SCENARIOS.md` e `docs/use_cases.md` detalham cenários de teste de ponta a ponta que envolvem a execução do binário compilado com várias flags. Um exemplo de relatório de execução simulada programaticamente pode ser encontrado em `docs/execution_reports/`.
*   **Documentação Adicional:** Consulte `docs/testing_approach.md` para uma descrição mais detalhada da filosofia e das camadas de teste.

Ao trabalhar nos testes ou no código que eles cobrem, é importante entender o papel dessas técnicas de mock para evitar confusão.

Esperamos que este guia seja útil! Boas contribuições!
```
