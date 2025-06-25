# Tarefa: REFACTOR-TEST-001 - Revisar wrappers de teste no Orchestrator

**ID da Tarefa:** REFACTOR-TEST-001
**Título Breve:** Analisar e refatorar wrappers de teste no Orchestrator
**Descrição Completa:** O arquivo `cli/orchestrator.go` contém vários métodos com sufixo `ForTest` (ex: `RunObserveModeForTest`, `CreateNetworkForTest`, `SetLoadWeightsFn`). Estes métodos são usados para expor funcionalidades internas ou permitir injeção de dependência para pacotes de teste externos (provavelmente `cli_test`). Esta tarefa consiste em:
1. Analisar cada um desses wrappers de teste.
2. Avaliar se a lógica subjacente no `Orchestrator` ou nos pacotes que ele chama pode ser refatorada para ser mais diretamente testável sem a necessidade desses wrappers. Isso pode envolver melhorias no design para injeção de dependência, extração de lógica para funções públicas em outros pacotes, ou tornar tipos internos mais acessíveis para teste (com cautela).
3. O objetivo é reduzir a superfície de API de teste específica e promover um design mais testável intrinsecamente.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** refactor/orchestrator-testability
**Critérios de Aceitação:**
- Cada método `XxxForTest` em `cli/orchestrator.go` é analisado.
- Para cada wrapper, uma decisão é tomada: manter, refatorar para eliminar, ou refatorar para melhorar.
- Onde aplicável, o código do `Orchestrator` ou dos pacotes relacionados é refatorado para melhorar a testabilidade.
- O número de wrappers de teste é reduzido, se possível, sem perder a capacidade de testar a funcionalidade.
**Notas/Decisões:**
- Melhorar a testabilidade pode levar a um design de código mais modular e desacoplado.
- Esta tarefa pode ser feita incrementalmente.
- A capacidade de executar testes (`TEST-002`) é um pré-requisito para validar que as refatorações não quebram os testes existentes ou que novos testes para a lógica refatorada passam.
