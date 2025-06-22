# Funcionalidade: Entrada, Saída e Representação de Dados (MVP)

Esta funcionalidade descreve como os dados são representados, como a entrada é fornecida à rede CrowNet e como a saída é interpretada no MVP.

## F3: Representação de Dígitos e Codificação de Entrada

### F3.1: Padrões de Dígitos Predefinidos
*   O sistema utiliza padrões binários (0 ou 1) predefinidos para os dígitos de 0 a 9.
*   Cada padrão é uma matriz de 5x7 pixels.
*   Esses padrões são fornecidos por um pacote dedicado (ex: `datagen`).

### F3.2: Apresentação do Padrão à Rede (`PresentPattern`)
*   Uma função específica (`PresentPattern`) é responsável por ativar os neurônios de entrada da rede com base em um padrão de dígito.
*   Os 35 neurônios de entrada designados correspondem aos 35 pixels do padrão 5x7.
*   Quando um padrão é apresentado:
    *   Os neurônios de entrada (os primeiros 35, conforme `InputNeuronIDs`) correspondentes a pixels "ativos" (valor 1.0 no padrão) são forçados a entrar no estado `FiringState`.
    *   Esses neurônios disparam imediatamente, gerando novos pulsos que são adicionados à lista `ActivePulses` da rede para serem processados nos ciclos de simulação subsequentes.
    *   Os neurônios de entrada correspondentes a pixels "inativos" (valor 0.0) não são diretamente afetados por esta função.

## F4: Interpretação da Saída

### F4.1: Neurônios de Saída Designados
*   A rede possui 10 neurônios cujos IDs estão armazenados na lista `OutputNeuronIDs`, designados especificamente como neurônios de saída.

### F4.2: Padrão de Resposta da Rede
*   A resposta da rede a um padrão de entrada é constituída pelo estado de ativação dos 10 neurônios de saída (identificados por `OutputNeuronIDs`).
*   A medida de ativação primária no MVP é o valor do `AccumulatedPulse` em cada um desses neurônios de saída, coletado após a rede executar um número definido de ciclos de "acomodação" (`cyclesToSettle`) desde a apresentação do input.
*   O objetivo do MVP é que, após o treinamento (exposição), esses vetores de ativação de 10 dimensões se tornem distintos e consistentes para diferentes dígitos de entrada.
*   **Nota:** O MVP foca na formação de representações internas distintas. A classificação explícita dessas representações para rótulos de dígitos (0-9) é uma etapa de análise/interpretação pós-MVP.

## F7: Saída do MVP e Observação

*   **Modo "expose":**
    *   Durante o processo de exposição/treinamento, o sistema pode fornecer atualizações periódicas no console, como:
        *   Número da época atual.
        *   Níveis atuais de Cortisol e Dopamina.
        *   Possivelmente, um resumo das mudanças nos pesos sinápticos (embora isso possa ser mais para depuração).
*   **Modo "observe":**
    *   Ao observar a resposta da rede a um dígito específico:
        *   O sistema exibe o padrão do dígito de entrada que foi apresentado.
        *   O sistema exibe o vetor de ativação resultante dos 10 neurônios de saída.
*   **Modo "sim":**
    *   Saídas gerais sobre o estado da simulação, como:
        *   Estatísticas de simulação (ex: ciclo atual).
        *   Frequência de disparo de neurônios monitorados (se configurado).
        *   Níveis de químicos.
*   **Depuração:** Flags de depuração (ex: `-debugChem`) podem habilitar impressões mais detalhadas sobre a produção e os níveis de neuroquímicos.
