# Caso de Uso: UC-OBSERVE - Observar Resposta da Rede a um Dígito

*   **ID:** UC-OBSERVE
*   **Ator Principal:** Usuário (interagindo via CLI)
*   **Breve Descrição:** O usuário executa a aplicação no modo `observe` para carregar pesos sinápticos previamente treinados, apresentar um padrão de dígito específico à rede e observar o padrão de ativação resultante nos neurônios de saída. Isso permite avaliar como a rede representa o dígito internamente.
*   **Pré-condições:**
    1.  A aplicação CrowNet está compilada e executável.
    2.  Um arquivo de pesos sinápticos (`-weightsFile`) treinado (geralmente por uma execução anterior do modo `expose`) DEVE existir e ser válido.
*   **Pós-condições (Sucesso):**
    1.  O padrão de ativação (ex: `AccumulatedPulse`) dos 10 neurônios de saída designados é exibido no console para o dígito apresentado.
    2.  Os pesos sinápticos carregados não são alterados.
    3.  O Usuário recebe feedback no console sobre a conclusão do processo.
*   **Pós-condições (Falha):**
    1.  Uma mensagem de erro é exibida no console.

## Fluxo Principal (Sucesso):

1.  **Usuário** inicia a aplicação CrowNet com o flag `-mode observe` e outros flags relevantes:
    *   `-digit <D>` (obrigatório para este modo, especificando o dígito 0-9)
    *   `-weightsFile <caminho_arquivo_pesos>` (obrigatório, para carregar os pesos treinados)
    *   `-cyclesToSettle <N_ciclos_acomodacao>` (obrigatório para este modo, ou usa padrão)
    *   `-neurons <N_total_neuronios>` (opcional, deve corresponder à configuração da rede que gerou os pesos)
2.  **Sistema** inicializa a rede neural:
    a.  Configura o número total de neurônios.
    b.  Distribui os tipos de neurônios e posiciona-os.
    c.  Carrega os pesos sinápticos do arquivo especificado por `-weightsFile`. Se falhar, o processo é interrompido (ver Fluxos de Exceção).
    d.  Inicializa os níveis de neuroquímicos.
    e.  **Temporariamente desativa** as dinâmicas de plasticidade Hebbiana, modulação química e sinaptogênese para garantir uma observação "limpa" e não alterar os pesos.
3.  **Sistema** informa ao Usuário o início da fase de observação, mostrando os parâmetros configurados.
4.  **Sistema** obtém o padrão binário para o `-digit` especificado.
5.  **Sistema** reseta as ativações dos neurônios e limpa pulsos ativos.
6.  **Sistema** apresenta o padrão do dígito aos neurônios de input designados.
7.  **Sistema** executa o ciclo de simulação (`RunCycle`) por `N_ciclos_acomodacao` vezes. Durante estes ciclos:
    a.  Neurônios atualizam seus estados (sem modulação química nos limiares).
    b.  Pulsos se propagam.
    c.  Pesos sinápticos NÃO são ajustados.
    d.  Níveis de neuroquímicos NÃO são atualizados dinamicamente nem aplicam efeitos.
    e.  Neurônios NÃO se movem.
8.  **Sistema** recupera o valor do `AccumulatedPulse` de cada um dos 10 neurônios de saída designados.
9.  **Sistema** exibe no console:
    a.  O dígito que foi apresentado.
    b.  O padrão de ativação dos neurônios de saída (lista de valores).
10. **Sistema** restaura as configurações originais das dinâmicas (plasticidade, químicos, sinaptogênese) para seus estados padrão (habilitados).
11. **Sistema** exibe uma mensagem de conclusão da sessão.

## Fluxos de Exceção:

*   **2.c.i. Falha ao carregar arquivo de pesos:**
    *   **Sistema** informa ao Usuário que o arquivo de pesos não pôde ser carregado (ex: não encontrado, formato inválido) e encerra com uma mensagem de erro fatal, pois os pesos são essenciais para este modo.
*   **4.a. Falha ao obter padrão de dígito (dígito inválido):**
    *   **Sistema** encerra a execução com uma mensagem de erro fatal.
*   **X.Y.i. Erro de configuração de flags (ex: tipo inválido, flag obrigatório ausente):**
    *   **Sistema** (via biblioteca `flag`) exibe uma mensagem de erro sobre o uso incorreto dos flags e encerra.
```
