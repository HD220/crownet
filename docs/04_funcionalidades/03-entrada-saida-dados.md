# Funcionalidade: Manipulação de Dados de Entrada e Saída na Rede CrowNet

## 1. Visão Geral

Esta funcionalidade descreve como os dados de entrada são preparados e apresentados à rede neural CrowNet, como a rede representa internamente sua resposta, e que tipo de informação o sistema fornece ao usuário sobre esses processos.

## 2. Entrada de Dados para a Rede

### 2.1. Representação dos Padrões de Entrada
*   O sistema utiliza um conjunto predefinido de padrões para representar os dígitos numéricos de 0 a 9.
*   Cada dígito é codificado como um padrão binário (matriz de pixels com valores 0 ou 1) em um formato de 5x7 pixels.
*   Estes padrões são fornecidos por um componente interno de geração de dados.

### 2.2. Apresentação de um Padrão à Rede
*   Quando um padrão de dígito é apresentado à rede:
    *   Os 35 neurônios de entrada designados (correspondentes aos 35 pixels do padrão 5x7) são ativados de acordo com o padrão.
    *   Neurônios de entrada correspondentes a pixels "ativos" (valor 1) no padrão são forçados a disparar uma vez, gerando pulsos que iniciam a atividade na rede.
    *   Neurônios de entrada correspondentes a pixels "inativos" (valor 0) não são diretamente estimulados neste processo.

## 3. Saída de Dados da Rede (Resposta Neuronal)

### 3.1. Neurônios de Saída Designados
*   Um conjunto específico de 10 neurônios na rede é designado como neurônios de saída. Sua atividade coletiva representa a resposta da rede a um determinado estímulo.

### 3.2. Interpretação da Resposta da Rede
*   A resposta da rede a um padrão de entrada é capturada pelo estado de ativação desses 10 neurônios de saída.
*   No MVP, a principal medida de ativação é o valor do potencial acumulado (o `AccumulatedPulse`) em cada um dos neurônios de saída.
*   Este vetor de ativação de 10 dimensões é geralmente coletado após a rede ter processado a entrada por um número definido de ciclos de simulação (um período de "acomodação").
*   O objetivo do treinamento da rede é que esses vetores de ativação se tornem distintos e consistentemente associados a cada dígito de entrada diferente.
*   *Nota: O MVP foca na capacidade da rede de formar representações internas distinguíveis. A tradução explícita dessas representações de volta para rótulos de dígitos (0-9) é considerada uma etapa de análise ou funcionalidade pós-MVP.*

## 4. Feedback da Aplicação ao Usuário (Saída no Console)

O sistema fornece informações ao usuário através do console, variando conforme o modo de operação:

*   **Durante o Treinamento (modo `expose`):**
    *   Informações de progresso, como o número da época atual.
    *   Níveis atuais dos neuroquímicos simulados (Cortisol e Dopamina).
    *   Opcionalmente (para depuração), resumos sobre mudanças significativas nos pesos sinápticos.
*   **Durante a Observação/Teste (modo `observe`):**
    *   Visualização do padrão do dígito de entrada que foi apresentado à rede.
    *   O vetor de ativação resultante dos 10 neurônios de saída.
*   **Durante a Simulação Geral (modo `sim`):**
    *   Estatísticas gerais da simulação, como o ciclo atual.
    *   Se configurado, a frequência de disparo de neurônios específicos que estão sendo monitorados.
    *   Níveis atuais dos neuroquímicos.
*   **Informações de Depuração:** Se flags de depuração estiverem ativas, o sistema pode imprimir informações mais detalhadas sobre processos internos, como a dinâmica dos neuroquímicos.

Este fluxo de dados e feedback permite que o usuário interaja com a rede, treine-a, observe seu comportamento e analise seus processos internos.
