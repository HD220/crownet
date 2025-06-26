# Tarefa: TEST-003.2.2 - Validar formato/conteúdo da saída do console no teste de integração do modo `observe`.

**ID da Tarefa:** TEST-003.2.2
**Título Breve:** Validação de saída do console para `observe` mode.
**Descrição Completa:** Estender os testes de integração do modo `observe` para validar o formato e, superficialmente, o conteúdo da saída de console que representa o padrão de ativação dos neurônios de saída. O teste deve verificar se a saída é gerada e se adere à estrutura esperada.
**Status:** Pendente
**Dependências (IDs):** TEST-003.2, TEST-003.2.1, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-observe-output
**Critérios de Aceitação:**
- Um teste de integração executa o comando `crownet observe` usando o setup e o arquivo de pesos de fixture de `TEST-003.2.1`.
- O teste captura a saída padrão (stdout) do comando.
- O teste verifica se a saída contém elementos esperados, como:
    - O cabeçalho "Output Neuron Activation Pattern".
    - Linhas correspondentes ao número de neurônios de saída configurados na simulação associada aos pesos de fixture.
    - Formato de barra ASCII (e.g., `[|||   ]`).
- O teste não precisa validar a *exatidão* do padrão de ativação em relação à lógica da rede neural, mas sim a presença e formatação da saída.
**Notas/Decisões:**
- Capturar stdout pode exigir o uso de `os/exec` e pipes, ou se estiver testando programaticamente, redirecionando `os.Stdout` temporariamente.
- A validação pode usar `strings.Contains` ou expressões regulares para verificar a estrutura da saída.
- Este teste complementa `TEST-003.2.1` focando na apresentação da informação.
