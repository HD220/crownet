# Tabela Principal de Tarefas - CrowNet

**Status:** Pendente, Em Andamento, Concluído, Bloqueado, Revisão, Cancelado
**Prioridade (P0-P4):** P0 (Crítica), P1 (Alta), P2 (Média), P3 (Baixa), P4 (Muito Baixa)

| ID da Tarefa | Título Breve da Tarefa                     | Status    | Dependências (IDs) | Prioridade | Responsável | Link para Detalhes                            | Notas Breves                                  |
|--------------|--------------------------------------------|-----------|--------------------|------------|-------------|-----------------------------------------------|-----------------------------------------------|
| SYS-MIG-001  | Migrar para novo formato de tarefas        | Subdividido | -                  | P0         | AgenteJules | [TSK-SYS-MIG-001.md](./tasks/TSK-SYS-MIG-001.md) | Sub-tarefas: .1 (Concluído), .2, .3.          |
| SYS-MIG-001.1| Criar arquivos de detalhe para tarefas     | Concluído | SYS-MIG-001        | P0         | AgenteJules | [TSK-SYS-MIG-001.1.md](./tasks/TSK-SYS-MIG-001.1.md) | 18 arquivos de detalhe criados.              |
| SYS-MIG-001.2| Criar novo índice TASKS.md                 | Concluído | SYS-MIG-001,SYS-MIG-001.1 | P0         | AgenteJules | [TSK-SYS-MIG-001.2.md](./tasks/TSK-SYS-MIG-001.2.md) | Novo arquivo de índice TASKS.md criado e populado. |
| SYS-MIG-001.3| (Interno) Adaptar agente ao novo formato   | Concluído | SYS-MIG-001,SYS-MIG-001.2 | P0         | AgenteJules | [TSK-SYS-MIG-001.3.md](./tasks/TSK-SYS-MIG-001.3.md) | Lógica interna do agente adaptada.          |
| TEST-002     | Executar testes, depurar e garantir passagem | Pendente  | -                  | P2         | AgenteJules | [TSK-TEST-002.md](./tasks/TSK-TEST-002.md)    | Tentativa de correção de build. Execução bloqueada por erro de ferramenta. |
| TEST-003     | Desenvolver testes de integração chave     | Pendente  | TEST-002           | P2         | AgenteJules | [TSK-TEST-003.md](./tasks/TSK-TEST-003.md)    | Testar modos CLI.      |
| FEATURE-003  | Configuração via arquivo (TOML/YAML)       | Concluído | -                  | P2         | AgenteJules | [TSK-FEATURE-003.md](./tasks/TSK-FEATURE-003.md)| Parcial: Docs e exemplo TOML. Implementação bloqueada por erro de ferramenta (go get). |
| REFACTOR-004 | Interfaces neuronais (Extensibilidade)   | Pendente  | -                  | P2         | AgenteJules | [TSK-REFACTOR-004.md](./tasks/TSK-REFACTOR-004.md)| Definir FiringCondition, etc.               |
| CHORE-003    | Criar Makefile (build, lint, teste)      | Pendente  | FEAT-001.2         | P2         | AgenteJules | [TSK-CHORE-003.md](./tasks/TSK-CHORE-003.md)  | Targets: build, lint, test.                 |
| DOC-003      | Revisão global dos GoDocs                  | Pendente  | -                  | P2         | AgenteJules | [TSK-DOC-003.md](./tasks/TSK-DOC-003.md)      | Para APIs públicas.                         |
| SEC-001      | Validar caminhos de arquivo para E/S       | Pendente  | -                  | P2         | AgenteJules | [TSK-SEC-001.md](./tasks/TSK-SEC-001.md)      | Para -weightsFile, -dbPath.                 |
| CHORE-004    | Verificar status de dependências externas  | Pendente  | -                  | P3         | AgenteJules | [TSK-CHORE-004.md](./tasks/TSK-CHORE-004.md)  | Foco: go-sqlite3.                           |
| UX-001       | Revisar usabilidade da CLI                 | Pendente  | -                  | P2         | AgenteJules | [TSK-UX-001.md](./tasks/TSK-UX-001.md)        | Mensagens de ajuda, erros, feedback.        |
| REFACTOR-005 | Interfaces para propagação de pulso        | Pendente  | PERF-002           | P3         | AgenteJules | [TSK-REFACTOR-005.md](./tasks/TSK-REFACTOR-005.md)| Extensibilidade de pulse.ProcessCycle.      |
| REFACTOR-006 | Interfaces para sinaptogênese            | Pendente  | -                  | P3         | AgenteJules | [TSK-REFACTOR-006.md](./tasks/TSK-REFACTOR-006.md)| Extensibilidade de applySynaptogenesis.     |
| PERF-004     | Realizar profiling básico da aplicação     | Pendente  | TEST-002           | P2         | AgenteJules | [TSK-PERF-004.md](./tasks/TSK-PERF-004.md)    | Modos 'expose' e 'sim'.                     |
| DOC-004      | Criar CONTRIBUTING.md                      | Pendente  | -                  | P3         | AgenteJules | [TSK-DOC-004.md](./tasks/TSK-DOC-004.md)      | Diretrizes para contribuição.               |
| CHORE-005    | Adicionar arquivo LICENSE                  | Concluído | -                  | P1         | AgenteJules | [TSK-CHORE-005.md](./tasks/TSK-CHORE-005.md)  | Ex: MIT ou Apache 2.0.                      |
| CHORE-006    | Configurar GitHub Issue Templates          | Pendente  | -                  | P3         | AgenteJules | [TSK-CHORE-006.md](./tasks/TSK-CHORE-006.md)  | Para bugs e feature requests.               |
| FEATURE-004  | Utilitário para log SQLite               | Pendente  | -                  | P3         | AgenteJules | [TSK-FEATURE-004.md](./tasks/TSK-FEATURE-004.md)| Exportar/visualizar dados.                  |
| UX-002       | Melhorar visualização output 'observe'     | Pendente  | -                  | P3         | AgenteJules | [TSK-UX-002.md](./tasks/TSK-UX-002.md)        | Considerar ASCII art.                       |
| REFACTOR-007 | Adicionar nil checks para SimParams        | Pendente  | -                  | P2         | AgenteJules | [TSK-REFACTOR-007.md](./tasks/TSK-REFACTOR-007.md)| Em funções críticas.                        |
