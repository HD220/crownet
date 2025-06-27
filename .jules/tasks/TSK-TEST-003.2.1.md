# Tarefa: TEST-003.2.1 - Setup para teste de integração do modo `observe` e execução básica.

**ID da Tarefa:** TEST-003.2.1
**Título Breve:** Setup e teste básico de execução para `observe` mode.
**Descrição Completa:** Configurar o ambiente necessário para os testes de integração do modo `observe` da CLI. Isso inclui a preparação de um arquivo de pesos sinápticos de fixture que produza uma saída conhecida para um dado padrão de entrada. Implementar um teste básico que executa o comando `crownet observe` com parâmetros mínimos válidos (usando o arquivo de pesos de fixture e um dígito específico) e verifica se o comando conclui com sucesso (exit code 0 ou sem erro programático) e se alguma saída inicial esperada é impressa no console.
**Status:** Concluído
**Dependências (IDs):** TEST-003.2, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** 2024-07-28
**Branch Git Proposta:** test/observe-basic-it
**Critérios de Aceitação:**
- Um arquivo de pesos de fixture (e.g., `fixture_weights.json`) é criado e adicionado ao repositório (possivelmente em um diretório `testdata`). Este arquivo deve ser simples e produzir uma resposta distinguível para um dígito de entrada específico.
- Um novo arquivo de teste de integração para o modo `observe` é criado.
- Funções helper para carregar o arquivo de pesos de fixture e configurar os parâmetros da CLI para o modo `observe` são implementadas.
- Um teste de integração inicial é escrito que:
    - Executa o comando `observe` com o arquivo de pesos de fixture e um dígito de entrada (ex: '1').
    - Verifica se o comando conclui com sucesso.
    - Verifica se o console imprime alguma mensagem inicial esperada (e.g., "Observing Network Response...").
- O teste limpa quaisquer artefatos temporários, se houver (embora `observe` geralmente não crie arquivos).
**Notas/Decisões:**
- O arquivo de pesos de fixture (`testdata/fixture_observe_weights.json`) foi criado com uma estrutura JSON simples e válida.
- A validação detalhada do padrão de ativação no console será feita em `TEST-003.2.2`. Este foca no fluxo de execução e setup.
- Teste `TestObserveCommand_BasicRun` implementado em `cmd/observe_integration_test.go`.
- Corrigido o caminho para o arquivo de fixture no teste para `../testdata/`.
- O teste executa o modo `observe` programaticamente e verifica a ausência de erros.
