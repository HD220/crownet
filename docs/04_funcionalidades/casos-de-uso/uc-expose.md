# Caso de Uso: UC-EXPOSE - Expor Rede a Padrões para Auto-Aprendizagem

*   **ID:** UC-EXPOSE
*   **Ator Principal:** Usuário (interagindo via CLI)
*   **Breve Descrição:** O usuário executa a aplicação no modo `expose` para apresentar repetidamente padrões de dígitos (0-9) à rede neural. Durante esta exposição, a rede ajusta seus pesos sinápticos através de plasticidade Hebbiana neuromodulada, permitindo que ela aprenda a formar representações internas distintas para cada dígito.
*   **Pré-condições:**
    1.  A aplicação CrowNet está compilada e executável.
    2.  (Opcional) Um arquivo de pesos sinápticos (`-weightsFile`) pode existir de uma execução anterior. Se não existir ou for inválido, a rede começará com pesos iniciais aleatórios.
*   **Pós-condições (Sucesso):**
    1.  A rede neural processou o número especificado de épocas de exposição aos padrões.
    2.  Os pesos sinápticos da rede foram modificados e salvos no arquivo especificado por `-weightsFile`.
    3.  (Opcional) Se o logging para SQLite (`-dbPath`) estiver ativado, o banco de dados conterá snapshots do estado da rede em intervalos definidos durante a exposição.
    4.  O Usuário recebe feedback no console sobre a conclusão do processo e o salvamento dos pesos.
*   **Pós-condições (Falha):**
    1.  Uma mensagem de erro é exibida no console.
    2.  Os pesos podem não ter sido salvos ou podem estar em um estado inconsistente se a falha ocorreu durante o salvamento.

## Fluxo Principal (Sucesso):

1.  **Usuário** inicia a aplicação CrowNet com o flag `-mode expose` e outros flags relevantes:
    *   `-epochs <N_epocas>` (obrigatório para este modo, ou usa padrão)
    *   `-lrBase <taxa_aprendizado_base>` (obrigatório para este modo, ou usa padrão)
    *   `-cyclesPerPattern <N_ciclos_por_padrao>` (obrigatório para este modo, ou usa padrão)
    *   `-weightsFile <caminho_arquivo_pesos>` (opcional, usa padrão se não especificado)
    *   `-neurons <N_total_neuronios>` (opcional, usa padrão se não especificado)
    *   `-dbPath <caminho_db>` (opcional, para logging SQLite)
    *   `-saveInterval <intervalo_ciclos_db>` (opcional, relevante se `-dbPath` fornecido)
    *   `-debugChem <true/false>` (opcional)
2.  **Sistema** inicializa a rede neural:
    a.  Configura o número total de neurônios.
    b.  Distribui os tipos de neurônios (Input, Output, Excitatory, Inhibitory, Dopaminergic).
    c.  Posiciona os neurônios no espaço 16D.
    d.  Tenta carregar os pesos sinápticos do arquivo especificado por `-weightsFile`. Se bem-sucedido, usa esses pesos; caso contrário, inicializa com pesos pequenos e aleatórios.
    e.  Inicializa os níveis de neuroquímicos (Cortisol, Dopamina).
    f.  Habilita as dinâmicas de plasticidade Hebbiana, modulação química e sinaptogênese.
    g.  (Se aplicável) Executa a lógica de `setupDopamineStimulationForExpose` para tentar garantir atividade dopaminérgica.
3.  **Sistema** informa ao Usuário o início da fase de exposição, mostrando os parâmetros configurados.
4.  **Sistema** itera pelo número de épocas (`N_epocas`):
    a.  Informa ao Usuário o início da época atual.
    b.  Para cada padrão de dígito (0 a 9):
        i.  Reseta as ativações dos neurônios e limpa quaisquer pulsos ativos remanescentes.
        ii. Apresenta o padrão do dígito atual aos neurônios de input designados (forçando-os a disparar).
        iii.Executa o ciclo de simulação (`RunCycle`) por `N_ciclos_por_padrao` vezes. Durante cada ciclo:
            1.  Neurônios atualizam seus estados.
            2.  Pulsos se propagam.
            3.  Pesos sinápticos são ajustados via plasticidade Hebbiana neuromodulada.
            4.  Níveis de neuroquímicos são atualizados e seus efeitos aplicados.
            5.  Neurônios se movem (sinaptogênese).
            6.  (Opcional) Se `-dbPath` estiver configurado e o intervalo de salvamento for atingido, o estado da rede é salvo no SQLite.
    c.  Informa ao Usuário a conclusão da época atual e exibe os níveis de Cortisol, Dopamina e um exemplo do fator de aprendizado efetivo.
5.  **Sistema** informa ao Usuário que a fase de exposição foi concluída.
6.  **Sistema** salva os pesos sinápticos finais no arquivo especificado por `-weightsFile`.
7.  **Sistema** informa ao Usuário se o salvamento dos pesos foi bem-sucedido.
8.  (Opcional) Se `-dbPath` estiver configurado e o salvamento final for aplicável, o estado da rede é salvo no SQLite.
9.  **Sistema** exibe uma mensagem de conclusão da sessão.

## Fluxos de Exceção:

*   **2.d.i. Falha ao carregar arquivo de pesos:**
    *   **Sistema** informa ao Usuário que o arquivo de pesos não pôde ser carregado (ex: não encontrado, formato inválido) e que a rede iniciará com pesos aleatórios. O processo continua.
*   **4.b.i. Falha ao obter padrão de dígito:**
    *   **Sistema** encerra a execução com uma mensagem de erro fatal.
*   **6.a. Falha ao salvar pesos:**
    *   **Sistema** informa ao Usuário que houve um erro ao salvar os pesos e encerra com uma mensagem de erro fatal.
*   **X.Y.i. Erro de configuração de flags (ex: tipo inválido):**
    *   **Sistema** (via biblioteca `flag`) exibe uma mensagem de erro sobre o uso incorreto dos flags e encerra.
*   **X.Y.ii. Falha na inicialização do banco de dados (se `-dbPath` fornecido):**
    *   **Sistema** encerra a execução com uma mensagem de erro fatal.
*   **X.Y.iii. Falha ao salvar no banco de dados (se `-dbPath` fornecido):**
    *   **Sistema** exibe um aviso no console, mas a simulação continua.
```
