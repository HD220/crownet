# Casos de Uso - CrowNet

## 1. Introdução

Este documento descreve os principais casos de uso para o sistema CrowNet, detalhando como diferentes atores interagem com o modelo de rede neural.

## 2. Atores

*   **Usuário Pesquisador:** Um indivíduo (pesquisador, estudante, desenvolvedor) que utiliza o CrowNet para configurar, executar simulações e analisar os resultados.
*   **Sistema:** O próprio software CrowNet.

## 3. Casos de Uso Principais (MVP)

### UC-001: Configurar e Inicializar uma Nova Simulação

*   **Ator:** Usuário Pesquisador
*   **Descrição:** O usuário configura os parâmetros iniciais da rede neural e inicia uma nova simulação.
*   **Pré-condições:** O software CrowNet está instalado e pronto para execução.
*   **Fluxo Principal:**
    1.  O Usuário Pesquisador define (ou usa defaults) parâmetros como:
        *   Número total de neurônios.
        *   Proporções de neurônios (Dopaminérgicos, Inibitórios, Excitatórios, Input, Output).
        *   Dimensões do espaço vetorial (fixo em 16D para o CrowNet).
        *   Parâmetros de inicialização para cortisol e dopamina.
        *   Configurações do gerador procedural para a posição dos neurônios.
    2.  O Usuário Pesquisador inicia o comando para criar e inicializar a rede.
    3.  O Sistema:
        a.  Gera as posições dos neurônios no espaço 16D.
        b.  Atribui tipos e propriedades iniciais aos neurônios.
        c.  Inicializa a glândula de cortisol.
        d.  Configura os neurônios de input e output.
        e.  Armazena o estado inicial da rede no banco de dados SQLite.
    4.  O Sistema informa ao usuário que a inicialização foi concluída.
*   **Pós-condições:** Uma nova rede neural é criada, inicializada e seu estado inicial é salvo. A simulação está pronta para ser executada.

### UC-002: Executar a Simulação da Rede Neural

*   **Ator:** Usuário Pesquisador
*   **Descrição:** O usuário executa a simulação da rede neural por um número definido de ciclos ou continuamente.
*   **Pré-condições:** Uma rede neural foi inicializada (UC-001).
*   **Fluxo Principal:**
    1.  O Usuário Pesquisador inicia o comando para executar a simulação.
    2.  O Usuário Pesquisador pode especificar o número de ciclos a serem simulados ou optar por uma execução contínua (interrompida manualmente).
    3.  O Sistema inicia o loop de simulação:
        a.  Para cada ciclo:
            i.  Processa a propagação de pulsos existentes.
            ii. Calcula interações entre neurônios (soma de potenciais).
            iii.Verifica quais neurônios disparam e gera novos pulsos.
            iv. Atualiza os estados dos neurônios (Repouso, Disparo, Refratário).
            v.  Atualiza os níveis de cortisol e dopamina.
            vi. Aplica a lógica de sinapogênese (movimentação dos neurônios).
            vii.Loga o estado da rede no banco de dados SQLite em intervalos configurados.
            viii.Processa inputs externos (se houver).
            ix. Gera outputs (frequência de disparo dos neurônios de output).
    4.  Se um número de ciclos foi especificado, a simulação para ao atingir esse número. Caso contrário, continua até ser interrompida pelo usuário.
    5.  O Sistema informa o status da simulação (em execução, concluída, interrompida).
*   **Pós-condições:** O estado da rede neural evolui ao longo do tempo. Os dados da simulação são registrados no banco de dados.

### UC-003: Fornecer Input para a Rede

*   **Ator:** Usuário Pesquisador
*   **Descrição:** O usuário fornece estímulos para os neurônios de input da rede durante a simulação.
*   **Pré-condições:** A simulação da rede neural está em execução (UC-002).
*   **Fluxo Principal:**
    1.  O Usuário Pesquisador utiliza uma interface (console para o MVP) para especificar os neurônios de input a serem estimulados e a intensidade/frequência do estímulo.
    2.  O Sistema recebe o input.
    3.  Durante os ciclos de simulação subsequentes, o Sistema converte esse input em "disparos" ou "aumento de potencial" nos neurônios de input designados.
    4.  Esses disparos se propagam pela rede, influenciando outros neurônios.
*   **Pós-condições:** A atividade da rede é influenciada pelo input fornecido pelo usuário.

### UC-004: Observar Output da Rede

*   **Ator:** Usuário Pesquisador
*   **Descrição:** O usuário observa a atividade dos neurônios de output da rede.
*   **Pré-condições:** A simulação da rede neural está em execução (UC-002).
*   **Fluxo Principal:**
    1.  O Sistema, a cada ciclo ou em intervalos definidos, calcula a frequência de disparo dos neurônios designados como output.
    2.  O Sistema exibe essa informação ao Usuário Pesquisador através de uma interface (console para o MVP).
    3.  O Usuário Pesquisador observa os padrões de output para entender a resposta da rede aos estímulos ou sua atividade espontânea.
*   **Pós-condições:** O usuário obtém informações sobre o estado e a atividade dos neurônios de output.

### UC-005: Analisar Dados da Simulação

*   **Ator:** Usuário Pesquisador
*   **Descrição:** O usuário acessa e analisa os dados da simulação armazenados no banco de dados.
*   **Pré-condições:** Uma ou mais simulações foram executadas e seus dados foram salvos no SQLite.
*   **Fluxo Principal:**
    1.  O Usuário Pesquisador utiliza ferramentas externas (clientes SQLite, scripts de análise de dados em Python, etc.) para acessar o banco de dados `crownet.db`.
    2.  O Usuário Pesquisador consulta tabelas contendo:
        *   Estado dos neurônios ao longo do tempo (posição, potencial, estado de ciclo).
        *   Histórico de pulsos.
        *   Níveis de cortisol e dopamina ao longo do tempo.
        *   Conectividade da rede (se aplicável e armazenada).
    3.  O Usuário Pesquisador realiza análises para entender a evolução da rede, padrões de atividade, efeitos de modulação, etc.
*   **Pós-condições:** O usuário obtém insights sobre o comportamento do modelo CrowNet através da análise dos dados coletados.

## 4. Casos de Uso Futuros (Pós-MVP)

*   **UC-006: Visualizar Dinamicamente a Rede:** O usuário observa uma representação gráfica 2D/3D da rede em tempo real, mostrando neurônios, suas conexões e atividade.
*   **UC-007: Interagir com a Rede via Interface Gráfica:** O usuário modifica parâmetros da simulação, estimula neurônios específicos ou altera a estrutura da rede através de uma GUI.
*   **UC-008: Salvar e Carregar Estado da Simulação:** O usuário salva o estado completo de uma simulação em andamento para retomá-la posteriormente.
*   **UC-009: Aplicar Mecanismos de Aprendizado:** O usuário ativa e configura regras de aprendizado (ex: Hebbiano) para observar como a rede se adapta e aprende com os estímulos.
