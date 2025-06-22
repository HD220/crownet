# Casos de Uso do Sistema CrowNet

Este documento detalha os principais casos de uso que o sistema de simulação de rede neural CrowNet suporta.

## 1. CU1: Treinamento da Rede para Reconhecimento de Padrões (Dígitos)

*   **Ator Principal:** Pesquisador / Usuário da CLI
*   **Descrição/Objetivo:** Treinar a rede neural para associar padrões de entrada específicos (atualmente, representações visuais de dígitos 0-9) com distintas representações internas, modificando os pesos sinápticos através de um processo de aprendizado Hebbiano. O objetivo é que a rede aprenda a "reconhecer" esses padrões.
*   **Pré-condições:**
    1.  Configuração da simulação (`config.SimulationParameters`) definida, incluindo taxa de aprendizado base, número de neurônios, dimensões dos padrões, etc.
    2.  Padrões de treinamento (dígitos) disponíveis (atualmente hardcoded em `datagen/digits.go`).
    3.  (Opcional) Um arquivo de pesos preexistente pode ser carregado para continuar o treinamento. Se não fornecido, a rede começa com pesos inicializados aleatoriamente.
*   **Fluxo Principal de Eventos:**
    1.  O usuário executa o CrowNet com o modo `expose`.
    2.  Flags da CLI especificam:
        *   `-mode expose`
        *   `-epochs <numero_de_epocas>`: Quantas vezes o conjunto completo de padrões será apresentado.
        *   `-cyclesPerPattern <ciclos>`: Quantos ciclos de simulação a rede roda para cada padrão apresentado.
        *   `-lrBase <taxa>`: A taxa de aprendizado base.
        *   `-weightsFile <caminho_arquivo>`: Arquivo para salvar os pesos treinados (e opcionalmente carregar pesos iniciais).
        *   `-neurons <numero>`: Número total de neurônios na rede.
        *   (Opcional) `-dbPath` e `-saveInterval` para logging do estado da rede durante o treinamento.
    3.  O sistema inicializa a rede (ou carrega pesos do `-weightsFile`).
    4.  Para cada época:
        a.  Para cada padrão de dígito (0-9):
            i.  O estado da rede (potenciais, pulsos) é resetado.
            ii. O padrão do dígito é apresentado aos neurônios de input da rede.
            iii. A rede executa `cyclesPerPattern` ciclos de simulação. Durante esses ciclos:
                *   Pulsos se propagam.
                *   Neurônios integram potenciais e disparam.
                *   Níveis neuroquímicos são atualizados.
                *   Efeitos neuroquímicos (na taxa de aprendizado, sinaptogênese, limiares) são aplicados.
                *   Aprendizado Hebbiano modifica os pesos sinápticos com base na atividade correlacionada.
                *   Sinaptogênese (movimento de neurônios) ocorre.
            iv. (Opcional) Se o logging SQLite estiver ativo, o estado da rede é salvo no intervalo especificado.
    5.  Ao final de todas as épocas, os pesos sinápticos finais da rede são salvos no arquivo especificado por `-weightsFile`.
*   **Pós-condições/Resultado Esperado:**
    1.  O arquivo especificado por `-weightsFile` contém os pesos sinápticos da rede, modificados pelo processo de aprendizado.
    2.  Se o logging SQLite estiver ativo, o banco de dados contém um registro da evolução do estado da rede.
    3.  O console exibe logs indicando o progresso das épocas e o estado final dos neuromoduladores.
*   **Fluxos Alternativos/Exceções:**
    *   Se `-weightsFile` for especificado para carregar pesos e o arquivo não existir, o sistema loga um aviso e continua com pesos inicializados aleatoriamente.
    *   Se houver erro ao salvar o arquivo de pesos final, o sistema termina com erro (ou retorna erro para o chamador `Run()`).
    *   Se os parâmetros da CLI forem inválidos (ex: épocas <= 0), a validação de configuração impede a execução.
*   **Exemplo de Execução CLI (Simplificado):**
    `./crownet -mode expose -neurons 70 -epochs 10 -cyclesPerPattern 5 -lrBase 0.01 -weightsFile trained_digits.json`
*   **Resultados Chave Esperados:**
    *   Criação/atualização do arquivo `trained_digits.json`.
    *   Logs no console indicando o progresso de cada época.

## 2. CU2: Observação/Inferência da Rede em um Padrão Conhecido (Pós-Treinamento)

*   **Ator Principal:** Pesquisador / Usuário da CLI
*   **Descrição/Objetivo:** Apresentar um padrão específico (um dígito) a uma rede neural previamente treinada e observar o padrão de ativação de seus neurônios de saída. Isso permite avaliar quão bem a rede "reconhece" ou "classifica" o padrão de entrada.
*   **Pré-condições:**
    1.  Um arquivo de pesos sinápticos treinado (`-weightsFile`) deve existir (geralmente resultado do CU1).
    2.  Configuração da simulação definida, consistente com a rede cujos pesos foram salvos.
*   **Fluxo Principal de Eventos:**
    1.  O usuário executa o CrowNet com o modo `observe`.
    2.  Flags da CLI especificam:
        *   `-mode observe`
        *   `-digit <numero_digito>`: O dígito (0-9) a ser apresentado.
        *   `-weightsFile <caminho_arquivo>`: Arquivo contendo os pesos treinados.
        *   `-cyclesToSettle <ciclos>`: Número de ciclos para a rede "acomodar" o padrão.
        *   `-neurons <numero>`: (Deve corresponder à rede dos pesos).
    3.  O sistema carrega os pesos do `-weightsFile`. Se falhar, termina com erro.
    4.  Dinâmicas de aprendizado/sinaptogênese são desabilitadas.
    5.  O padrão do dígito é apresentado à rede.
    6.  A rede executa `cyclesToSettle` ciclos.
    7.  O sistema imprime no console a ativação dos neurônios de saída.
*   **Pós-condições/Resultado Esperado:**
    1.  Saída no console mostrando "Dígito Apresentado: X" e uma lista de "OutNeurônio[i] (ID Y): Z.ZZZZ", onde Z.ZZZZ é o potencial acumulado.
    2.  Para uma rede bem treinada, o neurônio de saída associado ao dígito X deve ter a maior ativação.
*   **Fluxos Alternativos/Exceções:**
    *   Falha ao carregar `-weightsFile` resulta em erro.
    *   Dígito inválido ou erro ao obter padrão resulta em erro.
*   **Exemplo de Execução CLI:**
    `./crownet -mode observe -digit 7 -weightsFile trained_digits.json -neurons 70 -cyclesToSettle 10`
*   **Resultados Chave Esperados:**
    *   Linhas no console como: `OutNeurônio[7] (ID N): PotencialAlto` e outros neurônios com potencial mais baixo.

## 3. CU3: Simulação da Dinâmica da Rede sob Estímulo Contínuo e Logging de Estado

*   **Ator Principal:** Pesquisador / Usuário da CLI
*   **Descrição/Objetivo:** Executar uma simulação de forma livre, opcionalmente com estímulo contínuo, e registrar o estado da rede para análise.
*   **Pré-condições:** Configuração da simulação.
*   **Fluxo Principal de Eventos:**
    1.  Usuário executa com `-mode sim`.
    2.  Flags: `-cycles`, `-neurons`, opcionalmente `-stimInputID`, `-stimInputFreqHz`, `-dbPath`, `-saveInterval`.
    3.  Sistema inicializa a rede (opcionalmente carrega pesos se `-weightsFile` fornecido).
    4.  Aplica estímulo se configurado.
    5.  Executa simulação por N ciclos, aplicando todas as dinâmicas.
    6.  Logs periódicos no console.
    7.  Salva estado no DB em intervalos (se configurado).
    8.  Reporta frequência de output (se configurado) e estado final dos químicos.
*   **Pós-condições/Resultado Esperado:**
    1.  Logs no console.
    2.  Arquivo de banco de dados SQLite (`.db`) criado e populado se `-dbPath` e `-saveInterval` (ou ciclos > 0) forem usados.
*   **Fluxos Alternativos/Exceções:**
    *   ID de estímulo inválido resulta em erro (ou aviso e continuação sem o estímulo).
    *   Erro ao inicializar logger é fatal.
*   **Exemplo de Execução CLI:**
    `./crownet -mode sim -cycles 500 -neurons 100 -stimInputID 0 -stimInputFreqHz 20 -dbPath simulation_log.db -saveInterval 100`
*   **Resultados Chave Esperados:**
    *   Criação do arquivo `simulation_log.db`.
    *   Tabelas `NetworkSnapshots` e `NeuronStates` no DB com dados dos ciclos 100, 200, 300, 400, 500.
    *   Log no console indicando "Estímulo geral: Neurônio de Input 0 a 20.0 Hz."

## 4. CU4: Investigação do Efeito da Neuromodulação no Aprendizado

*   **Ator Principal:** Pesquisador
*   **Descrição/Objetivo:** Comparar a eficácia do aprendizado (CU1) sob diferentes condições neuroquímicas.
*   **Pré-condições:** Entendimento de como os parâmetros em `config.SimulationParameters` (ex: `DopamineInfluenceOnLR`, `CortisolInfluenceOnLR`) afetam a taxa de aprendizado efetiva.
*   **Fluxo Principal de Eventos:**
    1.  **Sessão A (Baseline):** Executar CU1 com `SimulationParameters` padrão. Salvar pesos (`weights_A.json`).
    2.  **Sessão B (Dopamina Alta):** Modificar `config.DefaultSimulationParameters()` (ou usar um arquivo de config se suportado) para aumentar `DopamineInfluenceOnLR` (ex: para um valor positivo alto) e/ou diminuir `CortisolInfluenceOnLR` (ex: para um valor menos negativo). Executar CU1 com os mesmos parâmetros de treino. Salvar pesos (`weights_B.json`).
    3.  **Sessão C (Cortisol Alto):** Modificar `SimulationParameters` para diminuir `DopamineInfluenceOnLR` e/ou aumentar o efeito supressor do `CortisolInfluenceOnLR`. Executar CU1. Salvar pesos (`weights_C.json`).
    4.  **Comparação:**
        a.  Analisar as diferenças nos arquivos de pesos.
        b.  Executar CU2 (modo `observe`) com cada conjunto de pesos em um conjunto de teste de dígitos e comparar a precisão/padrões de ativação.
*   **Pós-condições/Resultado Esperado:**
    1.  Diferentes arquivos de pesos.
    2.  Diferenças observáveis no desempenho de reconhecimento, correlacionadas com as modulações. Espera-se que "Dopamina Alta" melhore o aprendizado e "Cortisol Alto" o prejudique, com base nos fatores de influência.
*   **Exemplo de Execução CLI (Conceitual):**
    *   `./crownet -mode expose ... -weightsFile weights_A.json` (com SimParams padrão)
    *   *Modificar SimParams para Dopamina Alta (requer alteração no código ou config externa)*
    *   `./crownet -mode expose ... -weightsFile weights_B.json`
    *   `./crownet -mode observe -weightsFile weights_A.json ...`
    *   `./crownet -mode observe -weightsFile weights_B.json ...`
*   **Resultados Chave Esperados:**
    *   Taxa de aprendizado efetiva (visível nos logs do modo `expose` como "FatorLR Efetivo") deve ser maior na Sessão B e menor na Sessão C em comparação com A.
    *   Rede da Sessão B pode apresentar melhor performance no modo `observe`.

## 5. CU5: Investigação do Efeito da Neuromodulação no Limiar de Disparo

*   **Ator Principal:** Pesquisador
*   **Descrição/Objetivo:** Observar como os níveis de neuroquímicos alteram o `CurrentFiringThreshold` dos neurônios.
*   **Pré-condições:** Uma rede inicializada.
*   **Fluxo Principal de Eventos (Mais adequado para teste unitário/script dedicado):**
    1.  No código (ex: um teste ou script):
        a.  Criar `neurochemical.Environment` e `config.SimulationParameters`.
        b.  Ajustar `simParams.FiringThresholdIncreaseOnDopa` e `simParams.FiringThresholdIncreaseOnCort`.
        c.  Criar neurônios de teste.
        d.  Definir `env.DopamineLevel` e `env.CortisolLevel` para vários valores (zero, médio, alto).
        e.  Chamar `env.ApplyEffectsToNeurons()`.
        f.  Inspecionar `n.CurrentFiringThreshold` para cada neurônio.
*   **Pós-condições/Resultado Esperado:**
    1.  `CurrentFiringThreshold` varia conforme esperado:
        *   Aumenta com cortisol (se `FiringThresholdIncreaseOnCort` > 0).
        *   Diminui com dopamina (se `FiringThresholdIncreaseOnDopa` < 0).
        *   Efeito combinado é multiplicativo.
*   **Exemplo de Execução CLI (Não aplicável diretamente, mas para contexto):**
    *   Se o modo `sim` logasse `CurrentFiringThreshold` no DB, poderia ser observado lá após configurar diferentes taxas de produção/decaimento para cortisol/dopamina.
*   **Resultados Chave Esperados (de testes unitários/scripts):**
    *   Verificação de que os limiares são ajustados corretamente conforme a fórmula: `Base * (1 + CortisolFactor * NormCortisol) * (1 + DopaFactor * NormDopamine)`.
*   **Notas:** Este caso de uso é primariamente validado pelos testes unitários em `neurochemical_test.go` (`TestApplyEffectsToNeurons`).
```
