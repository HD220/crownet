# Requisitos: CrowNet MVP - Reconhecimento de Dígitos

Este documento detalha os requisitos funcionais e não funcionais para o MVP (Minimum Viable Product) da CrowNet, focado na tarefa de reconhecimento de dígitos manuscritos.

## 1. Requisitos Funcionais (RF)

### 1.1. Núcleo da Simulação CrowNet
*   **RF-CORE-001:** O sistema deve instanciar uma rede de neurônios em um espaço vetorial de 16 dimensões.
*   **RF-CORE-002:** A posição inicial dos neurônios deve ser gerada proceduralmente dentro do espaço 16D, utilizando um gerador de números aleatórios com uma semente configurável para reprodutibilidade.
*   **RF-CORE-003:** O sistema deve atribuir tipos aos neurônios (Dopaminérgico, Inibitório, Excitatório, Input, Output) de acordo com as proporções definidas:
    *   1% Dopaminérgicos
    *   30% Inibitórios
    *   69% Excitatórios
    *   (Nota: Neurônios de Input e Output são 5% cada. Definir se são tipos exclusivos ou se um neurônio pode ser, por exemplo, Excitatório E de Input). Para o MVP, considerar que Input/Output são papéis que podem ser atribuídos a uma fração dos neurônios Excitatórios/Inibitórios, ou são categorias distintas que somam ao total. O README sugere categorias distintas. Vamos assumir categorias distintas por enquanto, ajustando as outras proporções se necessário ou definindo que os 5% de input/output são subconjuntos dos outros tipos. **Decisão MVP: Input/Output são categorias funcionais distintas, e os neurônios restantes são distribuídos entre Dopaminérgicos, Inibitórios, Excitatórios.** A soma total ainda deve ser 100% dos neurônios. Ex: 5% Input, 5% Output, 1% Dopaminérgico, 27% Inibitório, 62% Excitatório (ajustando para somar 100%).
*   **RF-CORE-004:** O sistema deve implementar os quatro ciclos de estado para cada neurônio: Repouso, Disparo, Refratário Absoluto, Refratário.
*   **RF-CORE-005:** O sistema deve calcular a distância Euclidiana entre neurônios no espaço 16D.
*   **RF-CORE-006:** A propagação de pulso deve ocorrer a uma velocidade de 0.6 unidades de distância por ciclo de simulação.
*   **RF-CORE-007:** Um neurônio deve disparar um pulso se a soma dos potenciais de pulsos recebidos exceder seu limiar de disparo individual.
*   **RF-CORE-008:** O potencial acumulado de um neurônio (soma dos pulsos) deve decair gradualmente se não receber novos pulsos.
*   **RF-CORE-009:** Cada neurônio deve manter um registro do último ciclo em que disparou.
*   **RF-CORE-010:** O sistema deve implementar a sinapogênese:
    *   Neurônios se aproximam de neurônios que dispararam ou estavam em período refratário recentemente.
    *   Neurônios se afastam de neurônios que estavam em repouso recentemente.
    *   A taxa de movimentação (sinapogênese) deve ser influenciada pelos níveis de cortisol e dopamina.
*   **RF-CORE-011:** Deve existir uma glândula de cortisol localizada no centro do espaço vetorial.
*   **RF-CORE-012:** A produção de cortisol deve aumentar quando pulsos excitatórios atingem a glândula de cortisol.
*   **RF-CORE-013:** O nível de cortisol deve diminuir com o tempo se a glândula não for estimulada.
*   **RF-CORE-014:** O cortisol deve modular o limiar de disparo dos neurônios (inicialmente diminui, mas com níveis altos, aumenta) e a taxa de sinapogênese (níveis altos diminuem).
*   **RF-CORE-015:** Neurônios dopaminérgicos devem produzir dopamina.
*   **RF-CORE-016:** O nível de dopamina deve decair com o tempo.
*   **RF-CORE-017:** A dopamina deve aumentar o limiar de disparo dos neurônios e aumentar a taxa de sinapogênese.

### 1.2. Aplicação de Reconhecimento de Dígitos (MNIST)
*   **RF-APP-001:** O sistema deve ser capaz de carregar o dataset MNIST (imagens de treinamento e teste, e seus respectivos rótulos).
*   **RF-APP-002:** O sistema deve pré-processar as imagens MNIST (ex: normalização de pixels, achatamento para vetor).
*   **RF-APP-003:** O sistema deve codificar uma imagem MNIST pré-processada em um padrão de frequência de disparo para os neurônios de input designados. O método de codificação deve ser definido (ex: intensidade do pixel para frequência, mapeamento de regiões da imagem para neurônios de input).
*   **RF-APP-004:** O sistema deve possuir 10 neurônios de output, cada um correspondendo a um dígito (0-9).
*   **RF-APP-005:** O sistema deve decodificar a atividade dos neurônios de output (ex: neurônio com maior frequência de disparo ou maior potencial acumulado durante uma janela de tempo) para determinar o dígito classificado.
*   **RF-APP-006:** O sistema deve implementar um loop de treinamento que itera sobre o dataset de treinamento MNIST por um número configurável de épocas.
*   **RF-APP-007:** Durante o treinamento, após a classificação de uma imagem, o sistema deve aplicar um mecanismo de "recompensa" (se a classificação for correta) ou "punição" (se incorreta).
*   **RF-APP-008:** O mecanismo de recompensa/punição deve influenciar a dinâmica da rede, presumivelmente através da modulação da produção/níveis de cortisol e/ou dopamina, ou afetando diretamente a sinapogênese de forma a reforçar/enfraquecer caminhos.
*   **RF-APP-009:** O sistema deve ser capaz de avaliar o desempenho do modelo treinado no conjunto de teste do MNIST, calculando a taxa de acerto.

### 1.3. Interface de Linha de Comando (CLI) e Gerenciamento
*   **RF-CLI-001:** A CLI deve permitir ao usuário iniciar o treinamento da rede CrowNet com o dataset MNIST.
    *   Parâmetros: número de épocas, (opcional) caminho para carregar modelo preexistente, caminho para salvar modelo treinado.
*   **RF-CLI-002:** A CLI deve permitir ao usuário avaliar um modelo CrowNet treinado usando o conjunto de teste MNIST.
    *   Parâmetros: caminho para o modelo treinado.
*   **RF-CLI-003:** A CLI deve exibir o progresso do treinamento (ex: época atual, taxa de acerto no lote/época).
*   **RF-CLI-004:** A CLI deve exibir a taxa de acerto final no conjunto de teste após a avaliação.
*   **RF-CLI-005:** O sistema deve permitir salvar o estado completo da rede neural (posições dos neurônios, seus estados, parâmetros, níveis de neuroquímicos) em um arquivo utilizando SQLite.
*   **RF-CLI-006:** O sistema deve permitir carregar um estado de rede neural previamente salvo de um arquivo SQLite para continuar o treinamento ou para avaliação.

## 2. Requisitos Não Funcionais (RNF)

*   **RNF-PERF-001 (Acurácia):** O MVP deve atingir uma taxa de acerto de classificação de pelo menos 80% no conjunto de teste do MNIST após um processo de treinamento adequado.
*   **RNF-PERF-002 (Tempo de Treinamento):** O treinamento da rede no dataset MNIST até atingir a acurácia alvo deve ser concluído em um tempo razoável em hardware de desktop comum (ex: abaixo de 12 horas).
*   **RNF-USAB-001 (CLI):** A interface de linha de comando deve ser clara, com comandos e opções compreensíveis. O feedback fornecido ao usuário deve ser informativo.
*   **RNF-CONF-001 (Configurabilidade):** Parâmetros chave da simulação (ex: número total de neurônios, taxa de decaimento de neuroquímicos, magnitude dos efeitos dos neuroquímicos, taxa base de sinapogênese, limiares de disparo iniciais) devem ser configuráveis (ex: via arquivo de configuração ou argumentos CLI) para facilitar a experimentação.
*   **RNF-REPRO-001 (Reprodutibilidade):** Utilizando a mesma semente para o gerador de números aleatórios, a mesma configuração e o mesmo dataset, os resultados do treinamento (evolução da acurácia) e da avaliação devem ser reproduzíveis.
*   **RNF-PERS-001 (Persistência):** O estado salvo da rede deve poder ser carregado corretamente, restaurando a rede ao estado em que foi salva.
*   **RNF-CODE-001 (Modularidade):** O código-fonte deve ser organizado de forma modular (ex: separação entre núcleo da simulação, lógica da aplicação MNIST, interface CLI).
*   **RNF-CODE-002 (Testabilidade):** Componentes críticos do núcleo da simulação devem ser passíveis de testes unitários.
*   **RNF-DOC-001 (Comentários):** O código deve conter comentários explicando seções complexas e decisões de design.

## 3. Requisitos de Dados

*   **RD-001:** O sistema utilizará o dataset MNIST, que consiste em imagens em escala de cinza de 28x28 pixels de dígitos manuscritos (0-9) e seus respectivos rótulos.
*   **RD-002:** O sistema deve ser capaz de lidar com o formato padrão do dataset MNIST.

## 4. Dependências (Tecnologias do README)

*   **RDEP-001 (SQLite):** SQLite será usado para persistência do estado do modelo. Uma biblioteca Go para SQLite (ex: `database/sql` com driver `mattn/go-sqlite3`) será utilizada.
*   **RDEP-002 (Go):** A linguagem de programação principal será Go.
*   **RDEP-003 (ArrayFire - Pós-MVP):** Não será requisito para o MVP inicial.
*   **RDEP-004 (Robotgo - Pós-MVP):** Não será requisito para o MVP inicial para visualização.
*   **RDEP-005 (OpenNoise - Pós-MVP):** Não será requisito para o MVP inicial; um gerador `math/rand` com seed será usado.

Esta lista de requisitos servirá como base para o desenvolvimento e teste do MVP da CrowNet. Poderá ser refinada à medida que o desenvolvimento progride.
