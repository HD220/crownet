# Cenários de Teste de Sistema para CrowNet

Este documento descreve cenários de teste de ponta a ponta para validar a funcionalidade principal da simulação CrowNet através de sua interface de linha de comando (CLI).

## Pré-requisitos Comuns para Testes Manuais/Automatizados:

1.  **Binário Compilado:** O projeto `crownet` deve ser compilado para gerar um executável (ex: `./crownet`).
2.  **Ambiente Limpo:** Recomenda-se executar testes em um diretório limpo ou garantir que arquivos de execuções anteriores (pesos, DBs) sejam removidos ou nomeados exclusivamente para evitar interferência.
3.  **Configurações Padrão:** Muitos testes podem depender dos `DefaultSimulationParameters` definidos em `config/config.go`. Alterações nesses padrões podem exigir ajustes nos resultados esperados.

## Abreviaturas:
*   `EXEC`: Refere-se ao comando para executar o binário compilado (ex: `./crownet`).

---

## Testes do Modo `sim` (Simulação Geral)

### Teste SIM_01: Execução Básica do Modo `sim`
*   **Comando:** `EXEC -mode sim -cycles 10 -neurons 50`
*   **Configuração de Entrada:** Nenhuma específica além dos parâmetros da CLI.
*   **Comportamento Esperado:**
    1.  O programa executa por 10 ciclos.
    2.  O programa termina sem erros (código de saída 0).
    3.  Saída no console deve indicar o início e fim da simulação, e logs de ciclo (a cada 10 ciclos ou no último, conforme `cli/orchestrator.go`).
    4.  Um arquivo de banco de dados (padrão: `crownet_sim_run.db`) **não** deve ser criado se `-dbPath` não for especificado ou se `-saveInterval` for 0 (a menos que o padrão de `dbPath` não seja vazio e `saveInterval` seja >0 por padrão na CLIConfig para o modo sim). *Nota: A lógica atual de `initializeLogger` e `runSimulationLoop` pode criar o DB mesmo com `saveInterval=0` se `dbPath` for fornecido. Isso precisa ser verificado.*
*   **Verificação:** Código de saída, logs no console.

### Teste SIM_02: Modo `sim` com Logging SQLite
*   **Comando:** `EXEC -mode sim -cycles 25 -neurons 50 -dbPath sim_test.db -saveInterval 10`
*   **Configuração de Entrada:** Nenhuma.
*   **Comportamento Esperado:**
    1.  O programa executa por 25 ciclos.
    2.  O programa termina sem erros.
    3.  Saída no console indica logs de ciclo.
    4.  O arquivo `sim_test.db` é criado.
    5.  O banco de dados `sim_test.db` deve conter:
        *   Tabela `NetworkSnapshots` com pelo menos 2 registros (ciclos 10, 20) mais o registro final (ciclo 25).
        *   Tabela `NeuronStates` com `N*num_snapshots` registros, onde N é o número de neurônios (50).
*   **Verificação:** Código de saída, logs, existência e conteúdo básico do arquivo `sim_test.db` (verificar contagem de linhas nas tabelas).

### Teste SIM_03: Modo `sim` com Estímulo de Input Contínuo
*   **Comando:** `EXEC -mode sim -cycles 10 -neurons 30 -stimInputID 0 -stimInputFreqHz 10`
    *   (Assumindo que o neurônio com ID 0 é um neurônio de input válido na rede de 30 neurônios).
*   **Configuração de Entrada:** Nenhuma.
*   **Comportamento Esperado:**
    1.  O programa executa por 10 ciclos.
    2.  O programa termina sem erros.
    3.  Saída no console deve indicar "Estímulo geral: Neurônio de Input 0 a 10.0 Hz."
    4.  Logs de ciclo devem mostrar atividade de pulsos.
*   **Verificação:** Código de saída, logs no console.

### Teste SIM_04: Modo `sim` com ID de Estímulo Inválido
*   **Comando:** `EXEC -mode sim -cycles 10 -neurons 30 -stimInputID 999 -stimInputFreqHz 10`
*   **Configuração de Entrada:** Nenhuma.
*   **Comportamento Esperado:**
    1.  O programa deve terminar com erro (código de saída não zero).
    2.  Mensagem de erro no console indicando que o ID do neurônio de input para estímulo é inválido.
*   **Verificação:** Código de saída, mensagem de erro.

---

## Testes do Modo `expose` (Treinamento/Exposição a Padrões)

### Teste EXP_01: Execução Básica do Modo `expose` (Criação de Pesos)
*   **Comando:** `EXEC -mode expose -epochs 2 -cyclesPerPattern 5 -neurons 60 -weightsFile expose_weights_01.json -lrBase 0.01`
*   **Configuração de Entrada:** Nenhuma (o arquivo de pesos será criado).
*   **Comportamento Esperado:**
    1.  O programa executa por 2 épocas.
    2.  O programa termina sem erros.
    3.  Saída no console indica progresso das épocas e processamento de padrões.
    4.  O arquivo `expose_weights_01.json` é criado e contém uma estrutura JSON válida para pesos sinápticos.
*   **Verificação:** Código de saída, logs, existência e formato básico de `expose_weights_01.json`.

### Teste EXP_02: Modo `expose` Carregando Pesos Existentes
*   **Pré-condição:** Executar Teste EXP_01 para gerar `expose_weights_01.json`.
*   **Comando:** `EXEC -mode expose -epochs 1 -cyclesPerPattern 5 -neurons 60 -weightsFile expose_weights_01.json -lrBase 0.005`
*   **Configuração de Entrada:** Arquivo `expose_weights_01.json` da execução anterior.
*   **Comportamento Esperado:**
    1.  O programa indica no console que carregou os pesos existentes.
    2.  Executa por 1 época.
    3.  O programa termina sem erros.
    4.  O arquivo `expose_weights_01.json` é sobrescrito com os novos pesos treinados.
*   **Verificação:** Código de saída, logs (especialmente a mensagem de carregamento de pesos), data de modificação do arquivo de pesos.

### Teste EXP_03: Modo `expose` com Arquivo de Pesos Inexistente (Não Fatal)
*   **Comando:** `EXEC -mode expose -epochs 1 -cyclesPerPattern 5 -neurons 60 -weightsFile non_existent_expose.json`
*   **Configuração de Entrada:** Garantir que `non_existent_expose.json` não exista.
*   **Comportamento Esperado:**
    1.  O programa indica no console que o arquivo de pesos não foi encontrado e que está iniciando com pesos aleatórios.
    2.  Executa por 1 época.
    3.  O programa termina sem erros.
    4.  O arquivo `non_existent_expose.json` é criado.
*   **Verificação:** Código de saída, logs, criação do arquivo.

### Teste EXP_04: Modo `expose` com Parâmetros Inválidos
*   **Comando:** `EXEC -mode expose -epochs 0 -weightsFile error_expose.json` (épocas <= 0)
*   **Configuração de Entrada:** Nenhuma.
*   **Comportamento Esperado:**
    1.  O programa deve terminar com erro devido à validação de configuração (épocas > 0).
    2.  Mensagem de erro informativa.
*   **Verificação:** Código de saída, mensagem de erro.
*   **Outros Casos:** Testar com `-cyclesPerPattern 0`.

---

## Testes do Modo `observe` (Observação de Padrão)

### Teste OBS_01: Execução Básica do Modo `observe`
*   **Pré-condição:** Executar Teste EXP_01 para gerar `expose_weights_01.json`.
*   **Comando:** `EXEC -mode observe -digit 7 -weightsFile expose_weights_01.json -cyclesToSettle 10 -neurons 60`
*   **Configuração de Entrada:** Arquivo `expose_weights_01.json`.
*   **Comportamento Esperado:**
    1.  O programa indica que carregou os pesos.
    2.  Executa a observação para o dígito 7.
    3.  O programa termina sem erros.
    4.  Saída no console deve mostrar "Dígito Apresentado: 7" e o "Padrão de Ativação dos Neurônios de Saída".
*   **Verificação:** Código de saída, logs, saída da ativação.

### Teste OBS_02: Modo `observe` com Arquivo de Pesos Inexistente (Fatal)
*   **Comando:** `EXEC -mode observe -digit 3 -weightsFile non_existent_observe.json`
*   **Configuração de Entrada:** Garantir que `non_existent_observe.json` não exista.
*   **Comportamento Esperado:**
    1.  O programa deve terminar com erro.
    2.  Mensagem de erro indicando que o arquivo de pesos não foi encontrado e que é necessário para o modo `observe`.
*   **Verificação:** Código de saída, mensagem de erro.

### Teste OBS_03: Modo `observe` com Dígito Inválido
*   **Pré-condição:** Executar Teste EXP_01 para gerar `expose_weights_01.json`.
*   **Comando:** `EXEC -mode observe -digit 15 -weightsFile expose_weights_01.json`
*   **Configuração de Entrada:** Arquivo `expose_weights_01.json`.
*   **Comportamento Esperado:**
    1.  O programa deve terminar com erro devido à validação de configuração (dígito 0-9).
    2.  Mensagem de erro informativa.
*   **Verificação:** Código de saída, mensagem de erro.

---

## Testes Gerais de Configuração e CLI

### Teste CFG_01: Modo Inválido
*   **Comando:** `EXEC -mode invalidmode`
*   **Comportamento Esperado:**
    1.  O programa termina com erro.
    2.  Mensagem de erro indicando modo inválido.
*   **Verificação:** Código de saída, mensagem de erro.

### Teste CFG_02: Semente RNG
*   **Comando 1:** `EXEC -mode sim -cycles 5 -neurons 20 -seed 123 -dbPath seed_test1.db -saveInterval 1`
*   **Comando 2:** `EXEC -mode sim -cycles 5 -neurons 20 -seed 123 -dbPath seed_test2.db -saveInterval 1`
*   **Comando 3:** `EXEC -mode sim -cycles 5 -neurons 20 -seed 456 -dbPath seed_test3.db -saveInterval 1`
*   **Comando 4:** `EXEC -mode sim -cycles 5 -neurons 20 -seed 0 -dbPath seed_test4a.db -saveInterval 1` (seed 0 usa time)
*   **Comando 5:** `EXEC -mode sim -cycles 5 -neurons 20 -seed 0 -dbPath seed_test4b.db -saveInterval 1` (seed 0 usa time, pequena pausa entre execuções)
*   **Comportamento Esperado:**
    1.  Todos os comandos terminam sem erro.
    2.  Os arquivos `seed_test1.db` e `seed_test2.db` devem ser idênticos (ou conter dados de simulação idênticos, difícil de verificar bit a bit sem ferramentas, mas posições iniciais de neurônios e pesos iniciais deveriam ser os mesmos).
    3.  O arquivo `seed_test3.db` deve ser diferente de `seed_test1.db`/`seed_test2.db`.
    4.  Os arquivos `seed_test4a.db` e `seed_test4b.db` devem ser diferentes (devido à semente baseada no tempo).
*   **Verificação:** Existência dos arquivos. A verificação de conteúdo idêntico/diferente é mais complexa e pode exigir scripts externos ou inspeção manual dos dados (ex: posições iniciais dos neurônios, primeiros pesos gerados).

---

Este documento serve como um ponto de partida. Testes mais específicos podem ser adicionados conforme necessário, e os métodos de verificação podem ser automatizados com scripts ou frameworks de teste apropriados.
```
