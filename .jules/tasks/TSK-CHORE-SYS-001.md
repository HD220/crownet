# Tarefa: CHORE-SYS-001 - Corrigir processo de gerenciamento de tarefas e docs do agente

**ID da Tarefa:** CHORE-SYS-001
**Título Breve:** Corrigir processo de tarefas e docs do agente
**Descrição Completa:** Esta tarefa visa corrigir uma omissão no processo de gerenciamento de tarefas do Agente Jules. Especificamente, o agente não estava criando os arquivos de detalhe de tarefa em `.jules/tasks/` para novas tarefas identificadas e adicionadas ao índice `.jules/TASKS.md`. Além disso, os documentos de orientação do agente (`AGENTS.md` e `.jules/AGENT_WORKFLOW.md`) não continham instruções explícitas sobre este requisito, nem sobre a necessidade de ler os arquivos de detalhe ao iniciar uma tarefa.
Esta tarefa inclui:
1. Criação retroativa dos arquivos de detalhe para tarefas recentemente adicionadas.
2. Padronização dos links no `.jules/TASKS.md` para o formato `TSK-PREFIX-NNN.md`.
3. Atualização do `AGENTS.md` (raiz) para incluir uma seção sobre "Gerenciamento de Tarefas Detalhadas".
4. Atualização do `.jules/AGENT_WORKFLOW.md` para integrar instruções sobre a leitura e criação de arquivos de detalhe de tarefa no ciclo de trabalho.
**Status:** Em Andamento (será Concluído ao final deste ciclo)
**Dependências (IDs):** -
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P0 (meta-tarefa para corrigir processo)
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** 2024-07-27
**Branch Git Proposta:** chore/fix-task-management-process
**Critérios de Aceitação:**
- Todos os arquivos de detalhe para tarefas recentemente identificadas são criados em `.jules/tasks/` com conteúdo apropriado.
- Os links no arquivo `.jules/TASKS.md` para essas tarefas são corrigidos para o formato `TSK-PREFIX-NNN.md`.
- O arquivo `AGENTS.md` (raiz) é criado ou atualizado com instruções claras sobre o gerenciamento de arquivos de tarefa.
- O arquivo `.jules/AGENT_WORKFLOW.md` (prompt do agente) é atualizado (ou o texto para atualização é fornecido) com as novas instruções de processo.
**Notas/Decisões:**
- Tarefa crucial para a integridade e rastreabilidade do processo de desenvolvimento do agente.
- Assegura que o agente siga o sistema de gerenciamento de tarefas pretendido de forma completa.
