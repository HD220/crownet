# Tarefa: SEC-001 - Analisar e validar caminhos de arquivo para E/S.

**ID da Tarefa:** SEC-001
**Título Breve:** Analisar e validar caminhos de arquivo para E/S.
**Descrição Completa:** Revisar o código que lida com caminhos de arquivo fornecidos pelo usuário através das flags `-weightsFile` e `-dbPath`. O objetivo é garantir que os caminhos sejam tratados de forma segura para prevenir vulnerabilidades como Path Traversal, especialmente se a aplicação fosse executada em um ambiente com permissões mais elevadas ou se os nomes dos arquivos pudessem ser influenciados por fontes não confiáveis.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** security/filepath-validation
**Critérios de Aceitação:**
- A lógica de manipulação dos caminhos de `-weightsFile` e `-dbPath` (principalmente em `config/config.go` e `cli/orchestrator.go`) é revisada.
- Os caminhos são "limpos" usando `filepath.Clean` para normalizá-los.
- Verifica-se se os caminhos resultantes são relativos ao diretório de trabalho esperado ou se são caminhos absolutos seguros.
- Considerar se há necessidade de restringir a escrita/leitura de arquivos a um subdiretório específico do projeto por padrão, a menos que um caminho absoluto seja explicitamente permitido e validado.
- Documentar quaisquer alterações ou decisões sobre a segurança do tratamento de caminhos.
**Notas/Decisões:**
- O risco de exploração de Path Traversal em uma aplicação CLI local é geralmente baixo, mas é uma boa prática de higiene de segurança.
- Foco em garantir que a aplicação não escreva ou leia arquivos em locais inesperados ou perigosos.
- A validação pode envolver a verificação de que o caminho resolvido está dentro de um diretório base permitido.
