# Casos de Uso: CrowNet MVP - Reconhecimento de Dígitos

Este documento descreve os principais casos de uso para o MVP da CrowNet, focado no reconhecimento de dígitos manuscritos.

## Atores

*   **Usuário Desenvolvedor/Pesquisador (UDR):** Indivíduo com conhecimento técnico que interage com o sistema via CLI para treinar, avaliar e gerenciar os modelos da CrowNet.

## Casos de Uso Principais

### UC-001: Treinar a Rede Neural para Reconhecimento de Dígitos

*   **Ator Principal:** UDR
*   **Pré-condições:**
    *   O dataset MNIST está acessível pelo sistema.
    *   A CLI da CrowNet está instalada e operacional.
*   **Fluxo Principal:**
    1.  O UDR executa o comando na CLI para iniciar o treinamento da rede (ex: `crownet train --dataset mnist --epochs 50 --output-model model.db`).
    2.  O sistema inicializa uma nova rede CrowNet (ou carrega um modelo existente, se especificado).
        *   Gera posições dos neurônios.
        *   Define tipos e conexões iniciais (implícitas pela proximidade e dinâmica).
    3.  O sistema carrega o dataset de treinamento MNIST.
    4.  Para cada época de treinamento especificada:
        a.  Para cada imagem de dígito no conjunto de treinamento:
            i.  O sistema pré-processa a imagem.
            ii. O sistema codifica a imagem em um padrão de frequência de disparo para os neurônios de input da CrowNet.
            iii. O sistema executa um número definido de ciclos de simulação da CrowNet, permitindo que os pulsos se propaguem e a rede processe a entrada.
            iv. O sistema decodifica a atividade dos neurônios de output para obter o dígito classificado.
            v.  O sistema compara o dígito classificado com o rótulo verdadeiro.
            vi. O sistema aplica o mecanismo de recompensa (se acerto) ou punição (se erro) para modular a dinâmica da rede (níveis de cortisol/dopamina, influenciando a sinapogênese e limiares).
        b.  (Opcional) O sistema exibe o progresso da época (ex: número da época, taxa de acerto parcial).
    5.  Após todas as épocas, o sistema informa que o treinamento foi concluído.
    6.  Se um caminho de saída foi especificado, o sistema salva o estado da rede neural treinada (ex: `model.db`).
*   **Fluxos Alternativos:**
    *   **UC-001.A1: Dataset não encontrado:** O sistema informa ao UDR que o dataset MNIST não foi encontrado no local esperado e encerra a operação.
    *   **UC-001.A2: Erro durante o salvamento do modelo:** O sistema informa ao UDR sobre o erro ao tentar salvar o modelo.
*   **Pós-condições:**
    *   A rede CrowNet teve seus parâmetros (posições dos neurônios, implicitamente as "conexões" através da dinâmica espacial) ajustados com base nos dados de treinamento.
    *   Um arquivo de modelo contendo o estado da rede treinada é salvo, se solicitado.

### UC-002: Avaliar o Desempenho da Rede Neural Treinada

*   **Ator Principal:** UDR
*   **Pré-condições:**
    *   Um modelo CrowNet treinado está disponível (arquivo .db).
    *   O dataset MNIST (conjunto de teste) está acessível.
    *   A CLI da CrowNet está instalada e operacional.
*   **Fluxo Principal:**
    1.  O UDR executa o comando na CLI para avaliar um modelo treinado (ex: `crownet evaluate --model model.db --dataset mnist`).
    2.  O sistema carrega o estado da rede neural do arquivo de modelo especificado.
    3.  O sistema carrega o conjunto de teste do dataset MNIST.
    4.  Para cada imagem de dígito no conjunto de teste:
        a.  O sistema pré-processa a imagem.
        b.  O sistema codifica a imagem em um padrão de frequência de disparo para os neurônios de input.
        c.  O sistema executa um número definido de ciclos de simulação da CrowNet (sem aplicar o mecanismo de aprendizado/recompensa/punição).
        d.  O sistema decodifica a atividade dos neurônios de output para obter o dígito classificado.
        e.  O sistema armazena a predição e o rótulo verdadeiro.
    5.  Após processar todas as imagens de teste, o sistema calcula a taxa de acerto geral (predições corretas / total de imagens).
    6.  O sistema exibe a taxa de acerto e, possivelmente, outras métricas (ex: matriz de confusão simples).
*   **Fluxos Alternativos:**
    *   **UC-002.A1: Modelo não encontrado:** O sistema informa ao UDR que o arquivo de modelo especificado não foi encontrado e encerra a operação.
    *   **UC-002.A2: Dataset não encontrado:** O sistema informa ao UDR que o dataset MNIST não foi encontrado e encerra a operação.
*   **Pós-condições:**
    *   O UDR recebe uma medida quantitativa (taxa de acerto) do desempenho do modelo treinado no conjunto de teste.

### UC-003: Salvar um Modelo de Rede Neural Treinada

*   **Ator Principal:** Sistema (como parte do UC-001) ou UDR (explicitamente, se uma funcionalidade de "snapshot" for implementada).
*   **Pré-condições:**
    *   Existe um estado de rede neural treinado ou em treinamento na memória.
    *   Um nome/caminho de arquivo válido para o modelo é fornecido.
*   **Fluxo Principal:**
    1.  O sistema recebe o comando para salvar o estado atual da rede neural.
    2.  O sistema coleta todas as informações relevantes do estado da rede:
        *   Posições e tipos de todos os neurônios.
        *   Estados atuais dos neurônios (potencial, ciclo refratário, etc.).
        *   Níveis atuais de cortisol e dopamina.
        *   Parâmetros de configuração da rede.
        *   (Opcional) Estado do gerador de números aleatórios para reprodutibilidade.
    3.  O sistema serializa essas informações e as armazena no arquivo SQLite especificado.
    4.  O sistema confirma que o salvamento foi bem-sucedido.
*   **Fluxos Alternativos:**
    *   **UC-003.A1: Erro de permissão de escrita:** O sistema não consegue escrever no local especificado e informa o erro.
    *   **UC-003.A2: Disco cheio:** O sistema não consegue salvar por falta de espaço e informa o erro.
*   **Pós-condições:**
    *   Um arquivo .db contendo o estado da rede neural é criado/atualizado no sistema de arquivos.

### UC-004: Carregar um Modelo de Rede Neural Treinada

*   **Ator Principal:** UDR (como parte do UC-002 ou para continuar o treinamento)
*   **Pré-condições:**
    *   Um arquivo de modelo .db válido, previamente salvo, existe.
    *   A CLI da CrowNet está operacional.
*   **Fluxo Principal:**
    1.  O UDR especifica um arquivo de modelo para carregar (ex: ao iniciar avaliação ou continuar treinamento `crownet train --load-model model.db ...`).
    2.  O sistema lê o arquivo SQLite.
    3.  O sistema desserializa os dados e reconstrói o estado da rede neural na memória:
        *   Restaura posições, tipos e estados dos neurônios.
        *   Restaura níveis de cortisol e dopamina.
        *   Restaura parâmetros de configuração.
    4.  O sistema confirma que o carregamento foi bem-sucedido e a rede está pronta para uso (avaliação, mais treinamento).
*   **Fluxos Alternativos:**
    *   **UC-004.A1: Arquivo de modelo não encontrado:** O sistema informa o erro.
    *   **UC-004.A2: Arquivo de modelo corrompido ou inválido:** O sistema não consegue desserializar os dados e informa o erro.
*   **Pós-condições:**
    *   O estado da rede neural previamente salvo é carregado na memória, pronto para ser usado.

### UC-005: Realizar uma Predição Única (Opcional para MVP inicial, mas bom para demonstração)

*   **Ator Principal:** UDR
*   **Pré-condições:**
    *   Um modelo CrowNet treinado está carregado.
    *   Uma imagem de dígito (ex: de um arquivo) está disponível.
*   **Fluxo Principal:**
    1.  O UDR fornece uma imagem de dígito ao sistema via CLI (ex: `crownet predict --model model.db --image digit.png`).
    2.  O sistema pré-processa a imagem.
    3.  O sistema codifica a imagem em um padrão de frequência de disparo para os neurônios de input.
    4.  O sistema executa um número definido de ciclos de simulação da CrowNet.
    5.  O sistema decodifica a atividade dos neurônios de output para obter o dígito classificado.
    6.  O sistema exibe o dígito classificado.
*   **Pós-condições:**
    *   O UDR recebe a predição do modelo para a imagem fornecida.

Estes casos de uso definem as interações fundamentais com o MVP da CrowNet e guiarão o desenvolvimento das funcionalidades da CLI e da lógica interna do sistema.
