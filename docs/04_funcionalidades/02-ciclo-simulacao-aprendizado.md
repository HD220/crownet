# Funcionalidade: Dinâmicas do Ciclo de Simulação da Rede CrowNet

## 1. Visão Geral

O ciclo de simulação é o processo central que governa a evolução temporal da rede neural CrowNet. A cada ciclo, o estado de todos os neurônios é atualizado, os sinais (pulsos) são propagados, as conexões sinápticas são ajustadas através do aprendizado, os neurônios podem se mover (sinaptogênese), e o ambiente neuroquímico da rede é modificado. Esta interação contínua de processos permite que a rede processe informações e aprenda.

A simulação progride em passos de tempo discretos, chamados "ciclos". As seguintes sub-funcionalidades ocorrem sequencialmente ou em paralelo dentro de cada ciclo:

## 2. Atualização do Estado Neuronal

O estado individual de cada neurônio é reavaliado a cada ciclo:

*   **Potencial Acumulado:** O potencial elétrico interno de um neurônio (seu "acumulador de pulso") decai gradualmente ao longo do tempo se não houver novos estímulos. Ao receber pulsos de outros neurônios, este potencial aumenta ou diminui com base na força do pulso e no peso da sinapse correspondente.
*   **Limiar de Disparo:** O limiar de potencial que um neurônio precisa atingir para disparar um novo pulso é dinâmico, sendo influenciado pelos níveis de neuroquímicos simulados (cortisol e dopamina).
*   **Estados Operacionais:** Os neurônios transitam por diferentes estados com base em seu potencial e histórico de atividade:
    *   **Repouso (Resting):** Estado base, aguardando estímulo.
    *   **Disparo (Firing):** Quando o potencial excede o limiar, o neurônio dispara, emitindo pulsos.
    *   **Refratário (Refractory):** Após o disparo, o neurônio entra em um período refratário (absoluto e depois relativo) durante o qual não pode disparar novamente ou tem um limiar de disparo elevado, respectivamente.

## 3. Propagação de Pulsos Neuronais

Quando um neurônio dispara, ele emite pulsos que viajam pela rede:

*   **Natureza do Pulso:** Pulsos carregam um sinal base, que é positivo para neurônios excitatórios (tendendo a ativar outros neurônios) e negativo para neurônios inibitórios (tendendo a suprimir outros). Neurônios dopaminérgicos não emitem pulsos de sinal direto desta forma; seu efeito é puramente químico.
*   **Mecanismo de Propagação:** No MVP, os pulsos se propagam a partir do neurônio emissor em um modelo de expansão esférica a uma velocidade constante.
*   **Impacto nos Neurônios-Alvo:** Neurônios localizados dentro da "casca" de efeito de um pulso em um determinado ciclo são considerados "atingidos". O efeito do pulso no potencial acumulado do neurônio alvo é proporcional ao sinal base do pulso multiplicado pelo peso da conexão sináptica entre o neurônio emissor e o alvo.

## 4. Aprendizado Sináptico (Plasticidade Hebbiana Neuromodulada)

A rede aprende ajustando a força (pesos) de suas conexões sinápticas.

*   **Princípio de Hebb:** A regra fundamental é "neurônios que disparam juntos, fortalecem sua conexão". O sistema detecta a co-ativação de neurônios (disparos que ocorrem próximos no tempo, dentro de uma "janela de coincidência").
*   **Atualização de Peso:** Quando uma co-ativação é detectada entre um neurônio pré-sináptico e um pós-sináptico, o peso da sinapse entre eles é ajustado. O ajuste é proporcional a uma taxa de aprendizado efetiva e à atividade dos neurônios envolvidos.
*   **Decaimento de Peso:** Para promover a competição e evitar a saturação, os pesos sinápticos também sofrem um leve decaimento passivo ao longo do tempo ou a cada atualização.
*   **Limites de Peso:** Os pesos sinápticos são mantidos dentro de uma faixa de valores mínimos e máximos permitidos.
*   **Neuromodulação do Aprendizado:** A taxa de aprendizado efetiva não é constante. Ela é dinamicamente modulada pelos níveis simulados de:
    *   **Dopamina:** Níveis elevados de dopamina tendem a aumentar a taxa de aprendizado.
    *   **Cortisol:** Níveis elevados de cortisol tendem a suprimir (reduzir) a taxa de aprendizado. A taxa efetiva, no entanto, não é reduzida abaixo de um fator mínimo.

## 5. Sinaptogênese (Dinamismo Estrutural)

A estrutura física da rede não é estática; os neurônios podem se mover no espaço 16D.

*   **Movimento Orientado pela Atividade:**
    *   Neurônios tendem a ser atraídos por outros neurônios que estiveram recentemente ativos (disparando ou em período refratário).
    *   Neurônios tendem a ser repelidos por neurônios que estão em estado de repouso.
*   **Neuromodulação da Sinaptogênese:** A intensidade ou taxa deste movimento também é influenciada pelos níveis de neuroquímicos:
    *   **Dopamina:** Níveis elevados de dopamina tendem a aumentar a taxa de movimento.
    *   **Cortisol:** Níveis elevados de cortisol tendem a reduzir a taxa de movimento.
    Os efeitos da dopamina e do cortisol na taxa de sinaptogênese são combinados de forma multiplicativa.
*   Este movimento pode, ao longo do tempo, alterar as proximidades entre neurônios, influenciando indiretamente a formação e o desfazimento de conexões sinápticas efetivas.

## 6. Modulação Neuroquímica

O ambiente químico da rede, representado pelos níveis de Cortisol e Dopamina, influencia diversas dinâmicas neuronais.

*   **Produção e Decaimento:**
    *   **Cortisol:** É produzido por uma "glândula" virtual central quando esta é estimulada por pulsos excitatórios em sua vizinhança. O nível de cortisol decai passivamente a uma taxa percentual a cada ciclo e é limitado a um máximo.
    *   **Dopamina:** É liberada por neurônios do tipo Dopaminérgico quando estes disparam. O nível de dopamina também decai passivamente (geralmente mais rápido que o cortisol) e é limitado a um máximo.
*   **Efeitos Principais:**
    *   **Nos Limiares de Disparo:**
        *   Os níveis de neuroquímicos (cortisol e dopamina) modulam dinamicamente o `CurrentFiringThreshold` de cada neurônio, que é o valor que seu potencial acumulado precisa exceder para disparar. Este é derivado do `BaseFiringThreshold` do neurônio.
        *   **Efeito do Cortisol:** O `CurrentFiringThreshold` é ajustado multiplicativamente com base no nível normalizado de cortisol. A magnitude e a direção desse ajuste (aumento ou diminuição do limiar) são determinadas pelo parâmetro de simulação `FiringThresholdIncreaseOnCort`. Um valor positivo para este parâmetro significa que níveis mais altos de cortisol aumentam o limiar, tornando o neurônio menos propenso a disparar.
        *   **Efeito da Dopamina:** Similarmente, o nível normalizado de dopamina ajusta multiplicativamente o `CurrentFiringThreshold` (após a aplicação do efeito do cortisol). O parâmetro `FiringThresholdIncreaseOnDopa` controla este efeito.
        *   Ambos os efeitos são aplicados sequencialmente, e o `CurrentFiringThreshold` resultante é então sujeito a um valor mínimo absoluto. A lógica anterior que descrevia um efeito em "U" para o cortisol não reflete a implementação atual.
    *   **Na Taxa de Aprendizado Hebbiano:** Detalhado na Seção 4 (Aprendizado Sináptico).
    *   **Na Taxa de Sinaptogênese:** Detalhado na Seção 5 (Sinaptogênese).

## 7. Conclusão do Ciclo

Ao final de cada ciclo de simulação, todos esses processos foram aplicados, resultando em um novo estado global da rede. A repetição desses ciclos permite que a rede CrowNet exiba comportamentos complexos, processe informações de entrada e adapte sua estrutura e função através do aprendizado.
