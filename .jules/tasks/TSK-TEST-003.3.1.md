# Tarefa: TEST-003.3.1 - Teste de integração de execução básica do modo `sim`.

**ID da Tarefa:** TEST-003.3.1
**Título Breve:** Teste básico de execução para `sim` mode.
**Descrição Completa:** Implementar um teste de integração fundamental para o modo `sim` da CLI. Este teste deve verificar se o comando `crownet sim` pode ser executado com um conjunto mínimo de parâmetros (como número de ciclos e neurônios) e se conclui com sucesso, indicando que o loop de simulação principal é funcional.
**Status:** Concluído
**Dependências (IDs):** TEST-003.3, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** 2024-07-28
**Branch Git Proposta:** test/sim-basic-it
**Critérios de Aceitação:**
- Um novo arquivo de teste de integração para o modo `sim` é criado (ou adicionado a um existente).
- Funções helper para configurar o ambiente de teste para o modo `sim` são implementadas (e.g., configurações mínimas).
- Um teste de integração é escrito que:
    - Executa o comando `sim` com parâmetros para uma simulação curta (e.g., 10 ciclos, 50 neurônios).
    - Verifica se o comando conclui sem erros (exit code 0 ou sem `error` programático).
    - Opcionalmente, verifica alguma saída de console indicando o início e o fim da simulação.
- O teste não precisa validar a lógica científica da simulação nem a persistência de dados (isso será feito em `TEST-003.3.2`).
- O teste limpa quaisquer artefatos temporários, se houver.
**Notas/Decisões:**
- Focado em manter este teste rápido e leve, validando apenas o fluxo de execução.
- Estratégia de execução programática via `cli.NewOrchestrator().Run()` utilizada.
- Teste `TestSimCommand_BasicRun` implementado em `cmd/sim_integration_test.go`.
- Configuração do teste ajustada para desabilitar monitoramento de neurônio de saída (`MonitorOutputID: -2`) para evitar falhas em configuração mínima.
