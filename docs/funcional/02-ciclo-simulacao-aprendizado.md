# Funcionalidade: Ciclo de Simulação e Aprendizado (MVP)

Esta funcionalidade detalha o ciclo principal de simulação da rede CrowNet no MVP, incluindo a atualização dos estados dos neurônios, propagação de pulsos, aprendizado Hebbiano neuromodulado, sinaptogênese e modulação química.

## F2.1: Ciclo de Simulação Discreto (`RunCycle`)
*   A simulação progride em passos de tempo discretos, denominados "ciclos".
*   Cada chamada a `RunCycle` avança a simulação em um ciclo.

## F2.2: Atualizações de Estado dos Neurônios
*   **Decaimento da Acumulação de Pulso:** Em cada ciclo, o valor acumulado de pulsos em cada neurônio decai gradualmente se nenhum novo pulso for recebido.
*   **Estados Neuronais:** Os neurônios transitam entre estados: Repouso, Disparo, Refratário Absoluto, Refratário, com base em suas regras internas e tempo desde o último disparo.
*   **Acumulação de Pulso:** Neurônios acumulam o valor de pulsos recebidos (`BaseSignal * PesoSinaptico`).
*   **Limiar de Disparo (`CurrentFiringThreshold`):** O limiar que um neurônio precisa alcançar para disparar é dinâmico, influenciado por neuroquímicos.

## F2.3: Disparo de Neurônios
*   Um neurônio dispara quando seu `AccumulatedPulse` excede seu `CurrentFiringThreshold`.
*   Ao disparar, um neurônio gera novos pulsos.

## F2.4: Propagação de Pulso
*   Os pulsos se propagam a partir do neurônio emissor.
*   O MVP utiliza um modelo de expansão esférica:
    *   Pulsos têm um valor base (`BaseSignal`): +1.0 para excitatórios, -1.0 para inibitórios (neurônios dopaminérgicos têm BaseSignal 0, seu efeito é químico).
    *   Pulsos viajam a uma velocidade fixa (ex: `PulsePropagationSpeed` de `neuron/config.go`, como 0.6 unidades de distância por ciclo).
    *   A cada ciclo, um pulso varre uma "casca" esférica no espaço. Neurônios dentro desta casca (entre `pulseEffectStartDist` e `pulseEffectEndDist` da origem do pulso) são considerados "atingidos" naquele ciclo.
    *   Quando um pulso atinge outro neurônio, o neurônio receptor processa o `ValorBaseDoPulso * PesoDaSinapse`.

## F2.5: Plasticidade Hebbiana Neuromodulada (`ApplyHebbianPlasticity`)
*   Esta é a principal mecânica de aprendizado no MVP.
*   **Princípio Hebbiano:** "Neurônios que disparam juntos, conectam-se mais fortemente".
*   **Co-ativação:** O sistema identifica pares de neurônios (pré-sináptico e pós-sináptico) que demonstraram atividade correlacionada. A correlação é determinada se ambos os neurônios dispararam dentro de uma janela de tempo definida por `HebbianCoincidenceWindow` (ex: 0 para disparo no mesmo ciclo, 1 para até 1 ciclo de diferença).
*   **Cálculo da Taxa de Aprendizado Efetiva:**
    *   Uma `BaseLearningRate` (configurável) é definida.
    *   Esta taxa é modulada pelos níveis atuais de Dopamina e Cortisol (conforme detalhado em F2.7).
*   **Atualização de Peso Sináptico:** O peso da sinapse (`W_ij`) entre o neurônio pré-sináptico `i` e o pós-sináptico `j` é ajustado:
    *   `ΔW_ij = taxa_aprendizado_efetiva * atividade_pre_i * atividade_pos_j` (onde atividade é tipicamente 1.0 se disparou na janela).
    *   `W_ij_novo = W_ij_antigo + ΔW_ij`
    *   Aplica-se também um decaimento de peso (`HebbianWeightDecay`): `W_ij_novo = W_ij_novo - W_ij_novo * HebbianWeightDecay`.
*   **Limites de Peso:** Os pesos são mantidos dentro de uma faixa definida (`HebbianWeightMin`, `HebbianWeightMax`).

## F2.6: Sinaptogênese (Movimento de Neurônios)
*   Neurônios podem se mover no espaço 16D.
*   **Influência da Atividade:**
    *   Neurônios tendem a se aproximar de outros neurônios que estiveram recentemente ativos (disparando ou em período refratário).
    *   Neurônios tendem a se afastar de neurônios que estão em repouso.
*   **Modulação Química:** A taxa ou intensidade do movimento (sinaptogênese) também é modulada pelos níveis de Cortisol e Dopamina. (Ex: Cortisol alto reduz o movimento).
*   Este movimento pode influenciar a conectividade ao longo do tempo, alterando as proximidades entre neurônios.

## F2.7: Modulação Química (Cortisol e Dopamina)
*   **Produção de Cortisol:**
    *   Produzido por uma "glândula" virtual localizada na posição central do espaço (`CortisolGlandPosition`).
    *   A produção aumenta quando pulsos excitatórios atingem a vizinhança da glândula (definida por `CortisolGlandRadius`). A quantidade produzida é `CortisolProductionPerHit`.
    *   Decai a uma taxa percentual (`CortisolDecayRate`) por ciclo e é limitado por `CortisolMaxLevel`.
*   **Produção de Dopamina:**
    *   Produzida por neurônios do tipo `DopaminergicNeuron` quando estes entram no estado `FiringState`. A quantidade produzida é `DopamineProductionPerEvent`.
    *   Decai a uma taxa percentual (`DopamineDecayRate`) por ciclo (geralmente mais rápido que o cortisol) e é limitada por `DopamineMaxLevel`.
*   **Efeitos dos Neuroquímicos:**
    *   **Modulação dos Limiares de Disparo:**
        *   **Cortisol:** Altera o `CurrentFiringThreshold` dos neurônios. O efeito tem um perfil em forma de U:
            *   Abaixo de `CortisolMinEffectThreshold`: Sem efeito significativo.
            *   Entre `CortisolMinEffectThreshold` e `CortisolOptimalLowThreshold`: Redução progressiva do limiar (até `MaxThresholdReductionFactor`).
            *   Entre `CortisolOptimalLowThreshold` e `CortisolOptimalHighThreshold`: Redução máxima do limiar.
            *   Entre `CortisolOptimalHighThreshold` e `CortisolHighEffectThreshold`: Aumento progressivo do limiar de volta ao basal.
            *   Acima de `CortisolHighEffectThreshold`: Aumento progressivo do limiar acima do basal (até `ThresholdIncreaseFactorHigh`).
        *   **Dopamina:** Aumenta o `CurrentFiringThreshold` (que já pode ter sido modificado pelo cortisol). O aumento é progressivo com o nível de dopamina, até `DopamineThresholdIncreaseFactor`.
    *   **Modulação da Taxa de Aprendizado Hebbiano (Plasticidade):**
        *   A `BaseLearningRate` é multiplicada por um fator derivado dos níveis de Cortisol e Dopamina.
        *   **Dopamina:** Aumenta o fator de aprendizado (até `MaxDopamineLearningMultiplier` em `DopamineMaxLevel`).
        *   **Cortisol (Alto):** Suprime o fator de aprendizado (reduzindo-o até `CortisolLearningSuppressionFactor` em `CortisolMaxLevel`, se acima de `CortisolHighEffectThreshold`). A taxa de aprendizado efetiva não cai abaixo de `MinLearningRateFactor` do original.
    *   **Modulação da Taxa de Sinaptogênese:**
        *   O fator de movimento base na sinaptogênese é multiplicado por um fator derivado de Cortisol e Dopamina.
        *   **Cortisol (Alto):** Reduz o movimento (até `SynaptogenesisReductionFactor` em `CortisolMaxLevel`, se acima de `CortisolHighEffectThreshold`).
        *   **Dopamina:** Aumenta o movimento (até `DopamineSynaptogenesisIncreaseFactor` em `DopamineMaxLevel`).
        Os efeitos de cortisol e dopamina na sinaptogênese são combinados multiplicativamente.
