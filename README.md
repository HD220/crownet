# CrowNet: Simulador de Rede Neural Bio-inspirada (MVP)

**CrowNet** é uma aplicação de linha de comando escrita em Go que simula um modelo computacional de rede neural. Inspirada em processos biológicos, a simulação apresenta neurônios interagindo em um espaço vetorial de 16 dimensões, sinaptogênese (movimento de neurônios dependente da atividade) e neuromodulação por cortisol e dopamina simulados.

O **Minimum Viable Product (MVP)** atual foca em demonstrar a **autoaprendizagem Hebbiana neuromodulada**. A rede é exposta a padrões simples de dígitos (0-9) e visa auto-organizar seus pesos sinápticos para formar representações internas distintas para esses diferentes inputs. O processo de aprendizado (plasticidade) é influenciado pelo ambiente químico simulado.

## Visão Geral do MVP

*   **Objetivo Principal:** Demonstrar que o modelo CrowNet pode exibir comportamento de auto-organização através de plasticidade Hebbiana neuromodulada, onde a rede aprende a formar representações internas distintas para diferentes padrões de entrada (dígitos 0-9).
*   **Interface:** Linha de Comando (CLI).
*   **Linguagem:** Go.

## Principais Conceitos Implementados no MVP

*   **Espaço 16D:** Neurônios existem e se movem em um espaço vetorial de 16 dimensões.
*   **Tipos de Neurônios:** Excitatórios, Inibitórios, Dopaminérgicos, Input e Output.
*   **Propagação de Pulso:** Modelo simplificado de expansão esférica.
*   **Pesos Sinápticos:** Conexões explícitas com pesos que determinam a força da influência entre neurônios.
*   **Aprendizado Hebbiano Neuromodulado:**
    *   **Plasticidade Hebbiana:** Pesos ajustados com base na co-ativação de neurônios.
    *   **Neuromodulação:** Taxa de aprendizado e limiares de disparo modulados por níveis de Dopamina (aumenta plasticidade) e Cortisol (altos níveis suprimem plasticidade).
*   **Sinaptogênese:** Movimento de neurônios influenciado pela atividade da rede e modulado por químicos.
*   **Codificação de Entrada:** Padrões binários 5x7 para dígitos 0-9.
*   **Representação de Saída:** Padrões de ativação distintos nos 10 neurônios de output.

## Comandos Principais (CLI)

A aplicação é controlada através de subcomandos. Os principais são:

1.  **`sim`**: Executa uma simulação geral da rede com todas as dinâmicas ativas.
    *   Exemplo: `./crownet sim --cycles 1000 --neurons 150`
    *   Use `./crownet sim --help` para todas as flags.
2.  **`expose`**: Treina a rede expondo-a a padrões de dígitos.
    *   Exemplo: `./crownet expose --epochs 50 --weightsFile pesos.json`
    *   Use `./crownet expose --help` para todas as flags.
3.  **`observe`**: Testa uma rede treinada com um dígito específico.
    *   Exemplo: `./crownet observe --digit 7 --weightsFile pesos.json`
    *   Use `./crownet observe --help` para todas as flags.
4.  **`logutil export`**: Exporta dados de logs SQLite para CSV.
    *   Exemplo: `./crownet logutil export --dbPath sim.db --table NetworkSnapshots`
    *   Use `./crownet logutil export --help` para todas as flags.

Consulte o [Guia de Interface de Linha de Comando](./docs/03_guias/guia_interface_linha_comando.md) para detalhes completos sobre todos os comandos e flags.

## Tecnologias Utilizadas (MVP)

*   **Go:** Linguagem de implementação.
*   **JSON:** Para salvar e carregar os pesos sinápticos aprendidos.
*   **SQLite:** (Opcional) Para salvar snapshots detalhados do estado da simulação para análise.

## Documentação Detalhada

Para uma compreensão completa das funcionalidades, arquitetura técnica, requisitos e casos de uso do CrowNet MVP, por favor, consulte os documentos localizados no diretório `/docs`:

*   **`docs/01_visao_geral.md`**: Visão geral do projeto.
*   **`docs/02_arquitetura.md`**: Detalhes da arquitetura de software, pacotes e algoritmos.
*   **`docs/requisitos.md`**: Requisitos Funcionais e Não Funcionais do MVP.
*   **`docs/03_guias/`**: Guias de configuração, estilo de código e interface de linha de comando.
    *   `guia_configuracao_ambiente.md`
    *   `guia_estilo_codigo.md`
    *   `guia_interface_linha_comando.md`
*   **`docs/04_funcionalidades/`**: Descrições detalhadas de cada funcionalidade do sistema e seus casos de uso.
    *   `01-inicializacao-rede.md`
    *   `02-ciclo-simulacao-aprendizado.md`
    *   `03-entrada-saida-dados.md`
    *   `04-modos-operacao.md`
    *   `05-persistencia-dados.md`
    *   `casos-de-uso/uc-expose.md`
    *   `casos-de-uso/uc-observe.md`
    *   `casos-de-uso/uc-sim.md`
*   **Outros documentos relevantes em `/docs`**: `TESTING_SCENARIOS.md`, `use_cases.md` (pode ser redundante com os UCs em funcionalidades), `refactoring_log.md`, `testing_approach.md`.

## Utilitário de Log (`logutil`) (FEATURE-004)

A aplicação CrowNet inclui um utilitário para interagir com os arquivos de log SQLite gerados pelo modo `sim`.

### Exportar Dados do Log

Para exportar dados de tabelas específicas do arquivo de log para o formato CSV, use o modo `logutil` com o subcomando `export`.

**Uso:**

```bash
./crownet -mode logutil -logutil.subcommand export -logutil.dbPath <caminho_para_seu_log.db> -logutil.table <nome_da_tabela> [-logutil.output <arquivo_de_saida.csv>] [-logutil.format csv]
```

**Argumentos:**

*   `-mode logutil`: Ativa o modo de utilitário de log.
*   `-logutil.subcommand export`: Especifica a ação de exportação. (Atualmente, único subcomando suportado).
*   `-logutil.dbPath <caminho_para_seu_log.db>`: **Obrigatório.** Caminho para o arquivo de banco de dados SQLite gerado pela simulação.
*   `-logutil.table <nome_da_tabela>`: **Obrigatório.** Nome da tabela a ser exportada. Tabelas suportadas:
    *   `NetworkSnapshots`: Contém informações gerais sobre o estado da rede em cada ciclo de salvamento (níveis de neuroquímicos, fatores de modulação, etc.).
    *   `NeuronStates`: Contém o estado detalhado de cada neurônio em cada snapshot salvo (posição, potencial, estado de disparo, etc.).
*   `-logutil.output <arquivo_de_saida.csv>`: (Opcional) Caminho para o arquivo CSV de saída. Se omitido, a saída CSV será impressa no `stdout` (saída padrão), permitindo redirecionamento (ex: `> meu_arquivo.csv`).
*   `-logutil.format csv`: (Opcional) Formato de saída. Atualmente, apenas `csv` é suportado e é o valor padrão.

**Exemplos:**

1.  Exportar a tabela `NetworkSnapshots` de `simulation.db` para `snapshots.csv`:
    ```bash
    ./crownet -mode logutil -logutil.subcommand export -logutil.dbPath simulation.db -logutil.table NetworkSnapshots -logutil.output snapshots.csv
    ```

2.  Exportar a tabela `NeuronStates` de `run1.db` para `stdout` e redirecionar para `neuron_data.csv`:
    ```bash
    ./crownet -mode logutil -logutil.subcommand export -logutil.dbPath run1.db -logutil.table NeuronStates > neuron_data.csv
    ```

**Notas sobre a Saída CSV:**

*   **`NeuronStates`**:
    *   Os campos `Type` e `CurrentState` (que são armazenados como inteiros no banco de dados) são convertidos para suas representações de string (ex: "Excitatory", "Firing") para melhor legibilidade.
    *   Os campos `Position` e `Velocity` são exportados como strings JSON, conforme armazenados no banco de dados.

## Como Construir e Executar (Exemplo)

1.  **Construir:**
    ```bash
    go build .
    ```
2.  **Executar (exemplo comando `expose`):**
    ```bash
    ./crownet expose --neurons 150 --epochs 20 --lrBase 0.005 --cyclesPerPattern 5 --weightsFile my_digit_weights.json
    ```
3.  **Executar (exemplo comando `observe`):**
    ```bash
    ./crownet observe --digit 7 --weightsFile my_digit_weights.json --cyclesToSettle 5
    ```
4.  **Executar (exemplo comando `sim` com seed):**
    ```bash
    ./crownet --seed 42 sim --cycles 500 --dbPath simulation_log.db
    ```

Consulte o `guia_interface_linha_comando.md` (ou use `--help` nos comandos) para mais detalhes sobre os flags.

## Reprodutibilidade

Para garantir que as simulações possam ser repetidas com os mesmos resultados (útil para depuração e análise comparativa), a aplicação suporta uma flag de semente aleatória:

*   `-seed <int64>`: Forneça um valor inteiro (longo) específico para a semente. Todas as operações estocásticas na simulação (posicionamento inicial de neurônios, inicialização de pesos, etc.) serão derivadas desta semente.
*   Se o flag `-seed` não for fornecido ou for explicitamente `-seed 0`, a simulação usará uma semente baseada no tempo atual, resultando em variabilidade entre as execuções.
```
