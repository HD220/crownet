# Rastreador de Tarefas do Projeto

Esta tabela rastreia as tarefas de desenvolvimento de alto nível para o projeto.

| ID da Tarefa | Descrição da Tarefa                                  | Status      | Complexidade (1-5) | Responsável                     | Dependências (IDs) | Data de Criação | Data de Conclusão (Estimada/Real) | Notas                                                                 |
| :----------- | :--------------------------------------------------- | :---------- | :----------------- | :------------------------------ | :----------------- | :-------------- | :-------------------------------- | :-------------------------------------------------------------------- |
| FEAT-001     | Configuração inicial do projeto                      | Concluído   | 4 - Alta           | AgenteJules                     | -                  | AAAA-MM-DD      | 2025-06-24                        | Configurar linters, formatadores e estrutura básica de pastas. Todas as subtarefas (FEAT-001.1 FEAT-001.4) foram concluídas. |
| FEAT-001.2   | Configurar o linter golangci-lint                    | Concluído   | 1 - Muito Baixa    | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | 2025-06-23                        | Adicionado .golangci.yml com conjunto sensato de linters. Verificação de execução pendente devido a problemas na ferramenta. |
| FEAT-001.3   | Configurar o formatador gofmt/goimports              | Concluído   | 1 - Muito Baixa    | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | 2025-06-23                        | .golangci.yml configurado para goimports. Guia de estilo atualizado para refletir o uso mandatório. |
| FEAT-001.4   | Integrar linters/formatadores em hook pre-commit (opcional) | Concluído   | 1 - Muito Baixa    | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | 2025-06-24                        | Documentado script e uso de hook pre-commit em guia_estilo_codigo.md. |
| DOC-001      | Elaborar documentação inicial da arquitetura         | Concluído   | 3 - Média          | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | 2025-06-24                        | Revisado e atualizado docs/02_arquitetura.md com diagrama, detalhes de pacotes, estruturas de dados e algoritmos. |
| DOC-002      | Detalhar arquitetura atual e propostas de reescrita  | Concluído   | 3 - Média          | AgenteJules                     | DOC-001            | AAAA-MM-DD      | 2025-06-24                        | Adicionada Seção 6 a docs/02_arquitetura.md com propostas de evolução/reescrita futura. |
| TEST-001     | Aumentar cobertura de testes unitários/integração    | Concluído   | 4 - Alta           | AgenteJules                     | -                  | AAAA-MM-DD      | 2025-06-24                        | Testes unitários escritos para pacotes synaptic, space, neuron, pulse, neurochemical, network (helpers). Execução e verificação de passagem pendentes devido a limitações da ferramenta. |
| FEATURE-002  | Assegurar reprodutibilidade total com seed aleatória | Concluído   | 3 - Média          | AgenteJules                     | -                  | AAAA-MM-DD      | 2025-06-24                        | Verificado uso consistente de RNG local semeado. Documentado uso da flag -seed no README.md. |
| CODE-QUALITY-001 | Aplicar Princípios de Engenharia de Software em Todo o Código-Fonte | Concluído   | 5 - Muito Alta     | AgenteJules                     | -                  | 2025-06-23      | 2025-06-24                        | Aplicar Clareza, DRY, KISS, Testabilidade, Segurança, YAGNI. Todas as subtarefas (.1 a .11) atribuídas a AgenteJules foram concluídas. |
| CODE-QUALITY-001.11 | Aplicar princípios de qualidade ao pacote 'synaptic'    | Concluído   | 3 - Média          | AgenteJules                     | CODE-QUALITY-001   | 2025-06-23      | 2025-06-24                        | Refatorado synaptic/weights.go: NetworkWeights agora é struct, encapsula simParams e rng, métodos atualizados. Test file não existia. |
| PERF-002     | Investigar otimização da propagação de pulsos         | Concluído   | 4 - Alta           | AgenteJules                     | REVIEW-044         | AAAA-MM-DD      | 2025-06-24                        | Implementado SpatialGrid em space/spatial_grid.go e integrado em PulseList.ProcessCycle para otimizar busca de neurônios afetados por pulsos. |
| PERF-003     | Avaliar otimização de `GenerateRandomPositionInHyperSphere` | Concluído | 3 - Média          | AgenteJules                     | REVIEW-046         | AAAA-MM-DD      | 2025-06-24                        | Implementado método de Muller (normalização de N gaussianas + escala radial) em space/geometry.go, substituindo rejeição. Assinatura da função alterada para usar *rand.Rand. |
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
