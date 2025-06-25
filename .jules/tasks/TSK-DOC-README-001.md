# Tarefa: DOC-README-001 - Corrigir inconsistências no `README.md`

**ID da Tarefa:** DOC-README-001
**Título Breve:** Corrigir e atualizar `README.md`
**Descrição Completa:** Atualizar o arquivo `README.md` para corrigir inconsistências identificadas em relação ao comportamento atual do código e para refletir funcionalidades recentemente adicionadas ou alteradas. As principais correções foram focadas no `docs/03_guias/guia_interface_linha_comando.md`, e o `README.md` foi ajustado para ser consistente ou para apontar para o guia para detalhes mais precisos.
**Status:** Concluído
**Dependências (IDs):** -
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P1
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** 2024-07-27
**Branch Git Proposta:** docs/fix-readme-guide-consistency-DOC-README-001
**Critérios de Aceitação:**
- A descrição do comportamento da flag `-dbPath` no `guia_interface_linha_comando.md` (referenciado pelo README) está correta (não recria o arquivo).
- O estado da funcionalidade `-configFile` (TOML) está claro no `guia_interface_linha_comando.md` (implementado) e o `README.md` menciona TOML nas tecnologias.
- A descrição do esquema do BD SQLite para `Position` e `Velocity` no `guia_interface_linha_comando.md` está correta (JSON string).
- O `README.md` reflete a nova estrutura de CLI baseada em subcomandos (introduzida por REFACTOR-CLI-001).
**Notas/Decisões:**
- Muitas das correções detalhadas foram aplicadas ao `guia_interface_linha_comando.md`, e o `README.md` foi atualizado para ser consistente com essas mudanças em seu nível de abstração mais alto.
- A tarefa visava manter a documentação principal do usuário precisa.
