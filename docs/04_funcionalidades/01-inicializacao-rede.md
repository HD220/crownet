# Funcionalidade: Inicialização da Rede Neural CrowNet

## 1. Visão Geral

Esta funcionalidade descreve o processo pelo qual a rede neural CrowNet é configurada e inicializada antes do início de qualquer simulação ou processo de aprendizado. Uma inicialização correta é crucial para o funcionamento adequado da rede e para a reprodutibilidade dos experimentos.

O processo envolve a definição do número e tipos de neurônios, seu posicionamento no espaço simulado, a atribuição de suas propriedades intrínsecas e o estabelecimento das conexões sinápticas iniciais.

## 2. Configuração da População Neuronal

### 2.1. Número Total de Neurônios
*   O sistema permite que o usuário especifique o número total de neurônios que comporão a rede. Esta configuração é tipicamente fornecida através de um parâmetro na interface de linha de comando (CLI).

### 2.2. Tipos e Distribuição de Neurônios
*   A rede é composta por diferentes tipos de neurônios, cada um com papéis específicos:
    *   **Neurônios de Input (Entrada):** Um número fixo de neurônios (atualmente 35) é designado para receber os padrões de entrada externos (ex: representações de dígitos).
    *   **Neurônios de Output (Saída):** Um número fixo de neurônios (atualmente 10) é designado para representar a resposta da rede.
    *   **Neurônios Internos:** Os neurônios restantes são distribuídos entre os seguintes tipos, com base em proporções configuráveis ou padrões do sistema:
        *   **Excitatory (Excitatórios):** Propagam sinais que tendem a ativar outros neurônios.
        *   **Inhibitory (Inibitórios):** Propagam sinais que tendem a suprimir a atividade de outros neurônios.
        *   **Dopaminergic (Dopaminérgicos):** Modulam a atividade da rede e processos de aprendizado através da simulação de dopamina.

## 3. Disposição Espacial dos Neurônios

### 3.1. Espaço de Simulação
*   Todos os neurônios existem e interagem dentro de um espaço vetorial de 16 dimensões.

### 3.2. Posicionamento Inicial
*   A posição inicial de cada neurônio no espaço 16D é determinada aleatoriamente.
*   Para garantir uma distribuição espacial inicial organizada e evitar agrupamentos indesejados, o posicionamento aleatório pode ser sujeito a restrições radiais ou outras heurísticas específicas para cada tipo de neurônio.

## 4. Atribuição de Propriedades Neuronais

Cada neurônio na rede é inicializado com um conjunto de propriedades fundamentais:
*   **Identificador Único (ID):** Um ID numérico exclusivo que distingue o neurônio dos demais.
*   **Tipo Neuronal:** Uma designação do seu tipo funcional (Input, Output, Excitatory, Inhibitory, Dopaminergic).
*   **Estado Inicial:** O estado operacional inicial do neurônio, que é tipicamente "Repouso" (Resting).
*   **Limiar de Disparo Base:** Um valor base que o potencial acumulado do neurônio deve exceder para que ele dispare.
*   **Acumulador de Pulso:** Inicializado em zero, representa o potencial elétrico acumulado pelo neurônio.

## 5. Estabelecimento de Conexões Sinápticas Iniciais

As conexões entre os neurônios (sinapses) e a força dessas conexões (pesos sinápticos) são cruciais para o processamento de informação na rede.

### 5.1. Estrutura de Conectividade
*   O sistema estabelece uma matriz de pesos sinápticos que representa as conexões de um neurônio de origem para um neurônio de destino.
*   Inicialmente, a rede é configurada com uma conectividade potencial "all-to-all", o que significa que qualquer neurônio pode, a princípio, conectar-se a qualquer outro neurônio.

### 5.2. Valores Iniciais dos Pesos
*   Os valores iniciais dos pesos sinápticos são definidos como números pequenos e aleatórios, podendo ser positivos (excitatórios) ou negativos (inibitórios), dentro de uma faixa predefinida.
*   **Auto-conexões:** Conexões de um neurônio para ele mesmo são explicitamente proibidas ou inicializadas com peso zero.
*   **Considerações Específicas:**
    *   Embora a estrutura de dados permita conexões para neurônios de Input, a lógica de aprendizado e propagação de sinal do MVP geralmente não modifica ativamente os pesos que chegam aos neurônios de Input. Eles são primariamente fontes de sinal.

## 6. Resultado da Inicialização

Ao final do processo de inicialização, a rede CrowNet está pronta para a simulação. Ela consiste em uma população de neurônios devidamente configurados, posicionados e interconectados, com um estado inicial definido, preparada para processar entradas, aprender e evoluir ao longo dos ciclos de simulação.
