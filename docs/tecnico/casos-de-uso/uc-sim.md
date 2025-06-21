# Caso de Uso: UC-SIM - Executar Simulação Geral da Rede

*   **ID:** UC-SIM
*   **Ator Principal:** Usuário (interagindo via CLI)
*   **Breve Descrição:** O usuário executa a aplicação no modo `sim` para rodar uma simulação geral da rede CrowNet com todas as suas dinâmicas ativas (plasticidade Hebbiana, modulação química, sinaptogênese). Este modo é útil para observar comportamentos emergentes, testar a estabilidade da rede sob estímulos contínuos ou para logging detalhado para análise offline, sem o foco específico na tarefa de aprendizado de dígitos dos modos `expose` ou `observe`.
*   **Pré-condições:**
    1.  A aplicação CrowNet está compilada e executável.
    2.  (Opcional) Um arquivo de pesos sinápticos (`-weightsFile`) pode existir. Se fornecido e válido, a rede o utilizará; caso contrário, começará com pesos iniciais aleatórios.
*   **Pós-condições (Sucesso):**
    1.  A simulação é executada pelo número de ciclos especificado.
    2.  (Opcional) Se o logging para SQLite (`-dbPath`) estiver ativado, o banco de dados conterá snapshots do estado da rede.
    3.  (Opcional) Se um neurônio de output estiver sendo monitorado (`-monitorOutputID`), sua frequência de disparo é exibida.
    4.  O Usuário recebe feedback no console sobre o progresso e a conclusão da simulação.
*   **Pós-condições (Falha):**
    1.  Uma mensagem de erro é exibida no console.

## Fluxo Principal (Sucesso):

1.  **Usuário** inicia a aplicação CrowNet com o flag `-mode sim` e outros flags relevantes:
    *   `-cycles <N_ciclos_simulacao>` (obrigatório para este modo, ou usa padrão)
    *   `-weightsFile <caminho_arquivo_pesos>` (opcional)
    *   `-neurons <N_total_neuronios>` (opcional)
    *   `-stimInputID <ID_neuronio_estimulo>` (opcional, para estímulo contínuo)
    *   `-stimInputFreqHz <frequencia_estimulo>` (opcional, relevante se `-stimInputID` fornecido)
    *   `-monitorOutputID <ID_neuronio_monitorado>` (opcional, para monitorar frequência de saída)
    *   `-dbPath <caminho_db>` (opcional, para logging SQLite)
    *   `-saveInterval <intervalo_ciclos_db>` (opcional, relevante se `-dbPath` fornecido)
    *   `-debugChem <true/false>` (opcional)
2.  **Sistema** inicializa a rede neural:
    a.  Configura o número total de neurônios.
    b.  Distribui os tipos de neurônios e posiciona-os.
    c.  Se `-weightsFile` for fornecido, tenta carregar os pesos sinápticos. Se bem-sucedido, usa esses pesos; caso contrário (ou se não fornecido), inicializa com pesos pequenos e aleatórios.
    d.  Inicializa os níveis de neuroquímicos.
    e.  Habilita todas as dinâmicas: plasticidade Hebbiana, modulação química e sinaptogênese.
3.  **Sistema** informa ao Usuário o início da simulação geral, mostrando os parâmetros configurados.
4.  **Sistema** configura o estímulo de input contínuo, se `-stimInputID` e `-stimInputFreqHz` forem fornecidos e válidos.
5.  **Sistema** itera pelo número de ciclos de simulação (`N_ciclos_simulacao`):
    a.  Executa o ciclo de simulação (`RunCycle`):
        i.  Processa inputs de frequência (se configurado).
        ii. Neurônios atualizam seus estados.
        iii.Pulsos se propagam.
        iv. Pesos sinápticos são ajustados via plasticidade Hebbiana neuromodulada.
        v.  Níveis de neuroquímicos são atualizados e seus efeitos aplicados.
        vi. Neurônios se movem (sinaptogênese).
    b.  A cada 10 ciclos (ou similar), exibe no console uma linha de progresso (ciclo atual, níveis químicos, número de pulsos).
    c.  (Opcional) Se `-dbPath` estiver configurado e o intervalo de salvamento (`-saveInterval`) for atingido, o estado da rede é salvo no SQLite.
6.  (Opcional) Se `-dbPath` estiver configurado e o salvamento final for aplicável (ex: `saveInterval` era 0 ou o número total de ciclos não era múltiplo do intervalo), o estado final da rede é salvo no SQLite.
7.  (Opcional) Se `-monitorOutputID` for fornecido e válido, calcula e exibe a frequência de disparo do neurônio de saída monitorado.
8.  **Sistema** exibe os níveis finais de Cortisol e Dopamina.
9.  **Sistema** exibe uma mensagem de conclusão da sessão.

## Fluxos de Exceção:

*   **2.c.i. Falha ao carregar arquivo de pesos (se fornecido):**
    *   **Sistema** informa ao Usuário que o arquivo de pesos não pôde ser carregado e que a rede iniciará com pesos aleatórios. O processo continua.
*   **4.a. ID de neurônio de estímulo inválido:**
    *   **Sistema** informa ao Usuário que o ID do neurônio de estímulo é inválido e o estímulo não será aplicado. A simulação continua sem este estímulo específico.
*   **X.Y.i. Erro de configuração de flags (ex: tipo inválido):**
    *   **Sistema** (via biblioteca `flag`) exibe uma mensagem de erro sobre o uso incorreto dos flags e encerra.
*   **X.Y.ii. Falha na inicialização do banco de dados (se `-dbPath` fornecido):**
    *   **Sistema** encerra a execução com uma mensagem de erro fatal.
*   **X.Y.iii. Falha ao salvar no banco de dados (se `-dbPath` fornecido):**
    *   **Sistema** exibe um aviso no console, mas a simulação continua.
```
