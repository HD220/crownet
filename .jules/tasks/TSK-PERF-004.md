# Tarefa: PERF-004 - Realizar profiling básico da aplicação nos modos 'expose' e 'sim'.

**ID da Tarefa:** PERF-004
**Título Breve:** Realizar profiling básico da aplicação.
**Descrição Completa:** Utilizar as ferramentas de profiling do Go (pprof) para analisar o desempenho da aplicação CrowNet durante a execução dos modos `expose` e `sim` com configurações representativas (ex: número de neurônios, ciclos/épocas). O objetivo é identificar os principais gargalos de CPU e alocação de memória.
**Status:** Pendente
**Dependências (IDs):** TEST-002 (para ter cenários de execução estáveis e testados)
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** perf/initial-profiling
**Critérios de Aceitação:**
- Profiling de CPU é realizado para os modos `expose` e `sim` com configurações de teste significativas.
- Profiling de memória (heap) é realizado para os mesmos modos e configurações.
- Os resultados do pprof (ex: top funções, flame graphs) são analisados para identificar as áreas do código que consomem mais CPU e/ou alocam mais memória.
- Um breve relatório ou notas resumindo os achados é produzido.
- Se gargalos claros forem identificados, novas tarefas de otimização (PERF-XXX) podem ser propostas.
**Notas/Decisões:**
- Requer a capacidade de compilar e executar a aplicação com as flags de profiling habilitadas (ex: `net/http/pprof` ou `runtime/pprof`).
- Cenários de teste para profiling devem ser longos o suficiente para capturar comportamento representativo.
- Esta tarefa é investigativa e pode levar à criação de novas tarefas de otimização mais específicas.
- As otimizações de PERF-002 e PERF-003 já foram implementadas; este profiling ajudará a verificar seu impacto e encontrar novos alvos.
- **Resultados da Análise de Profiling Simulado (PERF-004.3 - 2024-07-28):**
    - **Nota:** A análise a seguir é conceitual, baseada no conhecimento do sistema e sem execução interativa de `go tool pprof`.
    - **Potenciais Hotspots de CPU (Comuns a `expose` e `sim`):**
        - Lógica de atualização de neurônios (`neuron.AdvanceState`, `neuron.DecayPotential`, `neuron.IntegrateIncomingPotential`) devido à execução por neurônio/ciclo.
        - Processamento de pulsos (`pulse.PulseList.ProcessCycle`), incluindo propagação, seleção de alvos (especialmente `spatialGrid.QuerySphereForCandidates`), e cálculo de impacto (`space.EuclideanDistance`, `SynapticWeights.GetWeight`).
        - Lógica de sinaptogênese (`ForceCalculator.CalculateForce`, `MovementUpdater.UpdateMovement`), se habilitada, devido a cálculos de distância e iterações entre pares de neurônios.
        - Geração de números aleatórios, se usada extensivamente em loops críticos.
    - **Potenciais Hotspots de CPU (Específicos do Modo):**
        - `expose`: `Network.PresentPattern` (manipulação de dados), `Network.applyHebbianLearning` (iterações entre pares de neurônios).
        - `sim`: `SQLiteLogger.LogNetworkState` (serialização JSON, transações DB), se o `SaveInterval` for frequente.
    - **Potenciais Fontes de Alocação de Memória:**
        - Criação/destruição de objetos `pulse.Pulse`.
        - Slices/mapas temporários em loops de simulação (e.g., listas de candidatos, vetores de força).
        - Serialização de estados de neurônios para logging SQLite.
    - **Recomendações Gerais para Otimização Futura (baseado na análise simulada):**
        - Revisar algoritmos em `spatialGrid` e cálculos de distância.
        - Investigar pooling de objetos para `pulse.Pulse` se o churn for alto.
        - Minimizar alocações em loops críticos.
        - Avaliar a frequência e eficiência do logging SQLite.
    - **Próximos Passos Sugeridos:** Se os perfis reais (quando analisados visualmente) confirmarem estes hotspots, criar tarefas PERF específicas para otimizar as áreas identificadas. Por exemplo, PERF-005: Otimizar cálculos de distância; PERF-006: Investigar pooling de objetos Pulse.
