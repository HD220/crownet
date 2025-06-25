# Tarefa: BUG-CORE-001 - Corrigir chamada de network.NewCrowNet

**ID da Tarefa:** BUG-CORE-001
**Título Breve:** Corrigir incompatibilidade na chamada de `network.NewCrowNet`
**Descrição Completa:** A função `network.NewCrowNet` em `network/network.go` espera um argumento `*config.AppConfig`. No entanto, a chamada em `cli/orchestrator.go` (dentro de `createNetwork`) estava passando argumentos individuais (`cliCfg.TotalNeurons`, `baseLearningRate`, etc.) em vez da struct `AppConfig`. Esta tarefa corrige a chamada em `cli/orchestrator.go` para passar a instância `o.AppCfg` corretamente.
**Status:** Concluído
**Dependências (IDs):** -
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P0
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** 2024-07-27
**Branch Git Proposta:** fix/bug-core-001-newcrownet-call
**Critérios de Aceitação:**
- A chamada para `network.NewCrowNet` em `cli/orchestrator.go` usa a assinatura correta, passando `*config.AppConfig`.
- A aplicação compila sem erros relacionados a esta chamada.
- (Idealmente) A inicialização da rede funciona conforme o esperado após a correção.
**Notas/Decisões:**
- Erro crítico que impedia a correta inicialização da rede.
- Teste manual da execução pós-correção foi bloqueado por limitações do ambiente.
