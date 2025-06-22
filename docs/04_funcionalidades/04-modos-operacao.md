# Funcionalidade: Modos de Operação da CLI CrowNet

## 1. Visão Geral

A aplicação de linha de comando (CLI) CrowNet opera através de diferentes modos, cada um projetado para uma finalidade específica na simulação e análise da rede neural. Estes modos são controlados principalmente pelo flag `-mode` na CLI.

Este documento fornece um sumário de cada modo, suas configurações chave e um link para o caso de uso detalhado correspondente.

## 2. Flags Globais Comuns

Alguns flags são comuns e podem ser aplicáveis a múltiplos modos de operação:

*   `-neurons <int>`: Especifica o número total de neurônios na rede.
*   `-weightsFile <string>`: Define o caminho para o arquivo JSON usado para salvar ou carregar os pesos sinápticos da rede.
*   `-dbPath <string>`: (Opcional) Caminho para um arquivo SQLite onde snapshots detalhados do estado da simulação podem ser registrados.
*   `-saveInterval <int>`: (Relevante se `-dbPath` for usado) Intervalo de ciclos para salvar o estado da rede no arquivo SQLite.
*   `-debugChem <bool>`: (Opcional) Habilita logs de depuração adicionais relacionados aos neuroquímicos.
*   `-cycles <int>`: (Uso varia por modo) Número geral de ciclos para uma simulação. Note que modos específicos como `expose` e `observe` usam flags de ciclo mais granulares.

## 3. Modo `expose`

*   **Propósito:** Expor a rede a sequências de padrões de dígitos (0-9) repetidamente. Este é o modo primário de "treinamento", permitindo que a rede auto-organize seus pesos sinápticos através da plasticidade Hebbiana neuromodulada.
*   **Dinâmicas Ativas:**
    *   Plasticidade Hebbiana (atualização de pesos).
    *   Modulação Química (produção, decaimento e efeitos de cortisol/dopamina).
    *   Sinaptogênese (movimento de neurônios).
*   **Flags Específicas Principais:**
    *   `-mode expose`
    *   `-epochs <int>`: Número de vezes que o conjunto completo de padrões de dígitos será apresentado.
    *   `-lrBase <float>`: Taxa de aprendizado base para a regra Hebbiana.
    *   `-cyclesPerPattern <int>`: Número de ciclos de simulação para cada padrão de dígito apresentado.
*   **Detalhes do Fluxo:** Para uma descrição completa do processo, interações e tratamento de erros, consulte o caso de uso:
    *   [**UC-EXPOSE: Expor Rede a Padrões para Auto-Aprendizagem](./casos-de-uso/uc-expose.md)**

## 4. Modo `observe`

*   **Propósito:** Carregar um conjunto de pesos sinápticos previamente aprendidos e apresentar um dígito específico à rede. O objetivo é observar o padrão de ativação resultante nos neurônios de saída, permitindo avaliar o que a rede aprendeu.
*   **Dinâmicas Ativas (Modificadas para Observação):**
    *   Plasticidade Hebbiana: Desativada (taxa de aprendizado zero).
    *   Sinaptogênese: Desativada.
    *   Modulação Química: Desativada (limiares neuronais em valores base, sem produção/efeito de químicos).
    *   *Esta configuração garante uma observação "limpa" da resposta da rede aos pesos fixos.*
*   **Flags Específicas Principais:**
    *   `-mode observe`
    *   `-digit <0-9>`: O dígito específico a ser apresentado.
    *   `-cyclesToSettle <int>`: Número de ciclos para permitir que a atividade da rede se estabilize antes de ler a saída.
*   **Detalhes do Fluxo:** Para uma descrição completa do processo, interações e tratamento de erros, consulte o caso de uso:
    *   [**UC-OBSERVE: Observar Resposta da Rede a um Padrão](./casos-de-uso/uc-observe.md)**

## 5. Modo `sim`

*   **Propósito:** Executar uma simulação geral da rede CrowNet com todas as suas dinâmicas intrínsecas ativas. Este modo é útil para observar comportamentos emergentes, testar configurações específicas sob estímulo contínuo, ou para logging detalhado e análise da evolução da rede ao longo do tempo.
*   **Dinâmicas Ativas:** Todas as dinâmicas estão normalmente ativas:
    *   Plasticidade Hebbiana.
    *   Modulação Química.
    *   Sinaptogênese.
*   **Flags Específicas Principais:**
    *   `-mode sim`
    *   `-cycles <int>`: (Usado aqui) Número total de ciclos para a simulação.
    *   `-stimInputID <int>`: (Opcional) ID de um neurônio de entrada para receber estímulo contínuo.
    *   `-stimInputFreqHz <float>`: (Relevante se `-stimInputID` usado) Frequência do estímulo.
    *   `-monitorOutputID <int>`: (Opcional) ID de um neurônio de saída para monitorar sua frequência de disparo.
*   **Detalhes do Fluxo:** Para uma descrição completa do processo, interações e tratamento de erros, consulte o caso de uso:
    *   [**UC-SIM: Executar Simulação Geral da Rede](./casos-de-uso/uc-sim.md)**

A combinação destes modos permite um ciclo completo de treinamento, teste e análise da rede neural CrowNet.
