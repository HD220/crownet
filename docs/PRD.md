# Documento de Requisitos do Produto (PRD) - CrowNet

## 1. Introdução

CrowNet é um modelo computacional de rede neural inspirado em processos biológicos. Ele simula a interação de neurônios em um espaço vetorial de 16 dimensões, incorporando dinâmicas como sinapogênese, propagação de pulso e a influência de substâncias neuroquímicas como cortisol e dopamina. O objetivo principal é criar um modelo procedural e otimizado de redes neurais biológicas, permitindo a exploração de comportamentos emergentes e aprendizado.

## 2. Objetivos do Produto

*   **Simulação Biologicamente Plausível:** Criar um modelo que capture aspectos chave do funcionamento de redes neurais reais.
*   **Desempenho Computacional:** Garantir que a simulação possa ser executada eficientemente, permitindo a exploração de redes de tamanho considerável e longos períodos de simulação.
*   **Exploração e Pesquisa:** Fornecer uma plataforma para pesquisadores e entusiastas estudarem princípios de neurociência computacional e inteligência artificial.
*   **Visualização e Interação:** Permitir a visualização do estado da rede e a interação em tempo real (objetivo de longo prazo).
*   **MVP Funcional:** Entregar uma primeira versão funcional que demonstre os principais mecanismos do modelo.

## 3. Público-Alvo

*   Pesquisadores em neurociência computacional.
*   Estudantes de inteligência artificial e neurociência.
*   Desenvolvedores interessados em modelos de IA bio-inspirados.
*   Hobbistas e entusiastas de simulações complexas.

## 4. Funcionalidades Chave (MVP)

*   **Inicialização da Rede:**
    *   Geração procedural de neurônios em um espaço 16D.
    *   Distribuição definida de tipos de neurônios (Dopaminérgicos, Inibitórios, Excitatórios, Input, Output).
    *   Localização central da glândula de cortisol.
*   **Dinâmica Neuronal:**
    *   Implementação dos 4 ciclos neuronais (Repouso, Disparo, Refratário Absoluto, Refratário).
    *   Propagação de pulso baseada em distância Euclidiana e velocidade definida.
    *   Cálculo de distância e busca de vizinhos otimizados.
    *   Neurônios disparam ao exceder o limiar de disparo; soma de pulsos decai gradualmente.
*   **Modulação Neuroquímica:**
    *   Simulação da produção e efeito do cortisol pela glândula central.
    *   Simulação da produção e efeito da dopamina por neurônios dopaminérgicos.
    *   Influência do cortisol e dopamina no limiar de disparo e sinapogênese.
*   **Sinapogênese:**
    *   Movimentação de neurônios baseada na atividade da rede (aproximação/afastamento).
*   **Input/Output:**
    *   Codificação de input e output baseada na frequência de pulsos.
    *   Interface simples para fornecer input e observar output (via console para o MVP).
*   **Persistência de Dados:**
    *   Armazenamento do estado da rede em um banco de dados SQLite para análise e monitoramento.
*   **Logging:**
    *   Registro do último ciclo de disparo de cada neurônio.

## 5. Critérios de Sucesso (MVP)

*   A simulação executa sem erros e demonstra os comportamentos básicos esperados (disparo de neurônios, propagação de pulsos).
*   Os efeitos do cortisol e da dopamina são observáveis na dinâmica da rede.
*   A sinapogênese resulta em alterações na topologia da rede ao longo do tempo.
*   Os dados da simulação são corretamente armazenados no banco de dados SQLite.
*   É possível fornecer um input simples e observar uma resposta correspondente no output.

## 6. Tecnologias (Conforme README)

*   Linguagem Principal: Go
*   Aceleração (Pós-MVP): ArrayFire
*   Banco de Dados: SQLite
*   Visualização (Pós-MVP): Robotgo
*   Geração Procedural (Inicial): `math/rand` ou similar em Go. (Pós-MVP: OpenNoise ou customizado)

## 7. Considerações Futuras (Pós-MVP)

*   Integração com ArrayFire para otimização de desempenho.
*   Desenvolvimento de uma interface gráfica para visualização e interação com Robotgo.
*   Implementação de mecanismos de aprendizado mais sofisticados (ex: Hebbian learning).
*   Uso de geradores de ruído avançados (OpenNoise) para inicialização da rede.
*   Testes e benchmarking de desempenho mais extensivos.
*   Criação de ferramentas de análise de dados da simulação.
