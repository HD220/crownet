# Tarefa: TEST-003.3.2 - Validar logging SQLite no teste de integração do modo `sim`.

**ID da Tarefa:** TEST-003.3.2
**Título Breve:** Validação de logging SQLite para `sim` mode.
**Descrição Completa:** Estender os testes de integração do modo `sim` para validar a funcionalidade de logging em SQLite. O teste deve executar uma simulação curta com o logging habilitado para um arquivo de banco de dados temporário e, em seguida, verificar se o arquivo foi criado e contém a estrutura de tabelas esperada.
**Status:** Pendente
**Dependências (IDs):** TEST-003.3, TEST-003.3.1, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-sim-sqlite
**Critérios de Aceitação:**
- Um teste de integração executa o comando `crownet sim` com logging SQLite habilitado (usando a flag `--dbPath` para um arquivo temporário).
- O teste verifica se o arquivo de banco de dados SQLite é criado no caminho especificado.
- O teste verifica se o arquivo de banco de dados não está vazio.
- O teste se conecta ao banco de dados criado e verifica a existência das tabelas principais (e.g., `NetworkSnapshots`, `NeuronStates`).
- O teste não precisa validar o conteúdo detalhado das tabelas, apenas sua presença e, possivelmente, se contêm alguma linha.
- O teste limpa o arquivo de banco de dados temporário após a execução.
**Notas/Decisões:**
- Usar um nome de arquivo de BD temporário e garantir sua remoção (e.g., usando `t.TempDir()` ou `os.Remove`).
- A interação com o BD pode ser feita usando o pacote `database/sql` e o driver `sqlite3`.
- Manter a simulação curta para que o teste seja rápido.
