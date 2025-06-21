# Funcionalidade: Modos de Operação da CLI (MVP)

Esta funcionalidade descreve os diferentes modos de operação da aplicação de linha de comando (CLI) CrowNet no MVP, controlados principalmente pelo flag `-mode`.

## F5.1: Aplicação de Linha de Comando
*   A interação com o CrowNet MVP é primariamente através de uma interface de linha de comando (`main.go`).

## F5.2: Modos Principais e Flags Comuns

### Flags Globais/Comuns (podem aplicar-se a múltiplos modos):
*   `-neurons <int>`: Número total de neurônios na rede.
*   `-cycles <int>`: Número total de ciclos para uma simulação ou para processar um padrão (dependendo do modo).
*   `-weightsFile <string>`: Caminho para o arquivo JSON para salvar ou carregar os pesos sinápticos.
*   `-dbPath <string>`: Caminho para o arquivo SQLite para logging detalhado da simulação (opcional).
*   `-saveInterval <int>`: Intervalo de ciclos para salvar o estado no SQLite (se `dbPath` for fornecido).
*   `-debugChem <bool>`: Habilita/desabilita logs de depuração para neuroquímicos.

### 1. Modo `expose` (`-mode expose`)
*   **Propósito:** Expor a rede a padrões de dígitos repetidamente para permitir a auto-organização dos pesos sinápticos através da plasticidade Hebbiana neuromodulada. Este é o modo de "treinamento" do MVP.
*   **Dinâmicas Ativas:** Todas as dinâmicas estão ativas:
    *   Plasticidade Hebbiana (atualização de pesos).
    *   Modulação Química (produção, decaimento e efeitos de cortisol/dopamina).
    *   Sinaptogênese (movimento de neurônios).
    *   Nota: Para auxiliar na modulação química durante a exposição, uma função (`setupDopamineStimulationForExpose`) tenta configurar um estímulo adicional para neurônios dopaminérgicos, se presentes.
*   **Flags Específicas:**
    *   `-epochs <int>`: Número de vezes que o conjunto completo de padrões de dígitos (0-9) será apresentado à rede.
    *   `-lrBase <float>`: Taxa de aprendizado base para a regra Hebbiana.
    *   `-cyclesPerPattern <int>`: Número de ciclos que a rede executa para cada padrão de dígito apresentado dentro de uma época. O flag `-cycles` global não é usado diretamente aqui.
*   **Processo:**
    1.  Carrega pesos de `-weightsFile` se existir e for válido, senão inicializa a rede com pesos aleatórios.
    2.  Itera por `-epochs`.
    3.  Em cada época, para cada dígito (0-9):
        a.  Reseta as ativações dos neurônios e limpa pulsos ativos.
        b.  Apresenta o padrão do dígito aos neurônios de entrada.
        c.  Executa a simulação (chama `RunCycle`) por `-cyclesPerPattern` ciclos. Durante estes ciclos, ocorrem aprendizado, mudanças químicas e sinaptogênese.
    4.  Salva os pesos finais em `-weightsFile`.

### 2. Modo `observe` (`-mode observe`)
*   **Propósito:** Carregar um conjunto de pesos previamente aprendidos e apresentar um dígito específico à rede para observar o padrão de ativação resultante nos neurônios de saída.
*   **Dinâmicas Ativas (Temporariamente Modificadas para Observação):**
    *   Para garantir uma observação "limpa" do padrão de ativação resultante dos pesos aprendidos (sem alterá-los ou introduzir variabilidade excessiva):
        *   **Plasticidade Hebbiana:** Desativada (a taxa de aprendizado efetiva se torna zero).
        *   **Sinaptogênese:** Desativada (`EnableSynaptogenesis` é temporariamente `false`).
        *   **Modulação Química:** Desativada (`EnableChemicalModulation` é temporariamente `false`). Isso significa que os limiares dos neurônios permanecem em seus valores base e não há produção/efeito de químicos durante os ciclos de acomodação.
*   **Flags Específicas:**
    *   `-digit <0-9>`: O dígito específico (0 a 9) a ser apresentado à rede.
    *   `-cyclesToSettle <int>`: Número de ciclos para permitir que a atividade se propague e se estabilize após a apresentação do dígito, antes de ler os neurônios de saída. O flag `-cycles` global não é usado diretamente aqui.
*   **Processo:**
    1.  Carrega os pesos sinápticos de `-weightsFile`. Falha se o arquivo não puder ser carregado.
    2.  Apresenta o padrão do `-digit` especificado.
    3.  Executa a simulação por `-cyclesToSettle` ciclos.
    4.  Exibe o padrão de ativação dos 10 neurônios de saída.

### 3. Modo `sim` (`-mode sim`)
*   **Propósito:** Executar uma simulação geral com todas as dinâmicas originais do CrowNet ativas, permitindo a observação de comportamentos emergentes da rede sob estímulo contínuo ou para análise detalhada. Não é focado especificamente na tarefa de aprendizado de dígitos, mas utiliza os mesmos mecanismos subjacentes.
*   **Dinâmicas Ativas:** Todas as dinâmicas estão ativas (Plasticidade Hebbiana, Modulação Química, Sinaptogênese).
*   **Flags Específicas:**
    *   `-stimInputID <int>`: ID de um neurônio de entrada para receber estímulo contínuo.
    *   `-stimInputFreqHz <float>`: Frequência (em Hz, convertida para probabilidade por ciclo) na qual o `-stimInputID` irá disparar.
    *   `-monitorOutputID <int>`: ID de um neurônio de saída cuja frequência de disparo será monitorada e reportada.
*   **Processo:**
    1.  Carrega pesos de `-weightsFile` se especificado, senão inicializa.
    2.  Executa a simulação por `-cycles` (global).
    3.  Se configurado, aplica estímulo contínuo ao `-stimInputID`.
    4.  Loga o estado da rede no SQLite se `-dbPath` for fornecido.
    5.  Reporta estatísticas, como a frequência do `-monitorOutputID` e níveis químicos.
