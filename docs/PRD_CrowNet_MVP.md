# Product Requirements Document: CrowNet MVP - Reconhecimento de Dígitos

## 1. Introdução

Este documento descreve os requisitos para o Minimum Viable Product (MVP) do projeto CrowNet. O CrowNet é um modelo computacional de rede neural inspirado em processos biológicos, simulando a interação de neurônios em um espaço vetorial de 16 dimensões. O MVP se concentrará em demonstrar a capacidade de aprendizado da CrowNet através de uma tarefa de classificação de dígitos manuscritos.

## 2. Objetivos do Produto

*   **Objetivo Principal:** Desenvolver uma implementação funcional da CrowNet capaz de aprender a classificar dígitos manuscritos (0-9) do dataset MNIST com uma taxa de acerto mínima de 80% no conjunto de teste.
*   **Objetivos Secundários:**
    *   Validar a arquitetura e os mecanismos de aprendizado propostos para a CrowNet (sinapogênese, modulação por cortisol/dopamina).
    *   Fornecer uma base para futuras expansões e aplicações da CrowNet.
    *   Permitir o armazenamento e carregamento do estado da rede neural para persistência do aprendizado.

## 3. Público-Alvo

*   **Usuário Primário:** Desenvolvedores e pesquisadores interessados em modelos de redes neurais biologicamente plausíveis e não convencionais.
*   **Usuário Secundário:** Entusiastas de inteligência artificial e aprendizado de máquina.

## 4. Escopo do MVP

### 4.1. Funcionalidades Incluídas:

*   **Núcleo da CrowNet:**
    *   Simulação de neurônios em um espaço vetorial 16D.
    *   Distribuição procedural inicial de neurônios (Dopaminérgicos, Inibitórios, Excitatórios, Input, Output) usando um gerador de números aleatórios com seed.
    *   Implementação dos ciclos neuronais (Repouso, Disparo, Refratário Absoluto, Refratário).
    *   Propagação de pulso baseada em distância Euclidiana e velocidade definida.
    *   Lógica de disparo de neurônios baseada em limiar.
    *   Mecanismo de sinapogênese (movimentação de neurônios).
    *   Simulação dos neuroquímicos Cortisol e Dopamina e seus efeitos na rede.
    *   Glândula de cortisol central.
*   **Aplicação de Reconhecimento de Dígitos:**
    *   Carregamento e pré-processamento do dataset MNIST.
    *   Codificação de imagens MNIST em padrões de frequência de disparo para os neurônios de input da CrowNet.
    *   Decodificação da atividade dos neurônios de output para determinar o dígito classificado.
    *   Implementação de um mecanismo de aprendizado emergente:
        *   Ajuste da dinâmica da rede (via cortisol/dopamina ou diretamente na sinapogênese) em resposta à correção da classificação (recompensa/punição).
*   **Interface e Gerenciamento:**
    *   Interface de Linha de Comando (CLI) para:
        *   Iniciar o processo de treinamento da rede com o dataset MNIST.
        *   Avaliar o desempenho da rede treinada em um conjunto de teste.
        *   Exibir a taxa de acerto e outras métricas relevantes (ex: perda, se aplicável).
        *   Salvar o estado da rede neural treinada em um arquivo (usando SQLite).
        *   Carregar um estado de rede neural previamente salvo.
*   **Armazenamento:**
    *   Uso do SQLite para persistir o estado da rede (posições dos neurônios, parâmetros, níveis de neuroquímicos no momento do salvamento).

### 4.2. Funcionalidades Excluídas (Pós-MVP):

*   Visualização gráfica avançada em tempo real com Robotgo.
*   Integração com OpenNoise para geração procedural complexa (um substituto mais simples será usado).
*   Otimização de desempenho com ArrayFire (a implementação inicial será em Go puro).
*   Interface gráfica de usuário (GUI).
*   Suporte para outros datasets além do MNIST.
*   Mecanismos de aprendizado supervisionado tradicionais (como backpropagation direto), a menos que o aprendizado emergente se mostre inviável e uma adaptação seja necessária.
*   Controle interativo da simulação via teclado/mouse (além da CLI).
*   Implementação da busca de vizinhos via "17 referenciais" e "resolução de equação quadrática" (será usada uma abordagem simplificada de raio de alcance para o MVP).

## 5. Requisitos

### 5.1. Requisitos Funcionais:

*   **RF-001:** O sistema deve permitir a inicialização de uma rede CrowNet com um número configurável de neurônios (dentro de limites razoáveis para desempenho do MVP).
*   **RF-002:** O sistema deve distribuir os tipos de neurônios (Dopaminérgicos, Inibitórios, Excitatórios, Input, Output) conforme as proporções especificadas no README.md.
*   **RF-003:** O sistema deve simular a propagação de pulsos entre neurônios com base na distância Euclidiana em 16D e velocidade de 0.6 unidades/ciclo.
*   **RF-004:** O sistema deve implementar os quatro ciclos de estado dos neurônios.
*   **RF-005:** O sistema deve implementar a lógica de disparo dos neurônios quando a soma dos pulsos recebidos exceder o limiar.
*   **RF-006:** O sistema deve implementar a sinapogênese, ajustando as posições dos neurônios.
*   **RF-007:** O sistema deve simular a produção e os efeitos do cortisol e da dopamina na rede.
*   **RF-008:** O sistema deve ser capaz de carregar imagens e rótulos do dataset MNIST.
*   **RF-009:** O sistema deve codificar uma imagem MNIST em um padrão de ativação (frequência de pulsos) para os neurônios de input.
*   **RF-010:** O sistema deve decodificar a atividade dos neurônios de output para produzir uma classificação de dígito (0-9).
*   **RF-011:** O sistema deve implementar um loop de treinamento que itere sobre o dataset MNIST.
*   **RF-012:** O sistema deve aplicar um mecanismo de recompensa/punição para modular o aprendizado com base na precisão da classificação.
*   **RF-013:** O sistema deve permitir salvar o estado atual da rede neural em um banco de dados SQLite.
*   **RF-014:** O sistema deve permitir carregar um estado de rede neural previamente salvo do SQLite.
*   **RF-015:** O sistema deve calcular e exibir a taxa de acerto da classificação no conjunto de treinamento e teste.
*   **RF-016:** A CLI deve fornecer comandos para treinar, avaliar, salvar e carregar o modelo.

### 5.2. Requisitos Não Funcionais:

*   **RNF-001 (Desempenho):** O sistema deve atingir uma taxa de acerto de classificação de pelo menos 80% no conjunto de teste do MNIST após o treinamento.
*   **RNF-002 (Desempenho):** O processo de treinamento para um número razoável de épocas (a ser determinado experimentalmente) deve ser concluído em um tempo viável em hardware de desktop moderno (ex: horas, não semanas).
*   **RNF-003 (Usabilidade):** A CLI deve ser intuitiva e fornecer feedback claro ao usuário sobre as operações em andamento.
*   **RNF-004 (Persistência):** Os modelos salvos devem poder ser recarregados corretamente para restaurar o estado da rede.
*   **RNF-005 (Configurabilidade):** Parâmetros chave da simulação (ex: número de neurônios, taxas de aprendizado implícitas nos efeitos dos neuroquímicos) devem ser configuráveis, possivelmente através de um arquivo de configuração ou argumentos de CLI, para facilitar a experimentação.
*   **RNF-006 (Reprodutibilidade):** Dada a mesma seed para geração de números aleatórios e a mesma sequência de dados de entrada, a simulação e o processo de aprendizado devem ser reproduzíveis.

## 6. Critérios de Sucesso do MVP

*   Atingir 80% ou mais de acurácia na classificação de dígitos do MNIST no conjunto de teste.
*   Todas as funcionalidades da CLI (treinar, avaliar, salvar, carregar) estão operacionais.
*   O modelo de aprendizado emergente demonstra capacidade de melhoria ao longo do tempo de treinamento.
*   O código é modular, razoavelmente bem documentado (comentários) e testável.

## 7. Considerações Futuras (Pós-MVP)

*   Implementação de otimizações de desempenho (ArrayFire).
*   Melhorias na geração procedural de neurônios (OpenNoise).
*   Desenvolvimento de ferramentas de visualização mais robustas (Robotgo).
*   Exploração de outras tarefas de aprendizado e datasets.
*   Refinamento dos modelos de neuroquímicos e sinapogênese.
*   Experimentação com diferentes arquiteturas de rede e distribuições de neurônios.

## 8. Riscos e Desafios

*   **Convergência do Aprendizado:** O maior risco é a dificuldade em fazer o modelo de aprendizado emergente convergir para a taxa de acerto desejada. Exigirá extensa experimentação e ajuste de parâmetros.
*   **Complexidade Computacional:** A simulação pode ser lenta, mesmo para o MVP, impactando a velocidade de experimentação.
*   **Interpretabilidade:** Entender por que a rede toma certas decisões ou por que o aprendizado está ou não funcionando pode ser difícil.
*   **Depuração:** Depurar uma simulação complexa com comportamento emergente é desafiador.

Este PRD servirá como guia para o desenvolvimento do MVP da CrowNet. Ele poderá ser atualizado conforme o projeto evolui e novos aprendizados são obtidos.
