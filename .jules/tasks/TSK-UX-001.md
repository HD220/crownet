# Tarefa: UX-001 - Revisar a usabilidade da CLI: mensagens de ajuda, erros e feedback.

**ID da Tarefa:** UX-001
**Título Breve:** Revisar usabilidade da CLI (ajuda, erros, feedback).
**Descrição Completa:** Realizar uma revisão da experiência do usuário (UX) da interface de linha de comando (CLI) do CrowNet. Isso envolve executar a aplicação com diversas combinações de flags (corretas e incorretas), analisar a clareza e utilidade das mensagens de ajuda (saída de `-h`), a informatividade das mensagens de erro, e a qualidade do feedback fornecido no console durante a execução dos diferentes modos.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** ux/cli-feedback-review
**Critérios de Aceitação:**
- A saída da flag de ajuda (`-h` ou `--help`) é clara, completa e lista todas as flags globais e específicas de modo com seus padrões.
- As mensagens de erro para flags inválidas, valores de flag incorretos ou combinações de flags problemáticas são informativas e guiam o usuário para a correção.
- O feedback no console durante a execução dos modos `sim`, `expose`, e `observe` é suficiente para entender o progresso e o estado da simulação.
- A terminologia usada nas mensagens é consistente com a documentação (`guia_interface_linha_comando.md`).
- Quaisquer inconsistências ou áreas de melhoria são documentadas (possivelmente como sub-tarefas ou issues).
**Notas/Decisões:**
- Comparar a experiência real com o documentado em `docs/03_guias/guia_interface_linha_comando.md`.
- O objetivo é garantir que a CLI seja o mais intuitiva e fácil de usar possível para novos usuários e para depuração.
- Considerar se a verbosidade do output é adequada ou se seriam úteis flags de `-verbose` ou `-quiet`.
