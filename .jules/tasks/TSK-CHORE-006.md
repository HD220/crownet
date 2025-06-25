# Tarefa: CHORE-006 - Configurar GitHub Issue Templates para bugs e feature requests.

**ID da Tarefa:** CHORE-006
**Título Breve:** Configurar GitHub Issue Templates.
**Descrição Completa:** Criar templates para issues do GitHub para padronizar o reporte de bugs e a solicitação de novas funcionalidades. Isso ajudará a garantir que os contribuidores forneçam todas as informações necessárias ao criar uma issue, facilitando a triagem e o trabalho subsequente.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** chore/setup-issue-templates
**Critérios de Aceitação:**
- Um diretório `.github/ISSUE_TEMPLATE/` é criado na raiz do projeto.
- Pelo menos dois arquivos de template Markdown são criados dentro deste diretório:
    - `bug_report.md`: Template para reportar bugs (ex: com seções para descrição, passos para reproduzir, comportamento esperado vs. atual, ambiente).
    - `feature_request.md`: Template para solicitar novas funcionalidades (ex: com seções para descrição do problema, solução proposta, alternativas consideradas).
- (Opcional) Um arquivo `config.yml` pode ser adicionado em `.github/ISSUE_TEMPLATE/` para customizar o seletor de templates de issue no GitHub.
- Os templates são claros e guiam o usuário a fornecer informações úteis.
**Notas/Decisões:**
- Consultar a documentação do GitHub sobre "Configuring issue templates for your repository".
- Manter os templates simples mas eficazes.
- Considerar adicionar labels padrão sugeridas nos templates.
