# Tarefa: FEATURE-004 - Desenvolver script/utilitário para exportar/visualizar dados do SQLite log.

**ID da Tarefa:** FEATURE-004
**Título Breve:** Desenvolver utilitário para exportar/visualizar log SQLite.
**Descrição Completa:** Criar um script ou utilitário simples (em Go ou Python) que possa ler o arquivo de banco de dados SQLite gerado pelo modo `sim` (`-dbPath`) e exportar os dados para formatos mais acessíveis (como CSV) ou gerar visualizações básicas/resumos textuais. O objetivo é facilitar a análise dos dados de simulação registrados.
**Status:** Pendente
**Dependências (IDs):**
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** feature/sqlite-log-exporter
**Critérios de Aceitação:**
- O utilitário pode se conectar a um arquivo de banco de dados SQLite do CrowNet.
- O utilitário oferece opções para exportar dados das tabelas `NetworkSnapshots` e `NeuronStates` para formato CSV.
- (Opcional) O utilitário pode gerar um resumo textual simples de uma simulação (ex: número de ciclos, variação de neuroquímicos, contagem média de disparos).
- (Opcional Estendido) O utilitário pode gerar plots básicos (ex: níveis de neuroquímicos ao longo do tempo, atividade de um neurônio específico).
- O utilitário é documentado com instruções de uso.
**Notas/Decisões:**
- A escolha entre Go ou Python para o script pode depender da facilidade de uso de bibliotecas de plotting ou manipulação de dados. Python (com pandas, matplotlib/seaborn) é forte nisso. Go pode ser usado para exportação CSV simples.
- Foco inicial na exportação para CSV, que é universalmente útil.
- A funcionalidade de logging para SQLite é descrita em RF-PERSIST-003.
- O utilitário pode ser uma ferramenta CLI separada ou um novo modo/flag na aplicação CrowNet principal. Um script separado é provavelmente mais simples inicialmente.
