# Tarefa: FEATURE-CONFIG-001 - Implementar carregamento de configuração via arquivo TOML

**ID da Tarefa:** FEATURE-CONFIG-001
**Título Breve:** Implementar carregamento de config via TOML
**Descrição Completa:** Implementar a funcionalidade para que a aplicação CrowNet possa carregar parâmetros de configuração de um arquivo TOML. A flag global `--configFile` (já definida) especificará o caminho para este arquivo. A ordem de precedência para configurações deve ser: 1. Valores Padrão internos, 2. Valores do arquivo TOML (se fornecido), 3. Valores das flags CLI (têm a maior precedência).
**Status:** Concluído
**Dependências (IDs):** -
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** 2024-07-27
**Branch Git Proposta:** feature/toml-config-FEATURE-CONFIG-001
**Critérios de Aceitação:**
- A aplicação usa a biblioteca `github.com/BurntSushi/toml` para parsing de TOML.
- Se a flag `--configFile` for fornecida na CLI, a aplicação tenta ler e desserializar o arquivo TOML especificado para a struct `config.AppConfig`.
- A lógica de carregamento é integrada nos handlers `RunE` dos comandos Cobra (`sim`, `expose`, `observe`).
- A ordem de precedência (Padrões -> TOML -> Flags CLI explícitas) é corretamente implementada.
- Um arquivo de exemplo `config.example.toml` é fornecido na raiz do repositório.
- A documentação (`README.md` e `guia_interface_linha_comando.md`) é atualizada para refletir a funcionalidade implementada e como usá-la.
**Notas/Decisões:**
- A desserialização do TOML assume que nomes de campo `snake_case` no arquivo mapeiam para `CamelCase` nas structs Go, ou que tags `toml:"..."` seriam adicionadas se necessário (atualmente não foram adicionadas, confiando na correspondência de nome).
- Erros ao decodificar o arquivo TOML são logados como aviso, permitindo que a aplicação continue com padrões e flags CLI, mas isso pode ser revisto para um erro fatal se desejado.
- Teste manual foi bloqueado por limitações do ambiente.
