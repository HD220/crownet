# Tarefa: PERF-004.1 - Setup e execução de profiling para modo `expose`.

**ID da Tarefa:** PERF-004.1
**Título Breve:** Profiling do modo `expose`.
**Descrição Completa:** Configurar e executar o profiling de CPU e memória para o modo `expose` da aplicação CrowNet. Isso envolve determinar a melhor forma de habilitar o profiling (via flags pprof, `runtime/pprof` ou `net/http/pprof`), executar o comando `expose` com uma carga de trabalho representativa e salvar os arquivos de perfil resultantes.
**Status:** Concluído
**Dependências (IDs):** PERF-004, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** 2024-07-28
**Branch Git Proposta:** feat/expose-profiling
**Critérios de Aceitação:**
- O método para habilitar profiling de CPU para o comando `expose` é implementado ou documentado.
- O método para habilitar profiling de memória (heap) para o comando `expose` é implementado ou documentado.
- O comando `expose` é executado com profiling habilitado usando uma configuração de teste (ex: 10 épocas, 50 neurônios, 10 ciclos/padrão).
- Arquivos de perfil de CPU (ex: `expose_cpu.pprof`) são gerados.
- Arquivos de perfil de memória (ex: `expose_mem.pprof`) são gerados.
- Os passos para gerar os perfis são documentados se não forem triviais.
**Notas/Decisões:**
- Para profiling de CPU, código adicionado para iniciar/parar o profiling em `cmd/expose.go` usando `runtime/pprof`.
- Para profiling de memória, um heap profile é escrito no final da execução bem-sucedida em `cmd/expose.go`.
- A carga de trabalho para teste manual foi mínima (1 epoch, 1 cycle/pattern, 50 neurons) para verificar a geração dos arquivos.
- `runtime/pprof` foi implementado com as flags `--cpuprofile` e `--memprofile` no comando `expose`.
- **Como Gerar Perfis (Modo Expose):**
    - Compile o binário: `go build -o crownet main.go`
    - Para perfil de CPU: Execute `./crownet expose [outras flags] --cpuprofile <arquivo_cpu.pprof>`
        - Exemplo: `./crownet expose --weightsFile weights.json --epochs 10 --neurons 100 --cpuprofile expose_cpu.pprof`
    - Para perfil de Memória (Heap): Execute `./crownet expose [outras flags] --memprofile <arquivo_mem.pprof>`
        - Exemplo: `./crownet expose --weightsFile weights.json --epochs 10 --neurons 100 --memprofile expose_mem.pprof`
    - Os arquivos de perfil (`.pprof`) podem ser analisados com `go tool pprof <binário> <arquivo_perfil>`.
- Geração de arquivos de perfil verificada manualmente.
