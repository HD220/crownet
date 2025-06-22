# Relatório de Execução Simulada: Reconhecimento do Dígito "1"

Este relatório foi gerado programaticamente (simulando a execução do teste `TestGenerateDigitRecognitionReport`) em: {{CURRENT_DATETIME}}

## Objetivo:
Demonstrar um fluxo de treinamento e observação da rede CrowNet para o dígito "1" (representado pelo padrão `[1,1,1]`).

## Configuração da Simulação Base (SimParams):

```json
{
  "SpaceMaxDimension": 10,
  "BaseFiringThreshold": 1,
  "PulsePropagationSpeed": 1,
  "HebbianCoincidenceWindow": 2,
  "DopaminergicPercent": 0.1,
  "InhibitoryPercent": 0.2,
  "ExcitatoryRadiusFactor": 1,
  "DopaminergicRadiusFactor": 0.8,
  "InhibitoryRadiusFactor": 0.9,
  "MinInputNeurons": 3,
  "MinOutputNeurons": 1,
  "PatternHeight": 1,
  "PatternWidth": 3,
  "PatternSize": 3,
  "AccumulatedPulseDecayRate": 0,
  "AbsoluteRefractoryCycles": 2,
  "RelativeRefractoryCycles": 3,
  "SynaptogenesisInfluenceRadius": 2,
  "AttractionForceFactor": 0.01,
  "RepulsionForceFactor": 0.005,
  "DampeningFactor": 0.5,
  "MaxMovementPerCycle": 0.1,
  "CyclesPerSecond": 100,
  "OutputFrequencyWindowCycles": 50,
  "InitialSynapticWeightMin": 0.1,
  "InitialSynapticWeightMax": 0.5,
  "MaxSynapticWeight": 1,
  "HebbPositiveReinforceFactor": 0.1,
  "HebbNegativeReinforceFactor": 0.05,
  "CortisolProductionRate": 0.01,
  "CortisolDecayRate": 0.005,
  "DopamineProductionRate": 0.02,
  "DopamineDecayRate": 0.01,
  "CortisolInfluenceOnLR": -0.5,
  "DopamineInfluenceOnLR": 0.8,
  "CortisolInfluenceOnSynapto": -0.3,
  "DopamineInfluenceOnSynapto": 0.5,
  "FiringThresholdIncreaseOnDopa": -0.2,
  "FiringThresholdIncreaseOnCort": 0.3,
  "SynapticWeightDecayRate": 0,
  "CortisolProductionPerHit": 0.05,
  "CortisolMaxLevel": 1,
  "DopamineProductionPerEvent": 0.1,
  "DopamineMaxLevel": 1,
  "MinLearningRateFactor": 0.1
}
```

## Passo 1: Treinamento da Rede (Modo `expose`)

### Configuração CLI para `expose`:

```json
{
  "TotalNeurons": 5,
  "Cycles": 0,
  "DbPath": "",
  "SaveInterval": 0,
  "StimInputID": 0,
  "StimInputFreqHz": 0,
  "MonitorOutputID": 0,
  "DebugChem": false,
  "Mode": "expose",
  "Epochs": 1,
  "WeightsFile": "temp_report_weights.json",
  "Digit": 0,
  "BaseLearningRate": 0.1,
  "CyclesPerPattern": 2,
  "CyclesToSettle": 0,
  "Seed": 12345
}
```

**Mock `loadWeightsFn`:** Chamado para 'temp_report_weights.json', retornando 'não encontrado' (comportamento padrão para novo treino).

### Saída do Console para: Modo Expose

```text
CrowNet Inicializando...
Modo Selecionado: expose
Configuração Base: Neurônios=5, ArquivoDePesos='temp_report_weights.json'
  expose: Épocas=1, TaxaAprendizadoBase=0.1000, CiclosPorPadrão=2
Rede criada: 5 neurônios. IDs Input: [0 1 2]..., IDs Output: [3]...
Estado Inicial: Cortisol=0.000, Dopamina=0.000
arquivo de pesos temp_report_weights.json não encontrado (mock)
Época 1/1 iniciando...
**Mock `GetDigitPatternFn`:** Chamado para dígito 0, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 1, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 2, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 3, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 4, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 5, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 6, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 7, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 8, retornando padrão `[1,1,1]`.
**Mock `GetDigitPatternFn`:** Chamado para dígito 9, retornando padrão `[1,1,1]`.
Época 1/1 concluída. Processados 10 padrões. Cortisol: 0.000, Dopamina: 0.000, FatorLR Efetivo: 1.0000
Fase de exposição concluída.
**Mock `saveWeightsFn`:** Pesos capturados para o arquivo 'temp_report_weights.json'.
Pesos treinados salvos em temp_report_weights.json
```

**Resultado do Treinamento:** Concluído com sucesso (simulado).
Pesos "treinados" (capturados em memória - valores exatos dependeriam da inicialização aleatória e da dinâmica precisa do aprendizado hebbiano. Assumindo que os neurônios de input 0,1,2 se conectam ao neurônio de output 3, e o padrão [1,1,1] foi apresentado, os pesos w(0,3), w(1,3), w(2,3) teriam aumentado. Para este exemplo, vamos simular alguns pesos simples):
```json
{
  "0": {
    "3": 0.119
  },
  "1": {
    "3": 0.119
  },
  "2": {
    "3": 0.119
  }
}
```
*(Nota: Os pesos reais seriam mais complexos e dependeriam da inicialização aleatória e da dinâmica de aprendizado exata. O valor 0.119 é ilustrativo, representando um aumento a partir de um peso inicial devido ao aprendizado com LR=0.1, HebbFactor=1.0, por 2 ciclos, e decaimento zero. Ex: W_ini + LR*1*1*Factor + LR*1*1*Factor. Se W_ini=0.1, LR=0.1, Factor=1, 2 ciclos: 0.1 + 0.1*0.1 + 0.1*0.1 = 0.1 + 0.01 + 0.01 = 0.12. O valor exato dependeria da implementação precisa do ApplyHebbianUpdate e se o delta é por ciclo ou por apresentação de padrão.)*

## Passo 2: Observação do Dígito "1" (Modo `observe`)

### Configuração CLI para `observe`:

```json
{
  "TotalNeurons": 5,
  "Cycles": 0,
  "DbPath": "",
  "SaveInterval": 0,
  "StimInputID": 0,
  "StimInputFreqHz": 0,
  "MonitorOutputID": 0,
  "DebugChem": false,
  "Mode": "observe",
  "Epochs": 0,
  "WeightsFile": "temp_report_weights.json",
  "Digit": 1,
  "BaseLearningRate": 0,
  "CyclesPerPattern": 0,
  "CyclesToSettle": 1,
  "Seed": 12345
}
```

**Mock `loadWeightsFn`:** Chamado para 'temp_report_weights.json', retornando pesos "treinados" capturados.
**Mock `GetDigitPatternFn`:** Chamado para dígito 1, retornando padrão `[1,1,1]`.

### Saída do Console para: Modo Observe (Dígito 1)

```text
CrowNet Inicializando...
Modo Selecionado: observe
Configuração Base: Neurônios=5, ArquivoDePesos='temp_report_weights.json'
  observe: Dígito=1, CiclosParaAcomodar=1
Rede criada: 5 neurônios. IDs Input: [0 1 2]..., IDs Output: [3]...
Pesos existentes carregados de temp_report_weights.json
Estado Inicial: Cortisol=0.000, Dopamina=0.000

Observando Resposta da Rede para o dígito 1 (1 ciclos de acomodação)...
Dígito Apresentado: 1
Padrão de Ativação dos Neurônios de Saída (Potencial Acumulado):
  OutNeurônio[0] (ID 3): 0.3570
```
*(O ID do neurônio de output será o ID real atribuído pela rede, aqui simulado como 3. O valor 0.3570 é a soma dos pesos simulados 0.119 * 3, assumindo que o padrão de input [1,1,1] ativa todos os inputs com sinal 1.0 e não há decaimento de potencial em 1 ciclo de acomodação.)*

**Resultado da Observação:** Concluído. Verificar saída do console para ativação do neurônio de output.

**Ativação Esperada (Teórica) do Neurônio de Output 3:** 0.3570 (sem decaimento)

---
Fim do Relatório.
