# Tarefa: DOC-GUIDE-CLI-001 - Corrigir inconsistências no Guia da CLI

**ID da Tarefa:** DOC-GUIDE-CLI-001
**Título Breve:** Corrigir e atualizar Guia da CLI
**Descrição Completa:** Sincronizar o arquivo `docs/03_guias/guia_interface_linha_comando.md` com o estado atual do código e outras documentações (como o `README.md`). As correções específicas incluíram:
1.  Comportamento da flag `-dbPath` (não recria o arquivo).
2.  Status e uso da flag `--configFile` para configuração TOML (agora implementada).
3.  Esquema correto da tabela `NeuronStates` no banco de dados SQLite (campos `Position` e `Velocity` como TEXT JSON, nomes de coluna `CurrentState` e `AccumulatedPotential`).
4.  Refletir a nova estrutura de CLI baseada em subcomandos (introduzida por REFACTOR-CLI-001), atualizando o formato geral dos comandos, listagem de flags e exemplos.
**Status:** Concluído
**Dependências (IDs):** DOC-README-001 (para consistência), REFACTOR-CLI-001 (para nova estrutura CLI)
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P1
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** 2024-07-27
**Branch Git Proposta:** docs/fix-readme-guide-consistency-DOC-README-001 (as alterações foram feitas como parte da DOC-README-001 e depois do REFACTOR-CLI-001)
**Critérios de Aceitação:**
- O `guia_interface_linha_comando.md` descreve com precisão o comportamento atual das flags `-dbPath` e `--configFile`.
- O esquema do banco de dados SQLite para `NeuronStates` está corretamente documentado.
- A estrutura de subcomandos da CLI, flags e exemplos estão atualizados e corretos.
**Notas/Decisões:**
- Esta tarefa garante que o guia detalhado da CLI seja uma fonte confiável de informação para os usuários.
- As alterações foram agrupadas com commits de DOC-README-001 e REFACTOR-CLI-001.
