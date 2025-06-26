# Tarefa: TEST-003.1.1 - Setup para teste de integração do modo `expose` e validação básica de execução.

**ID da Tarefa:** TEST-003.1.1
**Título Breve:** Setup e teste básico de execução para `expose` mode.
**Descrição Completa:** Esta tarefa envolve a criação da infraestrutura base para os testes de integração do modo `expose` da CLI. Isso inclui definir como os testes serão executados (programaticamente via `cli.Orchestrator` ou executando o binário), como gerenciar arquivos de configuração temporários e artefatos de saída (como arquivos de pesos). Um primeiro teste básico será implementado para garantir que o comando `crownet expose` pode ser invocado com um conjunto mínimo de parâmetros válidos e que ele executa sem erros, verificando o código de saída e possivelmente alguma saída de console esperada.
**Status:** Pendente
**Dependências (IDs):** TEST-003.1, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-expose-setup
**Critérios de Aceitação:**
- Um novo arquivo de teste (e.g., `cmd/expose_integration_test.go` ou similar) é criado.
- Funções helper para setup de ambiente de teste são implementadas (e.g., criação de diretórios temporários, escrita de arquivos de configuração TOML básicos).
- Um teste de integração inicial para o modo `expose` é escrito que:
    - Configura os parâmetros CLI necessários (e.g., `--epochs 1`, `--cyclesPerPattern 1`, `--neurons 10`, um `--weightsFile` temporário).
    - Executa o comando `expose`.
    - Verifica se o comando conclui com sucesso (exit code 0, sem `error` se executado programaticamente).
    - Verifica se alguma saída de console esperada (e.g., "Epoch 1/1 completed") é produzida (opcional, pode ser difícil de capturar/validar de forma robusta).
- O teste garante a limpeza de quaisquer arquivos temporários criados.
**Notas/Decisões:**
- Decidir sobre a estratégia de execução: invocar `cmd.Execute()` ou `cli.NewOrchestrator().Run()`. A segunda opção pode oferecer mais controle para testes.
- A validação de arquivos de pesos será feita na tarefa `TEST-003.1.2`. Este foca no fluxo de execução básico.
- Utilizar o pacote `os/exec` se for executar o binário compilado, ou chamadas diretas de função se for programático. Testes programáticos são geralmente mais fáceis de depurar e controlar.
