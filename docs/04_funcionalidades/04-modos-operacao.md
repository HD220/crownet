# Funcionalidade: Modos de Operação da CLI CrowNet

**NOTA IMPORTANTE: A interface de linha de comando (CLI) foi refatorada para usar subcomandos (ex: `crownet sim`, `crownet expose`) em vez da flag `-mode`. Esta página precisa de uma atualização completa para refletir a nova estrutura de comandos e flags detalhada em `docs/03_guias/guia_interface_linha_comando.md`. As descrições de propósito e dinâmicas dos modos abaixo ainda são relevantes, mas os exemplos de flags estão desatualizados.**

## 1. Visão Geral

A aplicação de linha de comando (CLI) CrowNet opera através de diferentes modos (agora subcomandos), cada um projetado para uma finalidade específica na simulação e análise da rede neural.

Este documento fornece um sumário de cada modo/comando, suas configurações chave e um link para o caso de uso detalhado correspondente. Consulte `docs/03_guias/guia_interface_linha_comando.md` para a sintaxe atualizada dos comandos e flags.

## 2. Flags Globais Comuns (Agora Flags Persistentes do Comando Raiz)

Algumas flags são aplicáveis globalmente:
*   `--configFile <string>`: Caminho para arquivo de configuração TOML.
*   `--seed <int64>`: Semente para aleatoriedade.

Outras flags que eram "globais" (ex: `--neurons`, `--weightsFile`, `--dbPath`, `--lrBase`) agora são específicas para os subcomandos de simulação (`sim`, `expose`, `observe`).

## 3. Comando `expose` (Anteriormente Modo `expose`)

*   **Propósito:** Expor a rede a sequências de padrões de dígitos (0-9) repetidamente. Este é o comando primário de "treinamento".
*   **Dinâmicas Ativas:** Plasticidade Hebbiana, Modulação Química, Sinaptogênese.
*   **Exemplo de Uso (Novo):** `./crownet expose --epochs 50 --lrBase 0.01 --weightsFile pesos.json`
*   **Flags Chave (Consulte `crownet expose --help` para lista completa e atualizada):**
    *   `--epochs`, `--lrBase`, `--cyclesPerPattern`, `--weightsFile`, `--neurons`, etc.
*   **Detalhes do Fluxo:** [**UC-EXPOSE: Expor Rede a Padrões para Auto-Aprendizagem](./casos-de-uso/uc-expose.md)**

## 4. Comando `observe` (Anteriormente Modo `observe`)

*   **Propósito:** Carregar pesos sinápticos previamente aprendidos e apresentar um dígito específico à rede.
*   **Dinâmicas Ativas (Modificadas para Observação):** Aprendizado, sinaptogênese e modulação química desativados.
*   **Exemplo de Uso (Novo):** `./crownet observe --digit 7 --weightsFile pesos.json`
*   **Flags Chave (Consulte `crownet observe --help`):**
    *   `--digit`, `--cyclesToSettle`, `--weightsFile`, `--neurons`, etc.
*   **Detalhes do Fluxo:** [**UC-OBSERVE: Observar Resposta da Rede a um Padrão](./casos-de-uso/uc-observe.md)**

## 5. Comando `sim` (Anteriormente Modo `sim`)

*   **Propósito:** Executar uma simulação geral da rede CrowNet com todas as suas dinâmicas intrínsecas ativas.
*   **Dinâmicas Ativas:** Plasticidade Hebbiana, Modulação Química, Sinaptogênese.
*   **Exemplo de Uso (Novo):** `./crownet sim --cycles 1000 --dbPath sim.db`
*   **Flags Chave (Consulte `crownet sim --help`):**
    *   `--cycles`, `--stimInputID`, `--stimInputFreqHz`, `--monitorOutputID`, `--dbPath`, `--neurons`, etc.
*   **Detalhes do Fluxo:** [**UC-SIM: Executar Simulação Geral da Rede](./casos-de-uso/uc-sim.md)**

## 6. Comando `logutil` (Novo)
*   **Propósito:** Fornecer utilitários para interagir com os logs SQLite.
*   **Subcomando `export`:** Exporta tabelas do log para CSV.
*   **Exemplo de Uso (Novo):** `./crownet logutil export --dbPath sim.db --table NetworkSnapshots`
*   **Flags Chave (Consulte `crownet logutil export --help`):**
    *   `--dbPath`, `--table`, `--format`, `--output`.

A combinação destes comandos permite um ciclo completo de treinamento, teste, análise e exportação de dados da rede neural CrowNet.
