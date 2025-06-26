# Tarefa: TEST-003.2 - Implementar testes de integração para CLI modo `observe`.

**ID da Tarefa:** TEST-003.2
**Título Breve:** Testes de integração para CLI `observe` mode.
**Descrição Completa:** Desenvolver testes de integração para o modo `observe` da CLI. Estes testes devem garantir que, ao carregar um arquivo de pesos conhecido e apresentar um padrão de entrada específico (e.g., um dígito), a saída da rede (padrão de ativação dos neurônios de saída) é consistente e conforme esperado.
**Status:** Pendente
**Dependências (IDs):** TEST-003, TEST-002
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-observe-mode
**Critérios de Aceitação:**
- Um teste de integração é implementado que executa o comando `crownet observe` com parâmetros válidos, incluindo um arquivo de pesos de fixture.
- O teste utiliza um padrão de entrada predefinido (ex: dígito '1').
- O teste verifica se a saída no console (representação do padrão de ativação) é gerada.
- Opcionalmente, o teste pode tentar parsear a saída do console para verificar se o número de neurônios de saída corresponde ao esperado ou se o formato geral está correto.
- O teste limpa quaisquer artefatos criados, se houver.
**Notas/Decisões:**
- Requer um arquivo de pesos de fixture (previamente treinado ou construído manualmente) que produza uma resposta conhecida para um dado input.
- A validação exata do "padrão de ativação correto" pode ser complexa; o teste pode focar em aspectos mais simples como a formatação da saída ou a ausência de erros durante a execução.
- Este teste foca na integração da CLI, carregamento de pesos e fluxo de observação.
