# Tarefa: TEST-003.3 - Implementar testes de integração para CLI modo `sim`.

**ID da Tarefa:** TEST-003.3
**Título Breve:** Testes de integração para CLI `sim` mode.
**Descrição Completa:** Desenvolver testes de integração para o modo `sim` da CLI. O objetivo é validar que a simulação geral pode ser executada por um número especificado de ciclos com um conjunto de parâmetros e, opcionalmente, que os logs são corretamente persistidos em um arquivo SQLite se configurado.
**Status:** Pendente
**Dependências (IDs):** TEST-003, TEST-002
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-sim-mode
**Critérios de Aceitação:**
- Um teste de integração é implementado que executa o comando `crownet sim` com parâmetros básicos (ex: número de ciclos, número de neurônios).
- O teste verifica se a simulação roda até o fim sem erros.
- Se o logging SQLite for habilitado via flags, o teste verifica se o arquivo de banco de dados é criado e se não está vazio.
- Opcionalmente, o teste pode realizar uma consulta simples ao BD para verificar a presença de tabelas esperadas (ex: `NetworkSnapshots`).
- O teste limpa quaisquer artefatos criados (ex: arquivo de BD temporário) após sua execução.
**Notas/Decisões:**
- Usar um número pequeno de ciclos e neurônios para manter o teste rápido.
- A validação do conteúdo do BD pode ser superficial, focando na criação e estrutura básica, não nos valores detalhados dos logs.
- Considerar o uso de um diretório temporário para o arquivo de BD.
