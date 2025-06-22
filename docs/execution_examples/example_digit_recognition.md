# Exemplo Prático: Treinamento e Observação para Reconhecimento de Dígitos

Este documento descreve um exemplo prático de como usar o CrowNet para treinar uma rede neural para reconhecer padrões de dígitos e, em seguida, observar sua resposta a esses padrões.

## 1. Objetivo do Exemplo

Demonstrar um fluxo de trabalho completo:
1.  Treinar uma rede neural com os padrões de dígitos 0-9 (modo `expose`).
2.  Salvar os pesos sinápticos aprendidos.
3.  Carregar os pesos treinados e apresentar um dígito específico (ex: "0") para observar a ativação dos neurônios de saída (modo `observe`).
4.  Apresentar um dígito diferente (ex: "1") para comparar a ativação e verificar a seletividade da rede.

## 2. Configuração Geral (Suposições)

Para este exemplo, vamos assumir os seguintes parâmetros CLI comuns onde aplicável (os valores exatos podem ser ajustados):
*   `-neurons 70`: Uma rede com um número razoável de neurônios (ex: 35 para input 7x5, 10 para output 0-9, e 25 internos).
*   `-seed 123`: Para reprodutibilidade.

Os parâmetros de simulação (`SimulationParameters`) serão os definidos em `config.DefaultSimulationParameters()`, a menos que especificado de outra forma (ex: `BaseLearningRate` pode ser ajustado).

## 3. Passo 1: Treinamento da Rede (Modo `expose`)

### Comando CLI:
```bash
./crownet -mode expose -neurons 70 -epochs 15 -cyclesPerPattern 10 -lrBase 0.01 -weightsFile digit_example_weights.json -seed 123
```
*   `-mode expose`: Especifica o modo de treinamento.
*   `-neurons 70`: Define o tamanho da rede.
*   `-epochs 15`: A rede verá o conjunto completo de dígitos 15 vezes.
*   `-cyclesPerPattern 10`: Para cada dígito apresentado, a rede simulará por 10 ciclos para permitir aprendizado.
*   `-lrBase 0.01`: Taxa de aprendizado base.
*   `-weightsFile digit_example_weights.json`: Os pesos aprendidos serão salvos neste arquivo.
*   `-seed 123`: Para resultados reproduzíveis.

### Saída do Console Esperada (Simulada - Resumida):
```
CrowNet Inicializando...
Modo Selecionado: expose
Configuração Base: Neurônios=70, ArquivoDePesos='digit_example_weights.json'
  expose: Épocas=15, TaxaAprendizadoBase=0.0100, CiclosPorPadrão=10
Rede criada: 70 neurônios. IDs Input: [0 1 2 3 4]..., IDs Output: [35 36 37 38 39 40 41 42 43 44]...
Estado Inicial: Cortisol=0.000, Dopamina=0.000
Época 1/15 iniciando...
Época 1/15 concluída. Processados 10 padrões. Cortisol: X.XXX, Dopamina: Y.YYY, FatorLR Efetivo: Z.ZZZ
Época 2/15 iniciando...
...
Época 15/15 iniciando...
Época 15/15 concluída. Processados 10 padrões. Cortisol: A.AAA, Dopamina: B.BBB, FatorLR Efetivo: C.CCC
Fase de exposição concluída.
Pesos treinados salvos em digit_example_weights.json

Sessão CrowNet finalizada. Duração total: HH:MM:SS.
```
*(X, Y, Z, A, B, C são valores numéricos que dependeriam da dinâmica exata da simulação)*

### Arquivo de Pesos Gerado (`digit_example_weights.json`):
*   **Descrição:** Um arquivo JSON contendo os pesos sinápticos da rede. A estrutura interna seria um mapa de IDs de neurônios de origem para outro mapa de IDs de neurônios de destino e seus respectivos pesos.
*   **Exemplo de Trecho (Estrutural, não valores reais):**
    ```json
    {
      "0": { // ID do Neurônio Pré-sináptico
        "35": 0.634, // ID Pós-sináptico: Peso
        "36": 0.128,
        // ... outros neurônios pós-sinápticos
      },
      "1": {
        "35": 0.050,
        "37": 0.871,
        // ...
      }
      // ... outros neurônios pré-sinápticos
    }
    ```

## 4. Passo 2: Observação do Dígito "0" (Modo `observe`)

### Comando CLI:
```bash
./crownet -mode observe -neurons 70 -digit 0 -weightsFile digit_example_weights.json -cyclesToSettle 20 -seed 123
```
*   `-mode observe`: Especifica o modo de observação.
*   `-digit 0`: Apresenta o padrão do dígito "0".
*   `-weightsFile digit_example_weights.json`: Carrega os pesos treinados na Etapa 1.
*   `-cyclesToSettle 20`: Permite que a rede processe o padrão por 20 ciclos.

### Saída do Console Esperada (Simulada - Foco na Ativação):
```
CrowNet Inicializando...
Modo Selecionado: observe
Configuração Base: Neurônios=70, ArquivoDePesos='digit_example_weights.json'
  observe: Dígito=0, CiclosParaAcomodar=20
Rede criada: 70 neurônios. IDs Input: [...]..., IDs Output: [N0 N1 N2 N3 N4 N5 N6 N7 N8 N9]...
Pesos existentes carregados de digit_example_weights.json
Estado Inicial: Cortisol=0.000, Dopamina=0.000

Observando Resposta da Rede para o dígito 0 (20 ciclos de acomodação)...
Dígito Apresentado: 0
Padrão de Ativação dos Neurônios de Saída (Potencial Acumulado):
  OutNeurônio[0] (ID N0): 1.8754  // <-- Espera-se alta ativação para o neurônio do dígito 0
  OutNeurônio[1] (ID N1): 0.2345
  OutNeurônio[2] (ID N2): 0.1987
  OutNeurônio[3] (ID N3): 0.3012
  OutNeurônio[4] (ID N4): 0.0987
  OutNeurônio[5] (ID N5): 0.1500
  OutNeurônio[6] (ID N6): 0.2500
  OutNeurônio[7] (ID N7): 0.1000
  OutNeurônio[8] (ID N8): 0.4500
  OutNeurônio[9] (ID N9): 0.0500

Sessão CrowNet finalizada. Duração total: HH:MM:SS.
```
*(IDs N0-N9 são placeholders para os IDs reais dos neurônios de saída. Os valores de potencial são ilustrativos, mas espera-se que o neurônio associado ao dígito "0" tenha o valor mais alto.)*

## 5. Passo 3: Observação do Dígito "1" (Modo `observe`)

### Comando CLI:
```bash
./crownet -mode observe -neurons 70 -digit 1 -weightsFile digit_example_weights.json -cyclesToSettle 20 -seed 123
```
*   `-digit 1`: Apresenta o padrão do dígito "1". Os outros parâmetros são os mesmos da observação anterior.

### Saída do Console Esperada (Simulada - Foco na Ativação):
```
CrowNet Inicializando...
Modo Selecionado: observe
Configuração Base: Neurônios=70, ArquivoDePesos='digit_example_weights.json'
  observe: Dígito=1, CiclosParaAcomodar=20
Rede criada: 70 neurônios. IDs Input: [...]..., IDs Output: [N0 N1 N2 N3 N4 N5 N6 N7 N8 N9]...
Pesos existentes carregados de digit_example_weights.json
Estado Inicial: Cortisol=0.000, Dopamina=0.000

Observando Resposta da Rede para o dígito 1 (20 ciclos de acomodação)...
Dígito Apresentado: 1
Padrão de Ativação dos Neurônios de Saída (Potencial Acumulado):
  OutNeurônio[0] (ID N0): 0.3123
  OutNeurônio[1] (ID N1): 1.9502  // <-- Espera-se alta ativação para o neurônio do dígito 1
  OutNeurônio[2] (ID N2): 0.2876
  OutNeurônio[3] (ID N3): 0.1500
  OutNeurônio[4] (ID N4): 0.4000
  OutNeurônio[5] (ID N5): 0.0800
  OutNeurônio[6] (ID N6): 0.1200
  OutNeurônio[7] (ID N7): 0.3500
  OutNeurônio[8] (ID N8): 0.2000
  OutNeurônio[9] (ID N9): 0.1100

Sessão CrowNet finalizada. Duração total: HH:MM:SS.
```
*(Espera-se que o neurônio associado ao dígito "1" agora tenha o valor mais alto, e o neurônio do dígito "0" tenha um valor significativamente menor do que no Passo 2.)*

## 6. Conclusão do Exemplo

Este exemplo demonstra o fluxo básico de treinamento de uma rede CrowNet para uma tarefa de reconhecimento de padrões e a subsequente observação de seu desempenho. A análise detalhada dos arquivos de pesos ou dos logs do banco de dados (se o modo `sim` com logging fosse usado) poderia fornecer insights mais profundos sobre a dinâmica da rede.
