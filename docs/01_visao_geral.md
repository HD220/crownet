# Visão Geral do Projeto CrowNet

## 1. O que é o CrowNet?

**CrowNet** é uma aplicação de linha de comando (CLI) desenvolvida em Go que simula um modelo computacional de rede neural bio-inspirada. Ele se propõe a modelar e simular o comportamento de neurônios que interagem em um espaço vetorial de alta dimensionalidade (16D), incorporando conceitos como sinaptogênese (movimento de neurônios dependente da atividade) e neuromodulação através de substâncias químicas simuladas como cortisol e dopamina.

## 2. Problema que Resolve

O principal problema que o CrowNet MVP (Minimum Viable Product) visa resolver é demonstrar e investigar a **autoaprendizagem Hebbiana neuromodulada**. O objetivo é mostrar que a rede neural simulada pode, através da interação de seus componentes e das regras de aprendizado implementadas, auto-organizar seus pesos sinápticos. Isso resulta na formação de representações internas distintas para diferentes padrões de entrada (atualmente, padrões de dígitos de 0 a 9), onde o processo de aprendizado (plasticidade sináptica) é dinamicamente influenciado pelo ambiente neuroquímico simulado.

## 3. Escopo Principal do MVP

O escopo do atual MVP do CrowNet inclui:

*   **Interface de Linha de Comando (CLI):** A interação com o simulador é feita exclusivamente através de comandos e flags no terminal.
*   **Simulação Espacial 16D:** Neurônios existem e podem se mover dentro de um espaço vetorial de 16 dimensões.
*   **Tipos Neuronais Diversificados:** Implementação de neurônios excitatórios, inibitórios, dopaminérgicos, de entrada (input) e de saída (output).
*   **Propagação de Pulso Simplificada:** Modelo de propagação de sinal neural baseado em expansão esférica.
*   **Conexões Sinápticas Ponderadas:** Relações explícitas entre neurônios com pesos que modulam a força da conexão.
*   **Aprendizado Hebbiano Neuromodulado:**
    *   **Plasticidade Hebbiana:** Ajuste de pesos sinápticos com base na co-ativação de neurônios pré e pós-sinápticos.
    *   **Neuromodulação:** A taxa de aprendizado e os limiares de disparo dos neurônios são modulados pelos níveis simulados de Dopamina (que tende a aumentar a plasticidade) e Cortisol (que, em altos níveis, pode suprimir a plasticidade).
*   **Sinaptogênese Rudimentar:** Movimento dos neurônios influenciado pela atividade da rede e modulado por fatores neuroquímicos.
*   **Codificação de Entrada Específica:** Utilização de padrões binários (formato 5x7 pixels) para representar dígitos de 0 a 9 como estímulo para a rede.
*   **Representação de Saída Observável:** A rede deve produzir padrões de ativação distintos em um conjunto dedicado de neurônios de saída, correspondentes aos diferentes dígitos de entrada aprendidos.
*   **Modos de Operação:**
    *   `expose`: Para treinar a rede com os padrões de dígitos.
    *   `observe`: Para testar uma rede previamente treinada.
    *   `sim`: Para simulações mais genéricas e observação de dinâmicas.

## 4. Tecnologia Chave

As principais tecnologias e formatos utilizados no CrowNet MVP são:

*   **Linguagem de Programação:** Go (Golang).
*   **Persistência de Pesos:** Arquivos JSON são utilizados para salvar e carregar os pesos sinápticos aprendidos pela rede.
*   **Logging Detalhado (Opcional):** SQLite é usado para registrar snapshots detalhados do estado da simulação para análises posteriores, se ativado.
*   **Gerenciamento de Dependências:** Go Modules.
