# Tarefa: CHORE-003 - Criar Makefile para build, lint e teste automatizados.

**ID da Tarefa:** CHORE-003
**Título Breve:** Criar Makefile para build, lint e teste automatizados.
**Descrição Completa:** Desenvolver um `Makefile` na raiz do projeto para simplificar e padronizar os processos de build, linting e execução de testes. Isso melhorará a experiência do desenvolvedor e facilitará a integração contínua.
**Status:** Pendente
**Dependências (IDs):** FEAT-001.2 (configuração do golangci-lint)
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** chore/add-makefile
**Critérios de Aceitação:**
- Um arquivo `Makefile` é criado na raiz do projeto.
- O Makefile inclui pelo menos os seguintes targets:
    - `build`: Compila o binário `crownet`.
    - `lint`: Executa `golangci-lint run ./...` usando o arquivo de configuração `.golangci.yml`.
    - `test`: Executa `go test ./... -v` (com verbose).
    - `clean`: Remove o binário compilado e arquivos de cache de teste (opcional).
    - `all`: Executa `lint`, `test`, e `build` (ou uma sequência lógica).
- Os targets funcionam corretamente no ambiente de desenvolvimento esperado.
- O `README.md` ou `CONTRIBUTING.md` é atualizado para mencionar o uso do Makefile.
**Notas/Decisões:**
- O target `build` deve produzir o executável `crownet` (ou nome similar) na raiz ou em um diretório `bin/`.
- Garantir que o target `lint` falhe se o linter encontrar problemas.
- Garantir que o target `test` falhe se algum teste não passar.
