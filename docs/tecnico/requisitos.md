# Requisitos Funcionais e Não Funcionais - CrowNet MVP

## 1. Introdução

Este documento detalha os Requisitos Funcionais (RF) e Não Funcionais (RNF) para o Minimum Viable Product (MVP) do projeto CrowNet, com foco na demonstração de autoaprendizagem Hebbiana neuromodulada.

## 2. Requisitos Funcionais (RF)

Os Requisitos Funcionais são derivados da documentação em `docs/funcional/`.

### RF-INIT: Inicialização da Rede
*   **RF-INIT-001:** O sistema deve permitir a configuração do número total de neurônios via parâmetro de CLI.
*   **RF-INIT-002:** O sistema deve criar um mínimo de 35 neurônios de Input e 10 neurônios de Output.
*   **RF-INIT-003:** Os neurônios restantes devem ser distribuídos entre os tipos Excitatory, Inhibitory e Dopaminergic com base em porcentagens configuráveis.
*   **RF-INIT-004:** Cada neurônio deve ser posicionado aleatoriamente em um espaço vetorial 16D, dentro de restrições radiais específicas para seu tipo.
*   **RF-INIT-005:** Cada neurônio deve ser inicializado com um ID único, tipo, estado inicial (Repouso), acumulador de pulso zero e limiar de disparo base.
*   **RF-INIT-006:** O sistema deve inicializar pesos sinápticos "all-to-all" (exceto auto-conexões) com valores pequenos e aleatórios entre todos os pares de neurônios. Auto-conexões devem ter peso zero.

### RF-SIM: Ciclo de Simulação e Dinâmicas da Rede
*   **RF-SIM-001:** A simulação deve progredir em ciclos de tempo discretos.
*   **RF-SIM-002:** Em cada ciclo, o acumulador de pulso de cada neurônio deve decair gradualmente.
*   **RF-SIM-003:** Cada neurônio deve seguir uma máquina de estados (Repouso, Disparo, Refratário Absoluto, Refratário) com durações configuráveis para os estados refratários.
*   **RF-SIM-004:** Um neurônio deve disparar se seu acumulador de pulso exceder seu limiar de disparo atual e não estiver em estado refratário absoluto.
*   **RF-SIM-005:** Pulsos devem se propagar esfericamente a partir de um neurônio em disparo, com velocidade fixa.
*   **RF-SIM-006:** Neurônios dentro da "casca" de efeito de um pulso em um ciclo devem ser considerados atingidos.
*   **RF-SIM-007:** O efeito de um pulso em um neurônio atingido deve ser `ValorBaseDoPulso * PesoDaSinapseOrigemDestino`.
*   **RF-SIM-008:** O `ValorBaseDoPulso` deve ser +1.0 para neurônios excitatórios/input/output, -1.0 para inibitórios, e 0.0 para dopaminérgicos (cujo efeito é químico).

### RF-LEARN: Aprendizado Hebbiano Neuromodulado
*   **RF-LEARN-001:** O sistema deve implementar a plasticidade Hebbiana para ajustar pesos sinápticos.
*   **RF-LEARN-002:** A co-ativação de um neurônio pré-sináptico e um pós-sináptico (disparo dentro de uma `HebbianCoincidenceWindow`) deve levar ao fortalecimento de seu peso sináptico.
*   **RF-LEARN-003:** Uma taxa de aprendizado base (`BaseLearningRate`) deve ser configurável.
*   **RF-LEARN-004:** A taxa de aprendizado efetiva deve ser modulada pelos níveis de Dopamina (aumenta) e Cortisol (altos níveis suprimem).
*   **RF-LEARN-005:** Os pesos sinápticos devem sofrer um leve decaimento (`HebbianWeightDecay`) a cada atualização.
*   **RF-LEARN-006:** Os pesos sinápticos devem ser limitados a uma faixa mínima e máxima (`HebbianWeightMin`/`Max`).

### RF-CHEM: Modulação Química (Cortisol e Dopamina)
*   **RF-CHEM-001:** O sistema deve simular os níveis de Cortisol e Dopamina.
*   **RF-CHEM-002 (Cortisol):** Cortisol deve ser produzido quando pulsos excitatórios atingem a vizinhança de uma glândula central.
*   **RF-CHEM-003 (Dopamina):** Dopamina deve ser produzida quando neurônios dopaminérgicos disparam.
*   **RF-CHEM-004:** Ambos os neuroquímicos devem decair percentualmente a cada ciclo e ser limitados a um nível máximo.
*   **RF-CHEM-005:** Cortisol deve modular os limiares de disparo dos neurônios (efeito em forma de U).
*   **RF-CHEM-006:** Dopamina deve modular os limiares de disparo dos neurônios (aumento).
*   **RF-CHEM-007:** Cortisol (altos níveis) deve reduzir a taxa de aprendizado Hebbiano e a taxa de sinaptogênese.
*   **RF-CHEM-008:** Dopamina deve aumentar a taxa de aprendizado Hebbiano e a taxa de sinaptogênese.
*   **RF-CHEM-009:** Os efeitos de Cortisol e Dopamina nos limiares e na sinaptogênese devem ser combinados (multiplicativamente para sinaptogênese, sequencialmente para limiares).

### RF-SYNAPTO: Sinaptogênese
*   **RF-SYNAPTO-001:** O sistema deve simular o movimento de neurônios no espaço 16D (sinaptogênese).
*   **RF-SYNAPTO-002:** Neurônios devem ser atraídos por neurônios ativos (disparando/refratários) e repelidos por neurônios em repouso.
*   **RF-SYNAPTO-003:** A magnitude do movimento (velocidade) deve ser modulada pelos níveis de Cortisol e Dopamina.
*   **RF-SYNAPTO-004:** O movimento deve ser amortecido e limitado a uma velocidade máxima por ciclo.
*   **RF-SYNAPTO-005:** Os neurônios devem permanecer dentro dos limites do espaço definido.

### RF-IO: Entrada e Saída de Dados
*   **RF-IO-001:** O sistema deve utilizar padrões binários 5x7 predefinidos para os dígitos 0-9.
*   **RF-IO-002:** A apresentação de um padrão de dígito deve forçar os 35 neurônios de input correspondentes a disparar uma vez.
*   **RF-IO-003:** A resposta da rede a um padrão de entrada deve ser o vetor de `AccumulatedPulse` dos 10 neurônios de output designados, após um período de acomodação.
*   **RF-IO-004:** O sistema deve exibir informações de progresso e estado no console, específicas para cada modo de operação.
*   **RF-IO-005 (Modo sim):** Permitir estímulo contínuo a um neurônio de input com frequência definida.
*   **RF-IO-006 (Modo sim):** Permitir monitoramento da frequência de disparo de um neurônio de output.

### RF-PERSIST: Persistência de Dados
*   **RF-PERSIST-001:** O sistema deve ser capaz de salvar os pesos sinápticos da rede em um arquivo JSON.
*   **RF-PERSIST-002:** O sistema deve ser capaz de carregar pesos sinápticos de um arquivo JSON para inicializar a rede.
*   **RF-PERSIST-003:** O sistema deve, opcionalmente (controlado por flag), salvar snapshots completos do estado da rede (neurônios, químicos) em um banco de dados SQLite.
*   **RF-PERSIST-004:** O logging para SQLite deve ocorrer em intervalos de ciclos configuráveis.
*   **RF-PERSIST-005:** O arquivo de banco de dados SQLite deve ser recriado a cada execução que utiliza esta opção.

### RF-MODE: Modos de Operação da CLI
*   **RF-MODE-001 (expose):** Implementar um modo para expor a rede a sequências de padrões de dígitos por múltiplas épocas, permitindo o aprendizado Hebbiano. Todas as dinâmicas (aprendizado, químicos, sinaptogênese) devem estar ativas. Deve carregar pesos existentes se disponíveis e salvar pesos ao final.
*   **RF-MODE-002 (observe):** Implementar um modo para carregar pesos treinados, apresentar um único padrão de dígito e exibir o padrão de ativação dos neurônios de saída. Dinâmicas de aprendizado, sinaptogênese e modulação química devem ser temporariamente desativadas durante os ciclos de acomodação neste modo.
*   **RF-MODE-003 (sim):** Implementar um modo de simulação geral onde todas as dinâmicas da rede estão ativas, permitindo estímulo contínuo e logging para SQLite.

## 3. Requisitos Não Funcionais (RNF)

### RNF-USAB: Usabilidade (CLI)
*   **RNF-USAB-001:** A interface de linha de comando deve ser clara, com flags e parâmetros de modo bem definidos.
*   **RNF-USAB-002:** A saída no console deve ser informativa, indicando o progresso da simulação e os resultados relevantes para cada modo.
*   **RNF-USAB-003:** Mensagens de erro devem ser claras e ajudar o usuário a identificar problemas de configuração ou uso.

### RNF-PERF: Desempenho
*   **RNF-PERF-001 (MVP):** A simulação de uma rede com ~100-200 neurônios por algumas centenas de ciclos (ex: 10 épocas * 10 dígitos * 5 ciclos/dígito = 500 ciclos) deve ser concluída em tempo razoável em hardware de desktop comum (ex: poucos minutos), permitindo iteração.
*   **RNF-PERF-002:** A reescrita do código deve evitar gargalos óbvios de desempenho, especialmente em cálculos de distância e iterações sobre neurônios/pulsos.

### RNF-MAINT: Manutenibilidade
*   **RNF-MAINT-001:** O código Go reescrito deve seguir as convenções padrão do Go.
*   **RNF-MAINT-002:** O código deve ser bem organizado em pacotes com responsabilidades claras.
*   **RNF-MAINT-003:** O código deve aderir aos princípios de Código Limpo (nomes significativos, funções pequenas, baixo acoplamento onde possível).
*   **RNF-MAINT-004:** O código deve seguir as regras do Object Calisthenics, conforme especificado no prompt do projeto.
*   **RNF-MAINT-005:** A documentação técnica e funcional deve ser mantida atualizada com o código.

### RNF-REL: Confiabilidade
*   **RNF-REL-001:** A simulação deve ser determinística se a semente do gerador de números aleatórios for fixada (embora o MVP atual não fixe explicitamente a semente global, as inicializações aleatórias devem ser consistentes dentro de uma execução). *Nota para reescrita: considerar tornar a semente configurável para reprodutibilidade total.*
*   **RNF-REL-002:** O sistema deve lidar graciosamente com arquivos de pesos ou de banco de dados ausentes ou malformados (ex: ao tentar carregar).

### RNF-CONF: Configurabilidade
*   **RNF-CONF-001:** Parâmetros chave da simulação (número de neurônios, taxas de aprendizado, parâmetros químicos, etc.) devem ser configuráveis (via flags no MVP; via arquivos de configuração na reescrita ideal).

### RNF-EXT: Extensibilidade (para a Reescrita)
*   **RNF-EXT-001:** A arquitetura reescrita deve facilitar a adição de novos tipos de neurônios, neuroquímicos ou regras de aprendizado no futuro.

### RNF-TEST: Testabilidade (para a Reescrita)
*   **RNF-TEST-001:** O código reescrito deve ser estruturado para permitir testes unitários de componentes críticos (ex: lógica neuronal, algoritmos de aprendizado, propagação de pulso).
```
