# Funcionalidade: Persistência de Dados (MVP)

Esta funcionalidade descreve como os dados importantes da simulação CrowNet são persistidos e carregados no MVP, especificamente os pesos sinápticos e, opcionalmente, o estado completo da rede.

## F6: Persistência

### F6.1: Salvar e Carregar Pesos Sinápticos
*   **Formato:** Os pesos sinápticos aprendidos pela rede são salvos e carregados em formato JSON.
*   **Estrutura do JSON:** O arquivo JSON contém um objeto onde cada chave é o ID de um neurônio de origem (como string). O valor para cada ID de origem é outro objeto, onde cada chave é o ID de um neurônio de destino (como string) e o valor é o peso sináptico (float).
    ```json
    {
      "0": { // ID do neurônio de origem
        "1": 0.7512,
        "2": -0.3200,
        // ... outros destinos para o neurônio 0
        "99": 0.0543
      },
      "1": { // ID do neurônio de origem
        "0": 0.6800,
        "2": 0.9876,
        // ...
      }
      // ... mais neurônios origem
    }
    ```
*   **Uso:**
    *   No modo `expose`, os pesos finais são salvos no arquivo especificado por `-weightsFile` ao final do processo de treinamento. Se um arquivo de pesos já existir no início e for válido, ele é carregado como ponto de partida; caso contrário, a rede começa com pesos iniciais aleatórios.
    *   No modo `observe`, os pesos são carregados do arquivo especificado por `-weightsFile`. A operação falhará se o arquivo não for encontrado ou for inválido.
    *   No modo `sim`, os pesos podem ser opcionalmente carregados de `-weightsFile`. Este modo não salva os pesos automaticamente ao final.

### F6.2: Logging Opcional do Estado da Rede em SQLite
*   **Propósito:** Para análise detalhada e depuração, o estado completo da rede pode ser salvo em um banco de dados SQLite. Isso é útil para os modos `sim` e `expose` (se configurado).
*   **Conteúdo do Snapshot:** Um snapshot da rede no banco de dados consiste em:
    *   **Tabela `NetworkSnapshots`:**
        *   `SnapshotID` (chave primária)
        *   `CycleCount` (ciclo da simulação em que o snapshot foi tirado)
        *   `Timestamp`
        *   `CortisolLevel`
        *   `DopamineLevel`
    *   **Tabela `NeuronStates`:** (um registro por neurônio, por snapshot)
        *   `StateID` (chave primária)
        *   `SnapshotID` (chave estrangeira para `NetworkSnapshots`)
        *   `NeuronID`
        *   `Position0` a `Position15` (coordenadas do neurônio no espaço 16D)
        *   `Velocity0` a `Velocity15` (componentes do vetor de velocidade do neurônio)
        *   `Type` (tipo do neurônio: Excitatory, Inhibitory, etc.)
        *   `State` (estado atual: Resting, Firing, etc.)
        *   `AccumulatedPulse`
        *   `BaseFiringThreshold`
        *   `CurrentFiringThreshold`
        *   `LastFiredCycle`
        *   `CyclesInCurrentState`
    *   Nota: Os pesos sinápticos em si não são tipicamente duplicados no SQLite a cada snapshot, pois o arquivo JSON é o meio primário para sua persistência. O foco do SQLite é o estado dinâmico.
*   **Controle:**
    *   Habilitado pelo fornecimento de um caminho de arquivo no flag `-dbPath <string>`. O arquivo de banco de dados é recriado a cada execução que utiliza esta opção.
    *   A frequência do logging é controlada pelo flag `-saveInterval <int>`, que especifica de quantos em quantos ciclos o estado da rede é salvo no banco de dados. Se `saveInterval` for 0 ou negativo, o logging pode ocorrer apenas no final ou não ocorrer, dependendo da implementação.
*   **Uso:**
    *   Principalmente no modo `sim` para registrar a evolução da rede sob dinâmicas gerais.
    *   Pode ser usado no modo `expose` para capturar a trajetória de aprendizado.
    *   Os dados no SQLite são destinados à análise offline usando ferramentas de banco de dados ou scripts.

## Considerações Adicionais
*   **Atomicidade:** Operações de salvamento (especialmente JSON) devem ser atômicas para evitar arquivos corrompidos se a simulação for interrompida.
*   **Tratamento de Erros:** O sistema deve lidar graciosamente com erros de arquivo (ex: arquivo não encontrado, permissões de escrita).
