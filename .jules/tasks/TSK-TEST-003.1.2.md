# Tarefa: TEST-003.1.2 - Validar manipulação de arquivo de pesos no teste de integração do modo `expose`.

**ID da Tarefa:** TEST-003.1.2
**Título Breve:** Validação de arquivo de pesos para `expose` mode.
**Descrição Completa:** Estender os testes de integração do modo `expose` para validar especificamente a criação e modificação de arquivos de pesos sinápticos. O teste deve verificar se um novo arquivo de pesos é criado quando não existe um, e se um arquivo existente é atualizado após a execução do treinamento.
**Status:** Pendente
**Dependências (IDs):** TEST-003.1, TEST-003.1.1, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-expose-weights
**Critérios de Aceitação:**
- Um teste de integração verifica se o comando `crownet expose` cria um novo arquivo de pesos no caminho especificado se ele não existir.
    - O teste verifica a existência do arquivo e se ele não está vazio.
    - O teste verifica se o arquivo é um JSON válido.
- Um teste de integração verifica se o comando `crownet expose` modifica um arquivo de pesos existente.
    - O teste pode criar um arquivo de pesos de fixture simples, obter seu timestamp de modificação, executar `expose`, e então verificar se o timestamp de modificação do arquivo mudou e se o conteúdo foi alterado.
- Os testes utilizam o setup e as funções helper definidas em `TEST-003.1.1`.
- Os testes limpam os arquivos de pesos temporários criados.
**Notas/Decisões:**
- A validação do *conteúdo* dos pesos (i.e., se o aprendizado ocorreu corretamente) é complexa e fora do escopo deste teste de integração de manipulação de arquivo. Focar na existência, formato e modificação do arquivo.
- Usar arquivos de pesos pequenos e configurações de simulação rápidas para os testes.
