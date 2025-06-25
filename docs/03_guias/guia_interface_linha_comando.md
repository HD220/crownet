# Guia de Interface de Linha de Comando (CLI) - CrowNet MVP

Este documento descreve o estilo e a estrutura da interface de linha de comando para a aplicação CrowNet MVP, incluindo o formato dos comandos, a saída esperada no console e a estrutura dos arquivos de dados relacionados.

## 1. Formato Geral dos Comandos

A aplicação é executada como um único binário (`crownet` após compilação). Suas operações são controladas por subcomandos e flags associadas.

**Estrutura Geral:**
```bash
./crownet [comando_global_flags] <comando> [subcomando_opcional] [flags_do_comando]
```
Exemplos:
```bash
./crownet --seed 12345 sim --cycles 1000 --neurons 150
./crownet logutil export --dbPath ./run.db --table NetworkSnapshots
```
Para obter ajuda sobre um comando específico:
```bash
./crownet <comando> --help
./crownet <comando> <subcomando> --help
```

## 2. Flags Globais Persistentes

Estas flags podem ser aplicadas ao comando raiz (`crownet`) e são herdadas por seus subcomandos:

*   `--configFile <string>`: Caminho para um arquivo de configuração TOML. Se especificado, os valores deste arquivo são carregados e podem ser sobrescritos por flags de comando. (Padrão: "", nenhum arquivo carregado por padrão) Veja `config.example.toml` para a estrutura.
*   `--seed <int64>`: Semente para o gerador de números aleatórios (0 usa o tempo atual). (Padrão: 0)

## 3. Comandos Principais e Suas Flags

A funcionalidade que antes era controlada pela flag `-mode` agora é acessada através de subcomandos diretos de `crownet`.

### 3.1. Comando `sim`

Executa uma simulação geral da rede.
**Uso:** `./crownet sim [flags]`

**Flags para `sim`:**
*   `-n, --neurons <int>`: Número total de neurônios na rede. (Padrão: 200)
*   `-w, --weightsFile <string>`: Caminho para o arquivo JSON para salvar/carregar os pesos sinápticos. (Padrão: "crownet_weights.json")
*   `--dbPath <string>`: Caminho para o arquivo SQLite para logging. Se o arquivo não existir, será criado. Se já existir, será aberto. (Padrão: "crownet_sim_run.db")
*   `--saveInterval <int>`: Intervalo de ciclos para salvar no SQLite (0 desabilita saves periódicos). (Padrão: 100)
*   `--lrBase <float64>`: Taxa de aprendizado base. (Padrão: 0.01)
*   `-c, --cycles <int>`: Total de ciclos de simulação. (Padrão: 1000)
*   `--stimInputID <int>`: ID do neurônio de entrada para estímulo contínuo (-1: primeiro, -2: desabilitado). (Padrão: -1)
*   `--stimInputFreqHz <float64>`: Frequência (Hz) para estímulo contínuo (0.0 desabilita). (Padrão: 0.0)
*   `--monitorOutputID <int>`: ID do neurônio de saída para monitorar frequência (-1: primeiro, -2: desabilitado). (Padrão: -1)
*   `--debugChem <bool>`: Habilita logs de depuração para neuroquímicos. (Padrão: false)

### 3.2. Comando `expose`

Treina a rede expondo-a a padrões de dígitos.
**Uso:** `./crownet expose [flags]`

**Flags para `expose`:**
*   `-n, --neurons <int>`: Total de neurônios. (Padrão: 200)
*   `-w, --weightsFile <string>`: Arquivo para salvar/carregar pesos. **Obrigatório para salvar após o treino.** (Padrão: "crownet_weights.json")
*   `--lrBase <float64>`: Taxa de aprendizado base. (Padrão: 0.01)
*   `-e, --epochs <int>`: Número de épocas de exposição. (Padrão: 50)
*   `--cyclesPerPattern <int>`: Ciclos por apresentação de padrão. (Padrão: 20)
*   `--dbPath <string>`: (Opcional) Caminho para SQLite para logging durante o treino.
*   `--saveInterval <int>`: (Opcional) Intervalo de ciclos para salvar no BD durante o treino.
*   `--debugChem <bool>`: Habilita logs de depuração para neuroquímicos. (Padrão: false)

### 3.3. Comando `observe`

Observa a resposta da rede a um dígito específico.
**Uso:** `./crownet observe [flags]`

**Flags para `observe`:**
*   `-n, --neurons <int>`: Total de neurônios (deve corresponder à rede dos pesos). (Padrão: 200)
*   `-w, --weightsFile <string>`: Arquivo para carregar pesos. **Obrigatório.** (Padrão: "crownet_weights.json")
*   `-d, --digit <0-9>`: O dígito a ser apresentado. (Padrão: 0)
*   `--cyclesToSettle <int>`: Número de ciclos para acomodação da rede. (Padrão: 50)
*   `--debugChem <bool>`: Habilita logs de depuração para neuroquímicos. (Padrão: false)

### 3.4. Comando `logutil`

Utilitários para interagir com logs SQLite.
**Uso:** `./crownet logutil <subcomando> [flags]`

#### 3.4.1. Subcomando `logutil export`
Exporta dados de tabelas do log para CSV.
**Uso:** `./crownet logutil export [flags]`

**Flags para `logutil export`:**
*   `-d, --dbPath <string>`: Caminho para o arquivo SQLite DB. **Obrigatório.**
*   `-t, --table <string>`: Tabela a ser exportada ('NetworkSnapshots' ou 'NeuronStates'). **Obrigatório.**
*   `-f, --format <string>`: Formato de saída (atualmente apenas 'csv'). (Padrão: "csv")
*   `-o, --output <string>`: Arquivo de saída (stdout se não especificado).

## 4. Arquivo de Configuração TOML (Opcional)

A aplicação pode ser configurada usando um arquivo TOML (especificado pela flag global `--configFile`). Consulte o arquivo `config.example.toml` na raiz do repositório para um exemplo detalhado da estrutura e dos campos disponíveis.

**Ordem de Precedência da Configuração:**
1.  **Valores Padrão Internos:** Definidos no código da aplicação.
2.  **Arquivo de Configuração TOML:** Se `-configFile` for fornecido e o arquivo for válido, seus valores sobrescrevem os padrões.
3.  **Flags da Linha de Comando (CLI):** Quaisquer flags fornecidas na execução sobrescrevem os valores do arquivo de configuração e os padrões.

**Estrutura do Arquivo TOML:**
O arquivo TOML pode conter duas seções principais: `[cli]` e `[sim_params]`.

*   **Seção `[cli]`:** Contém parâmetros que também podem ser definidos por flags CLI.
    ```toml
    [cli]
    mode = "sim"
    total_neurons = 250
    seed = 12345
    weights_file = "custom_weights.json"
    base_learning_rate = 0.015
    # ... outras flags CLI como cycles, db_path, etc.
    ```

*   **Seção `[sim_params]`:** Contém parâmetros detalhados da simulação (correspondentes à estrutura `SimulationParameters` no código).
    ```toml
    [sim_params]
    space_max_dimension = 12.0
    base_firing_threshold = 1.0
    pulse_propagation_speed = 1.0
    accumulated_pulse_decay_rate = 0.1
    # ... todos os outros SimulationParameters
    ```

**Exemplo de Uso:**
```bash
./crownet -configFile my_config.toml -mode observe -digit 5
```
Neste exemplo:
1.  A aplicação carrega os padrões.
2.  Lê `my_config.toml`. Se `my_config.toml` definir `mode = "sim"`, ele será sobrescrito.
3.  As flags CLI `-mode observe` e `-digit 5` sobrescrevem quaisquer valores para `mode` e `digit` definidos no arquivo `my_config.toml` ou nos padrões.

Se o arquivo `config.toml` não for encontrado ou for inválido (e a flag `-configFile` for especificada), um erro será emitido. Se `-configFile` não for usada, a aplicação procede com padrões e flags CLI.

## 5. Formato da Saída no Console por Modo

### 4.1. Saída de Inicialização Comum (Todos os Modos)

```text
CrowNet Initializing...
Selected Mode: <modo_selecionado>
Base Configuration: Neurons=<N_neurons>, WeightsFile='<arquivo_pesos>'
  // Linha específica do modo, exemplos:
  expose: Epochs=<N_epocas>, BaseLR=<taxa_aprendizado>, CyclesPerPattern=<ciclos_padrao>
  observe: Digit=<digito>, CyclesToSettle=<ciclos_acomodacao>
  sim: TotalCycles=<N_ciclos>, DBPath='<arquivo_db>', SaveInterval=<intervalo_save>
  sim: GeneralStimulus: InputID=<ID_input> at <frequencia> Hz (se estímulo geral estiver ativo no modo sim)
Network created: <N_total_neurons> neurons. Input IDs: [<id1>, <id2>, ...preview]..., Output IDs: [<id_out1>, <id_out2>, ...preview]...
Initial State: Cortisol=<nivel_C>, Dopamine=<nivel_D>
```

### 4.2. Modo `sim` (`-mode sim`)

*   **Flags Específicas:**
    *   `-cycles <int>`: Número total de ciclos de simulação para este modo. (Padrão: 1000)
    *   `-stimInputID <int>`: ID de um neurônio de entrada para estímulo contínuo (-1 para primeiro disponível, -2 para desabilitar). (Padrão: -1)
    *   `-stimInputFreqHz <float64>`: Frequência (Hz) para o estímulo contínuo (0.0 para desabilitar). (Padrão: 0.0)
    *   `-monitorOutputID <int>`: ID de um neurônio de saída para monitorar frequência (-1 para primeiro disponível, -2 para desabilitar). (Padrão: -1)

*   **Saída no Console:**
    ```text
    Running General Simulation for <N_ciclos> cycles...
    General stimulus: Input Neuron <ID_input> at <frequencia> Hz. (se aplicável)

    // Progresso a cada 10 ciclos:
    Cycle <ciclo_atual>/<N_ciclos>: C:<cortisol_lvl> D:<dopamine_lvl> SynModF:<syn_factor> Pulses:<N_pulsos_ativos>

    // Avisos de salvamento (se dbPath configurado):
    Network state for cycle <ciclo_save> saved to database (SnapshotID: <id_snapshot>).
    Warning during periodic save: <erro> (se ocorrer)
    Warning during final save: <erro> (se ocorrer)

    Frequency for Output Neuron <ID_output>: <freq_out> Hz (over last <janela_ciclos> cycles). (se monitorOutputID válido)
    Final State: Cortisol=<nivel_C_final>, Dopamine=<nivel_D_final>

    CrowNet session finished.
    ```

### 4.3. Modo `expose` (`-mode expose`)

*   **Flags Específicas:**
    *   `-epochs <int>`: Número de épocas de exposição aos padrões. (Padrão: 50)
    *   `-lrBase <float64>`: Taxa de aprendizado base para plasticidade Hebbiana. (Padrão: 0.01)
    *   `-cyclesPerPattern <int>`: Número de ciclos de simulação por apresentação de padrão. (Padrão: 20)

*   **Saída no Console:**
    ```text
    Starting Exposure Phase for <N_epocas> epochs (BaseLR: <taxa_lr>, CyclesPerPattern: <ciclos_padrao>)...
    [SETUP-EXPOSE] Attempting to set up dopamine stimulation... (mensagens sobre setup de estímulo dopaminérgico)
    Loaded existing weights from <arquivo_pesos> (se carregado com sucesso)
    Could not load weights from <arquivo_pesos> (<erro>). Starting with initial random weights. (se falhar ao carregar)

    Epoch <epoca_atual>/<N_epocas> starting...
    Epoch <epoca_atual>/<N_epocas> completed. Processed <N_padroes_epoca> patterns. Cortisol:<C_lvl> Dopamine:<D_lvl> Eff. LR Factor (example): <fator_lr_efetivo>

    Exposure phase completed.
    Saved trained weights to <arquivo_pesos>
    Failed to save weights to <arquivo_pesos>: <erro> (se ocorrer)

    CrowNet session finished.
    ```

### 4.4. Modo `observe` (`-mode observe`)

*   **Flags Específicas:**
    *   `-digit <0-9>`: O dígito a ser apresentado. (Padrão: 0)
    *   `-cyclesToSettle <int>`: Número de ciclos para acomodação da rede. (Padrão: 50)

*   **Saída no Console:**
    ```text
    Observing Network Response for digit <digito_obs> (<ciclos_acomodacao> settle cycles)...
    Loaded weights from <arquivo_pesos> for observation.
    Failed to load weights from <arquivo_pesos> for observation: <erro>. Expose the network first. (se falhar ao carregar)

    Digit Presented: <digito_obs>
    Output Neuron Activation Pattern (Accumulated Potential):
      OutputNeuron[ 0] (ID NNNN) | [BARRA_ASCII_0       ] | VALOR_0
      OutputNeuron[ 1] (ID MMMM) | [BARRA_ASCII_1       ] | VALOR_1
      ...
      OutputNeuron[ 9] (ID KKKK) | [BARRA_ASCII_9       ] | VALOR_9

    CrowNet session finished.
    ```
    Onde:
    *   `OutputNeuron[idx] (ID NNNN)`: Identifica o neurônio de saída (índice na lista de saída e seu ID global).
    *   `[BARRA_ASCII_X       ]`: É uma representação visual da ativação do neurônio. O comprimento da barra (preenchida com `|`) é proporcional à ativação do neurônio em relação aos outros neurônios de saída. Uma barra mais longa indica maior ativação relativa. O comprimento total da barra entre colchetes é fixo (ex: 20 caracteres).
    *   `VALOR_X`: É o valor numérico bruto do potencial acumulado do neurônio (ex: `0.5234`).

### 4.5. Mensagens de Erro Comuns

*   Erros fatais (que encerram a aplicação) são prefixados com a data/hora e geralmente usam `log.Fatalf`, resultando em:
    `YYYY/MM/DD HH:MM:SS Failed to <operação>: <detalhes_do_erro>`
*   Avisos ou erros não fatais são impressos diretamente no console, podendo ser prefixados com "Warning:".

## 5. Estrutura do Arquivo JSON de Pesos (`-weightsFile`)

O arquivo JSON armazena os pesos sinápticos como um objeto principal. Cada chave deste objeto é uma string representando o `ID` de um neurônio de origem. O valor associado a cada neurônio de origem é outro objeto, onde cada chave é uma string representando o `ID` de um neurônio de destino, e o valor é o peso sináptico (um número float).

Exemplo (`crownet_weights.json`):
```json
{
  "0": {
    "1": 0.751234,
    "2": -0.320012,
    "99": 0.054321
  },
  "1": {
    "0": 0.680001,
    "2": 0.987654
  }
  // ... mais neurônios de origem e seus respectivos pesos para neurônios de destino
}
```

## 6. Estrutura do Banco de Dados SQLite (`-dbPath`)

Se o logging para SQLite estiver ativado, duas tabelas principais são criadas:

*   **`NetworkSnapshots`**: Registra o estado global da rede em um ciclo específico.
    *   `SnapshotID` (INTEGER, PK, AI)
    *   `CycleCount` (INTEGER)
    *   `Timestamp` (DATETIME)
    *   `CortisolLevel` (REAL)
    *   `DopamineLevel` (REAL)

*   **`NeuronStates`**: Registra o estado detalhado de cada neurônio para um dado `SnapshotID`.
    *   `StateID` (INTEGER, PK, AI)
    *   `SnapshotID` (INTEGER, FK para `NetworkSnapshots.SnapshotID`)
    *   `NeuronID` (INTEGER)
    *   `Position` (TEXT): Coordenadas do neurônio (armazenadas como uma string JSON array).
    *   `Velocity` (TEXT): Componentes de velocidade do neurônio (armazenados como uma string JSON array).
    *   `Type` (INTEGER): Tipo numérico do neurônio.
    *   `CurrentState` (INTEGER): Estado numérico do neurônio (nome da coluna atualizado de `State` para `CurrentState` para corresponder ao código).
    *   `AccumulatedPotential` (REAL): (Nome da coluna atualizado de `AccumulatedPulse` para `AccumulatedPotential`).
    *   `BaseFiringThreshold` (REAL)
    *   `CurrentFiringThreshold` (REAL)
    *   `LastFiredCycle` (INTEGER)
    *   `CyclesInCurrentState` (INTEGER)

(Nota: Se o arquivo de banco de dados especificado por `-dbPath` não existir, ele será criado. Se já existir, será aberto.)
```
