# Tarefa: REFACTOR-007 - Revisão e adição de nil checks para *config.SimulationParameters.

**ID da Tarefa:** REFACTOR-007
**Título Breve:** Revisão e adição de nil checks para SimParams.
**Descrição Completa:** Realizar uma varredura no código-fonte para identificar todas as funções e métodos que recebem ou acessam `*config.SimulationParameters` (ou `*config.AppConfig` para então acessar `SimParams`). Adicionar verificações de `nil` robustas no início dessas funções/métodos para prevenir panics de desreferência de ponteiro nulo, especialmente em caminhos de código onde `SimParams` poderia, hipoteticamente ou devido a erros de programação futuros, não estar inicializado.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** refactor/nil-checks-simparams
**Critérios de Aceitação:**
- Funções críticas nos pacotes `network`, `neuron`, `pulse`, `neurochemical`, `space`, `synaptic`, `cli` que utilizam `*config.SimulationParameters` são revisadas.
- Verificações de `if params == nil { return err_ou_panic_controlado }` são adicionadas onde apropriado no início das funções.
- A estratégia de tratamento para um `SimParams` nulo é consistente (ex: retornar erro, logar e usar defaults seguros se possível, ou panic controlado com mensagem clara). Para a maioria das funções da simulação, um `SimParams` nulo é provavelmente um erro fatal de setup.
- Testes unitários (se existentes para as funções modificadas) são verificados para não serem afetados negativamente, ou são ajustados se necessário.
**Notas/Decisões:**
- Embora `NewCrowNet` e `NewAppConfig` inicializem `SimParams`, esta tarefa visa adicionar uma camada extra de robustez contra erros de programação ou estados inesperados.
- O foco é em funções públicas de pacotes e métodos de structs importantes.
- Para funções que não podem operar sem `SimParams`, retornar um erro é preferível a um panic descontrolado.
- Considerar se algumas funções poderiam operar com `DefaultSimulationParameters()` como fallback em caso de `nil`, embora isso possa mascarar problemas de configuração. Geralmente, falhar rápido é melhor.
