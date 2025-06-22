# Arquitetura de Software - CrowNet MVP

Este documento descreve a arquitetura de software do CrowNet MVP, detalhando a organização dos pacotes, as principais estruturas de dados e os algoritmos centrais.

## 1. Visão Geral da Arquitetura

O CrowNet MVP é uma aplicação de linha de comando (CLI) desenvolvida em Go. A arquitetura não segue um padrão formal como MVC ou Arquitetura Hexagonal de forma estrita. Em vez disso, é organizada em **pacotes modulares**, cada um com responsabilidades bem definidas, promovendo separação de conceitos e coesão.

A interação principal flui do pacote `main` (que lida com a CLI) para o pacote `network` (que orquestra a simulação), que por sua vez utiliza outros pacotes como `neuron`, `datagen`, `storage`, etc.

## 2. Estrutura de Pacotes (Existente e Proposta para Reescrita)

A análise do código existente e o plano de reescrita sugerem a seguinte estrutura de pacotes:

*   **`main`** (Ponto de Entrada)
    *   Responsável por:
        *   Parsear os argumentos da linha de comando (flags).
        *   Instanciar e configurar a rede (`CrowNet`).
        *   Orquestrar os diferentes modos de operação (`expose`, `observe`, `sim`) chamando as funcionalidades apropriadas do pacote `network` (ou de um futuro pacote `cli`).
    *   Depende de: `network`, `datagen`, `storage` (e futuros `config`, `cli`).

*   **`config`** (Proposto para Reescrita)
    *   Responsável por:
        *   Definir estruturas para armazenar todos os parâmetros de configuração da simulação (limiares, taxas, fatores de modulação, etc.).
        *   Carregar configurações de arquivos ou variáveis de ambiente (embora o MVP atual use apenas flags).
        *   Disponibilizar estas configurações de forma global ou injetada nos componentes que precisam delas.
    *   Não deve ter muitas dependências.

*   **`common` / `types`** (Proposto para Reescrita)
    *   Responsável por:
        *   Definir tipos primitivos encapsulados para maior segurança de tipo e clareza (ex: `NeuronID`, `CycleCount`, `Coordinate`, `SynapticWeight`).
        *   Definir estruturas de dados muito genéricas usadas em múltiplos pacotes (ex: `Point` para coordenadas 16D).
    *   Não deve ter muitas dependências.

*   **`datagen`** (Geração de Dados)
    *   Responsável por:
        *   Fornecer os padrões de entrada (ex: os padrões 5x7 para os dígitos 0-9).
    *   Dependências: Nenhuma significativa além de pacotes padrão.

*   **`neuron`** (Lógica do Neurônio Individual)
    *   Responsável por:
        *   Definir a estrutura `Neuron` e seus atributos (ID, Posição, Tipo, Estado, AcumuladorDePulso, Limiares, etc.).
        *   Implementar a máquina de estados do neurônio (Repouso, Disparo, Refratário).
        *   Lógica de recebimento de pulso e disparo.
        *   Decaimento do potencial acumulado.
    *   Dependências: `common`/`types` (para `Point`, tipos encapsulados), `neuron/config` (para constantes como períodos refratários).

*   **`pulse`** (Proposto para Reescrita, ou parte de `network`)
    *   Responsável por:
        *   Definir a estrutura `Pulse` e `PulseList`.
        *   `PulseList` gerencia a coleção de pulsos ativos, sua propagação, interações com neurônios e geração de novos pulsos.
        *   `Pulse` individualmente rastreia seu movimento e alcance.
    *   Dependências: `common`/`types`, `neuron` (para `Point`), `synaptic`, `space`.

*   **`space`** (Proposto para Reescrita, ou parte de `utils`)
    *   Responsável por:
        *   Funções relacionadas ao espaço 16D, como cálculo de distância Euclidiana.
        *   Potencialmente, otimizações de busca de vizinhos (embora o MVP atual use iteração).
    *   Dependências: `common`/`types`.

*   **`synaptic`** (Proposto para Reescrita, ou parte de `network`)
    *   Responsável por:
        *   Definir a estrutura para `SynapticWeights` (mapa de mapas).
        *   Funções para inicializar, obter e definir pesos.
    *   Dependências: `common`/`types` (para `NeuronID`, `SynapticWeight`).

*   **`neurochemical`** (Proposto para Reescrita, ou parte de `network`)
    *   Responsável por:
        *   Definir estruturas para `Cortisol` e `Dopamine`.
        *   Lógica de produção, decaimento e cálculo dos efeitos combinados dos neuroquímicos nos limiares e taxas (aprendizado, sinaptogênese).
    *   Dependências: `neuron` (para interagir com limiares), `network/config`.

*   **`network`** (Orquestração da Rede e Simulação)
    *   Responsável por:
        *   Definir a estrutura `CrowNet` que agrega todos os componentes da rede (neurônios, pulsos, pesos, químicos, configuração).
        *   Implementar o ciclo principal de simulação (`RunCycle`), orquestrando as atualizações de estado, propagação de pulso, aprendizado, sinaptogênese e modulação química.
        *   Lógica de aprendizado Hebbiano (`ApplyHebbianPlasticity`).
        *   Lógica de sinaptogênese (`applySynaptogenesis`).
        *   Gerenciamento de entrada/saída de padrões (`PresentPattern`, `GetOutputPatternForInput`).
        *   Interação com o sistema de persistência de pesos.
    *   Dependências: `neuron`, `datagen`, `storage`, `utils` (e os pacotes propostos como `pulse`, `synaptic`, `neurochemical`, `config`).

*   **`storage`** (Persistência de Dados)
    *   Responsável por:
        *   Salvar e carregar pesos sinápticos em formato JSON.
        *   Salvar o estado da rede em um banco de dados SQLite (opcional).
    *   Dependências: `network` (para obter o estado da rede), `os`, `encoding/json`, `database/sql`.

*   **`utils`** (Utilitários)
    *   Responsável por:
        *   Funções genéricas que podem ser usadas por múltiplos pacotes (ex: cálculo de distância Euclidiana, se não estiver em `space`).
    *   Não deve ter muitas dependências específicas do domínio.

*   **`cli`** (Proposto para Reescrita, para encapsular lógica de `main`)
    *   Responsável por:
        *   Toda a lógica de interação com a linha de comando que atualmente reside em `main.go`.
        *   Parsing de flags, configuração inicial da simulação com base nos flags, e invocação dos modos de operação.
    *   Dependências: `network`, `config`, `datagen`, `storage`.

## 3. Principais Estruturas de Dados

*   **`neuron.Point [16]float64`**: Representa uma coordenada no espaço 16D.
*   **`neuron.Neuron`**:
    *   `ID int`
    *   `Position neuron.Point`
    *   `Type neuron.NeuronType` (Enum: Excitatory, Inhibitory, Dopaminergic, Input, Output)
    *   `State neuron.NeuronState` (Enum: Resting, Firing, AbsoluteRefractory, RelativeRefractory)
    *   `AccumulatedPulse float64`
    *   `BaseFiringThreshold float64`
    *   `CurrentFiringThreshold float64`
    *   `LastFiredCycle int`
    *   `CyclesInCurrentState int`
    *   `Velocity neuron.Point`
*   **`network.Pulse`** (ou `pulse.Pulse`):
    *   `EmittingNeuronID int`
    *   `OriginPosition neuron.Point`
    *   `Value float64` (sinal base: +1.0 ou -1.0)
    *   `CreationCycle int`
    *   `CurrentDistance float64`
    *   `MaxTravelDistance float64`
*   **`pulse.PulseList`**:
    *   `pulses []*pulse.Pulse`
    *   Responsável por gerenciar a coleção de pulsos ativos, processar seu ciclo de vida (propagação, interação, remoção) e facilitar a criação de novos pulsos.
*   **`network.CrowNet`**:
    *   `Neurons []*neuron.Neuron`
    *   `ActivePulses *pulse.PulseList` // Anteriormente []*network.Pulse
    *   `InputNeuronIDs []int`, `OutputNeuronIDs []int`
    *   `CortisolLevel float64`, `DopamineLevel float64`
    *   `CycleCount int`
    *   `SynapticWeights map[int]map[int]float64` (De `NeuronID` para `NeuronID` para `peso`)
    *   `BaseLearningRate float64`
    *   Flags de configuração (`EnableSynaptogenesis`, `EnableChemicalModulation`, etc.)
    *   Estruturas de dados para I/O (mapas de frequência de input, histórico de output).

## 4. Principais Algoritmos

*   **Ciclo de Simulação (`CrowNet.RunCycle`)**:
    1.  Processar inputs de frequência (para modo `sim`).
    2.  Atualizar estados dos neurônios (decaimento de potencial, transições de estado refratário).
    3.  Propagar pulsos ativos:
        *   Atualizar distância.
        *   Identificar neurônios atingidos na "casca" de efeito do pulso.
        *   Para cada neurônio atingido, aplicar `ValorBasePulso * PesoSinaptico` ao `AccumulatedPulse`.
        *   Se neurônio atingido disparar, criar novo(s) pulso(s).
    4.  Atualizar níveis de Cortisol e Dopamina (produção e decaimento).
    5.  Aplicar efeitos do Cortisol e Dopamina (nos limiares de disparo e no fator de modulação da sinaptogênese).
    6.  Aplicar Plasticidade Hebbiana Neuromodulada (`ApplyHebbianPlasticity`).
    7.  Aplicar Sinaptogênese (`applySynaptogenesis`).
    8.  Incrementar `CycleCount`.

*   **Plasticidade Hebbiana Neuromodulada (`CrowNet.ApplyHebbianPlasticity`)**:
    1.  Calcular taxa de aprendizado efetiva: `BaseLearningRate * FatorModulacao(Dopamina, Cortisol)`.
    2.  Para cada conexão sináptica:
        *   Determinar atividade pré-sináptica e pós-sináptica (se dispararam dentro da `HebbianCoincidenceWindow`).
        *   Se co-ativos, calcular `ΔPeso = TaxaEfetiva * AtividadePre * AtividadePos`.
        *   Atualizar peso: `NovoPeso = PesoAntigo + ΔPeso`.
        *   Aplicar decaimento de peso: `NovoPeso = NovoPeso - NovoPeso * TaxaDecaimentoPeso`.
        *   Aplicar limites (Min/Max) ao peso.

*   **Sinaptogênese (`CrowNet.applySynaptogenesis`)**:
    1.  Para cada neurônio `n1`:
        *   Calcular vetor de força total de outros neurônios `n2` (atração por ativos, repulsão por em repouso), modulado pelo fator químico.
        *   Atualizar velocidade de `n1` (com amortecimento).
        *   Limitar magnitude da velocidade.
        *   Atualizar posição de `n1` com base na nova velocidade.
        *   Aplicar condições de contorno (manter dentro do espaço).

*   **Modulação Química (`updateCortisolLevel`, `applyCortisolEffects`, etc.)**:
    *   Produção baseada em eventos (hits na glândula, disparos de neurônios dopaminérgicos).
    *   Decaimento percentual por ciclo.
    *   Cálculo de fatores de modulação para limiares, aprendizado e sinaptogênese com base nos níveis atuais e parâmetros de efeito (limiares de efeito, fatores de aumento/redução).

## 5. Padrões de Projeto Identificados/Sugeridos

*   **Configuração Centralizada / Injeção de Dependência (para Reescrita):** Usar um pacote `config` para centralizar parâmetros e injetá-los onde necessário, em vez de constantes globais espalhadas.
*   **Máquina de Estados:** O comportamento do `neuron.Neuron` (Resting, Firing, Refractory) é um exemplo claro do padrão State.
*   **Strategy (Implícito):** Os diferentes modos de operação (`expose`, `observe`, `sim`) podem ser vistos como diferentes estratégias de execução da simulação, orquestradas em `main` (ou futuro `cli`).
*   **Tipos Primitivos Encapsulados (Value Object):** Para a reescrita, usar tipos como `NeuronID`, `CycleCount` em vez de `int` puro para melhorar a semântica e segurança de tipo.
*   **Coleções de Primeira Classe (para Reescrita):** Estruturas como `synaptic.NetworkWeights` (gerenciando pesos sinápticos) e `pulse.PulseList` (gerenciando a coleção de pulsos ativos e seu processamento) são exemplos deste padrão. A reescrita pode formalizar outras coleções, como `type NeuronCollection []*Neuron`.
*   **Builder (Potencial):** Para a inicialização complexa de `CrowNet`, um padrão Builder poderia ser considerado na reescrita para torná-la mais fluida, embora a função `NewCrowNet` atual já lide com isso.

A arquitetura atual é funcional para o MVP. A reescrita focará em refinar a modularidade, aplicar princípios de Código Limpo e Object Calisthenics, e formalizar alguns desses padrões para melhorar a manutenibilidade e testabilidade.
```
