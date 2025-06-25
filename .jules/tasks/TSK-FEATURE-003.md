# Tarefa: FEATURE-003 - Implementar configuração via arquivo (TOML/YAML).

**ID da Tarefa:** FEATURE-003
**Título Breve:** Implementar configuração via arquivo (TOML/YAML).
**Descrição Completa:** Modificar o pacote `config` para permitir que os parâmetros da simulação e da CLI sejam carregados a partir de um arquivo de configuração (ex: `config.toml` ou `config.yaml`). As flags da linha de comando devem ter precedência e poder sobrescrever os valores definidos no arquivo de configuração. Uma nova flag CLI (ex: `-configFile <path>`) pode ser introduzida para especificar o local do arquivo de configuração.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** feature/config-file-loading
**Critérios de Aceitação:**
- A aplicação pode carregar `SimulationParameters` e `CLIConfig` de um arquivo TOML ou YAML.
- Uma flag `-configFile` permite especificar o caminho para o arquivo de configuração.
- Se o arquivo de configuração não for especificado ou não existir, a aplicação deve usar os valores padrão e as flags CLI.
- Valores fornecidos via flags CLI sobrescrevem os valores correspondentes do arquivo de configuração.
- A documentação (ex: `guia_interface_linha_comando.md`) é atualizada para refletir a nova funcionalidade de arquivo de configuração.
- Testes unitários são adicionados para a lógica de carregamento e merge de configurações.
**Notas/Decisões:**
- Escolher um formato de arquivo (TOML é comum em Go, YAML também é uma opção).
- Bibliotecas como `spf13/viper` podem ser consideradas para facilitar o carregamento de múltiplos formatos e a sobrescrita de valores, ou implementar manualmente usando parsers padrão.
- A ordem de precedência deve ser: Defaults -> Arquivo de Configuração -> Flags CLI.
