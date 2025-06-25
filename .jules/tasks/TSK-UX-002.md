# Tarefa: UX-002 - Melhorar representação visual do output no modo 'observe'.

**ID da Tarefa:** UX-002
**Título Breve:** Melhorar visualização do output no modo 'observe'.
**Descrição Completa:** Aprimorar a forma como o padrão de ativação dos neurônios de saída é exibido no modo `observe`. Atualmente, é uma lista de valores de `AccumulatedPulse`. A proposta é adicionar uma representação visual mais intuitiva, como um ASCII art que mapeie os níveis de ativação dos 10 neurônios de saída para uma grade (ex: 2x5 ou similar) que possa se assemelhar visualmente a um dígito ou a um padrão reconhecível.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** feature/observe-mode-visual-output
**Critérios de Aceitação:**
- O modo `observe` exibe, além dos valores numéricos de ativação, uma representação ASCII art (ou similar baseada em texto) do padrão de output.
- A representação visual mapeia os 10 neurônios de saída para uma grade espacial (ex: 2 linhas de 5 neurônios, ou uma forma que lembre um display de 7 segmentos simplificado se os neurônios de saída tiverem uma semântica espacial).
- Diferentes níveis de ativação podem ser representados por diferentes caracteres ASCII (ex: ' ', '.', 'o', 'O', '#').
- A nova visualização é clara e ajuda a interpretar rapidamente o padrão de resposta da rede.
**Notas/Decisões:**
- A forma exata da grade visual (2x5, etc.) e o mapeamento de neurônios para posições na grade precisam ser definidos. Isso pode depender se há uma correspondência implícita entre os IDs dos neurônios de saída e "pixels" de um dígito.
- Se os 10 neurônios de saída não tiverem uma organização espacial predefinida para formar um dígito, a visualização pode ser uma simples barra de ativação ou um heatmap textual.
- A documentação (`guia_interface_linha_comando.md`) deve ser atualizada para descrever a nova saída visual.
- O objetivo é tornar o output do modo `observe` mais imediatamente interpretável.
