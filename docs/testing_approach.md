# Abordagem de Testes no Projeto CrowNet

Este documento descreve a estratégia e as diferentes camadas de teste utilizadas no projeto CrowNet para garantir a qualidade, corretude e robustez do código.

## 1. Filosofia Geral de Testes

O objetivo é ter uma pirâmide de testes saudável, com uma base sólida de testes unitários rápidos e isolados, complementada por testes de integração para verificar a interação entre componentes, e testes de sistema (ou de ponta a ponta) para validar os casos de uso completos da aplicação.

Durante a refatoração extensiva, um foco principal foi aumentar a testabilidade do código, o que permitiu a criação de muitos dos testes descritos abaixo.

## 2. Camadas de Teste

### 2.1. Testes Unitários (`_test.go` em cada pacote)

*   **Objetivo:** Verificar a corretude de unidades de código isoladas (funções, métodos, pequenas structs) dentro de cada pacote.
*   **Escopo:** Focado na lógica interna do pacote, com dependências externas (outros pacotes, sistema de arquivos, rede) geralmente mockadas ou controladas.
*   **Exemplos Implementados:**
    *   `common/types_test.go` (se houvesse lógica a testar, atualmente são só tipos)
    *   `config/config_test.go` (testes para validação de configuração)
    *   `neuron/neuron_test.go` (testa a lógica de estado, potencial, disparo do neurônio)
    *   `pulse/pulse_test.go` (testa a propagação de pulsos, gerenciamento em `PulseList`)
    *   `space/geometry_test.go` (testa funções de cálculo de distância, clamping, geração de posição)
    *   `synaptic/weights_test.go` (testa gerenciamento de pesos e aprendizado Hebbiano)
    *   `neurochemical/neurochemicals_test.go` (testa atualização de níveis químicos e seus efeitos)
    *   `datagen/digits_test.go` (testa a obtenção e validação de padrões de dígitos)
    *   `network/network_test.go` (testa funções de inicialização da rede, adição de neurônios, e algumas lógicas de I/O como `ConfigureFrequencyInput`, `GetOutputFrequency`, `PresentPattern`, `recordOutputFiring`).
    *   `storage/json_persistence_test.go` (testa serialização/desserialização de pesos).
    *   `storage/sqlite_logger_test.go` (testa criação de tabelas e logging de estado da rede para SQLite, usando DB em memória).

*   **Técnicas de Mocking/Isolamento Usadas:**
    *   **Dados de Entrada Controlados:** Fornecer entradas simples e previsíveis.
    *   **Structs de Teste:** Definir structs auxiliares dentro dos arquivos de teste quando necessário.
    *   **Verificação de Estado:** Checar os valores de retorno e o estado das structs após a execução da unidade testada.
    *   **Funções Helper de Teste:** Como `floatEquals` para comparação de floats.

### 2.2. Testes de "Quase Integração" para a CLI (`cli/orchestrator_test.go`)

*   **Objetivo:** Validar a lógica de orquestração do pacote `cli`, especialmente como ele lida com diferentes configurações, modos de operação e como propaga erros de chamadas a outros pacotes (como `storage`, `datagen`, `network`). Estes não são testes de ponta a ponta completos da CLI, mas testam o `Orchestrator` de forma mais integrada do que um teste unitário puro.
*   **Escopo:** Focado no `cli.Orchestrator` e suas interações diretas.
*   **Técnicas Empregadas:**
    *   **Injeção de Dependência via Campos de Função:**
        *   O `Orchestrator` foi refatorado para ter campos como `loadWeightsFn` e `saveWeightsFn`. No código de produção, eles apontam para as funções reais do pacote `storage`. Nos testes, eles são sobrescritos para retornar dados mockados ou simular erros, permitindo testar o comportamento do `Orchestrator` em resposta a essas condições.
    *   **Mocking de Variáveis de Função de Pacote:**
        *   A função `datagen.GetDigitPattern` foi transformada em uma variável de função (`datagen.GetDigitPatternFn`). Os testes em `cli_test` sobrescrevem temporariamente esta variável para fornecer padrões de entrada controlados, isolando o teste da implementação real de `datagen`.
    *   **Wrappers de Teste para Funções Não Exportadas:**
        *   Funções não exportadas do `Orchestrator` (como `runObserveMode`) são chamadas através de wrappers exportados (ex: `RunObserveModeForTest()`) definidos no próprio `orchestrator.go`. Isso permite que o pacote de teste (`cli_test`) invoque essas lógicas.
    *   **Captura de Saída do Console (`stdout`):**
        *   Uma função helper `captureStdoutReturnError` é usada para capturar o que seria impresso no console durante a execução de uma parte do `Orchestrator` (ex: o modo `observe`). Isso permite verificar se a saída informativa esperada é gerada.
    *   **Criação de Arquivos Temporários:** Para testes que envolvem arquivos (ex: carregar pesos), arquivos temporários são criados e limpos usando `t.TempDir()`.
*   **Exemplos Implementados:**
    *   Testes para caminhos de erro em `setupContinuousInputStimulus`.
    *   Testes para falhas de carregamento/salvamento de pesos nos modos `expose` e `observe`.
    *   Teste de verificação da saída do console para o modo `observe`.
    *   Teste para verificar a criação do arquivo de banco de dados no modo `sim`.

### 2.3. Testes de Sistema / Ponta a Ponta (Manuais ou Futuros Automatizados)

*   **Objetivo:** Validar o comportamento do sistema CrowNet como um todo, executando o binário compilado com diferentes argumentos da CLI e verificando os resultados finais (arquivos gerados, saídas principais no console, códigos de saída).
*   **Escopo:** Interação completa de todos os módulos, simulando o uso real pelo usuário.
*   **Documentação:** Os cenários para estes testes estão detalhados em:
    *   `TESTING_SCENARIOS.md`: Focado nos comandos da CLI e verificações de arquivos/saídas.
    *   `docs/use_cases.md`: Descreve os casos de uso funcionais, que informam os cenários de teste.
*   **Exemplo de Execução Simulada Documentada:**
    *   `docs/execution_reports/report_digit_1_recognition_simulation.md`: Este arquivo foi gerado *programaticamente* por um "teste" especial (`cli/orchestrator_report_test.go`) que simula um fluxo de caso de uso e formata os resultados em Markdown. Ele serve como um exemplo concreto de execução e o tipo de relatório que testes de sistema poderiam produzir.
*   **Status Atual:** Estes testes são atualmente manuais, baseados nos documentos acima. A automação futura poderia envolver scripts shell ou o uso do pacote `os/exec` em Go para invocar o binário e analisar seus resultados.

## 3. Papel dos Mocks e Simulações nos Testes

*   **Mocks Não Indicam Funcionalidade Ausente:** É crucial entender que o uso de mocks (como os `...Fn` no `Orchestrator` ou a sobrescrita de `datagen.GetDigitPatternFn`) nos pacotes `_test.go` é uma técnica para **isolar o código sob teste** e **criar condições de teste controladas e reproduzíveis**. Eles não significam que a funcionalidade real está faltando no código de produção. As implementações concretas (ex: `storage.LoadNetworkWeightsFromJSON`, `datagen.getDigitPatternInternal`) são as padrões usadas pela aplicação.
*   **Simulação vs. Execução Real:** O arquivo de relatório em `docs/execution_reports/` é gerado por uma *simulação programática* do fluxo do `Orchestrator`, não por uma execução real do binário CLI. Ele demonstra o que o código Go *deveria* fazer.

## 4. Cobertura de Teste

Embora não haja uma medição formal de cobertura de código configurada neste ambiente, o objetivo das fases de refatoração foi aumentar significativamente a cobertura de testes unitários para os módulos de lógica central. Os testes de "quase integração" para a CLI adicionam outra camada de verificação. Testes de sistema completos aumentariam ainda mais a confiança.

---

Esta abordagem de teste em múltiplas camadas visa fornecer um bom equilíbrio entre testes rápidos e focados (unitários) e validações de comportamento mais amplas (integração, sistema), adaptando-se às capacidades e ao estado de desenvolvimento do projeto.
```
