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
    *   Se houver erro ao salvar o arquivo de pesos final, o sistema termina com erro.
    *   Se os parâmetros da CLI forem inválidos (ex: épocas <= 0), a validação de configuração impede a execução.

## 2. CU2: Observação/Inferência da Rede em um Padrão Conhecido (Pós-Treinamento)

*   **Ator Principal:** Pesquisador / Usuário da CLI
*   **Descrição/Objetivo:** Apresentar um padrão específico (um dígito) a uma rede neural previamente treinada e observar o padrão de ativação de seus neurônios de saída. Isso permite avaliar quão bem a rede "reconhece" ou "classifica" o padrão de entrada.
*   **Pré-condições:**
    1.  Um arquivo de pesos sinápticos treinado (`-weightsFile`) deve existir (geralmente resultado do CU1).
    2.  Configuração da simulação definida.
*   **Fluxo Principal de Eventos:**
    1.  O usuário executa o CrowNet com o modo `observe`.
    2.  Flags da CLI especificam:
        *   `-mode observe`
        *   `-digit <numero_digito>`: O dígito (0-9) a ser apresentado.
        *   `-weightsFile <caminho_arquivo>`: Arquivo contendo os pesos treinados.
        *   `-cyclesToSettle <ciclos>`: Número de ciclos para a rede "acomodar" o padrão e estabilizar sua resposta.
        *   `-neurons <numero>`: Deve corresponder à configuração da rede cujos pesos foram salvos.
    3.  O sistema carrega os pesos sinápticos do `-weightsFile`. Se falhar, termina com erro.
    4.  As dinâmicas de aprendizado e sinaptogênese da rede são desabilitadas (`SetDynamicState(false, false, false)`). A modulação química pode ou não ser desabilitada, dependendo se o objetivo é ver a resposta "pura" ou modulada. Atualmente, é desabilitada.
    5.  O padrão do dígito especificado é obtido de `datagen`.
    6.  O estado da rede é resetado.
    7.  O padrão é apresentado aos neurônios de input.
    8.  A rede executa `cyclesToSettle` ciclos de simulação.
    9.  O potencial acumulado dos neurônios de saída é recuperado.
    10. O sistema imprime no console o dígito apresentado e o padrão de ativação (potencial acumulado) de cada neurônio de saída.
*   **Pós-condições/Resultado Esperado:**
    1.  A saída no console mostra a ativação dos neurônios de saída. Idealmente, para uma rede bem treinada, o neurônio de saída correspondente ao dígito apresentado terá a maior ativação.
*   **Fluxos Alternativos/Exceções:**
    *   Se `-weightsFile` não for encontrado ou for inválido, o sistema termina com erro.
    *   Se `-digit` for inválido, a validação de configuração impede a execução.
    *   Se houver erro ao obter o padrão do dígito ou ao apresentar o padrão à rede, o sistema termina com erro.

## 3. CU3: Simulação da Dinâmica da Rede sob Estímulo Contínuo e Logging de Estado

*   **Ator Principal:** Pesquisador / Usuário da CLI
*   **Descrição/Objetivo:** Executar uma simulação de forma livre da rede por um número especificado de ciclos, opcionalmente aplicando um estímulo contínuo a um neurônio de input específico, e registrar a evolução do estado da rede (neurônios, neuroquímicos) em um banco de dados para análise posterior.
*   **Pré-condições:**
    1.  Configuração da simulação definida.
*   **Fluxo Principal de Eventos:**
    1.  O usuário executa o CrowNet com o modo `sim`.
    2.  Flags da CLI especificam:
        *   `-mode sim`
        *   `-cycles <numero_de_ciclos>`
        *   `-neurons <numero>`
        *   (Opcional) `-stimInputID <id>` e `-stimInputFreqHz <freq>` para estímulo contínuo.
        *   (Opcional) `-dbPath <caminho_db>` e `-saveInterval <intervalo>` para logging SQLite.
        *   (Opcional) `-weightsFile <caminho_arquivo>` para carregar pesos iniciais (a rede não aprende ativamente neste modo por padrão, a menos que `isLearningEnabled` seja modificado programaticamente).
    3.  O sistema inicializa a rede (e opcionalmente carrega pesos).
    4.  Se o estímulo de input estiver configurado, ele é aplicado à rede.
    5.  Todas as dinâmicas da rede (aprendizado, sinaptogênese, modulação química) são ativadas por padrão para o modo `sim`.
    6.  A rede executa o número especificado de `-cycles`. Em cada ciclo:
        a.  Inputs de frequência são processados.
        b.  Estados dos neurônios são atualizados (potencial, disparo, refratário).
        c.  Pulsos ativos são processados.
        d.  Níveis neuroquímicos são atualizados e seus efeitos aplicados.
        e.  Aprendizado Hebbiano e sinaptogênese ocorrem.
        f.  (Opcional) Se o logging SQLite estiver ativo, o estado da rede é salvo no intervalo especificado.
        g.  O console exibe um log resumido do estado da rede periodicamente.
    7.  (Opcional) Se o logging SQLite estiver ativo e não houve salvamento no último ciclo devido ao intervalo, um salvamento final é feito.
    8.  (Opcional) Frequência de um neurônio de output monitorado é reportada.
    9.  O estado final dos neuroquímicos é impresso.
*   **Pós-condições/Resultado Esperado:**
    1.  Logs no console mostrando a progressão da simulação.
    2.  Se o logging SQLite estiver ativo, o arquivo de banco de dados especificado contém os snapshots do estado da rede.
*   **Fluxos Alternativos/Exceções:**
    *   Se `-stimInputID` for inválido, um erro é retornado/logado, e a simulação prossegue sem esse estímulo específico (ou termina, dependendo da implementação do tratamento de erro).
    *   Erro na inicialização do logger SQLite é fatal.

## 4. CU4: Investigação do Efeito da Neuromodulação no Aprendizado

*   **Ator Principal:** Pesquisador
*   **Descrição/Objetivo:** Comparar a eficácia ou velocidade do aprendizado (CU1) sob diferentes condições neuroquímicas simuladas.
*   **Pré-condições:**
    1.  Capacidade de influenciar os níveis de neuroquímicos ou seus efeitos durante o treinamento (modo `expose`). Atualmente, isso é feito através dos parâmetros em `config.SimulationParameters` que definem as taxas de produção/decaimento e os fatores de influência (ex: `DopamineInfluenceOnLR`, `CortisolInfluenceOnLR`).
*   **Fluxo Principal de Eventos:**
    1.  **Sessão de Treinamento A (Baseline):**
        a.  Executar CU1 com um conjunto de `SimulationParameters` que representem um estado neuroquímico "normal" ou de controle (ex: baixas taxas de produção, ou fatores de influência neutros/padrão). Salvar os pesos em `weights_baseline.json`.
    2.  **Sessão de Treinamento B (Modulada):**
        a.  Modificar os `SimulationParameters` relevantes (ex: aumentar `DopamineInfluenceOnLR` ou diminuir `CortisolInfluenceOnLR` (tornando-o menos negativo/supressor) para simular um estado pró-aprendizado; ou o oposto para um estado de aprendizado suprimido). Isso requer modificar `config.DefaultSimulationParameters()` ou ter uma forma de carregar `SimulationParameters` de um arquivo.
        b.  Executar CU1 com os mesmos parâmetros de época, ciclos, etc., da Sessão A, mas com os `SimulationParameters` modificados. Salvar os pesos em `weights_modulated.json`.
    3.  **Comparação:**
        a.  Comparar os arquivos `weights_baseline.json` e `weights_modulated.json` (ex: magnitude das mudanças, distribuição dos pesos).
        b.  Opcionalmente, executar CU2 (Modo `observe`) com ambos os conjuntos de pesos no mesmo conjunto de padrões de teste e comparar a precisão do reconhecimento.
*   **Pós-condições/Resultado Esperado:**
    1.  Dois (ou mais) conjuntos de arquivos de pesos representando diferentes condições neuroquímicas.
    2.  Observação de diferenças na estrutura dos pesos ou no desempenho da rede, correlacionadas com as modulações neuroquímicas simuladas.
*   **Notas:** Este caso de uso é mais conceitual e depende da capacidade do pesquisador de configurar e interpretar os `SimulationParameters`. A implementação atual da neuromodulação é simplificada; modelos mais complexos exigiriam mais parâmetros.

## 5. CU5: Investigação do Efeito da Neuromodulação no Limiar de Disparo

*   **Ator Principal:** Pesquisador
*   **Descrição/Objetivo:** Observar como os níveis de neuroquímicos (cortisol, dopamina) alteram o `CurrentFiringThreshold` dos neurônios, afetando sua excitabilidade.
*   **Pré-condições:**
    1.  Uma rede inicializada.
    2.  Capacidade de definir/influenciar os níveis de `CortisolLevel` e `DopamineLevel` no `neurochemical.Environment` e então chamar `ApplyEffectsToNeurons`.
*   **Fluxo Principal de Eventos (Simulado em Teste ou Código Específico):**
    1.  Criar uma instância de `neurochemical.Environment` e `config.SimulationParameters`.
    2.  Criar alguns neurônios de teste com `BaseFiringThreshold` conhecido.
    3.  **Cenário A (Controle):** Manter `env.CortisolLevel = 0`, `env.DopamineLevel = 0`. Chamar `env.ApplyEffectsToNeurons()`. Verificar se `CurrentFiringThreshold` é igual a `BaseFiringThreshold`.
    4.  **Cenário B (Alta Dopamina):** Definir `env.DopamineLevel` para um valor alto (ex: `simParams.DopamineMaxLevel`). Manter `env.CortisolLevel = 0`. Chamar `env.ApplyEffectsToNeurons()`. Verificar se `CurrentFiringThreshold` é modificado conforme esperado por `simParams.FiringThresholdIncreaseOnDopa` (ex: reduzido, se o fator for negativo).
    5.  **Cenário C (Alto Cortisol):** Definir `env.CortisolLevel` para um valor alto. Manter `env.DopamineLevel = 0`. Chamar `env.ApplyEffectsToNeurons()`. Verificar se `CurrentFiringThreshold` é modificado conforme esperado por `simParams.FiringThresholdIncreaseOnCort` (ex: aumentado, se o fator for positivo).
    6.  **Cenário D (Ambos Altos):** Definir ambos os níveis como altos e verificar o efeito combinado.
*   **Pós-condições/Resultado Esperado:**
    1.  Observação de que `CurrentFiringThreshold` dos neurônios muda em resposta aos níveis simulados de cortisol e dopamina, de acordo com as regras definidas em `ApplyEffectsToNeurons` e os parâmetros em `SimulationParameters`.
*   **Notas:** Este caso de uso é mais facilmente validado através de testes unitários do pacote `neurochemical` (como os já existentes) ou pequenos scripts de teste, em vez de uma execução completa da CLI, pois requer controle fino sobre os níveis químicos e inspeção direta dos estados dos neurônios.
```
