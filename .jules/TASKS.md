# Rastreador de Tarefas do Projeto

Esta tabela rastreia as tarefas de desenvolvimento de alto nível para o projeto.

| ID da Tarefa | Descrição da Tarefa                                  | Status      | Complexidade (1-5) | Responsável                     | Dependências (IDs) | Data de Criação | Data de Conclusão (Estimada/Real) | Notas                                                                 |
| :----------- | :--------------------------------------------------- | :---------- | :----------------- | :------------------------------ | :----------------- | :-------------- | :-------------------------------- | :-------------------------------------------------------------------- |
| FEAT-001     | Configuração inicial do projeto                      | Pendente    | 4 - Alta           | [NOME_DO_MEMBRO_DA_EQUIPE]      | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Configurar linters, formatadores e estrutura básica de pastas.        |
| DOC-001      | Elaborar documentação inicial da arquitetura         | Pendente    | 3 - Média          | AgenteJules                     | FEAT-001           | AAAA-MM-DD      | AAAA-MM-DD                        | Detalhar o padrão arquitetural escolhido em docs/02_arquitetura.md. |
| DOC-002      | Detalhar arquitetura atual e propostas de reescrita  | Pendente    | 3 - Média          | AgenteJules                     | DOC-001            | AAAA-MM-DD      | AAAA-MM-DD                        | Expandir docs/02_arquitetura.md para cobrir estado atual e planos futuros. |
| TEST-001     | Aumentar cobertura de testes unitários/integração    | Pendente    | 4 - Alta           | [NOME_DO_MEMBRO_DA_EQUIPE]      | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Focar em pacotes críticos: network, neuron, pulse, neurochemical, synaptic. |
| REFACTOR-001 | Revisar lógica de `network.calculateInternalNeuronCounts` | Pendente    | 2 - Baixa          | AgenteJules                     | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Simplificar ou tornar mais robusto o ajuste de contagem de neurônios internos. |
| FEATURE-002  | Assegurar reprodutibilidade total com seed aleatória | Pendente    | 3 - Média          | [NOME_DO_MEMBRO_DA_EQUIPE]      | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Verificar uso de rand local vs global e configurar explicitamente se necessário. |
| DOC-003      | Revisar e atualizar `docs/03_guias/guia_estilo_codigo.md` | Pendente  | 2 - Baixa          | AgenteJules                     | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Garantir que o guia de estilo de código está completo e atual. |
| REFACTOR-003 | Padronizar tratamento de erros em `cli/orchestrator.go` | Pendente  | 3 - Média          | AgenteJules                     | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Garantir consistência na propagação e tratamento de erros. |
| DOC-004      | Atualizar placeholder de link de arquitetura em AGENTS.md | Concluído   | 1 - Muito Baixa    | AgenteJules                     | -                  | AAAA-MM-DD      | AAAA-MM-DD                        | Substituído [LINK_PARA_DOCUMENTO_DE_ARQUITETURA] por docs/02_arquitetura.md. |
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
