# Tarefa: REFACTOR-MAIN-001 - Passar `os.Args` explicitamente para `NewAppConfig`

**ID da Tarefa:** REFACTOR-MAIN-001
**Título Breve:** Explicitar passagem de `os.Args` para configuração
**Descrição Completa:** Atualmente, a função `config.LoadCLIConfig` (que era chamada por `config.NewAppConfig` antes da refatoração para Cobra) depende implicitamente do `flag.CommandLine` global e de `os.Args`. Embora `NewAppConfig` e `LoadCLIConfig` não sejam mais os pontos de entrada principais para parsing de flags com Cobra, se forem mantidas para algum propósito (ex: testes, ou futura refatoração do carregamento TOML), elas devem ser modificadas para aceitar `args []string` explicitamente. A função `main.go` (antes da refatoração Cobra) chamava `config.NewAppConfig()` sem argumentos. Esta tarefa, se aplicável após a refatoração Cobra, visaria tornar qualquer lógica remanescente de parsing de flags ou construção de config mais explícita em suas dependências.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** refactor/explicit-osargs
**Critérios de Aceitação:**
- Se as funções `config.NewAppConfig(args []string)` e `config.LoadCLIConfig(fSet *flag.FlagSet, args []string)` forem mantidas e usadas, garantir que `main.go` (ou o chamador relevante) passe `os.Args[1:]` explicitamente.
- A lógica de parsing de flags (se ainda usada por essas funções) deve usar o `FlagSet` e os `args` fornecidos, não o estado global.
**Notas/Decisões:**
- Com a introdução do Cobra, a relevância direta desta tarefa diminuiu, pois Cobra gerencia seu próprio parsing de args.
- No entanto, se `config.NewAppConfig` for re-propósito para, por exemplo, orquestrar o carregamento de TOML e depois aplicar overrides de flags (já parseadas pelo Cobra), então a passagem explícita de dados ainda é uma boa prática.
- Esta tarefa pode ser reavaliada ou fundida com `FEATURE-CONFIG-001` se a lógica de `NewAppConfig` for central para o carregamento TOML.
- Por ora, a principal mudança (Cobra) já ocorreu. O `main.go` atual apenas chama `cmd.Execute()`.
- **Reavaliação:** Esta tarefa pode ser considerada obsoleta ou de prioridade muito baixa após REFACTOR-CLI-001, a menos que `config.NewAppConfig` e `config.LoadCLIConfig` sejam explicitamente reutilizadas.
