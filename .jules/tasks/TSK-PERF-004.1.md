# Tarefa: PERF-004.1 - Setup e execução de profiling para modo `expose`.

**ID da Tarefa:** PERF-004.1
**Título Breve:** Profiling do modo `expose`.
**Descrição Completa:** Configurar e executar o profiling de CPU e memória para o modo `expose` da aplicação CrowNet. Isso envolve determinar a melhor forma de habilitar o profiling (via flags pprof, `runtime/pprof` ou `net/http/pprof`), executar o comando `expose` com uma carga de trabalho representativa e salvar os arquivos de perfil resultantes.
**Status:** Pendente
**Dependências (IDs):** PERF-004, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** perf/profile-expose-mode
**Critérios de Aceitação:**
- O método para habilitar profiling de CPU para o comando `expose` é implementado ou documentado.
- O método para habilitar profiling de memória (heap) para o comando `expose` é implementado ou documentado.
- O comando `expose` é executado com profiling habilitado usando uma configuração de teste (ex: 10 épocas, 50 neurônios, 10 ciclos/padrão).
- Arquivos de perfil de CPU (ex: `expose_cpu.pprof`) são gerados.
- Arquivos de perfil de memória (ex: `expose_mem.pprof`) são gerados.
- Os passos para gerar os perfis são documentados se não forem triviais.
**Notas/Decisões:**
- Para profiling de CPU, pode ser necessário adicionar código para iniciar/parar o profiling em torno da execução do modo `expose`.
- Para profiling de memória, um heap profile pode ser escrito em um ponto específico (e.g., no final da execução).
- A carga de trabalho deve ser significativa o suficiente para gerar dados de perfil úteis, mas não excessivamente longa para a fase de coleta.
- Considerar se o profiling via `net/http/pprof` é viável/desejável para um CLI, ou se `runtime/pprof` é mais apropriado.
