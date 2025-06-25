# Tarefa: REFACTOR-CONFIG-001 - Agrupar `SimulationParameters`

**ID da Tarefa:** REFACTOR-CONFIG-001
**Título Breve:** Agrupar campos da struct `config.SimulationParameters`
**Descrição Completa:** A struct `config.SimulationParameters` é atualmente muito extensa, contendo um grande número de campos. Para melhorar a organização, legibilidade e manutenibilidade, esta tarefa propõe agrupar logicamente esses campos em sub-structs temáticas. Por exemplo, poderiam ser criadas structs como `LearningParameters`, `SynaptogenesisParameters`, `NeurochemicalParameters`, `SpatialParameters`, etc., que seriam então campos da struct `SimulationParameters` principal.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** refactor/group-simparams
**Critérios de Aceitação:**
- A struct `config.SimulationParameters` é refatorada para conter sub-structs agrupando parâmetros relacionados.
- Todo o código que acessa campos de `SimulationParameters` é atualizado para usar a nova estrutura (ex: `SimParams.Learning.HebbianWindow` em vez de `SimParams.HebbianCoincidenceWindow`).
- A desserialização de TOML (se implementada) e a função `DefaultSimulationParameters` são atualizadas para refletir a nova estrutura.
- A aplicação continua funcional e os parâmetros são lidos e usados corretamente.
**Notas/Decisões:**
- Esta é uma refatoração com impacto em muitas partes do código que acessam `SimParams`.
- Melhorará a organização do código de configuração.
- A escolha exata dos agrupamentos e nomes das sub-structs deve ser feita no início da implementação da tarefa.
