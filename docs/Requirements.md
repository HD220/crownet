# Requisitos Funcionais e Não Funcionais - CrowNet

## 1. Introdução

Este documento detalha os requisitos funcionais (o que o sistema deve fazer) e não funcionais (como o sistema deve ser) para o projeto CrowNet.

## 2. Requisitos Funcionais (RF)

### RF-CORE: Simulação da Rede Neural
*   **RF-CORE-001:** O sistema deve simular uma rede de neurônios em um espaço vetorial de 16 dimensões.
*   **RF-CORE-002:** O sistema deve permitir a configuração do número total de neurônios na rede.
*   **RF-CORE-003:** A posição inicial dos neurônios deve ser gerada proceduralmente (usando `math/rand` para o MVP).
*   **RF-CORE-004:** O sistema deve suportar os seguintes tipos de neurônios com distribuições percentuais configuráveis (defaults no README):
    *   Dopaminérgicos
    *   Inibitórios
    *   Excitatórios
    *   Input
    *   Output
*   **RF-CORE-005:** O sistema deve simular uma glândula de cortisol localizada no centro do espaço vetorial.

### RF-NEURON: Dinâmica dos Neurônios
*   **RF-NEURON-001:** Cada neurônio deve operar em um dos quatro ciclos: Repouso, Disparo, Refratário Absoluto, Refratário.
*   **RF-NEURON-002:** Um neurônio deve disparar um pulso quando a soma dos pulsos recebidos exceder seu limiar de disparo.
*   **RF-NEURON-003:** Quando um neurônio não recebe pulsos, a soma dos seus pulsos acumulados deve diminuir gradativamente até um valor basal.
*   **RF-NEURON-004:** Cada neurônio deve manter um registro do último ciclo em que disparou. Este registro só é atualizado no momento do disparo.
*   **RF-NEURON-005:** O limiar de disparo dos neurônios deve ser influenciado pelos níveis de cortisol e dopamina.

### RF-PULSE: Propagação de Pulso
*   **RF-PULSE-001:** A propagação de pulso entre neurônios deve ser baseada na distância Euclidiana em 16D.
*   **RF-PULSE-002:** A velocidade de propagação de pulso deve ser configurável (default: 0.6 unidades por ciclo).
*   **RF-PULSE-003:** O sistema deve identificar neurônios vizinhos dentro da área de alcance de um pulso.
    *   A busca de vizinhos deve ser realizada comparando a distância do neurônio emissor aos pontos referenciais (vértices de um hipercubo e o centro), descontando o raio de alcance do pulso.
*   **RF-PULSE-004:** Pulsos de neurônios excitatórios devem somar ao potencial dos neurônios alvo.
*   **RF-PULSE-005:** Pulsos de neurônios inibitórios devem subtrair do potencial dos neurônios alvo.
*   **RF-PULSE-006:** Pulsos de neurônios dopaminérgicos devem contribuir para o nível de dopamina na sua vizinhança.

### RF-CHEM: Modulação Química (Cortisol e Dopamina)
*   **RF-CHEM-001 (Cortisol):** A produção de cortisol pela glândula de cortisol deve ser aumentada por pulsos excitatórios que a atingem.
*   **RF-CHEM-002 (Cortisol):** Na ausência de estímulos, a quantidade de cortisol deve diminuir ao longo do tempo.
*   **RF-CHEM-003 (Cortisol):** O cortisol deve, inicialmente, diminuir o limiar de disparo dos neurônios. Após atingir um pico, níveis elevados de cortisol devem aumentar o limiar de disparo e reduzir a sinapogênese.
*   **RF-CHEM-004 (Dopamina):** A dopamina deve ser gerada por neurônios dopaminérgicos quando disparam.
*   **RF-CHEM-005 (Dopamina):** A dopamina deve aumentar o limiar de disparo dos neurônios e aumentar a taxa de sinapogênese.
*   **RF-CHEM-006 (Dopamina):** A dopamina deve ter uma taxa de decaimento ao longo do tempo, mais acentuada que a do cortisol.

### RF-SYNAPTO: Sinapogênese
*   **RF-SYNAPTO-001:** O sistema deve simular a sinapogênese, alterando a posição dos neurônios no espaço 16D.
*   **RF-SYNAPTO-002:** Neurônios devem se aproximar de neurônios que dispararam recentemente ou que estão em período refratário.
*   **RF-SYNAPTO-003:** Neurônios devem se afastar de neurônios que estão em estado de repouso.
*   **RF-SYNAPTO-004:** A taxa de sinapogênese deve ser influenciada pelos níveis de cortisol e dopamina.

### RF-IO: Input e Output
*   **RF-IO-001:** O sistema deve permitir a designação de um subconjunto de neurônios como neurônios de input.
*   **RF-IO-002:** O sistema deve permitir a designação de um subconjunto de neurônios como neurônios de output.
*   **RF-IO-003:** O input para a rede deve ser codificado pela frequência de disparo induzida nos neurônios de input.
*   **RF-IO-004:** O output da rede deve ser representado pela frequência de disparo dos neurônios de output.
*   **RF-IO-005:** Para o MVP, o sistema deve fornecer uma interface de console para fornecer inputs e observar outputs.

### RF-DATA: Persistência e Logging
*   **RF-DATA-001:** O sistema deve armazenar o estado da simulação (posições dos neurônios, estados, níveis químicos, etc.) em um banco de dados SQLite.
*   **RF-DATA-002:** O logging no banco de dados deve ocorrer em intervalos de ciclos configuráveis.
*   **RF-DATA-003:** O banco de dados deve ser nomeado `crownet.db` e armazenado no diretório `data/`.

### RF-SIM: Controle da Simulação
*   **RF-SIM-001:** O sistema deve permitir ao usuário iniciar e parar a simulação.
*   **RF-SIM-002:** O sistema deve permitir ao usuário especificar o número de ciclos de simulação a serem executados.

## 3. Requisitos Não Funcionais (RNF)

### RNF-PERF: Desempenho
*   **RNF-PERF-001:** O MVP deve ser capaz de simular uma rede de pelo menos 100 neurônios em tempo razoável em hardware de desktop comum. (O "tempo razoável" será definido melhor durante o desenvolvimento, mas visa permitir várias iterações de simulação por hora).
*   **RNF-PERF-002:** Os cálculos de distância e busca de vizinhos devem ser otimizados para evitar gargalos de desempenho.
*   **RNF-PERF-003 (Pós-MVP):** O sistema deverá ser capaz de ser acelerado utilizando ArrayFire para simulações maiores e mais rápidas.

### RNF-USAB: Usabilidade
*   **RNF-USAB-001 (MVP):** A interface de console para input/output e controle da simulação deve ser clara e fácil de usar.
*   **RNF-USAB-002 (Pós-MVP):** A interface gráfica (com Robotgo) deve fornecer uma visualização intuitiva do estado da rede e permitir interação fácil.

### RNF-REL: Confiabilidade
*   **RNF-REL-001:** A simulação deve ser determinística dado o mesmo conjunto de parâmetros iniciais e inputs.
*   **RNF-REL-002:** O sistema deve lidar graciosamente com configurações inválidas (ex: proporções de neurônios que não somam 100%).

### RNF-MAINT: Manutenibilidade
*   **RNF-MAINT-001:** O código fonte deve ser bem organizado, comentado e seguir as convenções da linguagem Go.
*   **RNF-MAINT-002:** O projeto deve incluir testes unitários para componentes críticos.

### RNF-SCAL: Escalabilidade
*   **RNF-SCAL-001 (Pós-MVP):** A arquitetura do sistema deve permitir o aumento do número de neurônios e a complexidade das interações sem grandes refatorações.

### RNF-PORT: Portabilidade
*   **RNF-PORT-001:** O núcleo da simulação em Go deve ser compilável e executável em múltiplos sistemas operacionais (Linux, macOS, Windows).
*   **RNF-PORT-002 (Pós-MVP):** A dependência de Robotgo e ArrayFire pode introduzir limitações de portabilidade que devem ser documentadas.

### RNF-EXT: Extensibilidade
*   **RNF-EXT-001:** Deve ser relativamente fácil adicionar novos tipos de neurônios, substâncias químicas ou regras de interação no futuro.

### RNF-SEC: Segurança
*   **RNF-SEC-001:** Não aplicável diretamente para o MVP, pois não há manipulação de dados sensíveis ou acesso externo além do controle do usuário local. Futuras versões com interfaces web ou remotas precisariam considerar aspectos de segurança.

## 4. Priorização (MVP)

Para o MVP, o foco principal será nos requisitos funcionais relacionados ao núcleo da simulação (RF-CORE, RF-NEURON, RF-PULSE, RF-CHEM, RF-SYNAPTO), input/output básico (RF-IO) e persistência de dados (RF-DATA). Requisitos não funcionais como desempenho básico (RNF-PERF-001) e manutenibilidade (RNF-MAINT) também são cruciais. Funcionalidades avançadas de visualização e otimizações pesadas (ArrayFire) são Pós-MVP.
