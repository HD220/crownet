# Funcionalidade: Inicialização da Rede (MVP)

Esta funcionalidade descreve como a rede neural CrowNet é inicializada para o Minimum Viable Product (MVP).

## Componentes Principais da Inicialização

### F1.1: Número Total de Neurônios
*   A rede é inicializada com um número total de neurônios configurável pelo usuário via parâmetro de linha de comando (ex: `-neurons`).

### F1.2: Tipos e Distribuição de Neurônios
*   A rede deve conter um mínimo de:
    *   35 neurônios de **Input** (para acomodar padrões de dígitos 5x7).
    *   10 neurônios de **Output**.
*   Os neurônios restantes são distribuídos entre os tipos:
    *   **Excitatory**
    *   **Inhibitory**
    *   **Dopaminergic**
*   As porcentagens para esses tipos restantes são configuráveis (com valores padrão definidos no sistema).

### F1.3: Posicionamento dos Neurônios
*   Os neurônios são posicionados em um espaço vetorial de 16 dimensões.
*   A posição inicial de cada neurônio é determinada aleatoriamente, mas dentro de restrições radiais específicas para cada tipo de neurônio (para garantir uma distribuição espacial inicial).

### F1.4: Propriedades Iniciais dos Neurônios
*   Cada neurônio é inicializado com:
    *   Um ID único.
    *   Seu tipo (Input, Output, Excitatory, Inhibitory, Dopaminergic).
    *   Um estado inicial (geralmente "Repouso").
    *   Limiares de disparo base.
    *   Acumulador de pulso inicializado (geralmente em zero).

### F1.5: Inicialização de Pesos Sinápticos
*   São inicializados pesos sinápticos explícitos entre neurônios.
*   A estrutura de pesos é uma representação de `ID_neuronio_origem` para `ID_neuronio_destino` resultando em um `valor_do_peso`.
*   Conexões são estabelecidas inicialmente de forma "all-to-all" (todos os neurônios podem potencialmente se conectar a todos os outros), com as seguintes especificidades:
    *   Pesos sinápticos existem de neurônios de Input para outros tipos de neurônios, e entre todos os outros tipos de neurônios não-Input.
    *   Os pesos iniciais são valores pequenos e aleatórios (positivos ou negativos, dentro de uma faixa definida).
    *   Auto-conexões (um neurônio para ele mesmo) têm peso sináptico inicial zero.
    *   Embora a estrutura de dados de pesos possa permitir a representação de pesos *para* neurônios de Input, a lógica de aprendizado e propagação não os utiliza ou modifica ativamente no contexto do MVP.
