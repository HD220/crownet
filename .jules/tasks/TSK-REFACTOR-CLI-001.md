# Tarefa: REFACTOR-CLI-001 - Melhorar estrutura de subcomandos CLI

**ID da Tarefa:** REFACTOR-CLI-001
**Título Breve:** Refatorar CLI para usar subcomandos verdadeiros
**Descrição Completa:** A CLI foi refatorada de um sistema baseado em flag `-mode` para uma estrutura de subcomando mais robusta e idiomática usando a biblioteca `spf13/cobra`. Agora os comandos são `crownet sim`, `crownet expose`, `crownet observe`, `crownet logutil export`, cada um com suas flags específicas. Flags globais como `--configFile` e `--seed` são persistentes.
**Status:** Concluído
**Dependências (IDs):** -
**Complexidade (1-5):** 4
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** 2024-07-27
**Branch Git Proposta:** refactor/cobra-cli-REFACTOR-CLI-001
**Critérios de Aceitação:**
- A aplicação usa `spf13/cobra` para gerenciar a CLI.
- Os modos anteriores (`sim`, `expose`, `observe`, `logutil`) são agora subcomandos.
- Flags são definidas nos comandos/subcomandos apropriados.
- A funcionalidade principal de cada modo permanece acessível através da nova estrutura CLI.
- Documentação (`README.md`, `guia_interface_linha_comando.md`) é atualizada para refletir a nova CLI.
**Notas/Decisões:**
- Melhora significativamente a usabilidade e extensibilidade da CLI.
- A dependência de `spf13/cobra` foi adicionada.
- Teste manual da nova CLI foi bloqueado.
