# Tarefa: TEST-003 - Desenvolver testes de integração para cenários chave.

**ID da Tarefa:** TEST-003
**Título Breve:** Desenvolver testes de integração para cenários chave.
**Descrição Completa:** Desenvolver testes de integração que cubram os cenários de ponta a ponta para os principais modos de operação da CLI (`expose`, `observe`, `sim`). Estes testes devem validar o fluxo completo da aplicação, desde a entrada de parâmetros da CLI até a saída esperada (ex: arquivos de pesos gerados, logs no console, dados no SQLite, padrões de ativação de neurônios de saída).
**Status:** Pendente
**Dependências (IDs):** TEST-002
**Complexidade (1-5):** 4
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-cli-scenarios
**Critérios de Aceitação:**
- Testes de integração implementados para o modo `expose`, validando a criação/atualização de arquivos de pesos.
- Testes de integração implementados para o modo `observe`, validando a correta apresentação de padrões de saída para inputs conhecidos e pesos carregados.
- Testes de integração implementados para o modo `sim`, validando a execução por um número de ciclos e, opcionalmente, a criação de logs no SQLite.
- Os testes utilizam configurações de simulação específicas e, se necessário, dados de entrada mockados ou predefinidos.
- Os testes verificam os resultados esperados e falham apropriadamente em caso de desvios.
**Notas/Decisões:**
- Estes testes podem ser mais lentos que os unitários e podem requerer a execução do binário compilado ou o uso de `cli.NewOrchestrator().Run()` programaticamente.
- Considerar o uso de arquivos de fixtures para pesos de entrada ou configurações complexas.
- A validação de dados no SQLite pode envolver a consulta direta ao banco de dados de teste.
