# Rastreador de Tarefas do Projeto

Esta tabela rastreia as tarefas de desenvolvimento de alto nível para o projeto.

| ID da Tarefa | Descrição da Tarefa                                  | Status      | Complexidade (1-5) | Responsável                     | Dependências (IDs) | Data de Criação | Data de Conclusão (Estimada/Real) | Notas                                                                 |
| :----------- | :--------------------------------------------------- | :---------- | :----------------- | :------------------------------ | :----------------- | :-------------- | :-------------------------------- | :-------------------------------------------------------------------- |
| TEST-002     | Executar testes unitários/integração, depurar e garantir passagem | Pendente    | 3 - Média          | AgenteJules                     | TEST-001           | 2025-06-24      | AAAA-MM-DD                        | Requer ambiente de execução funcional (go test ./...). Focar na estabilização dos testes criados em TEST-001. |
| TEST-003     | Desenvolver testes de integração para cenários chave de simulação | Pendente    | 4 - Alta           | AgenteJules                     | TEST-002           | 2025-06-24      | AAAA-MM-DD                        | Testar fluxo completo dos modos CLI (expose, observe, sim) com configs e dados específicos. |
| FEATURE-003  | Implementar carregamento de configuração via arquivo (TOML/YAML) | Pendente    | 3 - Média          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Modificar config/config.go. Flags CLI podem sobrescrever valores do arquivo. |
| REFACTOR-004 | Refatorar comportamento neuronal para usar interfaces (Extensibilidade) | Pendente    | 4 - Alta           | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Definir interfaces (FiringCondition, etc.) em neuron/neuron.go. |
| CHORE-003    | Criar Makefile para build, lint e teste automatizados | Pendente    | 2 - Baixa          | AgenteJules                     | FEAT-001.2         | 2025-06-24      | AAAA-MM-DD                        | Incluir targets 'make build', 'make lint', 'make test'. |
| DOC-003      | Revisão e aprimoramento global dos GoDocs para APIs públicas | Pendente    | 3 - Média          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Garantir documentação clara e completa para pacotes, funções e tipos exportados. |
| SEC-001      | Analisar e validar caminhos de arquivo para E/S (weights, SQLite DB) | Pendente    | 2 - Baixa          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Verificar limpeza/validação de caminhos para -weightsFile e -dbPath. |
| CHORE-004    | Verificar e documentar o status de manutenção das dependências externas | Pendente    | 1 - Muito Baixa    | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Foco em github.com/mattn/go-sqlite3. |
| UX-001       | Revisar a usabilidade da CLI: mensagens de ajuda, erros e feedback | Pendente    | 2 - Baixa          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Avaliar clareza das mensagens da CLI em diversos cenários. |
| REFACTOR-005 | Avaliar e propor interfaces para estratégias de propagação de pulso | Pendente    | 3 - Média          | AgenteJules                     | PERF-002           | 2025-06-24      | AAAA-MM-DD                        | Considerar extensibilidade de pulse.ProcessCycle para diferentes modelos de propagação. |
| REFACTOR-006 | Avaliar e propor interfaces para estratégias de sinaptogênese | Pendente    | 3 - Média          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Considerar extensibilidade de network.applySynaptogenesis para diferentes regras de movimento. |
| PERF-004     | Realizar profiling básico da aplicação nos modos 'expose' e 'sim' | Pendente    | 3 - Média          | AgenteJules                     | TEST-002           | 2025-06-24      | AAAA-MM-DD                        | Usar pprof para identificar gargalos de CPU/memória. Documentar resultados. |
| DOC-004      | Criar arquivo CONTRIBUTING.md com diretrizes para contribuição | Pendente    | 2 - Baixa          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Incluir estilo de código, fluxo de commit/PR, setup de ambiente. |
| CHORE-005    | Adicionar um arquivo LICENSE (ex: MIT ou Apache 2.0) | Pendente    | 1 - Muito Baixa    | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Pesquisar e adicionar texto da licença apropriada. |
| CHORE-006    | Configurar GitHub Issue Templates para bugs e feature requests | Pendente    | 1 - Muito Baixa    | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Criar arquivos de template em .github/ISSUE_TEMPLATE/. |
| FEATURE-004  | Desenvolver script/utilitário para exportar/visualizar dados do SQLite log | Pendente    | 3 - Média          | AgenteJules                     | RF-PERSIST-003     | 2025-06-24      | AAAA-MM-DD                        | Script Go/Python para CSVs, plots básicos ou resumo textual. |
| UX-002       | Melhorar representação visual do output no modo 'observe' | Pendente    | 2 - Baixa          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Considerar ASCII art para padrão 5x7 de output. |
| REFACTOR-007 | Revisão e adição de nil checks para *config.SimulationParameters | Pendente    | 2 - Baixa          | AgenteJules                     | -                  | 2025-06-24      | AAAA-MM-DD                        | Verificar e adicionar nil checks em funções críticas que usam SimParams. |

**Legenda de Status:**

*   **Pendente:** A tarefa ainda não foi iniciada.
*   **Em Andamento:** A tarefa está sendo trabalhada ativamente.
*   **Concluído:** A tarefa foi finalizada e verificada.
*   **Bloqueado:** A tarefa não pode progredir devido a dependências ou outros problemas.
*   **Revisão:** A tarefa foi concluída e está aguardando revisão/aprovação.

**Legenda de Complexidade:**

*   **1 (Muito Baixa):** Tarefa simples, rápida de executar.
*   **2 (Baixa):** Requer um pouco mais de esforço ou conhecimento.
*   **3 (Média):** Tarefa com complexidade moderada, pode envolver múltiplas etapas.
*   **4 (Alta):** Requer esforço significativo, pesquisa ou design.
*   **5 (Muito Alta):** Tarefa muito complexa, pode precisar ser subdividida.
