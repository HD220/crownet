# Funcionalidade: Persistência de Dados da Rede CrowNet

## 1. Visão Geral

A capacidade de salvar e carregar o estado da rede neural é fundamental no CrowNet. A persistência de dados permite que o progresso do aprendizado (pesos sinápticos) seja preservado entre sessões de simulação e que o estado detalhado da rede seja registrado para análise offline e depuração.

Esta funcionalidade abrange dois mecanismos principais:
*   Persistência dos pesos sinápticos em formato JSON.
*   Logging opcional do estado dinâmico completo da rede em um banco de dados SQLite.

## 2. Persistência de Pesos Sinápticos

Os pesos das conexões sinápticas, que são o principal resultado do processo de aprendizado da rede, são gerenciados da seguinte forma:

### 2.1. Formato e Estrutura
*   **Formato do Arquivo:** Os pesos sinápticos são salvos e carregados utilizando arquivos no formato JSON.
*   **Estrutura de Dados no JSON:** O arquivo JSON representa um mapa onde:
    *   Cada chave de nível superior é uma string representando o ID de um neurônio de origem.
    *   O valor associado a cada neurônio de origem é outro mapa, onde:
        *   Cada chave é uma string representando o ID de um neurônio de destino.
        *   O valor é o peso numérico (ponto flutuante) da sinapse entre o neurônio de origem e o neurônio de destino.
    *   Exemplo:
      ```json
      {
        "0": { "1": 0.75, "2": -0.32, /* ... */ },
        "1": { "0": 0.68, "2": 0.98, /* ... */ }
        /* ... mais neurônios de origem ... */
      }
      ```

### 2.2. Utilização nos Modos de Operação
*   **Modo `expose` (Treinamento):**
    *   Ao iniciar, se um arquivo de pesos (especificado via flag `-weightsFile`) existir e for válido, ele é carregado para inicializar a rede. Caso contrário, a rede começa com pesos aleatórios.
    *   Ao final do processo de treinamento, os pesos sinápticos finais da rede são salvos no arquivo especificado.
*   **Modo `observe` (Observação/Teste):**
    *   Os pesos sinápticos devem ser carregados de um arquivo especificado. A operação falhará se o arquivo não for encontrado ou for inválido.
*   **Modo `sim` (Simulação Geral):**
    *   Os pesos podem ser opcionalmente carregados de um arquivo no início da simulação. Este modo geralmente não salva os pesos automaticamente ao final, pois seu foco é a simulação de dinâmicas, não necessariamente o treinamento convergente.

## 3. Logging Opcional do Estado da Rede em SQLite

Para análises mais detalhadas da dinâmica da rede e para depuração, o sistema oferece a opção de registrar snapshots completos do estado da rede em um banco de dados SQLite.

### 3.1. Propósito e Ativação
*   **Objetivo:** Capturar a evolução temporal do estado de cada neurônio e dos níveis de neuroquímicos.
*   **Ativação:** O logging para SQLite é habilitado quando um caminho de arquivo de banco de dados é fornecido através do flag `-dbPath`. Se o arquivo já existir, ele é geralmente recriado a cada nova execução que utiliza esta opção.
*   **Frequência:** A frequência com que os snapshots são salvos é controlada pelo flag `-saveInterval` (número de ciclos entre cada salvamento).

### 3.2. Conteúdo do Snapshot no Banco de Dados
Um snapshot da rede no banco de dados SQLite normalmente inclui as seguintes informações, distribuídas em tabelas relacionais:

*   **Tabela `NetworkSnapshots` (ou similar):** Registra informações globais da rede por snapshot.
    *   Campos: `SnapshotID` (identificador único do snapshot), `CycleCount` (ciclo da simulação), `Timestamp`, `CortisolLevel`, `DopamineLevel`.
*   **Tabela `NeuronStates` (ou similar):** Registra o estado detalhado de cada neurônio no momento do snapshot.
    *   Campos: `StateID` (identificador único do estado), `SnapshotID` (referência ao snapshot), `NeuronID`, coordenadas de posição (ex: `Position0` a `Position15`), componentes de velocidade, tipo do neurônio, estado operacional (Repouso, Disparo, etc.), potencial acumulado, limiares de disparo (base e atual), ciclo do último disparo, ciclos no estado atual.

*Nota: Os pesos sinápticos em si geralmente não são duplicados no banco de dados SQLite a cada snapshot, pois o arquivo JSON é o meio primário para sua persistência. O foco do logging em SQLite é o estado dinâmico da rede.*

### 3.3. Utilização
*   **Modo `sim`:** Particularmente útil para registrar a evolução da rede sob dinâmicas gerais e estímulos específicos.
*   **Modo `expose`:** Pode ser usado para capturar a trajetória de aprendizado e a evolução dos estados neuronais durante o treinamento.
*   Os dados armazenados no SQLite são destinados à análise offline, utilizando ferramentas de consulta SQL, scripts de análise de dados (ex: Python com bibliotecas de SQLite e plotagem) ou outras ferramentas de visualização.

## 4. Considerações Importantes

*   **Atomicidade das Operações:** É importante que as operações de salvamento de arquivos (especialmente o JSON de pesos) sejam realizadas de forma atômica ou com mecanismos de proteção (ex: salvar em arquivo temporário e renomear) para minimizar o risco de arquivos corrompidos caso a simulação seja interrompida inesperadamente.
*   **Tratamento de Erros:** O sistema deve ser robusto a erros comuns de manipulação de arquivos, como arquivo não encontrado (para leitura), falha de permissão de escrita, ou formato de arquivo inválido, fornecendo feedback apropriado ao usuário.

A funcionalidade de persistência é vital para a utilidade e a capacidade de análise do simulador CrowNet.
