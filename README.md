# CrowNet

**CrowNet** é um modelo computacional de rede neural inspirado em processos biológicos, simula a interação de neurônios em um espaço vetorial de 16 dimensões. Através de dinâmicas como sinapogênese, propagação de pulso, e interação de substâncias químicas como **cortisol** e **dopamina**, o modelo visa replicar aspectos de redes neurais biológicas de forma procedural e otimizada para uso computacional.

## Estrutura do Modelo

### 1. **Localização dos Neurônios em 16 Dimensões**
- O modelo representa a rede de neurônios em um espaço vetorial de 16 dimensões.
- A posição dos neurônios é gerada proceduralmente, com base em um gerador de ruído como **OpenNoise** ou uma solução personalizada, para garantir que a rede inicializada não seja aleatória, mas tenha uma estrutura organizada.
  
### 2. **Distribuição dos Neurônios**
- **1% Dopaminérgicos** (Raio Maior: 60% do espaço)
- **30% Inibitórios** (Raio Menor: 10% do espaço)
- **69% Excitatórios** (Raio Médio: 30% do espaço)
- **5% Input**
- **5% Output**
  
A **glândula de cortisol** fica localizada no centro do espaço vetorial e serve como ponto de referência importante para a dinâmica da rede.

### 3. **Tecnologias Utilizadas**
- **ArrayFire**: Utilizado para aceleração de cálculos com uso de GPU, visando otimizar o processamento da rede neural.
- **SQLite**: Banco de dados para armazenar o estado do modelo e facilitar a análise e monitoramento da evolução da rede neural.
- **Robotgo**: Usado para visualizar o modelo na tela e controlar o teclado/mouse para interações em tempo real.
- **OpenNoise** (ou desenvolvimento próprio): Utilizado para a geração procedural da distribuição inicial dos neurônios, para encontrar as melhores configurações iniciais sem a necessidade de treinar do zero.

### 4. **Ciclos dos Neurônios**
O comportamento dos neurônios é modelado em 4 ciclos principais:
1. **Repouso**: O neurônio não está disparando.
2. **Disparo**: O neurônio dispara um pulso.
3. **Refratário Absoluto**: O neurônio não pode disparar, independente do estímulo.
4. **Refratário**: O neurônio está temporariamente inativo, mas pode ser estimulado novamente.

### 5. **Propagação de Pulso**
A propagação de pulso entre os neurônios é baseada na distância percorrida:
- **Velocidade de propagação**: 0.6 unidades por ciclo.
- **Distância máxima do espaço**: 8 unidades, o que leva cerca de 13,33 ciclos para o pulso percorrer o espaço de ponta a ponta.

### 6. **Cálculo de Distância**
- A distância entre os neurônios é calculada com base na **distância Euclidiana** em 16 dimensões.
- Busca de vizinhos feita usando a distancia do neuronio com os pontos referenciais descontando o raio, os que tiverem as distancias todas maiores que a distancia do neuronio emissor com o referencial-raio estão dentro da area de cobertura.

### 7. **Sinapogênese**
A sinapogênese é a taxa de movimentação dos neurônios no espaço, ajustada após a propagação de pulsos:
- Neurônios se aproximam daqueles que dispararam ou estavam em período refratário.
- Neurônios se afastam daqueles que estavam em repouso.

### 8. **Cortisol e Dopamina**
- **Cortisol**: A produção de cortisol é afetada pelos pulsos excitatórios que atingem a glândula de cortisol. O cortisol diminui o limiar de disparo dos neurônios inicialmente, mas ao atingir um pico, começa a reduzir o limiar, diminuindo também a sinapogênese. Caso não haja pulsos na glândula de cortisol, a quantidade de cortisol diminui com o tempo.
- **Dopamina**: A dopamina é gerada pelos neurônios dopaminérgicos e serve para aumentar o limiar de disparo dos neurônios e também aumentar a sinapogênese. A dopamina tem uma taxa de decaimento mais acentuada ao longo do tempo, ao contrário do cortisol.

### 9. **Codificação de Input/Output**
A codificação de **input** e **output** dos neurônios é baseada na **frequência de pulsos (Hz)**. A frequência é algo em torno de 10 ciclos por segundo (framerate), mas ajustes serão feitos conforme a necessidade do modelo. 

### 10. **Registro de Ciclos**
Cada neurônio mantém um registro do último ciclo em que disparou, e esse contador é atualizado apenas quando o neurônio efetivamente dispara, evitando atualizações desnecessárias a cada ciclo.

### 11. **Disparo dos Neurônios**
Os neurônios disparam quando a soma dos pulsos recebidos excede o **limiar de disparo**. Quando não recebem pulsos, a soma dos pulsos diminui gradativamente, voltando ao estado basal.

```go
loop {
  for(pulse of pulses){
    # Processamento da propagação dos pulsos
    1) Obter distancia do neuronio para os 17 referenciais e o raio
    2) Subtrair o raio das 17 distancias com o referencial
    3) Filtrar por neuronios onde a distancia com cada referencial é maior que a distancia calculada no passo anterior, exceto o proprio neuronio emissor
    4) Calcular o vetor 16D da posição dos neuronios baseado na distancia com os referenciais
       4.1) resolução da equação quadratica
    5) Calcular a distancia do neuronio emissor para os neuronios
    6) range de distancia de propagação do pulso para a iteração
       6.1) inicio = (int8(raio/0.6)*0.6) * iteração-1
       6.2) fim = (int8(raio/0.6)*0.6) * iteração
    7) Verificar se a faixa fim é maior que o raio, se for retorna
    8) Filtrar os neuronios que a distancia é maior ou igual ao inicio e menor que fim
    9) Se for soma soma 0.3 ao valor do neurono, se for inibição subitrai 0.3 do valor do neuronio se for dopamina soma a quanidade de dopamina do neuronio
    10) Verificar se o neuronio atingiu o limiar, se sim, criar novo pulso na fila para o neuronio
 }

}
```
