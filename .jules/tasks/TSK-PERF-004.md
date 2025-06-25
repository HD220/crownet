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
