# CrowNet: Simulador de Rede Neural Bio-inspirada (MVP)

**CrowNet** é uma aplicação de linha de comando escrita em Go que simula um modelo computacional de rede neural. Inspirada em processos biológicos, a simulação apresenta neurônios interagindo em um espaço vetorial de 16 dimensões, sinaptogênese (movimento de neurônios dependente da atividade) e neuromodulação por cortisol e dopamina simulados.

O **Minimum Viable Product (MVP)** atual foca em demonstrar a **autoaprendizagem Hebbiana neuromodulada**. A rede é exposta a padrões simples de dígitos (0-9) e visa auto-organizar seus pesos sinápticos para formar representações internas distintas para esses diferentes inputs. O processo de aprendizado (plasticidade) é influenciado pelo ambiente químico simulado.

## Visão Geral do MVP

*   **Objetivo Principal:** Demonstrar que o modelo CrowNet pode exibir comportamento de auto-organização através de plasticidade Hebbiana neuromodulada, onde a rede aprende a formar representações internas distintas para diferentes padrões de entrada (dígitos 0-9).
*   **Interface:** Linha de Comando (CLI).
*   **Linguagem:** Go.

## Principais Conceitos Implementados no MVP

*   **Espaço 16D:** Neurônios existem e se movem em um espaço vetorial de 16 dimensões.
*   **Tipos de Neurônios:** Excitatórios, Inibitórios, Dopaminérgicos, Input e Output.
*   **Propagação de Pulso:** Modelo simplificado de expansão esférica.
*   **Pesos Sinápticos:** Conexões explícitas com pesos que determinam a força da influência entre neurônios.
*   **Aprendizado Hebbiano Neuromodulado:**
    *   **Plasticidade Hebbiana:** Pesos ajustados com base na co-ativação de neurônios.
    *   **Neuromodulação:** Taxa de aprendizado e limiares de disparo modulados por níveis de Dopamina (aumenta plasticidade) e Cortisol (altos níveis suprimem plasticidade).
*   **Sinaptogênese:** Movimento de neurônios influenciado pela atividade da rede e modulado por químicos.
*   **Codificação de Entrada:** Padrões binários 5x7 para dígitos 0-9.
*   **Representação de Saída:** Padrões de ativação distintos nos 10 neurônios de output.

## Modos de Operação (CLI)

A aplicação suporta três modos principais, controlados pelo flag `-mode`:

1.  **`expose`**: Para treinar a rede, apresentando padrões de dígitos repetidamente, permitindo que o aprendizado Hebbiano ocorra.
    *   Flags chave: `-epochs`, `-lrBase`, `-cyclesPerPattern`, `-weightsFile`.
2.  **`observe`**: Para testar uma rede treinada, carregando pesos salvos e observando o padrão de saída da rede para um dígito específico.
    *   Flags chave: `-digit <0-9>`, `-weightsFile`, `-cyclesToSettle`.
3.  **`sim`**: Para rodar uma simulação geral com todas as dinâmicas ativas, útil para observação de comportamento ou logging detalhado.
    *   Flags chave: `-cycles`, `-stimInputID`, `-stimInputFreqHz`, `-dbPath`, `-saveInterval`.

## Tecnologias Utilizadas (MVP)

*   **Go:** Linguagem de implementação.
*   **JSON:** Para salvar e carregar os pesos sinápticos aprendidos.
*   **SQLite:** (Opcional) Para salvar snapshots detalhados do estado da simulação para análise.

## Documentação Detalhada

Para uma compreensão completa das funcionalidades, arquitetura técnica, requisitos e casos de uso do CrowNet MVP, por favor, consulte os documentos localizados no diretório `/docs`:

*   **`/docs/funcional/`**: Descrições detalhadas de cada funcionalidade do sistema.
    *   `01-inicializacao-rede.md`
    *   `02-ciclo-simulacao-aprendizado.md`
    *   `03-entrada-saida-dados.md`
    *   `04-modos-operacao.md`
    *   `05-persistencia-dados.md`
*   **`/docs/tecnico/`**: Documentação técnica, de estilo e arquitetural.
    *   `guia_interface_linha_comando.md`: Detalhes sobre a CLI, flags e formatos de saída.
    *   `arquitetura.md`: Visão geral da arquitetura de software, pacotes e algoritmos.
    *   `requisitos.md`: Requisitos Funcionais e Não Funcionais do MVP.
    *   `casos-de-uso/`: Descrições detalhadas dos cenários de uso para cada modo de operação.

## Como Construir e Executar (Exemplo)

1.  **Construir:**
    ```bash
    go build .
    ```
2.  **Executar (exemplo modo expose):**
    ```bash
    ./crownet -mode expose -neurons 150 -epochs 20 -lrBase 0.005 -cyclesPerPattern 5 -weightsFile my_digit_weights.json
    ```
3.  **Executar (exemplo modo observe):**
    ```bash
    ./crownet -mode observe -digit 7 -weightsFile my_digit_weights.json -cyclesToSettle 5
    ```

Consulte o `guia_interface_linha_comando.md` para mais detalhes sobre os flags.
```
