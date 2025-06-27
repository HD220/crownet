# Tarefa: PERF-004.2 - Setup e execução de profiling para modo `sim`.

**ID da Tarefa:** PERF-004.2
**Título Breve:** Profiling do modo `sim`.
**Descrição Completa:** Configurar e executar o profiling de CPU e memória para o modo `sim` da aplicação CrowNet. Similar a PERF-004.1, isso envolve habilitar o profiling, executar o comando `sim` com uma carga de trabalho representativa (e.g., número de ciclos, neurônios) e salvar os arquivos de perfil.
**Status:** Concluído
**Dependências (IDs):** PERF-004, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** 2024-07-28
**Branch Git Proposta:** feat/sim-profiling
**Critérios de Aceitação:**
- O método para habilitar profiling de CPU para o comando `sim` é implementado ou documentado.
- O método para habilitar profiling de memória (heap) para o comando `sim` é implementado ou documentado.
- O comando `sim` é executado com profiling habilitado usando uma configuração de teste (ex: 1000 ciclos, 200 neurônios).
- Arquivos de perfil de CPU (ex: `sim_cpu.pprof`) são gerados.
- Arquivos de perfil de memória (ex: `sim_mem.pprof`) são gerados.
- Os passos para gerar os perfis são documentados.
**Notas/Decisões:**
- As considerações sobre `runtime/pprof` vs `net/http/pprof` de PERF-004.1 também se aplicam aqui. `runtime/pprof` foi implementado em `cmd/sim.go`.
- A carga de trabalho para o modo `sim` deve ser escolhida para refletir um cenário de uso comum ou um que seja suspeito de problemas de desempenho.
- **Como Gerar Perfis (Modo Sim):**
    - Compile o binário: `go build -o crownet main.go`
    - Para perfil de CPU: Execute `./crownet sim [outras flags] --cpuprofile <arquivo_cpu.pprof>`
        - Exemplo: `./crownet sim --cycles 1000 --neurons 200 --cpuprofile sim_cpu.pprof`
    - Para perfil de Memória (Heap): Execute `./crownet sim [outras flags] --memprofile <arquivo_mem.pprof>`
        - Exemplo: `./crownet sim --cycles 1000 --neurons 200 --memprofile sim_mem.pprof`
    - Os arquivos de perfil (`.pprof`) podem ser analisados com `go tool pprof <binário> <arquivo_perfil>`.
- Geração de arquivos de perfil verificada manualmente.
