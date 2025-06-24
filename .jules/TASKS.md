# Rastreador de Tarefas do Projeto

Esta tabela rastreia as tarefas de desenvolvimento de alto nível para o projeto.

| ID da Tarefa | Descrição da Tarefa                                  | Status      | Complexidade (1-5) | Responsável                     | Dependências (IDs) | Data de Criação | Data de Conclusão (Estimada/Real) | Notas                                                                 |
| :----------- | :--------------------------------------------------- | :---------- | :----------------- | :------------------------------ | :----------------- | :-------------- | :-------------------------------- | :-------------------------------------------------------------------- |
| FEAT-001     | Configuração inicial do projeto                      | Subdividido | 4 - Alta           | [NOME_DO_MEMBRO_DA_EQUIPE]      | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Configurar linters, formatadores e estrutura básica de pastas. Subtarefas: FEAT-001.1, FEAT-001.2, FEAT-001.3, FEAT-001.4. |
| FEAT-001.2   | Configurar o linter golangci-lint                    | Pendente    | 1 - Muito Baixa    | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | AAAA-MM-DD                        | Adicionar .golangci.yml com linters sensatos.                      |
| FEAT-001.3   | Configurar o formatador gofmt/goimports              | Pendente    | 1 - Muito Baixa    | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | AAAA-MM-DD                        | Garantir uso de gofmt, considerar goimports, documentar.          |
| FEAT-001.4   | Integrar linters/formatadores em hook pre-commit (opcional) | Pendente    | 1 - Muito Baixa    | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | AAAA-MM-DD                        | Investigar/implementar hook pre-commit simples ou documentar uso manual. |
| DOC-001      | Elaborar documentação inicial da arquitetura         | Pendente    | 3 - Média          | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | AAAA-MM-DD                        | Detalhar o padrão arquitetural escolhido em docs/02_arquitetura.md. |
| DOC-002      | Detalhar arquitetura atual e propostas de reescrita  | Pendente    | 3 - Média          | AgenteJules                     | DOC-001            | AAAA-MM-DD      | AAAA-MM-DD                        | Expandir docs/02_arquitetura.md para cobrir estado atual e planos futuros. |
| TEST-001     | Aumentar cobertura de testes unitários/integração    | Pendente    | 4 - Alta           | [NOME_DO_MEMBRO_DA_EQUIPE]      | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Focar em pacotes críticos: network, neuron, pulse, neurochemical, synaptic. |
| FEATURE-002  | Assegurar reprodutibilidade total com seed aleatória | Pendente    | 3 - Média          | [NOME_DO_MEMBRO_DA_EQUIPE]      | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Verificar uso de rand local vs global e configurar explicitamente se necessário. |
| CODE-QUALITY-001 | Aplicar Princípios de Engenharia de Software em Todo o Código-Fonte | Subdividido | 5 - Muito Alta     | AgenteJules                     | -                  | 2025-06-23      | AAAA-MM-DD                        | Aplicar Clareza, DRY, KISS, Testabilidade, Segurança, YAGNI. Reescrever arquivos conforme necessário. |
| CODE-QUALITY-001.11 | Aplicar princípios de qualidade ao pacote 'synaptic'    | Pendente    | 3 - Média          | AgenteJules                     | CODE-QUALITY-001   | 2025-06-23      | AAAA-MM-DD                        | Inclui synaptic/weights.go, synaptic/weights_test.go. |
| PERF-002     | Investigar otimização da propagação de pulsos         | Pendente    | 4 - Alta           | AgenteJules                     | REVIEW-044         | AAAA-MM-DD      | AAAA-MM-DD                        | Otimizar `PulseList.ProcessCycle` com busca espacial. |
| PERF-003     | Avaliar otimização de `GenerateRandomPositionInHyperSphere` | Pendente | 3 - Média          | AgenteJules                     | REVIEW-046         | AAAA-MM-DD      | AAAA-MM-DD                        | Considerar método de Muller para altas dimensões. |
|              |                                                      |             |                    |                                 |                    |                 |                                   |                                                                       |

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
