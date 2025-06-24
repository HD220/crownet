# Arquitetura de Software - CrowNet MVP

Este documento descreve a arquitetura de software do CrowNet MVP, detalhando a organização dos pacotes, as principais estruturas de dados e os algoritmos centrais.

## 1. Visão Geral da Arquitetura

O CrowNet MVP é uma aplicação de linha de comando (CLI) desenvolvida em Go. A arquitetura não segue um padrão formal como MVC ou Arquitetura Hexagonal de forma estrita. Em vez disso, é organizada em **pacotes modulares**, cada um com responsabilidades bem definidas, promovendo separação de conceitos e coesão.

A interação principal flui do pacote `main` (que lida com a CLI) para o pacote `network` (que orquestra a simulação), que por sua vez utiliza outros pacotes como `neuron`, `datagen`, `storage`, etc.

### 1.1. Diagrama de Componentes de Alto Nível

O diagrama a seguir ilustra as interações entre os principais pacotes do sistema:

```ascii
                                     +-----------------+
                                     |      main       |
                                     |(CLI Entry Point)|
                                     +--------+--------+
                                              | (flags, mode)
                                              v
                                     +--------+--------+
                                     |      cli        |<--------------------+
                                     | (Orchestrator)  |                     |
                                     +--------+--------+                     |
                                              |                              | (config data)
              (simulation params)-------------+--------------(commands)      |
             /                                |                              |
            v                                 v                              |
+-----------+-----------+          +----------+----------+          +--------+--------+
|        config         |          |        network      |          |     storage     |
| (App/Sim Parameters)  |--------->|  (Core Simulation)  |<-------->| (JSON, SQLite)  |
+-----------------------+          +----------+----------+          +-----------------+
                                              ^
                                              |
                               +--------------+--------------+
                               | Dependencies of Network:   |
                               |                            |
                               |   +-------------------+    |
                               |   |      neuron       |    |
                               |   +-------------------+    |
                               |   |      pulse        |    |
                               |   +-------------------+    |
                               |   |   neurochemical   |    |
                               |   +-------------------+    |
                               |   |     synaptic      |    |
                               |   +-------------------+    |
                               |   |      space        |    |
                               |   +-------------------+    |
                               |   |     datagen       |    | // Input patterns
                               |   +-------------------+    |
                               |   |      common       |    | // Common types/utils
                               |   +-------------------+    |
                               +----------------------------+
```

**Legenda do Diagrama:**

*   **`main`**: Ponto de entrada da aplicação, processa argumentos da CLI.
*   **`cli`**: Orquestra a aplicação com base no modo e argumentos fornecidos. Utiliza `config` para obter parâmetros e instrui `network` para executar simulações.
*   **`config`**: Fornece dados de configuração (carregados de flags ou arquivos) para `cli` e `network`.
*   **`network`**: Componente central da aplicação. Recebe comandos e configurações, e utiliza pacotes especializados (`neuron`, `pulse`, `neurochemical`, `synaptic`, `space`, `datagen`, `common`) para realizar a simulação.
*   **`storage`**: Gerencia o salvamento e carregamento de dados (ex: pesos da rede, logs da simulação). `network` é seu cliente principal.
*   **Dependencies of `Network`**: Representa o conjunto de pacotes que fornecem as funcionalidades detalhadas para a simulação gerenciada por `network`. `datagen` fornece padrões de entrada, e `common` fornece tipos e utilitários compartilhados.

Este diagrama simplifica algumas interações para clareza, mas visa fornecer uma visão geral dos principais blocos arquiteturais e fluxos de dados.

## 2. Estrutura de Pacotes

A arquitetura do CrowNet é implementada através dos seguintes pacotes Go, cada um com responsabilidades distintas:

*   **`main`** (Ponto de Entrada)
    *   Localizado na raiz do projeto (`main.go`).
    *   Responsável por:
        *   Processar os argumentos da linha de comando (flags) utilizando o pacote `flag`.
        *   Invocar o pacote `cli` para orquestrar a execução com base nos argumentos fornecidos.
    *   Depende de: `cli`, `os`, `fmt`, `flag`.

*   **`cli`** (Orquestrador da Linha de Comando)
    *   Localizado em `cli/orchestrator.go`.
    *   Responsável por:
        *   Interpretar os argumentos da CLI e os modos de operação (`expose`, `observe`, `sim`).
        *   Coordenar a inicialização da configuração (`config`), da rede (`network`), e dos dados de entrada (`datagen`).
        *   Gerenciar o fluxo principal da simulação para o modo selecionado, incluindo carregar/salvar estados via `storage`.
        *   Apresentar feedback e resultados ao usuário no console.
    *   Depende de: `config`, `network`, `datagen`, `storage`, `fmt`, `log`, `time`.

*   **`config`** (Configuração da Aplicação e Simulação)
    *   Localizado em `config/config.go`.
    *   Responsável por:
        *   Definir as estruturas `AppConfig`, `CLIConfig`, e `SimulationParameters` que armazenam todos os parâmetros configuráveis.
        *   Carregar configurações a partir de flags da linha de comando.
        *   Validar os parâmetros de simulação.
        *   Disponibilizar estas configurações para outros pacotes.
    *   Depende de: `flag`, `fmt`, `os`, `time`, `encoding/json` (para potencial carregamento/salvamento futuro, não apenas flags).

*   **`common`** (Tipos e Utilitários Comuns)
    *   Localizado em `common/types.go`.
    *   Responsável por:
        *   Definir tipos de dados básicos e constantes usados em múltiplos pacotes (ex: `Point` para coordenadas 16D, `Dimensions`).
        *   Pode incluir funções utilitárias muito genéricas, se necessário.
    *   Não deve ter dependências significativas de outros pacotes do projeto.

*   **`datagen`** (Geração de Dados de Entrada)
    *   Localizado em `datagen/digits.go`.
    *   Responsável por:
        *   Fornecer os padrões de entrada para a simulação, especificamente os padrões binários 5x7 para os dígitos 0-9.
    *   Depende de: `config` (para validar dimensões dos padrões), `fmt`.

*   **`neuron`** (Lógica do Neurônio Individual)
    *   Localizado em `neuron/neuron.go` e `neuron/enums.go`.
    *   Responsável por:
        *   Definir a estrutura `Neuron` e seus atributos (ID, Posição, Tipo, Estado, AcumuladorDePulso, Limiares, Velocidade, etc.).
        *   Definir enums para `NeuronType` e `NeuronState`.
        *   Implementar a máquina de estados do neurônio (Repouso, Disparo, Refratário Absoluto, Refratário Relativo).
        *   Gerenciar a lógica de recebimento de pulso, atualização do potencial acumulado e decisão de disparo.
        *   Controlar o decaimento do potencial acumulado e o avanço pelos estados refratários.
    *   Depende de: `common` (para `Point`), `config` (para parâmetros como durações de estados, limiar base).

*   **`pulse`** (Gerenciamento e Propagação de Pulsos)
    *   Localizado em `pulse/pulse.go`.
    *   Responsável por:
        *   Definir a estrutura `Pulse` (origem, valor, posição atual, ciclo de criação, etc.) e `PulseList` para gerenciar a coleção de pulsos ativos.
        *   Implementar a lógica de propagação de pulsos através do espaço 16D, incluindo atualização de distância e raio de efeito.
        *   Determinar quais neurônios são atingidos por um pulso em um determinado ciclo.
    *   Depende de: `common` (para `Point`), `config` (para velocidade do pulso, raio máximo), `neuron`, `synaptic`, `space`.

*   **`space`** (Cálculos Espaciais)
    *   Localizado em `space/geometry.go` e `space/spatial_grid.go`.
    *   Responsável por:
        *   Fornecer funções utilitárias relacionadas ao espaço N-dimensional (atualmente 16D), como cálculo de distância Euclidiana e geração de posições aleatórias (`geometry.go`).
        *   Implementar estruturas de dados para indexação espacial, como `SpatialGrid`, para otimizar buscas de proximidade (ex: encontrar neurônios próximos a um pulso) (`spatial_grid.go`).
    *   Depende de: `math`, `math/rand`, `common` (para `Point`), `neuron` (para `SpatialGrid` armazenar neurônios), `fmt`.

*   **`synaptic`** (Gerenciamento de Pesos Sinápticos)
    *   Localizado em `synaptic/weights.go`.
    *   Responsável por:
        *   Definir a estrutura para armazenar e gerenciar os pesos sinápticos da rede (ex: `SynapticWeights` como `map[neuron.NeuronID]map[neuron.NeuronID]float64`).
        *   Fornecer funções para inicializar, obter, definir e modificar pesos sinápticos.
    *   Depende de: `neuron` (para `NeuronID`), `math/rand`.

*   **`neurochemical`** (Simulação de Neuroquímicos)
    *   Localizado em `neurochemical/neurochemicals.go`.
    *   Responsável por:
        *   Definir estruturas e lógica para simular os níveis de neuroquímicos (Cortisol, Dopamina).
        *   Gerenciar a produção (baseada em eventos da rede) e o decaimento desses químicos.
        *   Aplicar os efeitos modulatórios dos neuroquímicos nos limiares de disparo dos neurônios e nas taxas de aprendizado e sinaptogênese.
    *   Depende de: `config` (para parâmetros de efeito químico), `neuron` (para interagir com neurônios).

*   **`network`** (Orquestração da Rede Neural e Simulação Central)
    *   Localizado em `network/network.go`, `network/synaptogenesis.go`, `network/io_control.go`.
    *   Responsável por:
        *   Definir a estrutura `CrowNet` que encapsula todos os componentes da rede (neurônios, lista de pulsos ativos, pesos sinápticos, níveis químicos, configuração da simulação).
        *   Implementar o ciclo principal de simulação (`RunCycle`), orquestrando as atualizações de estado dos neurônios, propagação de pulsos, aprendizado Hebbiano, sinaptogênese e modulação química.
        *   Gerenciar a apresentação de padrões de entrada (`PresentPattern`) e a obtenção da ativação de saída (`GetOutputActivation`).
        *   Interagir com o pacote `storage` para persistência de pesos e logs.
    *   Depende de: `config`, `neuron`, `pulse`, `space`, `synaptic`, `neurochemical`, `datagen`, `storage`, `common`, `fmt`, `log`, `math`, `math/rand`, `sort`, `strings`, `time`.

*   **`storage`** (Persistência de Dados)
    *   Localizado em `storage/json_persistence.go` e `storage/sqlite_logger.go`.
    *   Responsável por:
        *   Salvar os pesos sinápticos da rede em arquivos JSON (`SaveSynapticWeights`).
        *   Carregar pesos sinápticos de arquivos JSON (`LoadSynapticWeights`).
        *   Opcionalmente, registrar snapshots detalhados do estado da rede (neurônios, químicos) em um banco de dados SQLite (`LogNetworkState`).
    *   Depende de: `config`, `neuron`, `synaptic`, `neurochemical`, `os`, `encoding/json`, `database/sql`, `github.com/mattn/go-sqlite3`, `fmt`, `log`, `path/filepath`, `time`.

## 3. Principais Estruturas de Dados

*   **`common.Point [16]float64`**: Representa uma coordenada no espaço 16D. Definido no pacote `common`.
*   **`neuron.Neuron`**:
    *   `ID neuron.NeuronID` (tipo encapsulado)
    *   `Position common.Point`
    *   `Type neuron.NeuronType` (Enum: Excitatory, Inhibitory, Dopaminergic, Input, Output)
    *   `State neuron.NeuronState` (Enum: Resting, Firing, AbsoluteRefractory, RelativeRefractory)
    *   `AccumulatedPulse float64`
    *   `BaseFiringThreshold float64` (configurável)
    *   `CurrentFiringThreshold float64` (modulado dinamicamente)
    *   `LastFiredCycle int`
    *   `CyclesInCurrentState int`
    *   `Velocity common.Point` (usado pela sinaptogênese)
*   **`pulse.Pulse`**:
    *   `EmittingNeuronID neuron.NeuronID`
    *   `OriginPosition common.Point`
    *   `Value float64` (sinal base: +1.0 para excitatório/input/output, -1.0 para inibitório)
    *   `CreationCycle int`
    *   `CurrentDistance float64`
    *   `MaxTravelRadius float64` (configurável, determina o alcance máximo do pulso)
*   **`pulse.PulseList`**:
    *   `Pulses []*pulse.Pulse` (slice de ponteiros para pulsos ativos)
    *   Responsável por gerenciar a coleção de pulsos ativos, processar seu ciclo de vida (propagação, interação, remoção) e facilitar a criação de novos pulsos.
*   **`synaptic.SynapticWeights`**:
    *   Internamente, provavelmente um `map[neuron.NeuronID]map[neuron.NeuronID]float64`.
    *   Encapsula a matriz de pesos sinápticos, fornecendo métodos para acesso e modificação.
*   **`network.CrowNet`**: Estrutura central que encapsula todo o estado e lógica da rede.
    *   `AppConfig *config.AppConfig` (contém `SimulationParameters`)
    *   `Neurons []*neuron.Neuron` (lista de todos os neurônios na rede)
    *   `NeuronMap map[neuron.NeuronID]*neuron.Neuron` (para acesso rápido aos neurônios por ID)
    *   `ActivePulses *pulse.PulseList` (gerenciador de pulsos ativos)
    *   `SynapticWeights *synaptic.SynapticWeights` (pesos sinápticos da rede)
    *   `InputNeuronIDSet map[neuron.NeuronID]struct{}` (conjunto de IDs de neurônios de entrada)
    *   `OutputNeuronIDSet map[neuron.NeuronID]struct{}` (conjunto de IDs de neurônios de saída)
    *   `Cortisol *neurochemical.Neurochemical` (estado do Cortisol)
    *   `Dopamine *neurochemical.Neurochemical` (estado da Dopamina)
    *   `CycleCount int` (contador de ciclos de simulação)
    *   `Stats *network.SimulationStats` (para coletar estatísticas da simulação)
    *   Outros campos para gerenciamento interno (ex: `rng *rand.Rand` para reprodutibilidade).

## 4. Principais Algoritmos

*   **Ciclo de Simulação (`CrowNet.RunCycle`)**: O coração da simulação, executado a cada passo de tempo.
    1.  **Processar Entradas Externas**: (Modo `sim`) Aplicar estímulos de frequência a neurônios de input.
    2.  **Atualizar Estados dos Neurônios**: Para cada neurônio:
        *   Decair `AccumulatedPulse`.
        *   Avançar estado na máquina de estados (Repouso, Disparo, Refratário Absoluto, Refratário Relativo) e `CyclesInCurrentState`.
    3.  **Processar Pulsos Ativos (`PulseList.ProcessCycle` otimizado com `SpatialGrid`)**:
        *   Para cada pulso ativo, sua propagação é atualizada.
        *   Em vez de verificar todos os neurônios, o `SpatialGrid` é consultado para obter uma lista de neurônios candidatos que estão espacialmente próximos à casca de efeito do pulso.
        *   Apenas para estes neurônios candidatos, é feita a verificação precisa de distância e, se aplicável, o efeito do pulso é calculado (`pulse.Value * SynapticWeight`) e integrado ao `AccumulatedPulse` do neurônio alvo.
        *   Pulsos que excederam seu `MaxTravelRadius` são removidos.
    4.  **Verificar Disparos Neuronais**: Para cada neurônio:
        *   Se `AccumulatedPulse > CurrentFiringThreshold` e não em estado refratário absoluto, o neurônio dispara.
        *   Mudar estado para `Firing`, registrar `LastFiredCycle`.
        *   Criar um novo `pulse.Pulse` originado deste neurônio e adicioná-lo à lista de pulsos ativos da rede.
    5.  **Atualizar Neuroquímicos (`neurochemical.Neurochemical.UpdateLevel`)**:
        *   Calcular produção de Cortisol (baseado em atividade perto da "glândula") e Dopamina (baseado em disparos de neurônios dopaminérgicos).
        *   Aplicar decaimento percentual aos níveis de Cortisol e Dopamina.
    6.  **Aplicar Efeitos dos Neuroquímicos (`neurochemical.ApplyEffectsToNeurons`, `neurochemical.GetModulationFactor`)**:
        *   Ajustar `CurrentFiringThreshold` de cada neurônio: O `BaseFiringThreshold` é modificado multiplicativamente pelos níveis de Cortisol e Dopamina, conforme os parâmetros `FiringThresholdIncreaseOnCort` e `FiringThresholdIncreaseOnDopa` em `SimulationParameters`. Por exemplo, `CurrentThreshold = BaseThreshold * (1 + CortisolEffectFactor) * (1 + DopamineEffectFactor)`. (Nota: Isto reflete uma modulação direta, não o efeito em "U" para o Cortisol que RF-CHEM-005 menciona - alinhando-se com a implementação descrita em `AGENTS.md`).
        *   Calcular fatores de modulação para aprendizado e sinaptogênese com base nos níveis químicos.
    7.  **Aplicar Plasticidade Hebbiana (`network.applyHebbianLearning`)**:
        *   Obter fator de modulação química para a taxa de aprendizado.
        *   Para cada conexão sináptica, se houver co-ativação recente (dentro de `HebbianCoincidenceWindow`) entre neurônio pré e pós-sináptico:
            *   Calcular `ΔPeso` usando `BaseLearningRate` (de `SimulationParameters`) e o fator de modulação.
            *   Atualizar peso: `NovoPeso = PesoAntigo + ΔPeso`.
        *   Aplicar decaimento de peso (`HebbianWeightDecay`).
        *   Garantir que os pesos permaneçam dentro dos limites `HebbianWeightMin`/`Max`.
    8.  **Aplicar Sinaptogênese (`network.applySynaptogenesis`)**:
        *   Se habilitado e em ciclo apropriado:
        *   Obter fator de modulação química para a taxa de movimento.
        *   Para cada neurônio, calcular a força líquida de atração/repulsão de outros neurônios (ativos atraem, inativos repelem).
        *   Atualizar `Velocity` do neurônio (com amortecimento e fator de modulação).
        *   Limitar `Velocity` a `MaxSynaptogenesisSpeed`.
        *   Atualizar `Position` do neurônio.
        *   Garantir que o neurônio permaneça dentro dos limites espaciais.
    9.  Incrementar `CycleCount` e coletar estatísticas.

*   **Modulação Química Detalhada**:
    *   **Limiares de Disparo**: Conforme descrito em `RunCycle` (item 6), os limiares são ajustados diretamente pelos níveis de Cortisol e Dopamina e seus respectivos fatores de influência (`FiringThresholdIncreaseOnCort`, `FiringThresholdIncreaseOnDopa`). Um valor positivo de `FiringThresholdIncreaseOnCort` significa que o Cortisol aumenta o limiar, tornando o disparo mais difícil. O mesmo se aplica à Dopamina. Os efeitos são combinados.
    *   **Taxa de Aprendizado e Sinaptogênese**: Os níveis de Cortisol e Dopamina são usados para calcular um `ModulationFactor` (tipicamente entre 0.0 e >1.0). Este fator escala a `BaseLearningRate` e a taxa de movimento da sinaptogênese. Altos níveis de Cortisol tendem a reduzir este fator, enquanto a Dopamina tende a aumentá-lo, conforme definido pelos parâmetros em `SimulationParameters`.

## 5. Padrões de Projeto Identificados/Sugeridos

*   **Configuração Centralizada / Injeção de Dependência (para Reescrita):** Usar um pacote `config` para centralizar parâmetros e injetá-los onde necessário, em vez de constantes globais espalhadas.
*   **Máquina de Estados:** O comportamento do `neuron.Neuron` (Resting, Firing, Refractory) é um exemplo claro do padrão State.
*   **Strategy (Implícito):** Os diferentes modos de operação (`expose`, `observe`, `sim`) podem ser vistos como diferentes estratégias de execução da simulação, orquestradas em `main` (ou futuro `cli`).
*   **Tipos Primitivos Encapsulados (Value Object):** Para a reescrita, usar tipos como `NeuronID`, `CycleCount` em vez de `int` puro para melhorar a semântica e segurança de tipo.
*   **Coleções de Primeira Classe (para Reescrita):** Estruturas como `synaptic.NetworkWeights` (gerenciando pesos sinápticos) e `pulse.PulseList` (gerenciando a coleção de pulsos ativos e seu processamento) são exemplos deste padrão. A reescrita pode formalizar outras coleções, como `type NeuronCollection []*Neuron`.
*   **Builder (Potencial):** Para a inicialização complexa de `CrowNet`, um padrão Builder poderia ser considerado na reescrita para torná-la mais fluida, embora a função `NewCrowNet` atual já lide com isso.

A arquitetura atual é funcional para o MVP. A reescrita focará em refinar a modularidade, aplicar princípios de Código Limpo e Object Calisthenics, e formalizar alguns desses padrões para melhorar a manutenibilidade e testabilidade.

## 6. Propostas de Evolução e Reescrita Futura

Esta seção descreve áreas potenciais para evolução da arquitetura e propostas de reescrita que podem melhorar ainda mais a manutenibilidade, extensibilidade, testabilidade, configurabilidade e desempenho do CrowNet. Estas propostas baseiam-se nos Requisitos Não Funcionais (especialmente RNF-EXT, RNF-CONF, RNF-TEST, RNF-REL) e na experiência adquirida durante o desenvolvimento e refatoração do MVP.

### 6.1. Gestão Avançada de Configuração
*   **Problema/Área:** Configuração atual primariamente via flags CLI.
*   **Proposta:** Implementar carregamento de parâmetros de simulação a partir de arquivos de configuração (ex: TOML, YAML), conforme RNF-CONF-001. Flags CLI poderiam especificar o arquivo ou sobrescrever valores.
*   **Benefícios:** Facilita o gerenciamento de configurações complexas, versionamento e compartilhamento. Reduz a necessidade de longas strings de comando.

### 6.2. Melhorias na Extensibilidade (RNF-EXT-001)
*   **Problema/Área:** Facilitar a adição de novos tipos de neurônios, efeitos neuroquímicos e regras de aprendizado.
*   **Propostas:**
    *   **Comportamento Neuronal via Interfaces:** Definir interfaces para comportamentos neuronais variáveis (ex: `FiringCondition`, `PulseEffectReceiver`). Neurônios concretos implementariam estas interfaces.
    *   **Registro de Efeitos Neuroquímicos:** Criar um sistema de registro onde diferentes neuroquímicos possam registrar suas funções de efeito específicas (modulação de limiar, modulação de taxa de aprendizado), em vez de lógica condicional centralizada.
    *   **Padrão Strategy para Regras de Aprendizado:** Se múltiplas rules de aprendizado forem previstas, refatorar a lógica de aprendizado para usar o padrão Strategy, permitindo a fácil substituição ou adição de algoritmos.
*   **Benefícios:** Aumenta a modularidade, reduz o acoplamento e simplifica a extensão das mecânicas centrais da simulação.

### 6.3. Aprimoramento da Testabilidade (RNF-TEST-001)
*   **Problema/Área:** Funções ou métodos complexos, especialmente no pacote `network`, que podem ser difíceis de testar unitariamente em isolamento.
*   **Propostas:**
    *   **Decomposição Adicional:** Avaliar funções complexas (ex: dentro de `network.RunCycle` ou seus helpers) para maior decomposição em unidades menores e testáveis.
    *   **Dependências Baseadas em Interfaces:** Para componentes chave como `PulseList`, `SynapticWeights`, e os gerenciadores de neuroquímicos, definir interfaces no pacote `network` (ou em um pacote de interfaces dedicado). `CrowNet` dependeria dessas interfaces, facilitando o uso de mocks em testes.
*   **Benefícios:** Permite testes unitários mais focados e robustos, melhora a confiança nas alterações e facilita refatorações seguras.

### 6.4. Reprodutibilidade Total e Gestão de Aleatoriedade (RNF-REL-001, FEATURE-002)
*   **Problema/Área:** Garantir que a simulação seja totalmente determinística com uma semente fornecida.
*   **Proposta:**
    *   Utilizar uma única instância de `rand.Rand` (semeada pela configuração) em toda a aplicação.
    *   Passar explicitamente esta instância `rng` para todas as funções que realizam operations aleatórias (posicionamento inicial, inicialização de pesos, etc.).
    *   Evitar o uso do `math/rand` global.
*   **Benefícios:** Assegura resultados idênticos para a mesma semente, fundamental para depuração, análise e validação de resultados.

### 6.5. Logging Avançado e Observabilidade
*   **Problema/Área:** Logging atual via SQLite é baseado em snapshots. Maior granularidade ou observabilidade em tempo real pode ser necessária.
*   **Propostas:**
    *   **Logging Estruturado de Eventos:** Implementar um sistema de logging de eventos mais detalhado para ocorrências chave (disparos neuronais específicos, alterações significativas de peso, etc.).
    *   **Métricas de Simulação:** Introduzir um sistema simples para coletar e expor/logar métricas de simulação ao longo do tempo (taxa média de disparo, níveis químicos, etc.).
*   **Benefícios:** Fornece insights mais profundos sobre a dinâmica da simulação e facilita a depuração de comportamentos complexos.

### 6.6. Otimizações de Desempenho (PERF-002, PERF-003)
*   **Problema/Área:** Potenciais gargalos em propagação de pulso e cálculos espaciais para redes maiores.
*   **Propostas:**
    *   **Otimização de `PulseList.ProcessCycle` (PERF-002):** Investigar e documentar a implementação de técnicas de indexação espacial (ex: k-d trees, grids) para otimizar a busca de neurônios vizinhos afetados por pulsos.
    *   **Otimização de `GenerateRandomPositionInHyperSphere` (PERF-003):** Avaliar e potencialmente implementar algoritmos mais eficientes para geração uniforme de pontos em hiperesferas (ex: método de Muller), se o método atual for um gargalo.
*   **Benefícios:** Permite simulações maiores e/ou mais rápidas, expandindo a capacidade de pesquisa da ferramenta.

### 6.7. Refinamentos Contínuos de Qualidade de Código (RNF-MAINT-003, RNF-MAINT-004)
*   **Problema/Área:** Manutenção contínua da aderência aos princípios de Código Limpo e Object Calisthenics.
*   **Proposta:** Realizar revisões periódicas do código com foco nestes princípios. Propor refatorações específicas e localizadas onde funções, structs ou pacotes possam ser simplificados, melhor encapsulados, ou ter sua legibilidade/manutenibilidade aprimorada.
*   **Benefícios:** Preserva e melhora a qualidade do código a longo prazo, facilitando a manutenção e futuras extensões.
```
