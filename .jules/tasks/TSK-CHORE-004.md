# Tarefa: CHORE-004 - Verificar e documentar o status de manutenção das dependências externas.

**ID da Tarefa:** CHORE-004
**Título Breve:** Verificar e documentar status de dependências.
**Descrição Completa:** Revisar as dependências externas do projeto listadas em `go.mod` (principalmente `github.com/mattn/go-sqlite3`) para verificar seu status de manutenção, licença e se existem vulnerabilidades conhecidas ou versões mais recentes recomendadas. Documentar brevemente esta verificação.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** chore/check-dependencies
**Critérios de Aceitação:**
- O repositório `github.com/mattn/go-sqlite3` é visitado para verificar atividade recente, issues abertas relevantes e informações sobre manutenção.
- Uma busca rápida por vulnerabilidades conhecidas para a versão utilizada é realizada (ex: consultando bancos de dados de vulnerabilidades como o GitHub Advisory Database ou Snyk).
- A licença da dependência é confirmada como compatível com o projeto.
- Uma breve nota sobre o status da dependência é adicionada à documentação do projeto (ex: em `docs/02_arquitetura.md` numa seção de dependências, ou num novo arquivo `DEPENDENCIES.md`).
**Notas/Decisões:**
- Esta é uma tarefa de "boa governança" para garantir que o projeto não dependa de bibliotecas abandonadas ou inseguras.
- Se forem encontradas versões mais recentes ou problemas, tarefas separadas de atualização ou mitigação podem ser criadas.
- Para `go-sqlite3`, é uma biblioteca CGO, então também envolve a dependência do SQLite em si no sistema de build/runtime.
